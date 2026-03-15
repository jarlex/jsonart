package jsonart

import (
	"reflect"
	"testing"
)

func assertPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic, got none", name)
		}
	}()
	f()
}

func TestValueTypes(t *testing.T) {
	v := NewValue()

	v.AsInt(42)
	if !v.IsInt() || !v.IsNumber() || v.IsString() {
		t.Errorf("Type mismatch for int")
	}
	if v.Int() != 42 {
		t.Errorf("Expected 42, got %v", v.Int())
	}
	assertPanic(t, "Int() as string", func() { v.String() })

	v.AsFloat(3.14)
	if v.IsInt() || !v.IsNumber() {
		t.Errorf("Type mismatch for float")
	}
	if v.Float() != 3.14 {
		t.Errorf("Expected 3.14, got %v", v.Float())
	}
	assertPanic(t, "Float() as int", func() { v.Int() })

	v.AsString("hello")
	if !v.IsString() {
		t.Errorf("Type mismatch for string")
	}
	if v.String() != "hello" {
		t.Errorf("Expected 'hello', got %v", v.String())
	}

	v.AsBool(true)
	if !v.IsBool() || !v.IsTrue() || v.IsFalse() {
		t.Errorf("Type mismatch for true")
	}

	v.AsBool(false)
	if !v.IsBool() || v.IsTrue() || !v.IsFalse() {
		t.Errorf("Type mismatch for false")
	}

	v.AsNull()
	if !v.IsNull() {
		t.Errorf("Type mismatch for null")
	}

	v.AsObject(map[string]*Value{"key": NewValue()})
	if !v.IsObject() {
		t.Errorf("Type mismatch for object")
	}
	if len(v.Object()) != 1 {
		t.Errorf("Expected object length 1")
	}
	assertPanic(t, "Object() as Array", func() { v.Array() })

	v.AsArray([]*Value{NewValue()})
	if !v.IsArray() {
		t.Errorf("Type mismatch for array")
	}
	if len(v.Array()) != 1 {
		t.Errorf("Expected array length 1")
	}
	assertPanic(t, "Array() as Object", func() { v.Object() })
}

func TestValueGetters(t *testing.T) {
	v := NewValue()
	v.AsObject(make(map[string]*Value))
	field := v.AddField("a")
	field.AsObject(make(map[string]*Value))
	sub := field.AddField("b")
	sub.AsInt(100)

	if got := v.Get("a", "b"); got == nil || got.Int() != 100 {
		t.Errorf("Get failed to retrieve deep value")
	}
	if got := v.Get("x"); got != nil {
		t.Errorf("Get expected nil for missing key, got %v", got)
	}
	if got := v.Get("a", "x"); got != nil {
		t.Errorf("Get expected nil for missing subkey, got %v", got)
	}

	ensured := v.Ensure("x", "y", "z")
	ensured.AsInt(200)

	if got := v.Get("x", "y", "z"); got == nil || got.Int() != 200 {
		t.Errorf("Ensure failed to build path")
	}
}

func TestValueAdd(t *testing.T) {
	v := NewValue()
	v.AsObject(make(map[string]*Value))
	v.AddField("x").AsInt(1)

	if len(v.Object()) != 1 {
		t.Errorf("Expected 1 field")
	}

	arr := NewValue()
	arr.AsArray(make([]*Value, 0))
	arr.AddElement().AsString("first")

	if len(arr.Array()) != 1 {
		t.Errorf("Expected 1 element")
	}
}

func TestValueInterface(t *testing.T) {
	v := NewValue()
	v.AsInt(42)
	if val := v.Value(); !reflect.DeepEqual(val, int64(42)) {
		t.Errorf("Value() for int: expected 42, got %v", val)
	}
	v.AsString("test")
	if val := v.Value(); !reflect.DeepEqual(val, "test") {
		t.Errorf("Value() for string: expected 'test', got %v", val)
	}
	v.AsNull()
	if val := v.Value(); val != nil {
		t.Errorf("Value() for null: expected nil, got %v", val)
	}
}

func TestSafeMethodsSuccess(t *testing.T) {
	t.Run("StringSafe on string", func(t *testing.T) {
		v := NewValue()
		v.AsString("hello")
		got, err := v.StringSafe()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello" {
			t.Errorf("expected \"hello\", got %q", got)
		}
	})

	t.Run("IntSafe on int", func(t *testing.T) {
		v := NewValue()
		v.AsInt(42)
		got, err := v.IntSafe()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 42 {
			t.Errorf("expected 42, got %d", got)
		}
	})

	t.Run("FloatSafe on float", func(t *testing.T) {
		v := NewValue()
		v.AsFloat(3.14)
		got, err := v.FloatSafe()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 3.14 {
			t.Errorf("expected 3.14, got %f", got)
		}
	})

	t.Run("FloatSafe on int (coercion)", func(t *testing.T) {
		v := NewValue()
		v.AsInt(42)
		got, err := v.FloatSafe()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 42.0 {
			t.Errorf("expected 42.0, got %f", got)
		}
	})

	t.Run("ObjectSafe on object", func(t *testing.T) {
		v := NewValue()
		v.AsObject(map[string]*Value{"a": NewValue()})
		got, err := v.ObjectSafe()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Errorf("expected map length 1, got %d", len(got))
		}
		if _, ok := got["a"]; !ok {
			t.Errorf("expected key \"a\" in map")
		}
	})

	t.Run("ArraySafe on array", func(t *testing.T) {
		v := NewValue()
		v.AsArray([]*Value{NewValue()})
		got, err := v.ArraySafe()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Errorf("expected slice length 1, got %d", len(got))
		}
	})

	t.Run("BoolSafe on true", func(t *testing.T) {
		v := NewValue()
		v.AsBool(true)
		got, err := v.BoolSafe()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != true {
			t.Errorf("expected true, got %v", got)
		}
	})

	t.Run("BoolSafe on false", func(t *testing.T) {
		v := NewValue()
		v.AsBool(false)
		got, err := v.BoolSafe()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != false {
			t.Errorf("expected false, got %v", got)
		}
	})
}

func TestSafeMethodsError(t *testing.T) {
	t.Run("StringSafe on int", func(t *testing.T) {
		v := NewValue()
		v.AsInt(42)
		got, err := v.StringSafe()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got != "" {
			t.Errorf("expected zero value \"\", got %q", got)
		}
		if msg := err.Error(); msg != "not a string value, got int64" {
			t.Errorf("unexpected error message: %s", msg)
		}
	})

	t.Run("IntSafe on float", func(t *testing.T) {
		v := NewValue()
		v.AsFloat(3.14)
		got, err := v.IntSafe()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got != 0 {
			t.Errorf("expected zero value 0, got %d", got)
		}
		if msg := err.Error(); msg != "not an int value, got float64" {
			t.Errorf("unexpected error message: %s", msg)
		}
	})

	t.Run("FloatSafe on string", func(t *testing.T) {
		v := NewValue()
		v.AsString("hello")
		got, err := v.FloatSafe()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got != 0 {
			t.Errorf("expected zero value 0, got %f", got)
		}
		if msg := err.Error(); msg != "not a number value, got string" {
			t.Errorf("unexpected error message: %s", msg)
		}
	})

	t.Run("ObjectSafe on array", func(t *testing.T) {
		v := NewValue()
		v.AsArray([]*Value{NewValue()})
		got, err := v.ObjectSafe()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got != nil {
			t.Errorf("expected nil map, got %v", got)
		}
	})

	t.Run("ArraySafe on object", func(t *testing.T) {
		v := NewValue()
		v.AsObject(map[string]*Value{"a": NewValue()})
		got, err := v.ArraySafe()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got != nil {
			t.Errorf("expected nil slice, got %v", got)
		}
	})

	t.Run("BoolSafe on string", func(t *testing.T) {
		v := NewValue()
		v.AsString("true")
		got, err := v.BoolSafe()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got != false {
			t.Errorf("expected zero value false, got %v", got)
		}
		if msg := err.Error(); msg != "not a bool value, got string" {
			t.Errorf("unexpected error message: %s", msg)
		}
	})
}

func TestBoolPanic(t *testing.T) {
	v := NewValue()
	v.AsString("x")
	assertPanic(t, "Bool() as string", func() { v.Bool() })
}
