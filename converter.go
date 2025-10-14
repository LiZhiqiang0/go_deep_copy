package copier

import (
	"copier/rt"
	"reflect"
	"strconv"
	"sync"
)

var mFuncMap sync.Map

// 缓存结构体信息
var structInfoCache *RCU

func init() {
	mFuncMap = sync.Map{}
	structInfoCache = NewRCU()
}

type rcuCacheInfo struct {
	ConvertFunc ConvertFunc
	IsFinal     bool
}

type ConvertFunc func(rt.Value, rt.Value) error

func LoadConvertFunc(v, t reflect.Type) (ConvertFunc, bool) {
	vType := rt.UnpackType(v)
	tType := rt.UnpackType(t)
	key := uint64(tType.Hash)<<32 + uint64(vType.Hash)
	if fi, ok := mFuncMap.Load(key); ok {
		return fi.(rcuCacheInfo).ConvertFunc, fi.(rcuCacheInfo).IsFinal
	}
	var (
		wg sync.WaitGroup
		f  ConvertFunc
	)
	wg.Add(1)
	fi, loaded := mFuncMap.LoadOrStore(key, rcuCacheInfo{
		ConvertFunc: func(v, t rt.Value) error {
			wg.Wait()
			return f(v, t)
		},
		IsFinal: false,
	})
	if loaded {
		return fi.(rcuCacheInfo).ConvertFunc, fi.(rcuCacheInfo).IsFinal
	}
	// Compute the real encoder and replace the indirect func with it.
	f = convertOp(v, t)
	wg.Done()
	mFuncMap.Store(key, rcuCacheInfo{ConvertFunc: f, IsFinal: true})

	return f, true
}

func convertOp(v, t reflect.Type) func(v, t rt.Value) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return cvtInt
		case reflect.Float32, reflect.Float64:
			return cvtIntFloat
		case reflect.String:
			return cvtIntString
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return cvtUint
		case reflect.Float32, reflect.Float64:
			return cvtUintFloat
		case reflect.String:
			return cvtUintString
		}

	case reflect.Float32, reflect.Float64:
		switch t.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return cvtFloatInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return cvtFloatUint
		case reflect.Float32, reflect.Float64:
			return cvtFloat
		}

	case reflect.Complex64, reflect.Complex128:
		switch t.Kind() {
		case reflect.Complex64, reflect.Complex128:
			return cvtComplex
		}

	case reflect.String:
		if t.Kind() == reflect.Slice {
			switch t.Elem().Kind() {
			case reflect.Uint8:
				return cvtStringBytes
			case reflect.Int32:
				return cvtStringRunes
			}
		}
		if t.Kind() == reflect.String {
			return cvtString
		}

	case reflect.Slice:
		if t.Kind() == reflect.String {
			switch v.Elem().Kind() {
			case reflect.Uint8:
				return cvtBytesString
			case reflect.Int32:
				return cvtRunesString
			}
		}
		if t.Kind() == reflect.Slice {
			return newSliceCvt(v, t)
		}
	case reflect.Struct:
		if t.Kind() == reflect.Struct {
			return newStructEncoder(v, t)
		}
	}
	return nil
}

// convertOp: intXX -> [u]intXX
func cvtInt(v, t rt.Value) error {
	value := *(*int64)(v.Ptr)
	if t.Typ.Kind() == rt.Uint {
		*(*uint64)(t.Ptr) = uint64(value)
	} else {
		*(*int64)(t.Ptr) = value
	}
	return nil
}

// convertOp: uintXX -> [u]intXX
func cvtUint(v, t rt.Value) error {
	value := *(*uint64)(v.Ptr)
	if t.Typ.Kind() == rt.Uint {
		*(*uint64)(t.Ptr) = value
	} else {
		*(*int64)(t.Ptr) = int64(value)
	}
	return nil
}

// convertOp: floatXX -> intXX
func cvtFloatInt(v, t rt.Value) error {
	var value float64
	if v.Typ.Kind() == rt.Float32 {
		value = float64(*(*float32)(v.Ptr))
	} else {
		value = *(*float64)(v.Ptr)
	}
	*(*int64)(t.Ptr) = int64(value)
	return nil
}

// convertOp: floatXX -> uintXX
func cvtFloatUint(v, t rt.Value) error {
	var value float64
	if v.Typ.Kind() == rt.Float32 {
		value = float64(*(*float32)(v.Ptr))
	} else {
		value = *(*float64)(v.Ptr)
	}
	*(*uint64)(t.Ptr) = uint64(value)
	return nil
}

// convertOp: intXX -> floatXX
func cvtIntFloat(v, t rt.Value) error {
	value := *(*int64)(v.Ptr)
	if t.Typ.Kind() == rt.Float32 {
		*(*float32)(t.Ptr) = float32(value)
	} else {
		*(*float64)(t.Ptr) = float64(value)
	}
	return nil
}

// convertOp: uintXX -> floatXX
func cvtUintFloat(v, t rt.Value) error {
	value := *(*uint64)(v.Ptr)
	if t.Typ.Kind() == rt.Float32 {
		*(*float32)(t.Ptr) = float32(value)
	} else {
		*(*float64)(t.Ptr) = float64(value)
	}
	return nil
}

// convertOp: floatXX -> floatXX
func cvtFloat(v, t rt.Value) error {
	var value float64
	if v.Typ.Kind() == rt.Float32 {
		value = float64(*(*float32)(v.Ptr))
	} else {
		value = *(*float64)(v.Ptr)
	}
	if t.Typ.Kind() == rt.Float32 {
		*(*float32)(t.Ptr) = float32(value)
	} else {
		*(*float64)(t.Ptr) = value
	}
	return nil
}

// convertOp: complexXX -> complexXX
func cvtComplex(v, t rt.Value) error {
	var value complex128
	if v.Typ.Kind() == rt.Complex64 {
		value = complex128(*(*complex64)(v.Ptr))
	} else {
		value = *(*complex128)(v.Ptr)
	}
	if t.Typ.Kind() == rt.Complex64 {
		*(*complex64)(t.Ptr) = complex64(value)
	} else {
		*(*complex128)(t.Ptr) = value
	}
	return nil
}

// convertOp: intXX -> string
func cvtIntString(v, t rt.Value) error {
	value := *(*int64)(v.Ptr)
	*(*string)(t.Ptr) = strconv.FormatInt(value, 10)
	return nil
}

// convertOp: String -> String
func cvtString(v, t rt.Value) error {
	value := *(*string)(v.Ptr)
	*(*string)(t.Ptr) = value
	return nil
}

// convertOp: uintXX -> string
func cvtUintString(v, t rt.Value) error {
	value := *(*uint64)(v.Ptr)
	*(*string)(t.Ptr) = strconv.FormatUint(value, 10)
	return nil
}

// convertOp: []byte -> string
func cvtBytesString(v, t rt.Value) error {
	value := *(*[]byte)(v.Ptr)
	*(*string)(t.Ptr) = string(value)
	return nil
}

// convertOp: string -> []byte
func cvtStringBytes(v, t rt.Value) error {
	value := *(*string)(v.Ptr)
	*(*[]byte)(t.Ptr) = []byte(value)
	return nil
}

// convertOp: []rune -> string
func cvtRunesString(v, t rt.Value) error {
	value := *(*[]rune)(v.Ptr)
	*(*string)(t.Ptr) = string(value)
	return nil
}

// convertOp: string -> []rune
func cvtStringRunes(v, t rt.Value) error {
	value := *(*string)(v.Ptr)
	*(*[]rune)(t.Ptr) = []rune(value)
	return nil
}

func newSliceCvt(v, t reflect.Type) ConvertFunc {
	elemConverter, isFinalEncoder := LoadConvertFunc(v.Elem(), t.Elem())
	cvt := &sliceConverter{
		elemConverter:  elemConverter,
		isFinalEncoder: isFinalEncoder,
		vElemType:      v.Elem(),
		tElemType:      t.Elem(),
	}
	return cvt.convert
}

// convertOp: []T -> *[N]T
//func cvtSliceArrayPtr(v, t rt.Value) error {
//	n := t.Elem().Len()
//	if n > v.Len() {
//		panic("reflect: cannot convert slice with length " + itoa.Itoa(v.Len()) + " to pointer to array with length " + itoa.Itoa(n))
//	}
//	h := (*unsafeheader.Slice)(v.ptr)
//	return Value{t.common(), h.Data, v.flag&^(flagIndir|flagAddr|flagKindMask) | flag(Pointer)}
//}

//
//// convertOp: []T -> [N]T
//func cvtSliceArray(v Value, t Type) Value {
//	n := t.Len()
//	if n > v.Len() {
//		panic("reflect: cannot convert slice with length " + itoa.Itoa(v.Len()) + " to array with length " + itoa.Itoa(n))
//	}
//	h := (*unsafeheader.Slice)(v.ptr)
//	typ := t.common()
//	ptr := h.Data
//	c := unsafe_New(typ)
//	typedmemmove(typ, c, ptr)
//	ptr = c
//
//	return Value{typ, ptr, v.flag&^(flagAddr|flagKindMask) | flag(Array)}
//}
//
//// convertOp: direct copy
//func cvtDirect(v Value, typ Type) Value {
//	f := v.flag
//	t := typ.common()
//	ptr := v.ptr
//	if f&flagAddr != 0 {
//		// indirect, mutable word - make a copy
//		c := unsafe_New(t)
//		typedmemmove(t, c, ptr)
//		ptr = c
//		f &^= flagAddr
//	}
//	return Value{t, ptr, v.flag.ro() | f} // v.flag.ro()|f == f?
//}
//
//// convertOp: concrete -> interface
//func cvtT2I(v Value, typ Type) Value {
//	target := unsafe_New(typ.common())
//	x := valueInterface(v, false)
//	if typ.NumMethod() == 0 {
//		*(*any)(target) = x
//	} else {
//		ifaceE2I(typ.common(), x, target)
//	}
//	return Value{typ.common(), target, v.flag.ro() | flagIndir | flag(Interface)}
//}
//
//// convertOp: interface -> interface
//func cvtI2I(v Value, typ Type) Value {
//	if v.IsNil() {
//		ret := Zero(typ)
//		ret.flag |= v.flag.ro()
//		return ret
//	}
//	return cvtT2I(v.Elem(), typ)
//}
