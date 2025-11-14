package go_deep_copy

import (
	"reflect"
	"strconv"
	"sync"
	"unsafe"

	"github.com/LiZhiqiang0/go_deep_copy/rt"
	"github.com/LiZhiqiang0/reflect2"
)

var mFuncMap *MapRCU

// 缓存结构体信息
var structInfoCache *LinerRCU

func init() {
	mFuncMap = NewMapRCU()
	structInfoCache = NewLinerRCU()
}

type rcuCacheInfo struct {
	ConvertFunc ConvertFunc
}

type ConvertFunc func(rt.Value, rt.Value) error

func LoadConvertFunc(v, t reflect2.Type) ConvertFunc {
	key := [2]uintptr{v.RType(), t.RType()}
	if fi, ok := mFuncMap.Load(key); ok {
		return fi.(rcuCacheInfo).ConvertFunc
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
	})
	if loaded {
		return fi.(rcuCacheInfo).ConvertFunc
	}
	f = convertOp(v, t)
	wrapFunc := ConvertFunc(func(v rt.Value, t rt.Value) error {
		if f == nil {
			return ErrNotSupported
		}
		if v.Typ.UnsafeIsNil(v.Ptr) {
			if t.Typ.Kind() == reflect.Ptr {
				*((*unsafe.Pointer)(t.Ptr)) = nil
				return nil
			}
			t.Typ.UnsafeSet(t.Ptr, t.Typ.UnsafeNew())
			return nil
		}
		return f(v, t)
	})
	wg.Done()
	mFuncMap.Store(key, rcuCacheInfo{ConvertFunc: wrapFunc})

	return wrapFunc
}

func convertOp(v, t reflect2.Type) func(v, t rt.Value) error {
	vKind := getKind(v)
	tKind := getKind(t)
	switch vKind {
	case reflect.Int:
		switch tKind {
		case reflect.Int, reflect.Uint:
			return cvtInt
		case reflect.Float32:
			return cvtIntFloat
		case reflect.String:
			return cvtIntString
		case reflect.Bool:
			return cvtIntBool

		}
	case reflect.Bool:
		switch tKind {
		case reflect.Int:
			return cvtBoolInt
		case reflect.Uint:
			return cvtBoolUint
		case reflect.Float32:
			return cvtBoolFloat
		case reflect.String:
			return cvtBoolString
		case reflect.Bool:
			return cvtBool

		}
	case reflect.Uint:
		switch tKind {
		case reflect.Int, reflect.Uint:
			return cvtUint
		case reflect.Float32:
			return cvtUintFloat
		case reflect.String:
			return cvtUintString
		case reflect.Bool:
			return cvtIntBool

		}

	case reflect.Float32:
		switch tKind {
		case reflect.Int:
			return cvtFloatInt
		case reflect.Uint:
			return cvtFloatUint
		case reflect.Float32:
			return cvtFloat
		case reflect.Bool:
			return cvtFloatBool

		}

	case reflect.Complex64, reflect.Complex128:
		switch tKind {
		case reflect.Complex64, reflect.Complex128:
			return cvtComplex
		}

	case reflect.String:
		switch tKind {
		case reflect.Slice:
			return cvtStringSlice
		case reflect.String:
			return cvtString
		case reflect.Int:
			return cvtStringInt
		case reflect.Uint:
			return cvtStringUint
		case reflect.Float32:
			return cvtStringFloat
		case reflect.Bool:
			return cvtStringBool

		}

	case reflect.Slice:
		switch tKind {
		case reflect.String:
			return cvtSliceToString
		case reflect.Slice:
			return cvtSliceToSlice
		case reflect.Array:
			return cvtSliceToArray

		}

	case reflect.Array:
		switch tKind {
		case reflect.Slice:
			return cvtArrayToSlice
		case reflect.Array:
			return cvtArray

		}
	case reflect.Struct:
		switch tKind {
		case reflect.Struct:
			return cvtStructToStruct

		case reflect.Map:
			return cvtStructToMap
		}
	case reflect.Map:
		switch tKind {
		case reflect.Struct:
			return cvtMapToStruct

		case reflect.Map:
			return cvtMapToMap
		}
	case reflect.Ptr:
		switch tKind {
		case reflect.Ptr:
			return cvtTToPtr
		default:
			return cvtPtrToT
		}
	case reflect.Interface:
		switch tKind {
		case reflect.Interface:
			return cvtIToI
		case reflect.Ptr:
			return cvtTToPtr
		default:
			return cvtIToT
		}
	}
	if tKind == reflect.Ptr {
		return cvtTToPtr
	}
	if tKind == reflect.Interface {
		return cvtTToI
	}
	return nil
}

func getKind(val reflect2.Type) reflect.Kind {
	kind := val.Kind()

	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uintptr:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}

// convertOp: intXX -> [u]intXX
func cvtInt(v, t rt.Value) error {
	value := v.Int()
	tType := getKind(t.Typ)
	if tType == reflect.Uint {
		t.SetUint(uint64(value))
	} else {
		t.SetInt(int64(value))
	}
	return nil
}

// convertOp: intXX -> bool
func cvtIntBool(v, t rt.Value) error {
	value := v.Int()
	t.SetBool(value != 0)
	return nil
}

// convertOp: uintXX -> [u]intXX
func cvtUint(v, t rt.Value) error {
	value := v.Uint()
	tType := getKind(t.Typ)
	if tType == reflect.Uint {
		t.SetUint(value)
	} else {
		t.SetInt(int64(value))
	}
	return nil
}

// convertOp: uintXX -> bool
func cvtUintBool(v, t rt.Value) error {
	value := v.Uint()
	t.SetBool(value != 0)
	return nil
}

// convertOp: floatXX -> intXX
func cvtFloatInt(v, t rt.Value) error {
	var value float64
	if v.Typ.Kind() == reflect.Float32 {
		value = float64(*(*float32)(v.Ptr))
	} else {
		value = *(*float64)(v.Ptr)
	}
	t.SetInt(int64(value))
	return nil
}

// convertOp: floatXX -> bool
func cvtFloatBool(v, t rt.Value) error {
	var value float64
	if v.Typ.Kind() == reflect.Float32 {
		value = float64(*(*float32)(v.Ptr))
	} else {
		value = *(*float64)(v.Ptr)
	}
	t.SetBool(value != 0)
	return nil
}

// convertOp: floatXX -> uintXX
func cvtFloatUint(v, t rt.Value) error {
	var value float64
	if v.Typ.Kind() == reflect.Float32 {
		value = float64(*(*float32)(v.Ptr))
	} else {
		value = *(*float64)(v.Ptr)
	}
	t.SetUint(uint64(value))
	return nil
}

// convertOp: intXX -> floatXX
func cvtIntFloat(v, t rt.Value) error {
	value := v.Int()
	if t.Typ.Kind() == reflect.Float32 {
		*(*float32)(t.Ptr) = float32(value)
	} else {
		*(*float64)(t.Ptr) = float64(value)
	}
	return nil
}

// convertOp: uintXX -> floatXX
func cvtUintFloat(v, t rt.Value) error {
	value := v.Uint()
	if t.Typ.Kind() == reflect.Float32 {
		*(*float32)(t.Ptr) = float32(value)
	} else {
		*(*float64)(t.Ptr) = float64(value)
	}
	return nil
}

// convertOp: floatXX -> floatXX
func cvtFloat(v, t rt.Value) error {
	var value float64
	if v.Typ.Kind() == reflect.Float32 {
		value = float64(*(*float32)(v.Ptr))
	} else {
		value = *(*float64)(v.Ptr)
	}
	if t.Typ.Kind() == reflect.Float32 {
		*(*float32)(t.Ptr) = float32(value)
	} else {
		*(*float64)(t.Ptr) = value
	}
	return nil
}

// convertOp: bool -> bool
func cvtBool(v, t rt.Value) error {
	value := *(*bool)(v.Ptr)
	*(*bool)(t.Ptr) = value
	return nil
}

// convertOp: bool -> intXX
func cvtBoolInt(v, t rt.Value) error {
	value := *(*bool)(v.Ptr)
	if value {
		t.SetInt(1)
	} else {
		t.SetInt(0)
	}
	return nil
}

// convertOp: bool -> uintXX
func cvtBoolUint(v, t rt.Value) error {
	value := *(*bool)(v.Ptr)
	if value {
		t.SetUint(1)
	} else {
		t.SetUint(0)
	}
	return nil
}

// convertOp: bool -> floatXX
func cvtBoolFloat(v, t rt.Value) error {
	value := *(*bool)(v.Ptr)
	if value {
		t.SetFloat(1)
	} else {
		t.SetFloat(0)
	}
	return nil
}

// convertOp: bool -> string
func cvtBoolString(v, t rt.Value) error {
	value := *(*bool)(v.Ptr)
	if value {
		*(*string)(t.Ptr) = "true"
	} else {
		*(*string)(t.Ptr) = "false"
	}
	return nil
}

// convertOp: complexXX -> complexXX
func cvtComplex(v, t rt.Value) error {
	var value complex128
	if v.Typ.Kind() == reflect.Complex64 {
		value = complex128(*(*complex64)(v.Ptr))
	} else {
		value = *(*complex128)(v.Ptr)
	}
	if t.Typ.Kind() == reflect.Complex64 {
		*(*complex64)(t.Ptr) = complex64(value)
	} else {
		*(*complex128)(t.Ptr) = value
	}
	return nil
}

// convertOp: intXX -> string
func cvtIntString(v, t rt.Value) error {
	value := v.Int()
	*(*string)(t.Ptr) = strconv.FormatInt(value, 10)
	return nil
}

// convertOp: String -> String
func cvtString(v, t rt.Value) error {
	value := *(*string)(v.Ptr)
	*(*string)(t.Ptr) = value
	return nil
}

// convertOp: String -> int
func cvtStringInt(v, t rt.Value) error {
	value := *(*string)(v.Ptr)
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	t.SetInt(intValue)
	return nil
}

// convertOp: String -> uint
func cvtStringUint(v, t rt.Value) error {
	value := *(*string)(v.Ptr)
	uintValue, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return err
	}
	t.SetUint(uintValue)
	return nil
}

// convertOp: String -> float
func cvtStringFloat(v, t rt.Value) error {
	value := *(*string)(v.Ptr)
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	t.SetFloat(floatValue)
	return nil
}

// convertOp: String -> bool
func cvtStringBool(v, t rt.Value) error {
	value := *(*string)(v.Ptr)
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	t.SetBool(boolValue)
	return nil
}

// convertOp: String -> Slice
func cvtStringSlice(v, t rt.Value) error {
	switch t.Typ.(reflect2.SliceType).Elem().Kind() {
	case reflect.Uint8:
		return cvtStringBytes(v, t)
	case reflect.Int32:
		return cvtStringRunes(v, t)
	default:
		return ErrNotSupported
	}
}

// convertOp: uintXX -> string
func cvtUintString(v, t rt.Value) error {
	value := v.Uint()
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

// convertOp: []T -> []T
func cvtSliceToSlice(v, t rt.Value) error {
	vType := v.Typ.(*reflect2.UnsafeSliceType)
	tType := t.Typ.(*reflect2.UnsafeSliceType)
	vElemType := v.Typ.(reflect2.SliceType).Elem()
	tElemType := t.Typ.(reflect2.SliceType).Elem()
	if vType.UnsafeIsNil(v.Ptr) {
		tType.UnsafeSetNil(t.Ptr)
		return nil
	}
	length := vType.UnsafeLengthOf(v.Ptr)
	tPtr := tType.UnsafeNew()
	for i := 0; i < length; i++ {
		elemConverter := LoadConvertFunc(vElemType, tElemType)
		tType.UnsafeGrow(tPtr, i+1)
		tElemPtr := tType.UnsafeGetIndex(tPtr, i)
		vElemPtr := vType.UnsafeGetIndex(v.Ptr, i)
		err := elemConverter(rt.Value{
			Ptr: vElemPtr,
			Typ: vElemType,
		}, rt.Value{
			Ptr: tElemPtr,
			Typ: tElemType,
		})
		if err != nil {
			return err
		}
	}
	tType.UnsafeSet(t.Ptr, tPtr)
	return nil
}

// convertOp: []T -> [N]T
func cvtSliceToArray(v, t rt.Value) error {
	vType := v.Typ.(*reflect2.UnsafeSliceType)
	tType := t.Typ.(*reflect2.UnsafeArrayType)
	vElemType := vType.Elem()
	tElemType := tType.Elem()
	vLength := vType.UnsafeLengthOf(v.Ptr)
	tLength := tType.Len()
	for i := 0; i < vLength && i < tLength; i++ {
		elemConverter := LoadConvertFunc(vElemType, tElemType)
		tElemPtr := tType.UnsafeGetIndex(t.Ptr, i)
		vElemPtr := vType.UnsafeGetIndex(v.Ptr, i)
		err := elemConverter(rt.Value{
			Ptr: vElemPtr,
			Typ: vElemType,
		}, rt.Value{
			Ptr: tElemPtr,
			Typ: tElemType,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// convertOp: Slice -> string
func cvtSliceToString(v, t rt.Value) error {
	switch v.Typ.(reflect2.SliceType).Elem().Kind() {
	case reflect.Uint8:
		return cvtBytesString(v, t)
	case reflect.Int32:
		return cvtRunesString(v, t)
	default:
		return ErrNotSupported
	}
}

// convertOp: [N]T -> []T
func cvtArrayToSlice(v, t rt.Value) error {
	vType := v.Typ.(*reflect2.UnsafeArrayType)
	tType := t.Typ.(*reflect2.UnsafeSliceType)
	vElemType := vType.Elem()
	tElemType := tType.Elem()
	vLength := vType.Len()
	tPtr := tType.UnsafeNew()
	for i := 0; i < vLength; i++ {
		elemConverter := LoadConvertFunc(vElemType, tElemType)
		tType.UnsafeGrow(tPtr, i+1)
		tElemPtr := tType.UnsafeGetIndex(tPtr, i)
		vElemPtr := vType.UnsafeGetIndex(v.Ptr, i)
		err := elemConverter(rt.Value{
			Ptr: vElemPtr,
			Typ: vElemType,
		}, rt.Value{
			Ptr: tElemPtr,
			Typ: tElemType,
		})
		if err != nil {
			return err
		}
	}
	tType.UnsafeSet(t.Ptr, tPtr)
	return nil
}

// convertOp: [N]T -> [N]T
func cvtArray(v, t rt.Value) error {
	vType := v.Typ.(*reflect2.UnsafeArrayType)
	tType := t.Typ.(*reflect2.UnsafeArrayType)
	vElemType := vType.Elem()
	tElemType := tType.Elem()
	vLength := vType.Len()
	tLength := tType.Len()
	for i := 0; i < vLength && i < tLength; i++ {
		elemConverter := LoadConvertFunc(vElemType, tElemType)
		tElemPtr := tType.UnsafeGetIndex(t.Ptr, i)
		vElemPtr := vType.UnsafeGetIndex(v.Ptr, i)
		err := elemConverter(rt.Value{
			Ptr: vElemPtr,
			Typ: vElemType,
		}, rt.Value{
			Ptr: tElemPtr,
			Typ: tElemType,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// convertOp: T -> interface{}
func cvtTToI(v rt.Value, t rt.Value) error {
	vKind := getKind(v.Typ)
	tPObj := (*interface{})(t.Ptr)
	var vObj interface{}
	switch vKind {
	case reflect.Int:
		vObj = v.Int()
	case reflect.Uint:
		vObj = v.Uint()
	case reflect.Float32:
		vObj = v.Float()
	case reflect.Bool:
		vObj = v.Bool()
	case reflect.Complex64, reflect.Complex128:
		vObj = v.Complex()
	case reflect.String:
		vObj = v.String()
	case reflect.Map, reflect.Array, reflect.Slice, reflect.Struct:
		vObj = v.Typ.UnsafeNew()
		vPtr := reflect2.PtrOf(vObj)
		cvtFunc := LoadConvertFunc(v.Typ, v.Typ)
		if cvtFunc == nil {
			return nil
		}
		err := cvtFunc(v, rt.Value{
			Typ: v.Typ,
			Ptr: vPtr,
		})
		if err != nil {
			return err
		}
		*tPObj = v.Typ.UnsafeIndirect(vPtr)
		return nil

	}
	*tPObj = vObj
	return nil
}

// convertOp: interface{} -> T
func cvtIToT(v rt.Value, t rt.Value) error {
	vObj := v.Typ.UnsafeIndirect(v.Ptr)
	v.Typ = reflect2.TypeOf(vObj)
	if v.Typ.Kind() == reflect.Ptr {
		v.Typ = v.Typ.(*reflect2.UnsafePtrType).Elem()
	}
	v.Ptr = reflect2.PtrOf(vObj)
	cvtFunc := LoadConvertFunc(v.Typ, t.Typ)
	return cvtFunc(v, t)
}

// convertOp: interface{} -> interface{}
func cvtIToI(v rt.Value, t rt.Value) error {
	vObj := v.Typ.UnsafeIndirect(v.Ptr)
	v.Typ = reflect2.TypeOf(vObj)
	if v.Typ.Kind() == reflect.Ptr {
		v.Typ = v.Typ.(*reflect2.UnsafePtrType).Elem()
	}
	v.Ptr = reflect2.PtrOf(vObj)
	cvtFunc := LoadConvertFunc(v.Typ, t.Typ)
	return cvtFunc(v, t)
}

func cvtTToPtr(v rt.Value, t rt.Value) error {
	if v.Typ.Kind() == reflect.Ptr && *((*unsafe.Pointer)(v.Ptr)) == nil {
		*((*unsafe.Pointer)(t.Ptr)) = nil
		return nil
	}
	t.Typ = t.Typ.(*reflect2.UnsafePtrType).Elem()
	cvtFunc := LoadConvertFunc(v.Typ, t.Typ)
	newPtr := t.Typ.UnsafeNew()
	err := cvtFunc(v, rt.Value{
		Ptr: newPtr,
		Typ: t.Typ,
	})
	if err != nil {
		return err
	}
	*((*unsafe.Pointer)(t.Ptr)) = newPtr
	return nil
}

func cvtPtrToT(v rt.Value, t rt.Value) error {
	v.Typ = v.Typ.(*reflect2.UnsafePtrType).Elem()
	cvtFunc := LoadConvertFunc(v.Typ, t.Typ)
	if cvtFunc == nil {
		return nil
	}
	if *((*unsafe.Pointer)(v.Ptr)) == nil {
		if t.Typ.Kind() == reflect.Ptr {
			*((*unsafe.Pointer)(t.Ptr)) = nil
			return nil
		}
		t.Typ.UnsafeSet(t.Ptr, t.Typ.UnsafeNew())
		return nil
	} else {
		vPtr := *((*unsafe.Pointer)(v.Ptr))
		err := cvtFunc(rt.Value{
			Ptr: vPtr,
			Typ: v.Typ,
		}, t)
		if err != nil {
			return err
		}
	}
	return nil
}

// convertOp: struct -> struct
func cvtStructToStruct(v rt.Value, t rt.Value) error {
	vInfo := loadStructFieldsInfo(v.Typ)
	tInfo := loadStructFieldsInfo(t.Typ)
	tFieldMap := tInfo.FieldMap
	vPtr := v.Ptr
	tPtr := t.Ptr
	for i := 0; i < len(vInfo.Fields); i++ {
		f := vInfo.Fields[i]

		tf, ok := tFieldMap[f.Name]
		if !ok {
			continue
		}
		fType := f.Field.Type()
		tfType := tf.Field.Type()
		// 直接使用指针 + 偏移，避免去将指针转换为对象
		childVPtr := pointerOffset(vPtr, f.Field.Offset())
		childTPtr := pointerOffset(tPtr, tf.Field.Offset())
		cvtFunc := LoadConvertFunc(fType, tfType)
		if cvtFunc == nil {
			continue
		}
		err := cvtFunc(rt.Value{
			Ptr: childVPtr,
			Typ: fType,
		}, rt.Value{
			Ptr: childTPtr,
			Typ: tfType,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// convertOp: map -> map
func cvtMapToMap(v rt.Value, t rt.Value) error {
	vType := v.Typ.(*reflect2.UnsafeMapType)
	tType := t.Typ.(*reflect2.UnsafeMapType)
	if tType.UnsafeIsNil(t.Ptr) {
		tType.UnsafeSet(t.Ptr, tType.UnsafeMakeMap(0))
	}
	if vType.UnsafeIsNil(v.Ptr) {
		tType.UnsafeSet(t.Ptr, tType.UnsafeNew())
		return nil
	}
	vKType := vType.Key()
	tKType := tType.Key()
	vElemType := vType.Elem()
	tElemType := tType.Elem()
	iter := vType.UnsafeIterate(v.Ptr)
	keyConverter := LoadConvertFunc(vKType, tKType)
	for iter.HasNext() {
		vKey, vElem := iter.UnsafeNext()
		elemConverter := LoadConvertFunc(vElemType, tElemType)
		if keyConverter == nil || elemConverter == nil {
			continue
		}
		tKey := tKType.UnsafeNew()
		tElem := tElemType.UnsafeNew()
		err := keyConverter(rt.Value{
			Ptr: vKey,
			Typ: vKType,
		}, rt.Value{
			Ptr: tKey,
			Typ: tKType,
		})
		if err != nil {
			return err
		}
		err = elemConverter(rt.Value{
			Ptr: vElem,
			Typ: vElemType,
		}, rt.Value{
			Ptr: tElem,
			Typ: tElemType,
		})
		if err != nil {
			return err
		}
		if tKType.UnsafeIsNil(tKey) {
			continue
		}
		tType.UnsafeSetIndex(t.Ptr, tKey, tElem)
	}
	return nil
}

// convertOp: struct -> map
func cvtStructToMap(v rt.Value, t rt.Value) error {
	tType := t.Typ.(*reflect2.UnsafeMapType)
	if tType.UnsafeIsNil(t.Ptr) {
		tType.UnsafeSet(t.Ptr, tType.UnsafeMakeMap(0))
	}
	tKType := tType.Key()
	if tKType.Kind() != reflect.String {
		return nil
	}
	tElemType := tType.Elem()
	vInfo := loadStructFieldsInfo(v.Typ)
	for i := 0; i < len(vInfo.Fields); i++ {
		f := vInfo.Fields[i].Field

		fType := f.Type()
		// 直接使用指针 + 偏移，避免去将指针转换为对象
		childVPtr := pointerOffset(v.Ptr, f.Offset())
		name := f.Name()
		tElem := tElemType.UnsafeNew()
		elemConverter := LoadConvertFunc(fType, tElemType)
		if elemConverter == nil {
			continue
		}
		err := elemConverter(rt.Value{
			Ptr: childVPtr,
			Typ: fType,
		}, rt.Value{
			Ptr: tElem,
			Typ: tElemType,
		})
		if err != nil {
			return err
		}
		tType.UnsafeSetIndex(t.Ptr, unsafe.Pointer(&name), tElem)
	}
	return nil
}

// convertOp: map -> struct
func cvtMapToStruct(v rt.Value, t rt.Value) error {
	vType := v.Typ.(*reflect2.UnsafeMapType)
	if vType.UnsafeIsNil(v.Ptr) {
		return nil
	}
	vKType := vType.Key()
	if vKType.Kind() != reflect.String {
		return nil
	}
	tInfo := loadStructFieldsInfo(t.Typ)
	tFieldMap := tInfo.FieldMap
	vElemType := vType.Elem()
	iter := vType.UnsafeIterate(v.Ptr)
	for iter.HasNext() {
		vKey, vElem := iter.UnsafeNext()
		key := *(*string)(vKey)
		tf, ok := tFieldMap[key]
		if !ok {
			continue
		}
		tfType := tf.Field.Type()
		cvtFunc := LoadConvertFunc(vElemType, tfType)
		if cvtFunc == nil {
			continue
		}
		childTPtr := pointerOffset(t.Ptr, tf.Field.Offset())
		err := cvtFunc(rt.Value{
			Ptr: vElem,
			Typ: vElemType,
		}, rt.Value{
			Ptr: childTPtr,
			Typ: tfType,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
}
