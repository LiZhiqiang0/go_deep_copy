# Go Deep Copy - 高性能深拷贝库

Go Deep Copy 是一个高性能的 Go 语言深拷贝库，专注于结构体、切片、Map 等复杂数据结构的深度复制。它采用反射和代码生成技术，提供了快速、安全、灵活的数据复制功能。性能优于github.com/jinzhu/copier和json序列化方式深拷贝

## 🌟 核心特性

### 深拷贝支持
- **结构体深拷贝**：完整复制嵌套结构体，确保数据独立性
- **切片深拷贝**：复制切片底层数据，避免共享底层数组
- **Map 深拷贝**：递归复制 Map 中的所有值
- **指针深拷贝**：正确处理指针字段，创建新的内存空间

### 智能类型转换
- **自动类型转换**：支持基本类型之间的自动转换（int ↔ int64 ↔ float64 ↔ string）
- **结构体互转**：不同结构体之间的字段映射和转换
- **Map ↔ 结构体**：双向转换，支持复杂嵌套结构
- **切片类型转换**：支持不同元素类型的切片转换

### 高级功能
- **字段映射**：支持不同名称字段之间的映射
- **高性能优化**：使用 unsafe 包和反射优化，比标准反射更快
- **并发安全**：支持并发环境下的深拷贝操作

## 🚀 快速开始

### 安装

```bash
go get -u github.com/LiZhiqiang0/go_deep_copy
```

### 基础使用

```go
package main

import (
    "fmt"
    "github.com/LiZhiqiang0/go_deep_copy"
)

type User struct {
    Name  string
    Age   int
    Email string
}

type Employee struct {
    Name  string
    Age   int
    Email string
}

func main() {
    // 结构体到结构体复制
    user := User{Name: "张三", Age: 30, Email: "zhangsan@example.com"}
    var employee Employee
    
    err := go_deep_copy.DeepCopy(&user, &employee)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Employee: %+v\n", employee)
    // 输出: Employee: {Name:张三 Age:30 Email:zhangsan@example.com}
}
```

## 📖 详细功能

### 1. 结构体深拷贝

```go
package main

import (
    "fmt"
    "github.com/LiZhiqiang0/go_deep_copy"
)

type Address struct {
    City    string
    Country string
}

type Person struct {
    Name    string
    Age     int
    Address *Address
}

func main() {
    person1 := &Person{
        Name: "李四",
        Age:  25,
        Address: &Address{
            City:    "北京",
            Country: "中国",
        },
    }
    
    var person2 Person
    err := go_deep_copy.DeepCopy(person1, &person2)
    if err != nil {
        panic(err)
    }
    
    // 修改原始数据，验证深拷贝
    person1.Address.City = "上海"
    fmt.Printf("person1.Address.City: %s\n", person1.Address.City) // 上海
    fmt.Printf("person2.Address.City: %s\n", person2.Address.City) // 北京
}
```

### 2. Map 与结构体互转

```go
package main

import (
    "fmt"
    "github.com/LiZhiqiang0/go_deep_copy"
)

type User struct {
    Name  string
    Age   int
    Email string
}

func main() {
    // Map 到结构体
    mapData := map[string]interface{}{
        "Name": "王五",
        "Age":  int64(35),
        "Email": "wangwu@example.com",
    }

    var user User
    err := go_deep_copy.DeepCopy(&mapData, &user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Map to User: %+v\n", user)

    // 结构体到 Map
    var result map[string]interface{}
    err = go_deep_copy.DeepCopy(&user, &result)
    if err != nil {
        panic(err)
    }
    fmt.Printf("User to Map: %+v\n", result)
}
```

### 3. 切片深拷贝

```go
package main

import (
    "fmt"
    "github.com/LiZhiqiang0/go_deep_copy"
)

type Product struct {
    ID    int
    Name  string
    Price float64
}

func main() {
    products1 := []Product{
        {ID: 1, Name: "手机", Price: 2999.99},
        {ID: 2, Name: "电脑", Price: 5999.99},
    }
    
    var products2 []Product
    err := go_deep_copy.DeepCopy(&products1, &products2)
    if err != nil {
        panic(err)
    }
    
    // 修改原始数据，验证深拷贝
    products1[0].Price = 1999.99
    fmt.Printf("products1[0].Price: %.2f\n", products1[0].Price) // 1999.99
    fmt.Printf("products2[0].Price: %.2f\n", products2[0].Price) // 2999.99
}
```

## ⚙️ 高级选项

### 字段名映射

```go
package main

import (
    "fmt"
    "github.com/LiZhiqiang0/go_deep_copy"
)

type Source struct {
    Name   string
    Secret string
    ID     int
}

type Target struct {
    Name     string
    Secret   string `go_deep_copy:"-"`        // 忽略此字段
    TargetID int    `go_deep_copy:"ID"`      // 映射到源结构的 ID 字段
}

func main() {
    source := Source{Name: "测试", Secret: "机密", ID: 1001}
    var target Target
    
    err := go_deep_copy.DeepCopy(&source, &target)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Target: %+v\n", target)
    // 输出: Target: {Name:测试 Secret: TargetID:1001}
}
```

## 🎯 性能优势

- **高性能反射**：使用 unsafe 包和反射优化，比标准反射更快
- **指针处理**：使用指针+偏移量寻址，避免了反射调用的开销
- **缓存机制**：转换函数和结构体反射信息缓存，避免重复反射操作

### 性能对比

以下是本库与copier库、json 序列化方式的性能对比：

```
BenchmarkCopyStruct/copier-10            244434     4784 ns/op    1272 B/op   60 allocs/op
BenchmarkCopyStruct/go_deep_copy-10      945814     1270 ns/op    240 B/op     8 allocs/op  
BenchmarkCopyStruct/json-10              597748     1958 ns/op    776 B/op    19 allocs/op
```

可以看到，本库在性能上相比copier库、json 序列化方式有显著提升，内存分配也更少。

## 📋 支持类型

### 基本类型
- 所有基本类型（int, float, string, bool 等）
- 基本类型的指针和切片
- 时间类型（time.Time）

### 复杂类型
- 结构体（支持嵌套）
- 切片和数组
- Map（支持嵌套）
- 接口类型
- 指针类型

