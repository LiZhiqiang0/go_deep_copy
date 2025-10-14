package copier_test

import (
	"copier"
	"encoding/json"
	"fmt"
	c1 "github.com/jinzhu/copier"
	"testing"
)

func BenchmarkCopyStruct(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	for x := 0; x < b.N; x++ {
		a := Employee{}
		c1.Copy(&a, &user)
		fmt.Println(a)
	}
}

func BenchmarkCopyStruct2(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []uint8{'x'}}
	for x := 0; x < b.N; x++ {
		a := Employee{}
		copier.Copy(&a, &user)
		fmt.Println(a)
	}
}

func BenchmarkNamaCopy(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	for x := 0; x < b.N; x++ {
		employee := &Employee{
			Name:      user.Name,
			NickName:  user.NickName,
			Age:       int64(user.Age),
			FakeAge:   int(user.FakeAge),
			DoubleAge: user.DoubleAge(),
		}

		for _, note := range user.Notes {
			employee.Notes = append(employee.Notes, note)
		}
		employee.Role(user.Role)
	}
}

func BenchmarkJsonMarshalCopy(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	for x := 0; x < b.N; x++ {
		data, _ := json.Marshal(user)
		var employee Employee
		json.Unmarshal(data, &employee)

		employee.DoubleAge = user.DoubleAge()
		employee.Role(user.Role)
	}
}
