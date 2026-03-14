# jsonart

Lightweight JSON parsing and navigation library for Go.

[![Go Version](https://img.shields.io/badge/Go-1.12%2B-blue)](https://golang.org/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

jsonart provides a simple and intuitive API for parsing JSON data and navigating through it using path-based access. It's designed to be lightweight with no external dependencies.

## Installation

```bash
go get github.com/jarlex/jsonart
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/jarlex/jsonart"
)

func main() {
    data := []byte(`{"name": "John", "age": 30}`)
    value, err := jsonart.Unmarshal(data)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Name:", value.Get("name").String())
    fmt.Println("Age:", value.Get("age").Int())
}
```

## Basic Usage

Parse any JSON data into a `*Value` and access its contents:

```go
// Parse JSON
data := []byte(`{
    "name": "Alice",
    "age": 28,
    "active": true,
    "scores": [95, 87, 92]
}`)
value, _ := jsonart.Unmarshal(data)

// Access simple values
name := value.Get("name").String()   // "Alice"
age := value.Get("age").Int()        // 28
active := value.Get("active").Bool() // true
```

## Navigation

### Get - Safe Path Navigation

The `Get()` method provides safe navigation through JSON structures. It returns `nil` if the path doesn't exist instead of panicking.

```go
data := []byte(`{
    "user": {
        "profile": {
            "name": "Bob"
        }
    }
}`)
value, _ := jsonart.Unmarshal(data)

// Navigate nested objects
name := value.Get("user", "profile", "name").String() // "Bob"

// Access array elements using string indices
score := value.Get("user", "profile", "scores", "0").Int() // 95

// Safe access - returns nil if path doesn't exist
missing := value.Get("user", "address", "city") // nil (no panic)
if missing != nil {
    fmt.Println(missing.String())
}
```

### Ensure - Create Paths Dynamically

The `Ensure()` method creates the path if it doesn't exist, allowing you to build JSON structures programmatically:

```go
// Create a new Value and build JSON dynamically
root := jsonart.NewValue()
root.AsObject(nil)

// Ensure creates intermediate objects if they don't exist
user := root.Ensure("user")
profile := user.Ensure("profile")
profile.Ensure("name").AsString("Charlie")

// Convert to native Go types
result := root.Value()
// result: map[string]interface{}{
//     "user": map[string]interface{}{
//         "profile": map[string]interface{}{
//             "name": "Charlie"
//         }
//     }
// }
```

### Difference Between Get and Ensure

| Method | Behavior |
|--------|----------|
| `Get(path...)` | Returns `nil` if path doesn't exist |
| `Ensure(path...)` | Creates the path if it doesn't exist |

```go
value, _ := jsonart.Unmarshal([]byte(`{}`))

// Get returns nil - path doesn't exist
result := value.Get("missing") // nil

// Ensure creates the path
created := value.Ensure("created") // Returns a new *Value
created.AsString("now it exists")
```

## Type Conversion

### Value() - Recursive Conversion to Native Go Types

The `Value()` method converts the JSON tree to native Go types recursively:

```go
data := []byte(`{
    "user": {
        "name": "Diana",
        "scores": [10, 20, 30]
    },
    "active": true
}`)
value, _ := jsonart.Unmarshal(data)

// Convert entire tree to Go types
result := value.Value()
// result: map[string]interface{}{
//     "user": map[string]interface{}{
//         "name": "Diana",
//         "scores": []interface{}{10, 20, 30}
//     },
//     "active": true
// }
```

### Type Extraction Methods

These methods extract values and **panic** if the type doesn't match:

```go
value, _ := jsonart.Unmarshal([]byte(`{
    "string": "hello",
    "int": 42,
    "float": 3.14,
    "bool": true,
    "array": [1, 2],
    "object": {"a": 1}
}`))

value.Get("string").String()  // "hello"
value.Get("int").Int()        // int64(42)
value.Get("float").Float()    // float64(3.14)
value.Get("bool").Bool()      // true
value.Get("array").Array()    // []*Value
value.Get("object").Object() // map[string]*Value
```

### Type Checking Methods

Check the JSON type before extracting to avoid panics:

```go
value, _ := jsonart.Unmarshal(data)

// Check type first
if value.Get("name").IsString() {
    name := value.Get("name").String()
}

if value.Get("count").IsInt() || value.Get("count").IsNumber() {
    count := value.Get("count").Int()
}

// Available type checks
value.IsObject()  // true if JSON object {}
value.IsArray()   // true if JSON array []
value.IsInt()     // true if integer
value.IsNumber()  // true if integer or float
value.IsBool()    // true if boolean
value.IsTrue()    // true if boolean true
value.IsFalse()   // true if boolean false
value.IsNull()    // true if null
value.IsString()  // true if string
```

### Handling Null Values

```go
value, _ := jsonart.Unmarshal([]byte(`{"optional": null}`))

// Check for null
if value.Get("optional").IsNull() {
    fmt.Println("Value is null")
}

// Null becomes nil when converted to Go types
result := value.Value() // map[string]interface{}{"optional": nil}
```

## API Reference

### Functions

#### Unmarshal

```go
func Unmarshal(data []byte) (value *Value, err error)
```

Parses JSON-encoded data and returns a `*Value`. Returns an error if the JSON is invalid.

```go
value, err := jsonart.Unmarshal([]byte(`{"key": "value"}`))
```

### Constructors

#### NewValue

```go
func NewValue() *Value
```

Creates a new empty `*Value`.

```go
root := jsonart.NewValue()
```

#### NULL

```go
var NULL
```

Constant representing a JSON null value.

```go
value, _ := jsonart.Unmarshal([]byte(`{"field": null}`))
value.Get("field").IsNull() // true
```

### Navigation Methods

| Method | Description |
|--------|-------------|
| `Get(path ...string) *Value` | Safe navigation, returns nil if path doesn't exist |
| `Ensure(path ...string) *Value` | Creates path if it doesn't exist |

### Type Checking Methods

| Method | Returns |
|--------|---------|
| `IsObject() bool` | true if value is a JSON object |
| `IsArray() bool` | true if value is a JSON array |
| `IsInt() bool` | true if value is an integer |
| `IsNumber() bool` | true if value is an integer or float |
| `IsBool() bool` | true if value is a boolean |
| `IsTrue() bool` | true if value is boolean `true` |
| `IsFalse() bool` | true if value is boolean `false` |
| `IsNull() bool` | true if value is null |
| `IsString() bool` | true if value is a string |

### Value Extraction Methods

**Note**: These methods panic if the type doesn't match.

| Method | Returns |
|--------|---------|
| `Object() map[string]*Value` | JSON object (panics if not object) |
| `Array() []*Value` | JSON array (panics if not array) |
| `Int() int64` | Integer value (panics if not int) |
| `Float() float64` | Float value (panics if not number) |
| `String() string` | String value (panics if not string) |
| `Value() interface{}` | Converts to native Go types recursively |

### Type Setters

| Method | Description |
|--------|-------------|
| `AsObject(value map[string]*Value)` | Sets value as a JSON object |
| `AsArray(value []*Value)` | Sets value as a JSON array |
| `AsInt(value int64)` | Sets value as an integer |
| `AsFloat(value float64)` | Sets value as a float |
| `AsBool(ok bool)` | Sets value as a boolean |
| `AsNull()` | Sets value as null |
| `AsString(value string)` | Sets value as a string |

### Builders

| Method | Description |
|--------|-------------|
| `AddField(key string) *Value` | Adds a field to an object and returns the new value |
| `AddElement() *Value` | Adds an element to an array and returns the new value |

## Practical Examples

### Extract Data from API Response

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/jarlex/jsonart"
)

func main() {
    // Simulated API response
    apiResponse := []byte(`{
        "status": "success",
        "data": {
            "users": [
                {"id": 1, "name": "Alice", "email": "alice@example.com"},
                {"id": 2, "name": "Bob", "email": "bob@example.com"}
            ],
            "pagination": {
                "page": 1,
                "total": 100
            }
        }
    }`)
    
    value, err := jsonart.Unmarshal(apiResponse)
    if err != nil {
        panic(err)
    }
    
    // Check status
    if value.Get("status").String() != "success" {
        panic("API request failed")
    }
    
    // Get users array
    users := value.Get("data", "users").Array()
    fmt.Printf("Found %d users\n", len(users))
    
    // Iterate through users
    for _, user := range users {
        fmt.Printf("User: %s <%s>\n", 
            user.Get("name").String(),
            user.Get("email").String())
    }
    
    // Get pagination info
    page := value.Get("data", "pagination", "page").Int()
    total := value.Get("data", "pagination", "total").Int()
    fmt.Printf("Page %d of %d total records\n", page, total)
}
```

### Iterate Over Array of Objects

```go
data := []byte(`{
    "products": [
        {"id": 1, "name": "Widget", "price": 29.99},
        {"id": 2, "name": "Gadget", "price": 49.99},
        {"id": 3, "name": "Gizmo", "price": 19.99}
    ]
}`)
value, _ := jsonart.Unmarshal(data)

// Get the array
products := value.Get("products").Array()

// Iterate and process
for i, product := range products {
    id := product.Get("id").Int()
    name := product.Get("name").String()
    price := product.Get("price").Float()
    
    fmt.Printf("%d. %s - $%.2f\n", id, name, price)
}
```

### Build JSON Programmatically

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/jarlex/jsonart"
)

func main() {
    // Create a new JSON document from scratch
    root := jsonart.NewValue()
    root.AsObject(nil)
    
    // Add user data
    user := root.Ensure("user")
    user.Ensure("name").AsString("Eve")
    user.Ensure("age").AsInt(25)
    
    // Add preferences as an array
    prefs := user.Ensure("preferences")
    prefs.AsArray(nil)
    prefs.AddElement().AsString("dark_mode")
    prefs.AddElement().AsString("notifications")
    
    // Add metadata
    root.Ensure("created_at").AsString("2024-01-15")
    
    // Convert to native Go types and marshal to JSON string
    native := root.Value()
    jsonBytes, _ := json.Marshal(native)
    fmt.Println(string(jsonBytes))
}
```

Output:
```json
{"created_at":"2024-01-15","user":{"age":25,"name":"Eve","preferences":["dark_mode","notifications"]}}
```

### Safe Type Checking and Extraction

```go
func getStringValue(v *jsonart.Value, path ...string) (string, bool) {
    val := v.Get(path...)
    if val == nil || !val.IsString() {
        return "", false
    }
    return val.String(), true
}

func getIntValue(v *jsonart.Value, path ...string) (int64, bool) {
    val := v.Get(path...)
    if val == nil || !val.IsInt() {
        return 0, false
    }
    return val.Int(), true
}

func main() {
    data := []byte(`{
        "name": "Frank",
        "age": "thirty",  // This is a string, not an int!
        "count": 42
    }`)
    value, _ := jsonart.Unmarshal(data)
    
    // Safe extraction with type checking
    if name, ok := getStringValue(value, "name"); ok {
        fmt.Println("Name:", name) // "Frank"
    }
    
    // This won't panic - age is a string, not an int
    if count, ok := getIntValue(value, "age"); !ok {
        fmt.Println("age is not an integer") // This will print
    }
    
    // This works - count is actually an int
    if count, ok := getIntValue(value, "count"); ok {
        fmt.Println("Count:", count) // 42
    }
}
```

## License

MIT License - see LICENSE file for details.
