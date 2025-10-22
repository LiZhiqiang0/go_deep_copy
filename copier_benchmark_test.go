package copier_test

import (
	"copier"
	"encoding/json"
	c1 "github.com/jinzhu/copier"
	"testing"
)

func BenchmarkCopyStruct(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	//user := map[string]interface{}{
	//	"Name":     "Jinzhu",
	//	"NickName": "jinzhu",
	//	"Age":      18,
	//	"FakeAge":  &fakeAge,
	//	"Role":     "Admin",
	//	"Notes":    []string{"hello world", "welcome"},
	//	"Flags":    []byte{'x'},
	//}
	runs := []struct {
		name string
		f    func()
	}{
		{"std",
			func() {
				a := User{}
				c1.CopyWithOption(&a, &user, c1.Option{DeepCopy: true})
				// println(a)
			},
		},
		{"my",
			func() {
				a := User{}
				copier.Copy(&a, &user)
				//fmt.Println(a)
			},
		},
		{"json",
			func() {
				data, _ := json.Marshal(user)
				a := User{}
				json.Unmarshal(data, &a)
				// println(a)
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
