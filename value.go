package jsonart

import (
	"fmt"
	"strconv"
)

// Value represents a JSON value that can hold any JSON type:
// object, array, string, number (int64 or float64), boolean, or null.
type Value struct {
	value interface{}
}

// null is the internal type used to represent a JSON null value.
type null int

// NULL is the sentinel value representing JSON null.
var NULL = null(0)

// NewValue creates a new Value with a nil underlying value.
func NewValue() *Value {
	return &Value{nil}
}

// AsObject sets the value to a JSON object. If value is nil, an empty object is created.
func (v *Value) AsObject(value map[string]*Value) {
	if value == nil {
		v.value = map[string]*Value{}
	} else {
		v.value = value
	}
}

// AsArray sets the value to a JSON array. If value is nil, an empty array is created.
func (v *Value) AsArray(value []*Value) {
	if value == nil {
		v.value = []*Value{}
	} else {
		v.value = value
	}
}

// AsInt sets the value to a JSON integer (int64).
func (v *Value) AsInt(value int64) {
	v.value = value
}

// AsFloat sets the value to a JSON floating-point number (float64).
func (v *Value) AsFloat(value float64) {
	v.value = value
}

// AsBool sets the value to a JSON boolean.
func (v *Value) AsBool(ok bool) {
	v.value = ok
}

// AsNull sets the value to JSON null.
func (v *Value) AsNull() {
	v.value = NULL
}

// AsString sets the value to a JSON string.
func (v *Value) AsString(value string) {
	v.value = value
}

// AddField adds a new field with the given key to the object and returns
// the newly created Value. Panics if the value is not an object.
func (v *Value) AddField(key string) *Value {
	if values, ok := v.value.(map[string]*Value); ok {
		value := NewValue()
		values[key] = value
		return value
	}
	panic(fmt.Sprintf("not an object value, got %T", v.value))
}

// AddElement appends a new element to the array and returns the newly
// created Value. Panics if the value is not an array.
func (v *Value) AddElement() *Value {
	if values, ok := v.value.([]*Value); ok {
		value := NewValue()
		v.value = append(values, value)
		return value
	}
	panic(fmt.Sprintf("not an array value, got %T", v.value))
}

// IsObject reports whether the value is a JSON object.
func (v *Value) IsObject() bool {
	switch v.value.(type) {
	case map[string]*Value:
		return true
	default:
		return false
	}
}

// IsArray reports whether the value is a JSON array.
func (v *Value) IsArray() bool {
	switch v.value.(type) {
	case []*Value:
		return true
	default:
		return false
	}
}

// IsInt reports whether the value is a JSON integer (int64).
func (v *Value) IsInt() bool {
	switch v.value.(type) {
	case int64:
		return true
	default:
		return false
	}
}

// IsNumber reports whether the value is a JSON number (int64 or float64).
func (v *Value) IsNumber() bool {
	switch v.value.(type) {
	case int64, float64:
		return true
	default:
		return false
	}
}

// IsBool reports whether the value is a JSON boolean.
func (v *Value) IsBool() bool {
	switch v.value.(type) {
	case bool:
		return true
	default:
		return false
	}
}

// IsTrue reports whether the value is a JSON boolean with value true.
func (v *Value) IsTrue() bool {
	switch v.value.(type) {
	case bool:
		return v.value == true
	default:
		return false
	}
}

// IsFalse reports whether the value is a JSON boolean with value false.
func (v *Value) IsFalse() bool {
	switch v.value.(type) {
	case bool:
		return v.value == false
	default:
		return false
	}
}

// IsNull reports whether the value is JSON null.
func (v *Value) IsNull() bool {
	switch v.value.(type) {
	case null:
		return true
	default:
		return false
	}
}

// IsString reports whether the value is a JSON string.
func (v *Value) IsString() bool {
	switch v.value.(type) {
	case string:
		return true
	default:
		return false
	}
}

// Get traverses the value tree following the given path of keys (for objects)
// or string indices (for arrays). Returns nil if any key is missing or the
// path cannot be followed.
func (v *Value) Get(path ...string) *Value {
	value := v
	index := 0
	var obj map[string]*Value
	var arr []*Value
	var err error
	var ok bool
	for _, key := range path {
		if obj, ok = value.value.(map[string]*Value); ok {
			if value, ok = obj[key]; !ok {
				return nil
			}
		} else if arr, ok = value.value.([]*Value); ok {
			index, err = strconv.Atoi(key)
			if err != nil {
				return nil
			}
			if index < 0 || len(arr) < index+1 {
				return nil
			}
			value = arr[index]
		} else {
			return nil
		}
	}
	return value
}

// Ensure traverses the value tree following the given path, creating
// intermediate objects as needed. Returns the Value at the end of the path.
// Panics if a non-object value is encountered that cannot be converted.
func (v *Value) Ensure(path ...string) *Value {
	temp := v
	var ok bool
	var obj map[string]*Value
	for _, field := range path {
		if !temp.IsObject() {
			temp.AsObject(nil)
		}
		if obj, ok = temp.value.(map[string]*Value); ok {
			if temp, ok = obj[field]; !ok {
				temp = NewValue()
				obj[field] = temp
			}
		} else {
			panic("unexpected non-object in Ensure path")
		}
	}
	return temp
}

// Object returns the underlying map for an object value.
// Panics if the value is not an object.
func (v *Value) Object() map[string]*Value {
	if value, ok := v.value.(map[string]*Value); ok {
		return value
	}
	panic(fmt.Sprintf("not an object value, got %T", v.value))
}

// Array returns the underlying slice for an array value.
// Panics if the value is not an array.
func (v *Value) Array() []*Value {
	if value, ok := v.value.([]*Value); ok {
		return value
	}
	panic(fmt.Sprintf("not an array value, got %T", v.value))
}

// Int returns the underlying int64 for an integer value.
// Panics if the value is not an integer.
func (v *Value) Int() int64 {
	if value, ok := v.value.(int64); ok {
		return value
	}
	panic(fmt.Sprintf("not an int value, got %T", v.value))
}

// Float returns the underlying float64 for a number value.
// If the value is an int64, it is converted to float64.
// Panics if the value is not a number.
func (v *Value) Float() float64 {
	if value, ok := v.value.(int64); ok {
		return float64(value)
	}
	if value, ok := v.value.(float64); ok {
		return float64(value)
	}
	panic(fmt.Sprintf("not a number value, got %T", v.value))
}

// String returns the underlying string for a string value.
// Panics if the value is not a string.
func (v *Value) String() string {
	if value, ok := v.value.(string); ok {
		return value
	}
	panic(fmt.Sprintf("not a string value, got %T", v.value))
}

// Value returns the underlying Go value as an interface{}.
// Objects are returned as map[string]interface{}, arrays as []interface{},
// null as nil, and primitives as their Go types.
func (v *Value) Value() interface{} {
	if v == nil {
		return nil
	}
	switch v.value.(type) {
	case null, nil:
		return nil
	case string, bool, int64, float64:
		return v.value
	case map[string]*Value:
		if values, ok := v.value.(map[string]*Value); ok {
			out := map[string]interface{}{}
			for key, value := range values {
				out[key] = value.Value()
			}
			return out
		}
		return nil
	case []*Value:
		if values, ok := v.value.([]*Value); ok {
			out := []interface{}{}
			for _, value := range values {
				out = append(out, value.Value())
			}
			return out
		}
		return nil
	default:
		return nil
	}
}
