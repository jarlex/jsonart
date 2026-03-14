package jsonart

import (
	"strings"
	"testing"
)

func TestParseStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		err      bool
	}{
		{"empty string", `""`, "", false},
		{"simple string", `"hello"`, "hello", false},
		{"escapes", `"\"\\\/\b\f\n\r\t"`, "\"\\/\b\f\n\r\t", false},
		{"unicode simple", `"\u20AC"`, "€", false},
		{"unicode surrogate pair", `"\uD83D\uDE00"`, "😀", false},
		{"unclosed string", `"abc`, "", true},
		{"invalid escape", `"\x"`, "", true},
		{"invalid unicode", `"\u123"`, "", true}, // missing a char
		{"string with internal space", `"hello world"`, "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Unmarshal([]byte(tt.input))
			if tt.err {
				if err == nil {
					t.Errorf("Expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error for %q: %v", tt.input, err)
				return
			}
			if !v.IsString() || v.String() != tt.expected {
				t.Errorf("Expected %q, got %v", tt.expected, v.Value())
			}
		})
	}
}

func TestParseNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{} // int64 or float64
		err      bool
	}{
		{"zero", `0`, int64(0), false},
		{"positive int", `42`, int64(42), false},
		{"negative int", `-42`, int64(-42), false},
		{"float", `3.14`, float64(3.14), false},
		{"negative float", `-3.14`, float64(-3.14), false},
		{"exponential", `1e3`, float64(1000), false},
		{"exponential capital", `1E3`, float64(1000), false},
		{"exponential with plus", `1e+3`, float64(1000), false},
		{"exponential with minus", `1e-3`, float64(0.001), false},
		{"fraction and exponent", `1.5e2`, float64(150), false},
		{"leading zero error", `01`, nil, true},
		{"minus only", `-`, nil, true},
		{"trailing dot", `1.`, nil, true},
		{"exponent no digits", `1e`, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Unmarshal([]byte(tt.input))
			if tt.err {
				if err == nil {
					t.Errorf("Expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error for %q: %v", tt.input, err)
				return
			}

			switch exp := tt.expected.(type) {
			case int64:
				if !v.IsInt() || v.Int() != exp {
					t.Errorf("Expected int %v, got %v (type %v)", exp, v.Value(), v.IsInt())
				}
			case float64:
				if v.IsInt() || !v.IsNumber() || v.Float() != exp {
					t.Errorf("Expected float %v, got %v", exp, v.Value())
				}
			}
		})
	}
}

func TestParseLiterals(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*Value) bool
		err   bool
	}{
		{"true", `true`, func(v *Value) bool { return v.IsTrue() }, false},
		{"false", `false`, func(v *Value) bool { return v.IsFalse() }, false},
		{"null", `null`, func(v *Value) bool { return v.IsNull() }, false},
		{"invalid true", `truX`, nil, true},
		{"invalid false", `falX`, nil, true},
		{"invalid null", `nulX`, nil, true},
		{"bare word", `foo`, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Unmarshal([]byte(tt.input))
			if tt.err {
				if err == nil {
					t.Errorf("Expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if !tt.check(v) {
				t.Errorf("Check failed for %q", tt.input)
			}
		})
	}
}

func TestParseObjects(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   bool
	}{
		{"empty object", `{}`, false},
		{"single string value", `{"key": "value"}`, false},
		{"multiple values", `{"a": 1, "b": true, "c": null}`, false},
		{"nested object", `{"a": {"b": 2}}`, false},
		{"missing comma", `{"a": 1 "b": 2}`, true},
		{"missing colon", `{"a" 1}`, true},
		{"trailing comma", `{"a": 1,}`, true},
		{"unclosed object", `{"a": 1`, true},
		{"non-string key", `{1: 1}`, true},
		{"unquoted key", `{a: 1}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Unmarshal([]byte(tt.input))
			if tt.err && err == nil {
				t.Errorf("Expected error for %q", tt.input)
			} else if !tt.err && err != nil {
				t.Errorf("Unexpected error for %q: %v", tt.input, err)
			}
		})
	}
}

func TestParseArrays(t *testing.T) {
	tests := []struct {
		name  string
		input string
		err   bool
	}{
		{"empty array", `[]`, false},
		{"single element", `[1]`, false},
		{"multiple elements", `[1, 2, "three", true, null]`, false},
		{"nested array", `[[1, 2], [3, 4]]`, false},
		{"missing comma", `[1 2]`, true},
		{"trailing comma", `[1,]`, true},
		{"unclosed array", `[1, 2`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Unmarshal([]byte(tt.input))
			if tt.err && err == nil {
				t.Errorf("Expected error for %q", tt.input)
			} else if !tt.err && err != nil {
				t.Errorf("Unexpected error for %q: %v", tt.input, err)
			}
		})
	}
}

func TestDeepNesting(t *testing.T) {
	const depth = 500

	// Deep nested array
	arrStr := strings.Repeat("[", depth) + strings.Repeat("]", depth)
	_, err := Unmarshal([]byte(arrStr))
	if err != nil {
		t.Errorf("Unexpected error parsing deep array: %v", err)
	}

	// Deep nested object
	objStr := strings.Repeat(`{"a":`, depth) + "1" + strings.Repeat(`}`, depth)
	_, err = Unmarshal([]byte(objStr))
	if err != nil {
		t.Errorf("Unexpected error parsing deep object: %v", err)
	}
}

func TestWhitespace(t *testing.T) {
	input := " \t\n\r{ \t\n\r\"key\" \t\n\r: \t\n\r1 \t\n\r} \t\n\r"
	v, err := Unmarshal([]byte(input))
	if err != nil {
		t.Errorf("Failed to parse with whitespace: %v", err)
	}
	if !v.IsObject() || v.Object()["key"].Int() != 1 {
		t.Errorf("Incorrect parse result for whitespace input")
	}
}

func TestTrailingGarbage(t *testing.T) {
	inputs := []string{
		`{} {}`,
		`1 2`,
		`"string" foo`,
		`true ,`,
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := Unmarshal([]byte(input))
			if err == nil {
				t.Errorf("Expected error due to trailing garbage for: %q", input)
			}
		})
	}
}
