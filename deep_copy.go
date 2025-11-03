package go_deep_copy

import (
	"github.com/LiZhiqiang0/go_deep_copy/rt"
	"github.com/modern-go/reflect2"
	"reflect"
	"unsafe"
)

// DeepCopy deep copy things
func DeepCopy(fromValue interface{}, toValue interface{}) (err error) {
	return deepCopy(fromValue, toValue)
}

func deepCopy(fromValue interface{}, toValue interface{}) (err error) {
	var (
		from = indirect(reflect.ValueOf(fromValue))
		to   = indirect(reflect.ValueOf(toValue))
	)

	if !to.CanAddr() {
		return ErrInvalidCopyDestination
	}

	// Return is from value is invalid
	if !from.IsValid() {
		return ErrInvalidCopyFrom
	}

	var fromPtr unsafe.Pointer
	if from.CanAddr() {
		fromPtr = unsafe.Pointer(from.UnsafeAddr())
	} else {
		fromPtr = reflect2.PtrOf(fromValue)
	}
	toPtr := unsafe.Pointer(to.UnsafeAddr())
	fromType2 := reflect2.Type2(from.Type())
	toType2 := reflect2.Type2(to.Type())
	cvtFunc := LoadConvertFunc(fromType2, toType2)
	return cvtFunc(rt.Value{
		Typ: fromType2,
		Ptr: fromPtr,
	}, rt.Value{
		Typ: toType2,
		Ptr: toPtr,
	})
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func indirectType(reflectType reflect.Type) (_ reflect.Type, isPtr bool) {
	for reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
		isPtr = true
	}
	return reflectType, isPtr
}
