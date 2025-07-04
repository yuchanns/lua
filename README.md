# lua
Lua Go binding in purego

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

	"go.yuchanns.xyz/lua/lua54"
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

	// Load a Lua script
	if err := L.DoString(`print("Hello, Lua!")`); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Call a Go function from Lua
	L.PushGoFunction(func(x float64) float64 {
		return x * 2
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

### Run Tests

```bash
make lua54

cd lua54 && go test -v .
```
