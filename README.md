# lua
Lua Go binding in purego

## Caution

⚠️This library is **working in progress** 🚧 And APIs are not stable yet, maybe cause breaking changes many times. I make it public only for unlimited GitHub Actions minutes. It is not recommended to use at this moment.

## Installation

```bash
go get go.yuchanns.xyz/lua
```

## Usage

```go
package main

import (
	_ "embed"
	"fmt"
	"os"

	"go.yuchanns.xyz/lua"
)

//go:embed liblua54.so
var luaLib []byte

func main() {
	f, err := os.CreateTemp("", "liblua54.*.so")
	if err != nil {
		fmt.Println("Error creating temp file:", err)
		return
	}
	_, err = f.Write(luaLib)
	if err != nil {
		fmt.Println("Error writing to temp file:", err)
		return
	}
	f.Close()
	f.Chmod(os.ModePerm)
	defer os.Remove(f.Name())

	// Create a new Lua library instance
	lib, err := lua.New(f.Name())
	if err != nil {
		fmt.Println("Error creating Lua library:", err)
		return
	}
	defer lib.Close()

	// Create a new Lua state
	L, err := lib.NewState()
	if err != nil {
		fmt.Println("Error creating Lua state:", err)
		return
	}
	defer L.Close()

	L.OpenLibs()

	// Load a Lua script
	if err := L.DoString(`print("Hello, Lua!")`); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Call a Go function from Lua
	L.PushCFunction(func(L *lua.State) int {
		x := L.CheckNumber(1)
		L.PushNumber(x * 2)
		return 1
	})
	if err := L.SetGlobal("double_number"); err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err := L.DoString(`print(double_number(21))`); err != nil {
		fmt.Println("Error:", err)
		return
	}
}
```

## Development

### Clone

```bash
git clone --recurse-submodules https://github.com/yuchanns/lua
```

### Build Dependencies

We use [luamake](https://github.com/actboy168/luamake) to build Lua.

```bash
luamake
```

### Run Tests

```bash
cd lua54 && go test -v .
```
