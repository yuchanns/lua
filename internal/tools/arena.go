package tools

import (
	"errors"
	"sync/atomic"
	"unsafe"
)

// Arena is a memory arena for testing Lua memory allocation from Go.
// It provides a custom memory allocator that manages memory in chunks,
// allowing for efficient allocation and deallocation of memory blocks.
//
// Arena is intentionally designed to not support concurrent access.
// Each Lua state should have its own Arena instance since Lua states
// are not thread-safe by default and memory allocation typically occurs
// in a single-threaded context.
//
// The Arena implements a simple memory management strategy:
//   - Memory is allocated in large chunks
//   - Free blocks are tracked and reused when possible
//   - Memory can be reallocated in-place when conditions permit
//   - All memory can be reset or freed at once
type Arena struct {
	chunks      []*chunk    // List of memory chunks allocated by this arena
	current     *chunk      // Currently active chunk for new allocations
	freeBlocks  []freeBlock // List of freed memory blocks available for reuse
	activeCalls int32       // Counter to detect concurrent access attempts
}

type chunk struct {
	data   []byte
	offset int
}

type freeBlock struct {
	offset uintptr
	size   int
}

type blockHeader struct {
	size  int
	magic uint32
}

const (
	magic      = 0xCAFEBABE
	headerSize = int(unsafe.Sizeof(blockHeader{}))
)

func NewArena() *Arena {
	a := &Arena{}
	a.addChunk(64 * 1024)
	return a
}

func (a *Arena) checkConcurrency() func() {
	if atomic.AddInt32(&a.activeCalls, 1) > 1 {
		atomic.AddInt32(&a.activeCalls, -1)
		panic("Arena: concurrent access detected")
	}
	return func() {
		atomic.AddInt32(&a.activeCalls, -1)
	}
}

func (a *Arena) addChunk(size int) {
	c := &chunk{
		data: make([]byte, size),
	}
	a.chunks = append(a.chunks, c)
	a.current = c
}

func (a *Arena) ReAlloc(ptr unsafe.Pointer, size int) unsafe.Pointer {
	defer a.checkConcurrency()()

	if ptr == nil {
		return a.allocNew(size)
	}

	if size == 0 {
		a.freeInternal(ptr)
		return nil
	}

	oldSize, err := a.getBlockSize(ptr)
	if err != nil {
		return nil
	}

	size = (size + 7) &^ 7

	if size == oldSize {
		return ptr
	}

	if size > oldSize {
		if newPtr := a.tryExpandInPlace(ptr, oldSize, size); newPtr != nil {
			return newPtr
		}
	}

	if size < oldSize {
		return a.shrinkInPlace(ptr, oldSize, size)
	}

	return a.reallocWithCopy(ptr, oldSize, size)
}

func (a *Arena) allocNew(size int) unsafe.Pointer {
	size = (size + 7) &^ 7
	totalSize := size + headerSize

	for i, block := range a.freeBlocks {
		if block.size >= totalSize {
			a.freeBlocks = append(a.freeBlocks[:i], a.freeBlocks[i+1:]...)

			if block.size > totalSize+headerSize+8 {
				newBlock := freeBlock{
					offset: block.offset + uintptr(totalSize),
					size:   block.size - totalSize,
				}
				a.freeBlocks = append(a.freeBlocks, newBlock)
			}

			return a.setupBlock(block.offset, size)
		}
	}

	if a.current.offset+totalSize > len(a.current.data) {
		chunkSize := 64 * 1024
		if totalSize > chunkSize {
			chunkSize = totalSize * 2
		}
		a.addChunk(chunkSize)
	}

	offset := uintptr(unsafe.Pointer(&a.current.data[a.current.offset]))
	a.current.offset += totalSize

	return a.setupBlock(offset, size)
}

func (a *Arena) tryExpandInPlace(ptr unsafe.Pointer, oldSize, newSize int) unsafe.Pointer {
	headerOffset := uintptr(ptr) - uintptr(headerSize)
	blockEnd := headerOffset + uintptr(headerSize+oldSize)
	expandSize := newSize - oldSize

	currentChunkStart := uintptr(unsafe.Pointer(&a.current.data[0]))
	currentChunkEnd := currentChunkStart + uintptr(a.current.offset)

	if blockEnd == currentChunkEnd {
		if a.current.offset+expandSize <= len(a.current.data) {
			a.current.offset += expandSize
			header := (*blockHeader)(unsafe.Pointer(headerOffset))
			header.size = newSize
			return ptr
		}
	}

	for i, block := range a.freeBlocks {
		if block.offset == blockEnd && block.size >= expandSize {
			if block.size == expandSize {
				a.freeBlocks = append(a.freeBlocks[:i], a.freeBlocks[i+1:]...)
			} else {
				a.freeBlocks[i].offset += uintptr(expandSize)
				a.freeBlocks[i].size -= expandSize
			}

			header := (*blockHeader)(unsafe.Pointer(headerOffset))
			header.size = newSize
			return ptr
		}
	}

	return nil
}

func (a *Arena) shrinkInPlace(ptr unsafe.Pointer, oldSize, newSize int) unsafe.Pointer {
	headerOffset := uintptr(ptr) - uintptr(headerSize)
	shrinkSize := oldSize - newSize

	header := (*blockHeader)(unsafe.Pointer(headerOffset))
	header.size = newSize

	if shrinkSize >= headerSize+8 {
		freeOffset := headerOffset + uintptr(headerSize+newSize)
		newFreeBlock := freeBlock{
			offset: freeOffset,
			size:   shrinkSize,
		}
		a.freeBlocks = append(a.freeBlocks, newFreeBlock)

		a.coalesceBlocks(len(a.freeBlocks) - 1)
	}

	return ptr
}

func (a *Arena) reallocWithCopy(ptr unsafe.Pointer, oldSize, newSize int) unsafe.Pointer {
	newPtr := a.allocNew(newSize)
	if newPtr == nil {
		return nil
	}

	copySize := oldSize
	if newSize < oldSize {
		copySize = newSize
	}

	oldSlice := (*[1 << 30]byte)(ptr)[:copySize:copySize]
	newSlice := (*[1 << 30]byte)(newPtr)[:copySize:copySize]
	copy(newSlice, oldSlice)

	a.freeInternal(ptr)

	return newPtr
}

func (a *Arena) getBlockSize(ptr unsafe.Pointer) (int, error) {
	if ptr == nil {
		return 0, errors.New("nil pointer")
	}

	headerOffset := uintptr(ptr) - uintptr(headerSize)
	header := (*blockHeader)(unsafe.Pointer(headerOffset))

	if header.magic != magic {
		return 0, errors.New("invalid pointer or corrupted memory")
	}

	return header.size, nil
}

func (a *Arena) setupBlock(offset uintptr, size int) unsafe.Pointer {
	header := (*blockHeader)(unsafe.Pointer(offset))
	header.size = size
	header.magic = magic

	return unsafe.Pointer(offset + uintptr(headerSize))
}

func (a *Arena) Free(ptr unsafe.Pointer) error {
	defer a.checkConcurrency()()
	return a.freeInternal(ptr)
}

func (a *Arena) freeInternal(ptr unsafe.Pointer) error {
	if ptr == nil {
		return nil
	}

	headerOffset := uintptr(ptr) - uintptr(headerSize)
	header := (*blockHeader)(unsafe.Pointer(headerOffset))

	if header.magic != magic {
		return errors.New("invalid pointer or corrupted memory")
	}

	block := freeBlock{
		offset: headerOffset,
		size:   header.size + headerSize,
	}

	a.freeBlocks = append(a.freeBlocks, block)

	a.coalesceBlocks(len(a.freeBlocks) - 1)

	header.magic = 0

	return nil
}

func (a *Arena) coalesceBlocks(index int) {
	if index < 0 || index >= len(a.freeBlocks) {
		return
	}

	current := a.freeBlocks[index]

	for i := 0; i < len(a.freeBlocks); i++ {
		if i == index {
			continue
		}

		other := a.freeBlocks[i]

		if current.offset+uintptr(current.size) == other.offset {
			a.freeBlocks[index].size += other.size
			a.freeBlocks = append(a.freeBlocks[:i], a.freeBlocks[i+1:]...)
			if i < index {
				index--
			}
			a.coalesceBlocks(index)
			return
		} else if other.offset+uintptr(other.size) == current.offset {
			a.freeBlocks[index].offset = other.offset
			a.freeBlocks[index].size += other.size
			a.freeBlocks = append(a.freeBlocks[:i], a.freeBlocks[i+1:]...)
			if i < index {
				index--
			}
			a.coalesceBlocks(index)
			return
		}
	}
}

func (a *Arena) FreeAll() {
	defer a.checkConcurrency()()

	a.chunks = nil
	a.current = nil
	a.freeBlocks = nil
}

func (a *Arena) Reset() {
	defer a.checkConcurrency()()

	for _, c := range a.chunks {
		c.offset = 0
	}
	if len(a.chunks) > 0 {
		a.current = a.chunks[0]
	}
	a.freeBlocks = a.freeBlocks[:0]
}

func (a *Arena) Size(ptr unsafe.Pointer) (int, error) {
	defer a.checkConcurrency()()
	return a.getBlockSize(ptr)
}

func (a *Arena) TotalAllocated() int {
	defer a.checkConcurrency()()

	total := 0
	for _, block := range a.freeBlocks {
		total += block.size
	}
	for _, c := range a.chunks {
		total += c.offset
	}
	return total
}

func (a *Arena) PeakMemory() int {
	defer a.checkConcurrency()()

	peak := 0
	for _, block := range a.freeBlocks {
		if block.size > peak {
			peak = block.size
		}
	}
	for _, c := range a.chunks {
		if c.offset > peak {
			peak = c.offset
		}
	}
	return peak
}

func (a *Arena) AllocCount() int {
	defer a.checkConcurrency()()

	count := 0
	for _, block := range a.freeBlocks {
		if block.size > 0 {
			count++
		}
	}
	for _, c := range a.chunks {
		if c.offset > 0 {
			count++
		}
	}
	return count
}
