package rt

import (
	"github.com/modern-go/reflect2"
	"reflect"
	"unsafe"
)

type Value struct {
	Typ reflect2.Type
	Ptr unsafe.Pointer
}

// SetBool sets v's underlying value.
// It panics if v's Kind is not [Bool] or if [Value.CanSet] returns false.
func (v Value) SetBool(x bool) {
	*(*bool)(v.Ptr) = x
}

// SetBytes sets v's underlying value.
// It panics if v's underlying value is not a slice of bytes.
func (v Value) SetBytes(x []byte) {
	*(*[]byte)(v.Ptr) = x
}

// setRunes sets v's underlying value.
// It panics if v's underlying value is not a slice of runes (int32s).
func (v Value) setRunes(x []rune) {
	*(*[]rune)(v.Ptr) = x
}

// SetComplex sets v's underlying value to x.
// It panics if v's Kind is not [Complex64] or [Complex128], or if [Value.CanSet] returns false.
func (v Value) SetComplex(x complex128) {
	switch k := v.Typ.Kind(); k {
	case reflect.Complex64:
		*(*complex64)(v.Ptr) = complex64(x)
	case reflect.Complex128:
		*(*complex128)(v.Ptr) = x
	}
}

// SetFloat sets v's underlying value to x.
// It panics if v's Kind is not [Float32] or [Float64], or if [Value.CanSet] returns false.
func (v Value) SetFloat(x float64) {
	switch k := v.Typ.Kind(); k {
	case reflect.Float32:
		*(*float32)(v.Ptr) = float32(x)
	case reflect.Float64:
		*(*float64)(v.Ptr) = x
	}
}

// SetInt sets v's underlying value to x.
// It panics if v's Kind is not [Int], [Int8], [Int16], [Int32], or [Int64], or if [Value.CanSet] returns false.
func (v Value) SetInt(x int64) {
	switch k := v.Typ.Kind(); k {
	case reflect.Int:
		*(*int)(v.Ptr) = int(x)
	case reflect.Int8:
		*(*int8)(v.Ptr) = int8(x)
	case reflect.Int16:
		*(*int16)(v.Ptr) = int16(x)
	case reflect.Int32:
		*(*int32)(v.Ptr) = int32(x)
	case reflect.Int64:
		*(*int64)(v.Ptr) = x
	}
}

// SetUint sets v's underlying value to x.
// It panics if v's Kind is not [Uint], [Uintptr], [Uint8], [Uint16], [Uint32], or [Uint64], or if [Value.CanSet] returns false.
func (v Value) SetUint(x uint64) {
	switch k := v.Typ.Kind(); k {
	case reflect.Uint:
		*(*uint)(v.Ptr) = uint(x)
	case reflect.Uint8:
		*(*uint8)(v.Ptr) = uint8(x)
	case reflect.Uint16:
		*(*uint16)(v.Ptr) = uint16(x)
	case reflect.Uint32:
		*(*uint32)(v.Ptr) = uint32(x)
	case reflect.Uint64:
		*(*uint64)(v.Ptr) = x
	case reflect.Uintptr:
		*(*uintptr)(v.Ptr) = uintptr(x)
	}
}

// SetPointer sets the [unsafe.Pointer] value v to x.
// It panics if v's Kind is not [UnsafePointer].
func (v Value) SetPointer(x unsafe.Pointer) {
	*(*unsafe.Pointer)(v.Ptr) = x
}

// SetString sets v's underlying value to x.
// It panics if v's Kind is not [String] or if [Value.CanSet] returns false.
func (v Value) SetString(x string) {
	*(*string)(v.Ptr) = x
}

func (v Value) Int() int64 {
	k := v.Typ.Kind()
	p := v.Ptr
	switch k {
	case reflect.Int:
		return int64(*(*int)(p))
	case reflect.Int8:
		return int64(*(*int8)(p))
	case reflect.Int16:
		return int64(*(*int16)(p))
	case reflect.Int32:
		return int64(*(*int32)(p))
	case reflect.Int64:
		return *(*int64)(p)
	}
	return 0
}

// Float returns v's underlying value, as a float64.
// It panics if v's Kind is not [Float32] or [Float64]
func (v Value) Float() float64 {
	k := v.Typ.Kind()
	switch k {
	case reflect.Float32:
		return float64(*(*float32)(v.Ptr))
	case reflect.Float64:
		return *(*float64)(v.Ptr)
	}
	return 0
}

// Uint returns v's underlying value, as a uint64.
// It panics if v's Kind is not [Uint], [Uintptr], [Uint8], [Uint16], [Uint32], or [Uint64].
func (v Value) Uint() uint64 {
	k := v.Typ.Kind()
	p := v.Ptr
	switch k {
	case reflect.Uint:
		return uint64(*(*uint)(p))
	case reflect.Uint8:
		return uint64(*(*uint8)(p))
	case reflect.Uint16:
		return uint64(*(*uint16)(p))
	case reflect.Uint32:
		return uint64(*(*uint32)(p))
	case reflect.Uint64:
		return *(*uint64)(p)
	case reflect.Uintptr:
		return uint64(*(*uintptr)(p))
	}
	return 0
}
