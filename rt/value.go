package rt

import (
	"reflect"
	"unsafe"
)

type Value struct {
	Typ  *Type
	Ptr  unsafe.Pointer
	Flag uintptr
}

func ReflectValueToValue(v *reflect.Value) *Value {
	return (*Value)(unsafe.Pointer(v))
}

func ValueToReflectValue(v *Value) *reflect.Value {
	return (*reflect.Value)(unsafe.Pointer(v))
}

func ReflectTypeToType(v *reflect.Type) *Type {
	return (*Type)(unsafe.Pointer(v))
}
