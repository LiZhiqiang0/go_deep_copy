# Go Deep Copy - High-Performance Deep Copy Library

Go Deep Copy is a high-performance Go deep copy library focused on deep copying of complex data structures such as structs, slices, and Maps. It uses reflection and code generation techniques to provide fast, safe, and flexible data copying functionality. Performance is better than github.com/jinzhu/copier and JSON serialization methods.

## 🌟 Core Features

### Deep Copy Support
- **Struct Deep Copy**: Complete copying of nested structs, ensuring data independence
- **Slice Deep Copy**: Copy slice underlying data to avoid sharing underlying arrays
- **Map Deep Copy**: Recursively copy all values in Map
- **Pointer Deep Copy**: Correctly handle pointer fields and create new memory space

### Smart Type Conversion
- **Automatic Type Conversion**: Support automatic conversion between basic types (int ↔ int64 ↔ float64 ↔ string)
- **Struct Interconversion**: Field mapping and conversion between different structs
- **Map ↔ Struct**: Bidirectional conversion, supporting complex nested structures
- **Slice Type Conversion**: Support slice conversion of different element types

### Advanced Features
- **Field Mapping**: Support mapping between fields with different names
- **High-Performance Optimization**: Use unsafe package and reflection optimization, faster than standard reflection
- **Concurrent Safety**: Support deep copy operations in concurrent environments

## 🚀 Quick Start

### Installation

```bash
go get -u github.com/LiZhiqiang0/go_deep_copy
```

### Basic Usage

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
    // Struct to struct copy
    user := User{Name: "John", Age: 30, Email: "john@example.com"}
    var employee Employee
    
    err := go_deep_copy.DeepCopy(&user, &employee)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Employee: %+v\n", employee)
    // Output: Employee: {Name:John Age:30 Email:john@example.com}
}
```

## 📖 Detailed Features

### 1. Struct Deep Copy

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
        Name: "Alice",
        Age:  25,
        Address: &Address{
            City:    "New York",
            Country: "USA",
        },
    }
    
    var person2 Person
    err := go_deep_copy.DeepCopy(person1, &person2)
    if err != nil {
        panic(err)
    }
    
    // Modify original data to verify deep copy
    person1.Address.City = "Los Angeles"
    fmt.Printf("person1.Address.City: %s\n", person1.Address.City) // Los Angeles
    fmt.Printf("person2.Address.City: %s\n", person2.Address.City) // New York
}
```

### 2. Map and Struct Interconversion

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
    // Map to struct
    mapData := map[string]interface{}{
        "Name": "Bob",
        "Age":  int64(35),
        "Email": "bob@example.com",
    }

    var user User
    err := go_deep_copy.DeepCopy(&mapData, &user)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Map to User: %+v\n", user)

    // Struct to Map
    var result map[string]interface{}
    err = go_deep_copy.DeepCopy(&user, &result)
    if err != nil {
        panic(err)
    }
    fmt.Printf("User to Map: %+v\n", result)
}
```

### 3. Slice Deep Copy

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
        {ID: 1, Name: "Phone", Price: 2999.99},
        {ID: 2, Name: "Computer", Price: 5999.99},
    }
    
    var products2 []Product
    err := go_deep_copy.DeepCopy(&products1, &products2)
    if err != nil {
        panic(err)
    }
    
    // Modify original data to verify deep copy
    products1[0].Price = 1999.99
    fmt.Printf("products1[0].Price: %.2f\n", products1[0].Price) // 1999.99
    fmt.Printf("products2[0].Price: %.2f\n", products2[0].Price) // 2999.99
}
```

## ⚙️ Advanced Options

### Field Name Mapping

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
    Secret   string `go_deep_copy:"-"`        // Ignore this field
    TargetID int    `go_deep_copy:"ID"`      // Map to source struct's ID field
}

func main() {
    source := Source{Name: "Test", Secret: "Confidential", ID: 1001}
    var target Target
    
    err := go_deep_copy.DeepCopy(&source, &target)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Target: %+v\n", target)
    // Output: Target: {Name:Test Secret: TargetID:1001}
}
```

## 🎯 Performance Advantages

- **High-Performance Reflection**: Uses unsafe package and reflection optimization, faster than standard reflection
- **Pointer Handling**: Uses pointer + offset addressing, avoiding reflection call overhead
- **Cache Mechanism**: Caches conversion functions and struct reflection information to avoid repeated reflection operations

### Performance Comparison

Here is the performance comparison between this library and copier library, JSON serialization:

```
BenchmarkCopyStruct/copier-10             850699             14030 ns/op            4248 B/op        178 allocs/op
BenchmarkCopyStruct/go_deep_copy-10      5380696              2182 ns/op             672 B/op         18 allocs/op
BenchmarkCopyStruct/json-10              2142386              5604 ns/op            1641 B/op         37 allocs/op
```

As you can see, this library has significant performance improvements compared to copier library and JSON serialization, with fewer memory allocations.

## 📋 Supported Types

### Basic Types
- All basic types (int, float, string, bool, etc.)
- Pointers and slices of basic types
- Time types (time.Time)

### Complex Types
- Structs (support nesting)
- Slices and arrays
- Maps (support nesting)
- Interface types
- Pointer types