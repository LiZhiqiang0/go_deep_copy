package copier_test

import (
	"database/sql"
	"time"
)

type TypeStruct1 struct {
	Field1 string
	Field2 string
	Field3 TypeStruct2
	Field4 *TypeStruct2
	Field5 []*TypeStruct2
	Field6 []TypeStruct2
	Field7 []*TypeStruct2
	Field8 []TypeStruct2
	Field9 []string
}

type TypeStruct2 struct {
	Field1 int
	Field2 string
	Field3 []TypeStruct2
	Field4 *TypeStruct2
	Field5 *TypeStruct2
	Field9 string
}

type TypeStruct3 struct {
	Field1 interface{}
	Field2 string
	Field3 TypeStruct4
	Field4 *TypeStruct4
	Field5 []*TypeStruct4
	Field6 []*TypeStruct4
	Field7 []TypeStruct4
	Field8 []TypeStruct4
}

type TypeStruct4 struct {
	field1 int
	Field2 string
}

func (t *TypeStruct4) Field1(i int) {
	t.field1 = i
}

type TypeBaseStruct5 struct {
	A bool
	B byte
	C float64
	D int16
	E int32
	F int64
	G time.Time
	H string
}

type TypeSqlNullStruct6 struct {
	A sql.NullBool    `json:"a"`
	B sql.NullByte    `json:"b"`
	C sql.NullFloat64 `json:"c"`
	D sql.NullInt16   `json:"d"`
	E sql.NullInt32   `json:"e"`
	F sql.NullInt64   `json:"f"`
	G sql.NullTime    `json:"g"`
	H sql.NullString  `json:"h"`
}
