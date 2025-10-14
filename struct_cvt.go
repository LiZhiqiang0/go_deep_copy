package copier

import (
	"copier/rt"
	"reflect"
)

func loadStructFieldsInfo(vt reflect.Type) *StructInfo {
	typ := rt.UnpackType(vt)
	if structInfo, ok := structInfoCache.Get(typ); ok {
		return structInfo.(*StructInfo)
	}
	structInfo := getStructInfo(vt)
	// 并发写；无法处理递归
	structInfoCache.Set(typ, structInfo)
	return structInfo
}

func getStructInfo(typ reflect.Type) (structInfo *StructInfo) {
	stdStructInfo := TypeFields(typ)
	return convertStdFieldInfo(&stdStructInfo, typ)
}

func newStructEncoder(vt, rt reflect.Type) ConvertFunc {
	vInfo := loadStructFieldsInfo(vt)
	tInfo := loadStructFieldsInfo(rt)
	info := &StructCvt{
		vStructInfo: vInfo,
		tStructInfo: tInfo,
	}
	return info.convert
}

type StructCvt struct {
	vStructInfo *StructInfo
	tStructInfo *StructInfo
}

// val *Value
func (se *StructCvt) convert(v rt.Value, t rt.Value) error {
	tFieldMap := make(map[string]*StructField, len(se.tStructInfo.list))
	for i := 0; i < len(se.tStructInfo.list); i++ {
		tFieldMap[se.tStructInfo.list[i].name] = &se.tStructInfo.list[i]
	}
	for i := 0; i < len(se.vStructInfo.list); i++ {
		f := &se.vStructInfo.list[i]
		tf, ok := tFieldMap[f.name]
		if !ok {
			continue
		}
		// 直接使用指针 + 偏移，避免去将指针转换为对象
		childVPtr := pointerOffset(v.Ptr, f.offset)
		childTPtr := pointerOffset(t.Ptr, tf.offset)
		cvtFunc, _ := LoadConvertFunc(f.typ, tf.typ)

		err := cvtFunc(rt.Value{
			Ptr:  childVPtr,
			Typ:  f.goType,
			Flag: uintptr(f.goType.Kind()),
		}, rt.Value{
			Ptr:  childTPtr,
			Typ:  tf.goType,
			Flag: uintptr(tf.goType.Kind()),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func convertStdFieldInfo(stdInfo *structFields, typ reflect.Type) (structInfo *StructInfo) {
	if stdInfo == nil {
		return nil
	}
	structInfo = &StructInfo{
		list: make([]StructField, 0, len(stdInfo.list)),
		typ:  typ,
	}
	for _, field := range stdInfo.list {
		curTyp := typ
		// curTyp 可能是 ptr  而 field.typ 不会
		var offset uintptr
		for _, id := range field.index {
			offset += curTyp.Field(id).Offset
			curTyp = curTyp.Field(id).Type
		}

		structInfo.list = append(structInfo.list, StructField{
			name:        field.name,
			nameBytes:   field.nameBytes,
			nameNonEsc:  field.nameNonEsc,
			nameEscHTML: field.nameEscHTML,
			tag:         field.tag,
			index:       field.index,
			typ:         curTyp,
			goType:      rt.UnpackType(field.typ),
			omitEmpty:   field.omitEmpty,
			quoted:      field.quoted,
			offset:      offset,
		})
	}
	return structInfo
}
