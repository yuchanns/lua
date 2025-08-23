package lua

import (
	"fmt"
	"reflect"
	"unsafe"
)

func toLightUserData(ud any) (p unsafe.Pointer) {
	switch v := ud.(type) {
	case unsafe.Pointer:
		p = v
	default:
		val := reflect.ValueOf(ud)
		if val.Kind() != reflect.Ptr {
			panic(fmt.Sprintf("expected a pointer, got %T", ud))
		}
		p = unsafe.Pointer(val.Pointer())
	}
	return
}
