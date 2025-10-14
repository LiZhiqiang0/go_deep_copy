package copier

import (
	"copier/rt"
	"github.com/modern-go/reflect2"
	"reflect"
	"unsafe"
)

type sliceConverter struct {
	elemConverter  ConvertFunc
	isFinalEncoder bool // 是否是最终版本encoder

	vElemType reflect.Type
	tElemType reflect.Type
}

type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func (ae *sliceConverter) convert(v, t reflect2.Type) error {
	vHeader := (*SliceHeader)(v.Ptr)
	length := vHeader.Len
	vSize := ae.vElemType.Size()
	tSize := ae.tElemType.Size()
	tSlice := reflect.MakeSlice(reflect.SliceOf(ae.tElemType), length, length)
	tPtr := rt.ReflectValueToValue(&tSlice)
	t = *tPtr
	tHeader := (*SliceHeader)(t.Ptr)
	for i := 0; i < length; i++ {
		if !ae.isFinalEncoder {
			ae.elemConverter, ae.isFinalEncoder = LoadConvertFunc(ae.vElemType, ae.tElemType)
		}
		tIndexPtr := pointerOffset(tHeader.Data, uintptr(i)*tSize)
		err := ae.elemConverter(rt.Value{
			Ptr:  pointerOffset(vHeader.Data, uintptr(i)*vSize),
			Typ:  rt.UnpackType(ae.vElemType),
			Flag: uintptr(ae.vElemType.Kind()),
		}, rt.Value{
			Ptr:  tIndexPtr,
			Typ:  rt.UnpackType(ae.tElemType),
			Flag: uintptr(ae.tElemType.Kind()),
		})
		if err != nil {
			return err
		}
	}
	reflectT := rt.ValueToReflectValue(&t)
	(*reflectT).Set(tSlice)
	return nil
}

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
}
