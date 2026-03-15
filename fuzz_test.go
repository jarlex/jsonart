package jsonart

import (
	"math"
	"reflect"
	"testing"
)

// valuesEqual compares two Value trees for structural equivalence.
// Uses .Value() to convert to interface{} and reflect.DeepEqual for comparison.
// Handles float64 precision differences by comparing with a tolerance.
func valuesEqual(a, b *Value) bool {
	av := a.Value()
	bv := b.Value()
	if reflect.DeepEqual(av, bv) {
		return true
	}
	return deepEqualWithFloatTolerance(av, bv)
}

// deepEqualWithFloatTolerance recursively compares two values, treating
// float64 values as equal if they are within a small relative tolerance.
func deepEqualWithFloatTolerance(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch av := a.(type) {
	case float64:
		switch bv := b.(type) {
		case float64:
			if math.IsNaN(av) && math.IsNaN(bv) {
				return true
			}
			if av == bv {
				return true
			}
			diff := math.Abs(av - bv)
			max := math.Max(math.Abs(av), math.Abs(bv))
			if max == 0 {
				return diff < 1e-15
			}
			return diff/max < 1e-12
		case int64:
			// float64(1000) from "1e3" marshals as "1000" which re-parses as int64
			return av == float64(bv)
		default:
			return false
		}
	case int64:
		switch bv := b.(type) {
		case int64:
			return av == bv
		case float64:
			// symmetric: int64 compared with float64
			return float64(av) == bv
		default:
			return false
		}
	case bool:
		bv, ok := b.(bool)
		if !ok {
			return false
		}
		return av == bv
	case string:
		bv, ok := b.(string)
		if !ok {
			return false
		}
		return av == bv
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		if len(av) != len(bv) {
			return false
		}
		for k, va := range av {
			vb, exists := bv[k]
			if !exists {
				return false
			}
			if !deepEqualWithFloatTolerance(va, vb) {
				return false
			}
		}
		return true
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !deepEqualWithFloatTolerance(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(a, b)
	}
}

// addJSONSeeds adds a diverse corpus of JSON examples to the fuzz target.
func addJSONSeeds(f *testing.F) {
	f.Helper()
	seeds := []string{
		// Primitives
		`null`,
		`true`,
		`false`,
		// Numbers: integers
		`0`,
		`42`,
		`-42`,
		// Numbers: floats
		`3.14`,
		`-3.14`,
		// Numbers: scientific notation
		`1e3`,
		`1E3`,
		`1e+3`,
		`1e-3`,
		`1.5e2`,
		// Strings
		`""`,
		`"hello"`,
		`"hello world"`,
		`"\"\\\/\b\f\n\r\t"`,
		`"\u20AC"`,
		`"\uD83D\uDE00"`,
		// Containers: empty
		`{}`,
		`[]`,
		// Containers: non-empty
		`{"key":"value"}`,
		`[1,2,3]`,
		// Mixed types
		`[1,"two",true,null]`,
		`{"a":1,"b":true,"c":null}`,
		// Nested
		`{"a":{"b":2}}`,
		`[[1,2],[3,4]]`,
		// Complex
		`{"users":[{"name":"Alice","age":30},{"name":"Bob","age":25}],"count":2}`,
		// Whitespace variations
		" \t\n\r{ \t\n\r\"key\" \t\n\r: \t\n\r1 \t\n\r} \t\n\r",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
}

// FuzzUnmarshal tests that Unmarshal never panics on arbitrary input.
func FuzzUnmarshal(f *testing.F) {
	addJSONSeeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		v, err := Unmarshal(data)
		if err != nil {
			return
		}
		if v == nil {
			t.Fatal("Unmarshal returned nil value with nil error")
		}
		// Verify the value is usable
		_ = v.Value()
	})
}

// FuzzRoundTrip tests that valid JSON survives Unmarshal -> Marshal -> Unmarshal
// with structural equivalence preserved.
func FuzzRoundTrip(f *testing.F) {
	addJSONSeeds(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		v1, err := Unmarshal(data)
		if err != nil {
			return // skip invalid JSON
		}

		out, err := Marshal(v1)
		if err != nil {
			t.Fatalf("Marshal failed on valid parsed value: %v", err)
		}

		v2, err := Unmarshal(out)
		if err != nil {
			t.Fatalf("Unmarshal failed on Marshal output %q: %v", string(out), err)
		}

		if !valuesEqual(v1, v2) {
			t.Fatalf("Round-trip mismatch:\n  input:    %q\n  marshaled: %q\n  v1: %v\n  v2: %v",
				string(data), string(out), v1.Value(), v2.Value())
		}
	})
}

// FuzzMarshalString tests that arbitrary strings survive the round-trip
// through Value -> Marshal -> Unmarshal.
func FuzzMarshalString(f *testing.F) {
	seeds := []string{
		"",
		"hello",
		"hello world",
		"hello \"world\"",
		"back\\slash",
		"new\nline",
		"tab\there",
		"\x00",
		"\x01\x02\x03\x04\x05",
		"euro\u20ACsign",
		"\U0001F600",
		"/forward/slash",
		"\b\f\r",
		"a\"b\\c\nd\te\x00f",
		"\xff\xfe",
		"日本語テキスト",
		"very long string " + "abcdefghij" + "abcdefghij" + "abcdefghij" + "abcdefghij" + "abcdefghij",
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		s := string(data)
		v := NewValue()
		v.AsString(s)

		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("Marshal failed for string: %v", err)
		}

		v2, err := Unmarshal(out)
		if err != nil {
			t.Fatalf("Unmarshal failed on marshaled string %q: %v", string(out), err)
		}

		if v2.String() != s {
			t.Fatalf("String round-trip mismatch:\n  original: %q\n  got:      %q\n  json:     %q",
				s, v2.String(), string(out))
		}
	})
}
