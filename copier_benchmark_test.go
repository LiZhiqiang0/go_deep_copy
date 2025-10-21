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
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	runs := []struct {
		name string
		f    func()
	}{
		{"std",
			func() {
				a := Employee{}
				c1.Copy(&a, &user)
			},
		},
		{"my",
			func() {
				a := Employee{}
				copier.Copy(&a, &user)
				fmt.Println(a)
			},
		},
		{"json",
			func() {
				data, _ := json.Marshal(user)
				a := Employee{}
				json.Unmarshal(data, &a)
			},
		},
	}
	for _, r := range runs {
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.f()
			}
			b.StopTimer()
		})
	}
}

func BenchmarkCopyStructStd(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	runs := []struct {
		name string
		f    func()
	}{
		{"std",
			func() {
				a := Employee{}
				c1.Copy(&a, &user)
			},
		},
	}
	for _, r := range runs {
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.f()
			}
			b.StopTimer()
		})
	}
}

func BenchmarkCopyStructMy(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	runs := []struct {
		name string
		f    func()
	}{
		{"my",
			func() {
				a := Employee{}
				copier.Copy(&a, &user)
			},
		},
	}
	for _, r := range runs {
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.f()
			}
			b.StopTimer()
		})
	}
}

func BenchmarkNamaCopy(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	for x := 0; x < b.N; x++ {
		employee := &Employee{
			Name:      user.Name,
			NickName:  user.NickName,
			Age:       int64(user.Age),
			FakeAge:   int(*user.FakeAge),
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
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	runs := []struct {
		name string
		f    func()
	}{
		{"json",
			func() {
				data, _ := json.Marshal(user)
				var employee Employee
				json.Unmarshal(data, &employee)

				employee.DoubleAge = user.DoubleAge()
				employee.Role(user.Role)
			},
		},
	}
	for _, r := range runs {
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.f()
			}
			b.StopTimer()
		})
	}
}
