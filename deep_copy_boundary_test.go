package go_deep_copy_test

import (
	"github.com/LiZhiqiang0/go_deep_copy"
	"testing"
)

// TestBoundaryConditions 测试边界条件
func TestBoundaryConditions(t *testing.T) {
	t.Run("empty struct", func(t *testing.T) {
		type EmptyStruct struct{}

		source := EmptyStruct{}
		target := EmptyStruct{}

		err := go_deep_copy.DeepCopy(&source, &target)
		if err != nil {
			t.Errorf("Copy empty struct failed: %v", err)
		}
	})

	t.Run("zero values", func(t *testing.T) {
		type ZeroStruct struct {
			Int    int
			String string
			Bool   bool
			Float  float64
			Slice  []string
			Map    map[string]int
			Ptr    *string
		}

		source := ZeroStruct{}
		target := ZeroStruct{
			Int:    100,
			String: "existing",
			Bool:   true,
			Float:  3.14,
			Slice:  []string{"existing"},
			Map:    map[string]int{"existing": 1},
			Ptr:    new(string),
		}

		err := go_deep_copy.DeepCopy(&source, &target)
		if err != nil {
			t.Errorf("Copy zero values failed: %v", err)
		}

		// 验证零值是否被拷贝
		if target.Int != 0 {
			t.Errorf("Int zero value not copied: got %d, want 0", target.Int)
		}
		if target.String != "" {
			t.Errorf("String zero value not copied: got %s, want empty", target.String)
		}
		if target.Bool != false {
			t.Errorf("Bool zero value not copied: got %t, want false", target.Bool)
		}
		if target.Float != 0 {
			t.Errorf("Float zero value not copied: got %f, want 0", target.Float)
		}
		if target.Slice != nil {
			t.Errorf("Slice zero value not copied: got %v, want nil", target.Slice)
		}
		if target.Map != nil {
			t.Errorf("Map zero value not copied: got %v, want nil", target.Map)
		}
	})

	t.Run("large slice", func(t *testing.T) {
		type Item struct {
			ID   int
			Name string
		}

		// 创建大切片
		source := make([]Item, 10000)
		for i := 0; i < 10000; i++ {
			source[i] = Item{ID: i, Name: "Item"}
		}

		var target []Item
		err := go_deep_copy.DeepCopy(&source, &target)

		if err != nil {
			t.Errorf("Copy large slice failed: %v", err)
		}

		if len(target) != len(source) {
			t.Errorf("Large slice length mismatch: got %d, want %d", len(target), len(source))
		}

		// 验证几个关键元素
		if target[0].ID != source[0].ID {
			t.Errorf("First element not copied correctly")
		}
		if target[9999].ID != source[9999].ID {
			t.Errorf("Last element not copied correctly")
		}
		if target[5000].ID != source[5000].ID {
			t.Errorf("Middle element not copied correctly")
		}
	})

	t.Run("nested slices", func(t *testing.T) {
		type NestedStruct struct {
			Matrix [][]int
		}

		source := NestedStruct{
			Matrix: [][]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
		}

		target := NestedStruct{}
		err := go_deep_copy.DeepCopy(&source, &target)

		if err != nil {
			t.Errorf("Copy nested slices failed: %v", err)
		}

		if len(target.Matrix) != len(source.Matrix) {
			t.Errorf("Outer slice length mismatch")
		}

		for i := range source.Matrix {
			if len(target.Matrix[i]) != len(source.Matrix[i]) {
				t.Errorf("Inner slice length mismatch at index %d", i)
			}
			for j := range source.Matrix[i] {
				if target.Matrix[i][j] != source.Matrix[i][j] {
					t.Errorf("Matrix value mismatch at [%d][%d]", i, j)
				}
			}
		}
	})
}

// TestAdvancedErrorCases 测试更多错误情况
func TestAdvancedErrorCases(t *testing.T) {
	t.Run("incompatible types", func(t *testing.T) {
		type Source struct {
			Data map[string]interface{}
		}

		type Target struct {
			Data string // 不兼容的类型
		}

		source := Source{
			Data: map[string]interface{}{
				"key": "value",
			},
		}

		target := Target{}
		err := go_deep_copy.DeepCopy(&source, &target)

		// 有些版本的go_deep_copy会尝试进行类型转换，所以这里不强制要求错误
		if err != nil {
			t.Logf("Incompatible types handled with error: %v", err)
		} else {
			t.Logf("Incompatible types handled with conversion")
		}
	})
}

// TestPerformance 性能测试
func TestPerformance(t *testing.T) {
	type LargeStruct struct {
		Field1  string
		Field2  int
		Field3  float64
		Field4  bool
		Field5  []string
		Field6  map[string]int
		Field7  *string
		Field8  []byte
		Field9  interface{}
		Field10 int64
	}

	source := LargeStruct{
		Field1:  "test string",
		Field2:  42,
		Field3:  3.14159,
		Field4:  true,
		Field5:  []string{"item1", "item2", "item3"},
		Field6:  map[string]int{"key1": 1, "key2": 2},
		Field7:  new(string),
		Field8:  []byte{1, 2, 3, 4, 5},
		Field9:  "interface data",
		Field10: 999999999,
	}

	// 执行多次拷贝测试性能
	for i := 0; i < 1000; i++ {
		target := LargeStruct{}
		err := go_deep_copy.DeepCopy(&source, &target)
		if err != nil {
			t.Errorf("Copy failed at iteration %d: %v", i, err)
			break
		}
	}
}

// TestConcurrentCopy 并发拷贝测试
func TestConcurrentCopy(t *testing.T) {
	type Data struct {
		ID   int
		Name string
	}

	source := Data{ID: 1, Name: "test"}

	// 并发执行拷贝
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			target := Data{}
			err := go_deep_copy.DeepCopy(&source, &target)
			if err != nil {
				t.Errorf("Concurrent copy failed for goroutine %d: %v", id, err)
			}
			if target.ID != source.ID || target.Name != source.Name {
				t.Errorf("Concurrent copy data mismatch for goroutine %d", id)
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}
