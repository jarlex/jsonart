package jsonart

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestMarshalStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", `""`},
		{"simple string", "hello", `"hello"`},
		{"with spaces", "hello world", `"hello world"`},
		{"double quote", `hello "world"`, `"hello \"world\""`},
		{"backslash", `back\slash`, `"back\\slash"`},
		{"newline", "hello\nworld", `"hello\nworld"`},
		{"carriage return", "hello\rworld", `"hello\rworld"`},
		{"tab", "hello\tworld", `"hello\tworld"`},
		{"backspace", "hello\x08world", `"hello\bworld"`},
		{"form feed", "hello\x0Cworld", `"hello\fworld"`},
		{"null byte", "hello\x00world", `"hello\u0000world"`},
		{"control char 0x01", "a\x01b", `"a\u0001b"`},
		{"control char 0x1f", "a\x1fb", `"a\u001fb"`},
		{"forward slash not escaped", "a/b", `"a/b"`},
		{"unicode euro sign", "€", `"€"`},
		{"unicode emoji", "😀", `"😀"`},
		{"mixed escapes", "a\"b\\c\nd\te", `"a\"b\\c\nd\te"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue()
			v.AsString(tt.input)
			out, err := Marshal(v)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if string(out) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(out))
			}
		})
	}
}

func TestMarshalStringControlCharsExhaustive(t *testing.T) {
	// Test all control characters 0x00-0x1F
	named := map[byte]string{
		0x08: `\b`,
		0x09: `\t`,
		0x0A: `\n`,
		0x0C: `\f`,
		0x0D: `\r`,
	}
	for c := byte(0); c < 0x20; c++ {
		v := NewValue()
		v.AsString(string([]byte{c}))
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error for control char 0x%02x: %v", c, err)
			continue
		}
		got := string(out)
		if esc, ok := named[c]; ok {
			expected := `"` + esc + `"`
			if got != expected {
				t.Errorf("Control char 0x%02x: expected %s, got %s", c, expected, got)
			}
		} else {
			// Should be \u00XX
			if !strings.Contains(got, `\u00`) {
				t.Errorf("Control char 0x%02x: expected \\u00XX escape, got %s", c, got)
			}
		}
	}
}

func TestMarshalNumbers(t *testing.T) {
	t.Run("integers", func(t *testing.T) {
		tests := []struct {
			name     string
			input    int64
			expected string
		}{
			{"zero", 0, "0"},
			{"positive", 42, "42"},
			{"negative", -1, "-1"},
			{"large positive", 9223372036854775807, "9223372036854775807"},
			{"large negative", -9223372036854775808, "-9223372036854775808"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				v := NewValue()
				v.AsInt(tt.input)
				out, err := Marshal(v)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if string(out) != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, string(out))
				}
			})
		}
	})

	t.Run("floats", func(t *testing.T) {
		tests := []struct {
			name     string
			input    float64
			expected string
		}{
			{"pi", 3.14, "3.14"},
			{"negative pi", -3.14, "-3.14"},
			{"zero point zero", 0.0, "0"},
			{"small decimal", 0.001, "0.001"},
			{"one", 1.0, "1"},
			{"negative one", -1.0, "-1"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				v := NewValue()
				v.AsFloat(tt.input)
				out, err := Marshal(v)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if string(out) != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, string(out))
				}
			})
		}
	})
}

func TestMarshalLiterals(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Value)
		expected string
	}{
		{"true", func(v *Value) { v.AsBool(true) }, "true"},
		{"false", func(v *Value) { v.AsBool(false) }, "false"},
		{"null", func(v *Value) { v.AsNull() }, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue()
			tt.setup(v)
			out, err := Marshal(v)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if string(out) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(out))
			}
		})
	}
}

func TestMarshalObjects(t *testing.T) {
	t.Run("empty object", func(t *testing.T) {
		v := NewValue()
		v.AsObject(nil)
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if string(out) != "{}" {
			t.Errorf("Expected {}, got %s", string(out))
		}
	})

	t.Run("single key", func(t *testing.T) {
		v := NewValue()
		v.AsObject(nil)
		v.AddField("key").AsString("value")
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if string(out) != `{"key":"value"}` {
			t.Errorf("Expected {\"key\":\"value\"}, got %s", string(out))
		}
	})

	t.Run("multiple keys structural", func(t *testing.T) {
		v := NewValue()
		v.AsObject(nil)
		v.AddField("a").AsInt(1)
		v.AddField("b").AsBool(true)
		v.AddField("c").AsNull()
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		// Re-parse and verify structurally (map order non-deterministic)
		reparsed, err := Unmarshal(out)
		if err != nil {
			t.Errorf("Failed to re-parse marshal output: %v", err)
			return
		}
		if !reparsed.IsObject() {
			t.Errorf("Expected object after re-parse")
			return
		}
		if reparsed.Get("a").Int() != 1 {
			t.Errorf("Expected a=1, got %v", reparsed.Get("a").Value())
		}
		if !reparsed.Get("b").IsTrue() {
			t.Errorf("Expected b=true")
		}
		if !reparsed.Get("c").IsNull() {
			t.Errorf("Expected c=null")
		}
	})

	t.Run("nested objects", func(t *testing.T) {
		v := NewValue()
		v.AsObject(nil)
		inner := v.AddField("outer")
		inner.AsObject(nil)
		inner.AddField("inner").AsInt(42)
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if string(out) != `{"outer":{"inner":42}}` {
			t.Errorf("Expected {\"outer\":{\"inner\":42}}, got %s", string(out))
		}
	})

	t.Run("key with special characters", func(t *testing.T) {
		v := NewValue()
		v.AsObject(nil)
		v.AddField("hello \"world\"").AsInt(1)
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if string(out) != `{"hello \"world\"":1}` {
			t.Errorf("Expected escaped key, got %s", string(out))
		}
	})
}

func TestMarshalArrays(t *testing.T) {
	t.Run("empty array", func(t *testing.T) {
		v := NewValue()
		v.AsArray(nil)
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if string(out) != "[]" {
			t.Errorf("Expected [], got %s", string(out))
		}
	})

	t.Run("single element", func(t *testing.T) {
		v := NewValue()
		v.AsArray(nil)
		v.AddElement().AsInt(1)
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if string(out) != "[1]" {
			t.Errorf("Expected [1], got %s", string(out))
		}
	})

	t.Run("multiple elements mixed types", func(t *testing.T) {
		v := NewValue()
		v.AsArray(nil)
		v.AddElement().AsInt(1)
		v.AddElement().AsString("two")
		v.AddElement().AsBool(true)
		v.AddElement().AsNull()
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if string(out) != `[1,"two",true,null]` {
			t.Errorf("Expected [1,\"two\",true,null], got %s", string(out))
		}
	})

	t.Run("nested arrays", func(t *testing.T) {
		v := NewValue()
		v.AsArray(nil)
		inner1 := v.AddElement()
		inner1.AsArray(nil)
		inner1.AddElement().AsInt(1)
		inner1.AddElement().AsInt(2)
		inner2 := v.AddElement()
		inner2.AsArray(nil)
		inner2.AddElement().AsInt(3)
		inner2.AddElement().AsInt(4)
		out, err := Marshal(v)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
			return
		}
		if string(out) != "[[1,2],[3,4]]" {
			t.Errorf("Expected [[1,2],[3,4]], got %s", string(out))
		}
	})
}

func TestMarshalErrors(t *testing.T) {
	t.Run("nil Value", func(t *testing.T) {
		_, err := Marshal(nil)
		if err == nil {
			t.Errorf("Expected error for nil *Value")
			return
		}
		if !strings.Contains(err.Error(), "nil") {
			t.Errorf("Expected error mentioning 'nil', got: %v", err)
		}
	})

	t.Run("uninitialized Value", func(t *testing.T) {
		v := &Value{}
		_, err := Marshal(v)
		if err == nil {
			t.Errorf("Expected error for uninitialized Value")
			return
		}
		if !strings.Contains(err.Error(), "uninitialized") {
			t.Errorf("Expected error mentioning 'uninitialized', got: %v", err)
		}
	})

	t.Run("NaN float", func(t *testing.T) {
		v := NewValue()
		v.AsFloat(math.NaN())
		_, err := Marshal(v)
		if err == nil {
			t.Errorf("Expected error for NaN")
			return
		}
		if !strings.Contains(err.Error(), "NaN") {
			t.Errorf("Expected error mentioning 'NaN', got: %v", err)
		}
	})

	t.Run("positive Inf", func(t *testing.T) {
		v := NewValue()
		v.AsFloat(math.Inf(1))
		_, err := Marshal(v)
		if err == nil {
			t.Errorf("Expected error for +Inf")
			return
		}
		if !strings.Contains(err.Error(), "infinite") {
			t.Errorf("Expected error mentioning 'infinite', got: %v", err)
		}
	})

	t.Run("negative Inf", func(t *testing.T) {
		v := NewValue()
		v.AsFloat(math.Inf(-1))
		_, err := Marshal(v)
		if err == nil {
			t.Errorf("Expected error for -Inf")
			return
		}
		if !strings.Contains(err.Error(), "infinite") {
			t.Errorf("Expected error mentioning 'infinite', got: %v", err)
		}
	})

	t.Run("nil in array", func(t *testing.T) {
		v := NewValue()
		v.value = []*Value{nil}
		_, err := Marshal(v)
		if err == nil {
			t.Errorf("Expected error for nil in array")
			return
		}
		if !strings.Contains(err.Error(), "array") {
			t.Errorf("Expected error mentioning 'array', got: %v", err)
		}
	})

	t.Run("nil in object", func(t *testing.T) {
		v := NewValue()
		v.value = map[string]*Value{"key": nil}
		_, err := Marshal(v)
		if err == nil {
			t.Errorf("Expected error for nil in object")
			return
		}
		if !strings.Contains(err.Error(), "object") {
			t.Errorf("Expected error mentioning 'object', got: %v", err)
		}
	})

	t.Run("uninitialized nested in array", func(t *testing.T) {
		v := NewValue()
		v.AsArray(nil)
		v.AddElement() // left uninitialized
		_, err := Marshal(v)
		if err == nil {
			t.Errorf("Expected error for uninitialized value in array")
			return
		}
		if !strings.Contains(err.Error(), "uninitialized") {
			t.Errorf("Expected error mentioning 'uninitialized', got: %v", err)
		}
	})

	t.Run("uninitialized nested in object", func(t *testing.T) {
		v := NewValue()
		v.AsObject(nil)
		v.AddField("key") // left uninitialized
		_, err := Marshal(v)
		if err == nil {
			t.Errorf("Expected error for uninitialized value in object")
			return
		}
		if !strings.Contains(err.Error(), "uninitialized") {
			t.Errorf("Expected error mentioning 'uninitialized', got: %v", err)
		}
	})
}

func TestMarshalRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple string", `"hello world"`},
		{"string with escapes", `"hello \"world\" \n \t"`},
		{"integer", `42`},
		{"negative integer", `-100`},
		{"float", `3.14`},
		{"true", `true`},
		{"false", `false`},
		{"null", `null`},
		{"empty object", `{}`},
		{"empty array", `[]`},
		{"simple array", `[1,2,3]`},
		{"mixed array", `[1,"two",true,null]`},
		{"nested arrays", `[[1,2],[3,4]]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse original
			v1, err := Unmarshal([]byte(tt.input))
			if err != nil {
				t.Fatalf("Failed to unmarshal input %q: %v", tt.input, err)
			}
			// Marshal
			out, err := Marshal(v1)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}
			// Re-parse
			v2, err := Unmarshal(out)
			if err != nil {
				t.Fatalf("Failed to unmarshal output %q: %v", string(out), err)
			}
			// Compare structurally
			if !reflect.DeepEqual(v1.Value(), v2.Value()) {
				t.Errorf("Round-trip mismatch: input=%q, output=%q", tt.input, string(out))
			}
		})
	}
}

func TestMarshalRoundTripObjects(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single key object", `{"key":"value"}`},
		{"multi key object", `{"a":1,"b":true,"c":null}`},
		{"nested object", `{"a":{"b":2}}`},
		{"complex structure", `{"users":[{"name":"Alice","age":30},{"name":"Bob","age":25}],"count":2}`},
		{"all types", `{"s":"text","i":42,"f":3.14,"b":true,"n":null,"a":[1,2],"o":{}}`},
		{"empty containers", `{"a":[],"b":{}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse original
			v1, err := Unmarshal([]byte(tt.input))
			if err != nil {
				t.Fatalf("Failed to unmarshal input %q: %v", tt.input, err)
			}
			// Marshal
			out, err := Marshal(v1)
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}
			// Re-parse
			v2, err := Unmarshal(out)
			if err != nil {
				t.Fatalf("Failed to unmarshal output %q: %v", string(out), err)
			}
			// Compare structurally via Value() (avoids map ordering issues)
			if !reflect.DeepEqual(v1.Value(), v2.Value()) {
				t.Errorf("Round-trip mismatch:\n  input:  %q\n  output: %q", tt.input, string(out))
			}
		})
	}
}

func TestMarshalRoundTripSpecialStrings(t *testing.T) {
	// Strings that exercise escaping through the round-trip
	strings := []string{
		"",
		"hello",
		"hello \"world\"",
		"back\\slash",
		"new\nline",
		"tab\there",
		"null\x00byte",
		"euro€sign",
		"emoji😀here",
		"\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f",
	}

	for _, s := range strings {
		t.Run(s, func(t *testing.T) {
			v1 := NewValue()
			v1.AsString(s)
			out, err := Marshal(v1)
			if err != nil {
				t.Fatalf("Failed to marshal string: %v", err)
			}
			v2, err := Unmarshal(out)
			if err != nil {
				t.Fatalf("Failed to unmarshal output %q: %v", string(out), err)
			}
			if v2.String() != s {
				t.Errorf("Round-trip string mismatch: expected %q, got %q", s, v2.String())
			}
		})
	}
}

func TestMarshalComplex(t *testing.T) {
	t.Run("deeply nested", func(t *testing.T) {
		// Build: {"a":{"b":{"c":[1,true,"deep"]}}}
		v := NewValue()
		v.AsObject(nil)
		b := v.AddField("a")
		b.AsObject(nil)
		c := b.AddField("b")
		c.AsObject(nil)
		arr := c.AddField("c")
		arr.AsArray(nil)
		arr.AddElement().AsInt(1)
		arr.AddElement().AsBool(true)
		arr.AddElement().AsString("deep")

		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Re-parse and check
		reparsed, err := Unmarshal(out)
		if err != nil {
			t.Fatalf("Failed to re-parse: %v", err)
		}
		if reparsed.Get("a", "b", "c", "2").String() != "deep" {
			t.Errorf("Deep value mismatch")
		}
	})

	t.Run("empty containers within containers", func(t *testing.T) {
		v := NewValue()
		v.AsObject(nil)
		v.AddField("a").AsArray(nil)
		v.AddField("b").AsObject(nil)

		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		reparsed, err := Unmarshal(out)
		if err != nil {
			t.Fatalf("Failed to re-parse: %v", err)
		}
		if !reparsed.Get("a").IsArray() || len(reparsed.Get("a").Array()) != 0 {
			t.Errorf("Expected empty array for key 'a'")
		}
		if !reparsed.Get("b").IsObject() || len(reparsed.Get("b").Object()) != 0 {
			t.Errorf("Expected empty object for key 'b'")
		}
	})

	t.Run("mixed types in array", func(t *testing.T) {
		v := NewValue()
		v.AsArray(nil)
		v.AddElement().AsInt(42)
		v.AddElement().AsFloat(3.14)
		v.AddElement().AsString("text")
		v.AddElement().AsBool(true)
		v.AddElement().AsBool(false)
		v.AddElement().AsNull()
		inner := v.AddElement()
		inner.AsObject(nil)
		inner.AddField("nested").AsInt(1)

		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		reparsed, err := Unmarshal(out)
		if err != nil {
			t.Fatalf("Failed to re-parse: %v", err)
		}
		if !reparsed.IsArray() || len(reparsed.Array()) != 7 {
			t.Errorf("Expected 7-element array, got %v", len(reparsed.Array()))
		}
	})
}

func TestMarshalNumberRoundTrip(t *testing.T) {
	// Parse JSON with scientific notation, marshal, re-parse
	input := `{"int":42,"float":3.14,"neg":-1,"exp":1.5e2}`
	v1, err := Unmarshal([]byte(input))
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	out, err := Marshal(v1)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	v2, err := Unmarshal(out)
	if err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}
	if v2.Get("int").Int() != 42 {
		t.Errorf("Expected int=42, got %v", v2.Get("int").Value())
	}
	if v2.Get("float").Float() != 3.14 {
		t.Errorf("Expected float=3.14, got %v", v2.Get("float").Value())
	}
	if v2.Get("neg").Int() != -1 {
		t.Errorf("Expected neg=-1, got %v", v2.Get("neg").Value())
	}
	// exp: 1.5e2 = 150.0 (parsed as float64)
	if v2.Get("exp").Float() != 150 {
		t.Errorf("Expected exp=150, got %v", v2.Get("exp").Value())
	}
}
