package jsonart

import (
	"fmt"
	"math"
	"strconv"
)

type marshaler struct {
	buf []byte
}

// Marshal serializes a Value tree to compact JSON bytes.
// Returns an error if v is nil, uninitialized, or contains
// non-representable values (NaN, Inf).
func Marshal(v *Value) ([]byte, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot marshal nil *Value")
	}
	m := &marshaler{buf: make([]byte, 0, 1024)}
	if err := m.marshal(v); err != nil {
		return nil, err
	}
	return m.buf, nil
}

func (m *marshaler) marshal(v *Value) error {
	switch val := v.value.(type) {
	case nil:
		return fmt.Errorf("cannot marshal uninitialized Value")
	case null:
		_ = val
		m.buf = append(m.buf, "null"...)
	case bool:
		if val {
			m.buf = append(m.buf, "true"...)
		} else {
			m.buf = append(m.buf, "false"...)
		}
	case string:
		m.marshalString(val)
	case int64:
		m.buf = strconv.AppendInt(m.buf, val, 10)
	case float64:
		if err := m.marshalFloat(val); err != nil {
			return err
		}
	case map[string]*Value:
		if err := m.marshalObject(val); err != nil {
			return err
		}
	case []*Value:
		if err := m.marshalArray(val); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cannot marshal unknown type %T", v.value)
	}
	return nil
}

func (m *marshaler) marshalString(s string) {
	m.buf = append(m.buf, '"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"':
			m.buf = append(m.buf, '\\', '"')
		case c == '\\':
			m.buf = append(m.buf, '\\', '\\')
		case c == '\n':
			m.buf = append(m.buf, '\\', 'n')
		case c == '\r':
			m.buf = append(m.buf, '\\', 'r')
		case c == '\t':
			m.buf = append(m.buf, '\\', 't')
		case c == '\b':
			m.buf = append(m.buf, '\\', 'b')
		case c == '\f':
			m.buf = append(m.buf, '\\', 'f')
		case c < 0x20:
			m.buf = append(m.buf, '\\', 'u', '0', '0',
				hexChar(c>>4), hexChar(c&0x0F))
		default:
			m.buf = append(m.buf, c)
		}
	}
	m.buf = append(m.buf, '"')
}

func (m *marshaler) marshalFloat(f float64) error {
	if math.IsNaN(f) {
		return fmt.Errorf("cannot marshal NaN float value")
	}
	if math.IsInf(f, 0) {
		return fmt.Errorf("cannot marshal infinite float value")
	}
	m.buf = strconv.AppendFloat(m.buf, f, 'g', -1, 64)
	return nil
}

func (m *marshaler) marshalObject(obj map[string]*Value) error {
	m.buf = append(m.buf, '{')
	first := true
	for key, val := range obj {
		if val == nil {
			return fmt.Errorf("cannot marshal nil *Value for object key %q", key)
		}
		if !first {
			m.buf = append(m.buf, ',')
		}
		first = false
		m.marshalString(key)
		m.buf = append(m.buf, ':')
		if err := m.marshal(val); err != nil {
			return err
		}
	}
	m.buf = append(m.buf, '}')
	return nil
}

func (m *marshaler) marshalArray(arr []*Value) error {
	m.buf = append(m.buf, '[')
	for i, val := range arr {
		if val == nil {
			return fmt.Errorf("cannot marshal nil *Value in array at index %d", i)
		}
		if i > 0 {
			m.buf = append(m.buf, ',')
		}
		if err := m.marshal(val); err != nil {
			return err
		}
	}
	m.buf = append(m.buf, ']')
	return nil
}

func hexChar(c byte) byte {
	if c < 10 {
		return '0' + c
	}
	return 'a' + c - 10
}
