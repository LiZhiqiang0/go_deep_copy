package copier

import (
	"copier/rt"
	"github.com/modern-go/reflect2"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
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

func LoadConvertFunc(v, t reflect2.Type) (ConvertFunc, bool) {
	key := [2]uintptr{v.RType(), t.RType()}
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
		case reflect.Interface:
			return cvtTToI
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
		case reflect.Interface:
			return cvtTToI
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
		case reflect.Interface:
			return cvtTToI
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
			return cvtIntBool
		case reflect.Interface:
			return cvtTToI
		}

	case reflect.Complex64, reflect.Complex128:
		switch tKind {
		case reflect.Complex64, reflect.Complex128:
			return cvtComplex
		case reflect.Interface:
			return cvtTToI
		}

	case reflect.String:
		switch tKind {
		case reflect.Slice:
			switch t.(reflect2.SliceType).Elem().Kind() {
			case reflect.Uint8:
				return cvtStringBytes
			case reflect.Int32:
				return cvtStringRunes
			}
		case reflect.String:
			return cvtString
		case reflect.Int:
			return cvtStringInt
		case reflect.Uint:
			return cvtStringUint
		case reflect.Float32:
			return cvtStringFloat
		case reflect.Bool:
			return cvtIntBool
		case reflect.Interface:
			return cvtTToI
		}

	case reflect.Slice:
		switch tKind {
		case reflect.String:
			switch v.(reflect2.SliceType).Elem().Kind() {
			case reflect.Uint8:
				return cvtBytesString
			case reflect.Int32:
				return cvtRunesString
			}
		case reflect.Slice:
			return cvtSliceToSlice
		case reflect.Interface:
			return cvtTToI
		}
	case reflect.Struct:
		switch tKind {
		case reflect.Struct:
			return cvtStructToStruct
		case reflect.Interface:
			return cvtTToI
		case reflect.Map:
			return cvtStructToMap
		}
	case reflect.Map:
		switch tKind {
		case reflect.Struct:
			return cvtMapToStruct
		case reflect.Interface:
			return cvtTToI
		case reflect.Map:
			return cvtMapToMap
		}
	case reflect.Ptr:
		return cvtPtrToT
	}
	if tKind == reflect.Ptr {
		return cvtTToPtr
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
	elemConverter, isFinalEncoder := LoadConvertFunc(vElemType, tElemType)
	length := vType.UnsafeLengthOf(v.Ptr)
	tPtr := tType.UnsafeNew()
	for i := 0; i < length; i++ {
		if !isFinalEncoder {
			elemConverter, isFinalEncoder = LoadConvertFunc(vElemType, tElemType)
		}
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
	vElemType := v.Typ.(reflect2.SliceType).Elem()
	tElemType := t.Typ.(*reflect2.UnsafeSliceType).Elem()
	elemConverter, isFinalEncoder := LoadConvertFunc(vElemType, tElemType)
	vLength := vType.UnsafeLengthOf(v.Ptr)
	tLength := tType.Len()
	for i := 0; i < vLength && i < tLength; i++ {
		if !isFinalEncoder {
			elemConverter, isFinalEncoder = LoadConvertFunc(vElemType, tElemType)
		}
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

// convertOp: [N]T -> []T
func cvtArrayToSlice(v, t rt.Value) error {
	vType := v.Typ.(*reflect2.UnsafeArrayType)
	tType := t.Typ.(*reflect2.UnsafeSliceType)
	vElemType := v.Typ.(reflect2.SliceType).Elem()
	tElemType := t.Typ.(*reflect2.UnsafeSliceType).Elem()
	elemConverter, isFinalEncoder := LoadConvertFunc(vElemType, tElemType)
	vLength := vType.Len()
	tPtr := tType.UnsafeNew()
	for i := 0; i < vLength; i++ {
		if !isFinalEncoder {
			elemConverter, isFinalEncoder = LoadConvertFunc(vElemType, tElemType)
		}
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
	vElemType := v.Typ.(reflect2.SliceType).Elem()
	tElemType := t.Typ.(reflect2.SliceType).Elem()
	elemConverter, isFinalEncoder := LoadConvertFunc(vElemType, tElemType)
	vLength := vType.Len()
	tLength := tType.Len()
	for i := 0; i < vLength && i < tLength; i++ {
		if !isFinalEncoder {
			elemConverter, isFinalEncoder = LoadConvertFunc(vElemType, tElemType)
		}
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
	tObjPtr := v.Typ.UnsafeNew()
	v.Typ.UnsafeSet(tObjPtr, v.Ptr)
	*(*interface{})(t.Ptr) = v.Typ.UnsafeIndirect(tObjPtr)
	return nil
}

func cvtTToPtr(v rt.Value, t rt.Value) error {
	t.Typ = t.Typ.(*reflect2.UnsafePtrType).Elem()
	cvtFunc, _ := LoadConvertFunc(v.Typ, t.Typ)
	if cvtFunc == nil {
		return nil
	}
	if *((*unsafe.Pointer)(t.Ptr)) == nil {
		//pointer to null, we have to allocate memory to hold the value
		newPtr := t.Typ.UnsafeNew()
		err := cvtFunc(v, rt.Value{
			Ptr: newPtr,
			Typ: v.Typ,
		})
		if err != nil {
			return err
		}
		*((*unsafe.Pointer)(t.Ptr)) = newPtr
	} else {
		t.Ptr = *((*unsafe.Pointer)(t.Ptr))
		err := cvtFunc(v, t)
		if err != nil {
			return err
		}
	}
	return nil
}

func cvtPtrToT(v rt.Value, t rt.Value) error {
	v.Typ = v.Typ.(*reflect2.UnsafePtrType).Elem()
	cvtFunc, _ := LoadConvertFunc(v.Typ, t.Typ)
	if cvtFunc == nil {
		return nil
	}
	if *((*unsafe.Pointer)(v.Ptr)) == nil {
		return nil
	} else {
		vObj := v.Typ.UnsafeIndirect(v.Ptr)
		vPtr := reflect2.PtrOf(vObj)
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
	tFieldMap := make(map[string]reflect2.StructField, len(tInfo.Fields))
	for i := 0; i < len(tInfo.Fields); i++ {
		tFieldMap[tInfo.Fields[i].Field.Name()] = tInfo.Fields[i].Field
	}
	vPtr := v.Ptr
	tPtr := t.Ptr
	for i := 0; i < len(vInfo.Fields); i++ {
		f := vInfo.Fields[i].Field
		tf, ok := tFieldMap[f.Name()]
		if !ok {
			continue
		}
		fType := f.Type()
		tfType := tf.Type()
		// 直接使用指针 + 偏移，避免去将指针转换为对象
		childVPtr := pointerOffset(vPtr, f.Offset())
		childTPtr := pointerOffset(tPtr, tf.Offset())
		cvtFunc, _ := LoadConvertFunc(fType, tfType)
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
		return nil
	}
	vKType := vType.Key()
	tKType := tType.Key()
	vElemType := vType.Elem()
	tElemType := tType.Elem()
	elemConverter, isElemFinalEncoder := LoadConvertFunc(vElemType, tElemType)
	keyConverter, isKeyFinalEncoder := LoadConvertFunc(vKType, tKType)
	iter := vType.UnsafeIterate(v.Ptr)
	for iter.HasNext() {
		vKey, vElem := iter.UnsafeNext()
		if !isElemFinalEncoder {
			elemConverter, isElemFinalEncoder = LoadConvertFunc(vElemType, tElemType)
		}
		if !isKeyFinalEncoder {
			keyConverter, isKeyFinalEncoder = LoadConvertFunc(vKType, tKType)
		}
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
		name := f.StructField.Name
		tElem := tElemType.UnsafeNew()
		elemConverter, _ := LoadConvertFunc(fType, tElemType)
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
	tFieldMap := make(map[string]reflect2.StructField, len(tInfo.Fields))
	for i := 0; i < len(tInfo.Fields); i++ {
		tFieldMap[tInfo.Fields[i].Field.Name()] = tInfo.Fields[i].Field
	}
	vElemType := vType.Elem()
	iter := vType.UnsafeIterate(v.Ptr)
	for iter.HasNext() {
		vKey, vElem := iter.UnsafeNext()
		key := *(*string)(vKey)
		tf, ok := tFieldMap[key]
		if !ok {
			continue
		}
		tfType := tf.Type()
		cvtFunc, _ := LoadConvertFunc(vElemType, tfType)
		if cvtFunc == nil {
			continue
		}
		childTPtr := pointerOffset(t.Ptr, tf.Offset())
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

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
}
