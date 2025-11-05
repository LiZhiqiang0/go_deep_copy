package go_deep_copy_test

import (
	"encoding/json"
	"github.com/LiZhiqiang0/go_deep_copy"
	"github.com/jinzhu/copier"
	"testing"
)

type Book struct {
	BookId  int
	BookIds []int
	Title   string
	Titles  []string
	Price   float64
	Prices  []float64
	hot     bool
	Hots    []bool
	Author  Author
	Authors []Author
	Weights []int
}

type Author struct {
	Name string
	Age  int
	Male bool
}

var book = Book{
	BookId:  12125925,
	BookIds: []int{-2147483648, 2147483647},
	Title:   "未来简史-从智人到智神",
	Titles:  []string{"hello", "world"},
	Price:   40.8,
	Prices:  []float64{-0.1, 0.1},
	hot:     true,
	Hots:    []bool{true, true, true},
	Author:  author,
	Authors: []Author{author, author, author},
	Weights: nil,
}

var author = Author{
	Name: "json",
	Age:  99,
	Male: true,
}

func BenchmarkCopyStruct(b *testing.B) {

	runs := []struct {
		name string
		f    func()
	}{
		{"copier",
			func() {
				a := Book{}
				copier.CopyWithOption(&a, &book, copier.Option{DeepCopy: true})
			},
		},
		{"go_deep_copy",
			func() {
				a := Book{}
				go_deep_copy.DeepCopy(&book, &a)
			},
		},
		{"json",
			func() {
				data, _ := json.Marshal(book)
				a := Book{}
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
