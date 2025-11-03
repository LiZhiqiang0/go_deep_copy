package go_deep_copy_test

import (
	"encoding/json"
	"github.com/LiZhiqiang0/go_deep_copy"
	"github.com/jinzhu/copier"
	"testing"
)

func BenchmarkCopyStruct(b *testing.B) {
	var fakeAge int32 = 12
	user := User{Name: "Jinzhu", NickName: "jinzhu", Age: 18, FakeAge: &fakeAge, Role: "Admin", Notes: []string{"hello world", "welcome"}, Flags: []byte{'x'}}
	runs := []struct {
		name string
		f    func()
	}{
		{"copier",
			func() {
				a := User{}
				copier.CopyWithOption(&a, &user, copier.Option{DeepCopy: true})
			},
		},
		{"go_deep_copy",
			func() {
				a := User{}
				go_deep_copy.DeepCopy(&user, &a)
			},
		},
		{"json",
			func() {
				data, _ := json.Marshal(user)
				a := User{}
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
