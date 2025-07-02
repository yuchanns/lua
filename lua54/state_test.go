package lua_test

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/lua/internal/tools"
	"go.yuchanns.xyz/lua/lua54"
)

func (s *Suite) TestTrackingAlloc(assert *require.Assertions, t *testing.T) {
	arena := tools.NewArena()
	t.Cleanup(arena.FreeAll)

	L, err := s.lib.NewState(lua.WithAlloc(trackingAlloc, arena))
	assert.NoError(err)

	t.Cleanup(L.Close)

	err = L.DoString(`local t = {}; for i=1,1000 do t[i] = i end`)
	assert.NoError(err)

	t.Logf("Total Allocated Memory: %d bytes", arena.TotalAllocated())
	t.Logf("Peak Memory Usage: %d bytes", arena.PeakMemory())
	t.Logf("Allocation Count: %d", arena.AllocCount())
}

func (s *Suite) TestLimitedAlloc(assert *require.Assertions, t *testing.T) {
	arena := tools.NewArena()
	t.Cleanup(arena.FreeAll)

	limitedMem := &limitedMemory{Limit: 1024 * 1024, Arena: arena}

	L, err := s.lib.NewState(lua.WithAlloc(limitedAlloc, limitedMem))
	assert.NoError(err)

	t.Cleanup(L.Close)

	err = L.DoString(`local t = {}; for i=1,100000 do t[i] = string.rep('x', 100) end`)

	assert.Error(err) // Should fail due to allocation limit
}

func trackingAlloc(arena *tools.Arena, ptr unsafe.Pointer, osize, nsize int) (newPtr unsafe.Pointer) {
	if nsize == 0 {
		arena.Free(ptr)
		return
	}
	return arena.ReAlloc(ptr, nsize)
}

type limitedMemory struct {
	Limit int
	Arena *tools.Arena
}

func limitedAlloc(limitedMem *limitedMemory, ptr unsafe.Pointer, osize, nsize int) (newPtr unsafe.Pointer) {
	arena := limitedMem.Arena
	if nsize == 0 {
		arena.Free(ptr)
		return
	}
	newUsed := arena.AllocCount() - osize + nsize
	if newUsed > limitedMem.Limit {
		return // Allocation exceeds limit
	}
	return arena.ReAlloc(ptr, nsize)
}
