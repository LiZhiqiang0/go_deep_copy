package copier_test

import (
	"copier"
	"testing"
)

type User struct {
	Name string
	//Birthday *time.Time
	NickName string
	Role     string
	Age      int32
	FakeAge  *int32
	Notes    []string
	Flags    []byte
	Class    *Class
}

type Class struct {
	Name string
	ID   int64
}

func (user User) DoubleAge() int32 {
	return 2 * user.Age
}

type Employee struct {
	_User *User
	Name  string
	//Birthday  *time.Time
	NickName  string
	Age       int64
	FakeAge   int
	EmployeID int64
	DoubleAge int32
	SuperRule string
	Notes     []string
	Flags     []byte
	Class     interface{}
}

func (employee *Employee) Role(role string) {
	employee.SuperRule = "Super " + role
}

// TestStructToMap 测试结构体到map的转换
func TestStructToMap(t *testing.T) {
	t.Run("basic struct to map", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
			City string
		}

		person := Person{Name: "张三", Age: 25, City: "北京"}
		var result map[string]interface{}

		err := copier.Copy(&result, &person)
		if err != nil {
			t.Errorf("结构体转map失败: %v", err)
		}

		if result["Name"] != "张三" {
			t.Errorf("Name字段转换错误，期望: 张三，实际: %v", result["Name"])
		}
		if result["Age"] != int64(25) {
			t.Errorf("Age字段转换错误，期望: 25，实际: %v", result["Age"])
		}
		if result["City"] != "北京" {
			t.Errorf("City字段转换错误，期望: 北京，实际: %v", result["City"])
		}
	})

	t.Run("struct with pointer to map", func(t *testing.T) {
		type Product struct {
			Name  string
			Price *float64
			Stock int
		}

		price := 99.9
		product := Product{Name: "手机", Price: &price, Stock: 100}
		var result map[string]interface{}

		err := copier.Copy(&result, &product)
		if err != nil {
			t.Errorf("结构体(含指针)转map失败: %v", err)
		}

		if result["Name"] != "手机" {
			t.Errorf("Name字段转换错误，期望: 手机，实际: %v", result["Name"])
		}
		if result["Price"] != price {
			t.Errorf("Price字段转换错误，期望: %v，实际: %v", price, result["Price"])
		}
		if result["Stock"] != int64(100) {
			t.Errorf("Stock字段转换错误，期望: 100，实际: %v", result["Stock"])
		}
	})

	t.Run("nested struct to map", func(t *testing.T) {
		type Address struct {
			Province string
			City     string
		}

		type Student struct {
			Name    string
			Age     int
			Address Address
		}

		student := Student{
			Name: "李四",
			Age:  20,
			Address: Address{
				Province: "广东",
				City:     "深圳",
			},
		}
		var result map[string]interface{}

		err := copier.Copy(&result, &student)
		if err != nil {
			t.Errorf("嵌套结构体转map失败: %v", err)
		}

		if result["Name"] != "李四" {
			t.Errorf("Name字段转换错误，期望: 李四，实际: %v", result["Name"])
		}
		if result["Age"] != int64(20) {
			t.Errorf("Age字段转换错误，期望: 20，实际: %v", result["Age"])
		}
	})
}

// TestStructToMapOnly 测试结构体到map的转换
func TestStructToMapOnly(t *testing.T) {
	t.Run("struct with slice to map", func(t *testing.T) {
		type Skill struct {
			Name  string
			Level int
		}

		type Developer struct {
			Name     string
			Age      int
			Skills   []Skill
			IsActive bool
		}

		developer := Developer{
			Name: "程序员",
			Age:  28,
			Skills: []Skill{
				{Name: "Go", Level: 5},
				{Name: "Python", Level: 4},
			},
			IsActive: true,
		}

		var result map[string]interface{}
		err := copier.Copy(&result, &developer)
		if err != nil {
			t.Errorf("结构体(含切片)转map失败: %v", err)
		}

		if result["Name"] != "程序员" {
			t.Errorf("Name字段转换错误，期望: 程序员，实际: %v", result["Name"])
		}
		if result["Age"] != int64(28) {
			t.Errorf("Age字段转换错误，期望: 28，实际: %v", result["Age"])
		}
		if result["IsActive"] != true {
			t.Errorf("IsActive字段转换错误，期望: true，实际: %v", result["IsActive"])
		}
	})

	t.Run("struct with map field to map", func(t *testing.T) {
		type Config struct {
			Version  string
			Settings map[string]string
			Debug    bool
		}

		config := Config{
			Version: "1.0.0",
			Settings: map[string]string{
				"timeout": "30s",
				"retries": "3",
			},
			Debug: false,
		}

		var result map[string]interface{}
		err := copier.Copy(&result, &config)
		if err != nil {
			t.Errorf("结构体(含map字段)转map失败: %v", err)
		}

		if result["Version"] != "1.0.0" {
			t.Errorf("Version字段转换错误，期望: 1.0.0，实际: %v", result["Version"])
		}
		if result["Debug"] != false {
			t.Errorf("Debug字段转换错误，期望: false，实际: %v", result["Debug"])
		}
	})
}

// TestBasicCopy 测试基础拷贝功能
func TestBasicCopy(t *testing.T) {
	t.Run("struct to struct", func(t *testing.T) {
		user := User{
			Name:     "John",
			NickName: "johnny",
			Age:      25,
			Role:     "Admin",
			Notes:    []string{"note1", "note2"},
			Flags:    []byte{1, 2, 3},
		}

		employee := Employee{}
		err := copier.Copy(&employee, &user)

		if err != nil {
			t.Errorf("Copy failed: %v", err)
		}

		if employee.Name != user.Name {
			t.Errorf("Name not copied: got %s, want %s", employee.Name, user.Name)
		}
		if employee.NickName != user.NickName {
			t.Errorf("NickName not copied: got %s, want %s", employee.NickName, user.NickName)
		}
		if employee.Age != int64(user.Age) {
			t.Errorf("Age not copied correctly: got %d, want %d", employee.Age, user.Age)
		}
	})

	t.Run("slice to slice", func(t *testing.T) {
		users := []User{
			{Name: "John", Age: 25},
			{Name: "Jane", Age: 30},
		}

		var employees []Employee
		err := copier.Copy(&employees, &users)

		if err != nil {
			t.Errorf("Copy failed: %v", err)
		}

		if len(employees) != len(users) {
			t.Errorf("Slice length mismatch: got %d, want %d", len(employees), len(users))
		}

		for i := range users {
			if employees[i].Name != users[i].Name {
				t.Errorf("Name not copied at index %d: got %s, want %s", i, employees[i].Name, users[i].Name)
			}
			if employees[i].Age != int64(users[i].Age) {
				t.Errorf("Age not copied correctly at index %d: got %d, want %d", i, employees[i].Age, users[i].Age)
			}
		}
	})
}

// TestCopyWithPointer 测试指针字段拷贝
func TestCopyWithPointer(t *testing.T) {
	fakeAge := int32(30)
	user := User{
		Name:    "John",
		FakeAge: &fakeAge,
	}

	employee := Employee{}
	err := copier.Copy(&employee, &user)

	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}

	if employee.FakeAge != int(*user.FakeAge) {
		t.Errorf("FakeAge not copied correctly: got %d, want %d", employee.FakeAge, *user.FakeAge)
	}
}

// TestCopyWithNestedStruct 测试嵌套结构体拷贝
func TestCopyWithNestedStruct(t *testing.T) {
	user := User{
		Name: "John",
		Class: &Class{
			Name: "Math",
			ID:   101,
		},
	}

	employee := Employee{}
	err := copier.Copy(&employee, &user)

	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}

	if employee.Class == nil {
		t.Error("Class field not copied")
	}
	class, ok := employee.Class.(Class)
	if !ok {
		t.Error("Class field type assertion failed")
	}

	if class.Name != user.Class.Name {
		t.Errorf("Class Name not copied: got %s, want %s", class.Name, user.Class.Name)
	}
	if class.ID != user.Class.ID {
		t.Errorf("Class ID not copied: got %d, want %d", class.ID, user.Class.ID)
	}
}

// TestCopyWithTags 测试标签功能
func TestCopyWithTags(t *testing.T) {
	type SourceWithTags struct {
		Name   string
		Secret string
		ID     int
	}

	type TargetWithTags struct {
		Name     string
		Secret   string `copier:"-"`
		TargetID int    `copier:"ID"`
	}

	source := SourceWithTags{
		Name:   "John",
		Secret: "secret123",
		ID:     1001,
	}

	target := TargetWithTags{}

	err := copier.Copy(&target, &source)

	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}

	if target.Name != source.Name {
		t.Errorf("Name not copied: got %s, want %s", target.Name, source.Name)
	}

	if target.Secret != "" {
		t.Errorf("Secret should be ignored: got %s", target.Secret)
	}

	if target.TargetID != source.ID {
		t.Errorf("ID not mapped correctly: got %d, want %d", target.TargetID, source.ID)
	}
}

// TestCopyWithOption 测试拷贝选项
func TestCopyWithOption(t *testing.T) {
	type SimpleStruct struct {
		Name  string
		Age   int
		Empty string
	}

	source := SimpleStruct{
		Name: "John",
		Age:  25,
	}

	target := SimpleStruct{
		Name:  "Existing",
		Empty: "NotEmpty",
	}

	err := copier.Copy(&target, &source)

	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}

	if target.Name != source.Name {
		t.Errorf("Name should be overwritten: got %s, want %s", target.Name, source.Name)
	}
}

// TestCopyErrorCases 测试错误情况
func TestCopyErrorCases(t *testing.T) {
	t.Run("nil destination", func(t *testing.T) {
		source := User{Name: "John"}
		err := copier.Copy(nil, &source)

		if err == nil {
			t.Error("Expected error when destination is nil")
		}
	})

	t.Run("nil source", func(t *testing.T) {
		employee := Employee{}
		err := copier.Copy(&employee, nil)

		if err == nil {
			t.Error("Expected error when source is nil")
		}
	})

	t.Run("non-pointer destination", func(t *testing.T) {
		source := User{Name: "John"}
		employee := Employee{}
		err := copier.Copy(employee, &source)

		if err == nil {
			t.Error("Expected error when destination is not a pointer")
		}
	})
}

// TestCopyMap 测试map拷贝
func TestCopyMap(t *testing.T) {
	source := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	var target map[string]int64
	err := copier.Copy(&target, &source)

	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}

	if len(target) != len(source) {
		t.Errorf("Map length mismatch: got %d, want %d", len(target), len(source))
	}

	for k, v := range source {
		if target[k] != int64(v) {
			t.Errorf("Map value not copied correctly for key %s: got %d, want %d", k, target[k], v)
		}
	}
}

// TestCopyDifferentTypes 测试不同类型之间的拷贝
func TestCopyDifferentTypes(t *testing.T) {
	type IntStruct struct {
		Value int
	}

	type StringStruct struct {
		Value string
	}

	source := IntStruct{Value: 42}
	target := StringStruct{}

	err := copier.Copy(&target, &source)

	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}
	if target.Value != "42" {
		t.Errorf("Value not copied correctly: got %s, want %s", target.Value, "42")
	}
}

// TestDeepCopy 测试深拷贝
func TestDeepCopy(t *testing.T) {
	source := User{
		Name:  "John",
		Notes: []string{"note1", "note2"},
		Class: &Class{
			Name: "Math",
			ID:   101,
		},
	}

	var target User
	err := copier.CopyWithOption(&target, &source, copier.Option{DeepCopy: true})

	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}

	// 修改源数据，检查目标数据是否受影响
	source.Notes[0] = "modified"
	source.Class.Name = "Modified"

	if target.Notes[0] == "modified" {
		t.Error("Deep copy failed: slice was not deep copied")
	}

	if target.Class.Name == "Modified" {
		t.Error("Deep copy failed: nested struct was not deep copied")
	}
}

// TestMapToStruct 测试map到结构体的转换
func TestMapToStruct(t *testing.T) {
	t.Run("basic map to struct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
			City string
		}

		source := map[string]interface{}{
			"Name": "张三",
			"Age":  int64(25),
			"City": "北京",
		}

		var target Person
		err := copier.Copy(&target, &source)
		if err != nil {
			t.Errorf("Map转结构体失败: %v", err)
		}

		if target.Name != "张三" {
			t.Errorf("Name字段转换错误，期望: 张三，实际: %s", target.Name)
		}
		if target.Age != 25 {
			t.Errorf("Age字段转换错误，期望: 25，实际: %d", target.Age)
		}
		if target.City != "北京" {
			t.Errorf("City字段转换错误，期望: 北京，实际: %s", target.City)
		}
	})

	t.Run("map to struct with type conversion", func(t *testing.T) {
		type Product struct {
			Name  string
			Price float64
			Stock int64
		}

		source := map[string]interface{}{
			"Name":  "手机",
			"Price": int64(99),
			"Stock": int64(100),
		}

		var target Product
		err := copier.Copy(&target, &source)
		if err != nil {
			t.Errorf("Map转结构体(类型转换)失败: %v", err)
		}

		if target.Name != "手机" {
			t.Errorf("Name字段转换错误，期望: 手机，实际: %s", target.Name)
		}
		if target.Price != 99.0 {
			t.Errorf("Price字段转换错误，期望: 99.0，实际: %f", target.Price)
		}
		if target.Stock != 100 {
			t.Errorf("Stock字段转换错误，期望: 100，实际: %d", target.Stock)
		}
	})

	t.Run("map to struct with missing fields", func(t *testing.T) {
		type Employee struct {
			Name     string
			Age      int
			Salary   float64
			IsActive bool
		}

		source := map[string]interface{}{
			"Name": "李四",
			"Age":  int64(30),
		}

		var target Employee
		err := copier.Copy(&target, &source)
		if err != nil {
			t.Errorf("Map转结构体(缺失字段)失败: %v", err)
		}

		if target.Name != "李四" {
			t.Errorf("Name字段转换错误，期望: 李四，实际: %s", target.Name)
		}
		if target.Age != 30 {
			t.Errorf("Age字段转换错误，期望: 30，实际: %d", target.Age)
		}
		if target.Salary != 0 {
			t.Errorf("Salary字段应该是默认值，期望: 0，实际: %f", target.Salary)
		}
		if target.IsActive != false {
			t.Errorf("IsActive字段应该是默认值，期望: false，实际: %t", target.IsActive)
		}
	})

	t.Run("map with different key types", func(t *testing.T) {
		type Config struct {
			Version string
			Debug   bool
			Port    int
		}

		source := map[string]interface{}{
			"Version": "1.0.0",
			"Debug":   true,
			"Port":    int64(8080),
		}

		var target Config
		err := copier.Copy(&target, &source)
		if err != nil {
			t.Errorf("Map转结构体(不同key类型)失败: %v", err)
		}

		if target.Version != "1.0.0" {
			t.Errorf("Version字段转换错误，期望: 1.0.0，实际: %s", target.Version)
		}
		if target.Debug != true {
			t.Errorf("Debug字段转换错误，期望: true，实际: %t", target.Debug)
		}
		if target.Port != 8080 {
			t.Errorf("Port字段转换错误，期望: 8080，实际: %d", target.Port)
		}
	})
}

// TestMapToStructWithNested 测试嵌套结构体与map的转换
func TestMapToStructWithNested(t *testing.T) {
	t.Run("nested struct to map", func(t *testing.T) {
		type Address struct {
			Province string
			City     string
			Street   string
		}

		type Person struct {
			Name    string
			Age     int
			Address Address
		}

		person := Person{
			Name: "王五",
			Age:  35,
			Address: Address{
				Province: "上海",
				City:     "上海",
				Street:   "南京路",
			},
		}

		var result map[string]interface{}
		err := copier.Copy(&result, &person)
		if err != nil {
			t.Errorf("嵌套结构体转map失败: %v", err)
		}

		if result["Name"] != "王五" {
			t.Errorf("Name字段转换错误，期望: 王五，实际: %v", result["Name"])
		}

		// 检查嵌套的Address结构
		if addr, ok := result["Address"].(Address); ok {
			if addr.Province != "上海" {
				t.Errorf("Address.Province字段转换错误，期望: 上海，实际: %s", addr.Province)
			}
		} else {
			t.Error("Address字段类型转换失败")
		}
	})

	t.Run("map to nested struct", func(t *testing.T) {
		type Address struct {
			Province string
			City     string
			Street   string
		}

		type Person struct {
			Name    string
			Age     int
			Address Address
		}

		source := map[string]interface{}{
			"Name": "赵六",
			"Age":  int64(40),
			"Address": Address{
				Province: "广东",
				City:     "深圳",
				Street:   "科技园",
			},
		}

		var target Person
		err := copier.Copy(&target, &source)
		if err != nil {
			t.Errorf("Map转嵌套结构体失败: %v", err)
		}

		if target.Name != "赵六" {
			t.Errorf("Name字段转换错误，期望: 赵六，实际: %s", target.Name)
		}
		if target.Age != 40 {
			t.Errorf("Age字段转换错误，期望: 40，实际: %d", target.Age)
		}
		if target.Address.Province != "广东" {
			t.Errorf("Address.Province字段转换错误，期望: 广东，实际: %s", target.Address.Province)
		}
		if target.Address.City != "深圳" {
			t.Errorf("Address.City字段转换错误，期望: 深圳，实际: %s", target.Address.City)
		}
		if target.Address.Street != "科技园" {
			t.Errorf("Address.Street字段转换错误，期望: 科技园，实际: %s", target.Address.Street)
		}
	})
}

// TestComplexMapStructConversion 测试复杂的结构体与map转换
func TestComplexMapStructConversion(t *testing.T) {
	t.Run("struct with slice and map to map", func(t *testing.T) {
		type Skill struct {
			Name  string
			Level int
		}

		type Employee struct {
			Name     string
			Age      int
			Skills   []Skill
			Projects map[string]string
			Active   bool
		}

		employee := Employee{
			Name: "高级程序员",
			Age:  32,
			Skills: []Skill{
				{Name: "Go", Level: 5},
				{Name: "Java", Level: 4},
				{Name: "Python", Level: 3},
			},
			Projects: map[string]string{
				"proj1": "电商平台",
				"proj2": "支付系统",
			},
			Active: true,
		}

		var result map[string]interface{}
		err := copier.Copy(&result, &employee)
		if err != nil {
			t.Errorf("复杂结构体转map失败: %v", err)
		}

		if result["Name"] != "高级程序员" {
			t.Errorf("Name字段转换错误，期望: 高级程序员，实际: %v", result["Name"])
		}
		if result["Active"] != true {
			t.Errorf("Active字段转换错误，期望: true，实际: %v", result["Active"])
		}

		// 检查切片
		if skills, ok := result["Skills"].([]Skill); ok {
			if len(skills) != 3 {
				t.Errorf("Skills切片长度错误，期望: 3，实际: %d", len(skills))
			}
		} else {
			t.Error("Skills字段类型转换失败")
		}

		// 检查map
		if projects, ok := result["Projects"].(map[string]string); ok {
			if len(projects) != 2 {
				t.Errorf("Projects map长度错误，期望: 2，实际: %d", len(projects))
			}
		} else {
			t.Error("Projects字段类型转换失败")
		}
	})

	t.Run("complex map to struct", func(t *testing.T) {
		type Department struct {
			Name string
			Code int
		}

		type Company struct {
			Name       string
			Department Department
		}

		source := map[string]interface{}{
			"Name": "科技公司",
			"Department": Department{
				Name: "研发部",
				Code: 1001,
			},
		}

		var target Company
		err := copier.Copy(&target, &source)
		if err != nil {
			t.Errorf("复杂map转结构体失败: %v", err)
		}

		if target.Name != "科技公司" {
			t.Errorf("Name字段转换错误，期望: 科技公司，实际: %s", target.Name)
		}
		if target.Department.Name != "研发部" {
			t.Errorf("Department.Name字段转换错误，期望: 研发部，实际: %s", target.Department.Name)
		}
		if target.Department.Code != 1001 {
			t.Errorf("Department.Code字段转换错误，期望: 1001，实际: %d", target.Department.Code)
		}
	})
}
