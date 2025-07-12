package tools

import (
	"fmt"
	"reflect"
	"unsafe"
)

func ToLightUserData(ud any) (p unsafe.Pointer, err error) {
	switch v := ud.(type) {
	case unsafe.Pointer:
		p = v
	default:
		val := reflect.ValueOf(ud)
		if val.Kind() != reflect.Ptr {
			err = fmt.Errorf("expected a pointer, got %T", ud)
			return
		}
		p = unsafe.Pointer(val.Pointer())
	}
	return
}
