package jsonart

import (
	"bytes"
	"fmt"
	"strconv"
)

var (
	valueStart = "{, [, [0-9], -, t, f, n, \""
	bytesTrue  = []byte{'r', 'u', 'e'}
	bytesFalse = []byte{'a', 'l', 's', 'e'}
	bytesNull  = []byte{'u', 'l', 'l'}
)

const (
	stateNone = iota
	stateString
	stateArrayValueOrEnd
	stateArrayEndOrComma
	stateObjectKeyOrEnd
	stateObjectColon
	stateObjectEndOrComma
	stateObjectKey
)

type state struct {
	value  *Value
	state  int
	parent *state
}

type parser struct {
	data   []byte
	size   int
	offset int
	buf    []byte
	curr   *state
}

func (p *parser) unexpected(expect string) error {
	if p.offset < p.size {
		return fmt.Errorf("Unexpected token '%c' at: %d, expect: %s", p.data[p.offset], p.offset, expect)
	}
	return fmt.Errorf("Unexpected EOF, expect: %s", expect)
}

func (p *parser) unexpectedAt(expect string, offset int) error {
	if offset < p.size {
		return fmt.Errorf("Unexpected token '%c' at: %d, expect: %s", p.data[offset], offset, expect)
	}
	return fmt.Errorf("Unexpected EOF, expect: %s", expect)
}

func Unmarshal(data []byte) (value *Value, err error) {
	value = &Value{nil}
	root := &state{value, stateNone, nil}
	p := &parser{
		data:   data,
		size:   len(data),
		offset: 0,
		buf:    make([]byte, 0, 1024),
		curr:   root,
	}
	err = p.parse()
	return
}

func (p *parser) parseString() (string, error) {
	var hexDigits [4]int
	var pos int
	var hexIdx int
	var codePoint int

	p.offset++
	p.buf = p.buf[:0]
LOOP_STRING:
	for pos = p.offset; pos < p.size; pos++ {
		switch p.data[pos] {
		case '\\':
			pos++
			if pos == p.size {
				return "", p.unexpectedAt("escaped char", pos)
			}
			switch p.data[pos] {
			case 'U', 'u':
				pos++
				if p.size < pos+4 {
					return "", p.unexpectedAt("[0-F]", p.size)
				}
				for hexIdx = 0; hexIdx < 4; hexIdx++ {
					switch p.data[pos] {
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						hexDigits[hexIdx] = int(p.data[pos]) - 0x30
					case 'a', 'b', 'c', 'd', 'e', 'f':
						hexDigits[hexIdx] = int(p.data[pos]) - 0x57
					case 'A', 'B', 'C', 'D', 'E', 'F':
						hexDigits[hexIdx] = int(p.data[pos]) - 0x37
					default:
						return "", p.unexpectedAt("[0-F]", pos)
					}
					pos++
				}
				codePoint = (hexDigits[0] << 12) | (hexDigits[1] << 8) | (hexDigits[2] << 4) | (hexDigits[3])
				if codePoint > 0xD7FF && codePoint < 0xDC00 {
					if p.size < pos+6 {
						if p.size == pos || p.data[pos] != '\\' {
							return "", p.unexpectedAt("\\", p.size)
						}
						if p.size < pos+2 || (p.data[pos+1] != 'U' && p.data[pos+1] != 'u') {
							return "", p.unexpectedAt("Uu", pos+1)
						}
						return "", p.unexpectedAt("[0-F]", p.size)
					}
					if p.data[pos] != '\\' {
						return "", p.unexpectedAt("\\", pos)
					}
					pos++
					if p.data[pos] != 'U' && p.data[pos] != 'u' {
						return "", p.unexpectedAt("Uu", pos)
					}
					pos++
					for hexIdx = 0; hexIdx < 4; hexIdx++ {
						switch p.data[pos] {
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							hexDigits[hexIdx] = int(p.data[pos]) - 0x30
						case 'a', 'b', 'c', 'd', 'e', 'f':
							hexDigits[hexIdx] = int(p.data[pos]) - 0x57
						case 'A', 'B', 'C', 'D', 'E', 'F':
							hexDigits[hexIdx] = int(p.data[pos]) - 0x37
						default:
							return "", p.unexpectedAt("[0-F]", pos)
						}
						pos++
					}
					lowSurrogate := (hexDigits[0] << 12) | (hexDigits[1] << 8) | (hexDigits[2] << 4) | (hexDigits[3])
					if lowSurrogate < 0xDC00 || lowSurrogate > 0xDFFF {
						return "", p.unexpectedAt("[0xdc00 - 0xdfff]", pos-4)
					}
					codePoint = (((codePoint - 0xD800) << 10) | (lowSurrogate - 0xDC00)) + 0x10000
				}
				pos--
				if codePoint < 0x0080 {
					p.buf = append(p.buf, byte(codePoint))
				} else if codePoint < 0x0800 {
					p.buf = append(p.buf, 0xC0|byte(codePoint>>6), 0x80|byte(codePoint&0xBF))
				} else if codePoint < 0x10000 {
					p.buf = append(p.buf, 0xE0|byte(codePoint>>12), 0x80|byte((codePoint>>6)&0xBF), 0x80|byte(codePoint&0xBF))
				} else {
					p.buf = append(p.buf, 0xF0|byte(codePoint>>18), 0x80|byte((codePoint>>12)&0xBF), 0x80|byte((codePoint>>6)&0xBF), 0x80|byte(codePoint&0xBF))
				}
			case 't':
				p.buf = append(p.buf, '\t')
			case 'r':
				p.buf = append(p.buf, '\r')
			case 'n':
				p.buf = append(p.buf, '\n')
			case '"':
				p.buf = append(p.buf, '"')
			case '\\':
				p.buf = append(p.buf, '\\')
			case '/':
				p.buf = append(p.buf, '/')
			case 'b':
				p.buf = append(p.buf, 0x08)
			case 'f':
				p.buf = append(p.buf, 0x0C)
			default:
				return "", p.unexpectedAt("escape sequence", pos)
			}
		case '"':
			break LOOP_STRING
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
			16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31:
			return "", p.unexpectedAt("unicode", pos)
		default:
			p.buf = append(p.buf, p.data[pos])
		}
	}
	if pos == p.size {
		return "", p.unexpectedAt("\" to end string", pos)
	}
	p.offset = pos + 1
	return string(p.buf), nil
}

func (p *parser) parseNumber() (interface{}, error) {
	var pos int
	var digitCount int
	var numStart int
	var decimalPart []byte
	var exponentPart []byte

	decimalPart = nil
	exponentPart = nil
	numStart = p.offset
	if p.data[p.offset] == '-' {
		p.offset++
	}
LOOP_NUM_INT:
	for pos = p.offset; pos < p.size; pos++ {
		switch p.data[pos] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			continue
		default:
			break LOOP_NUM_INT
		}
	}
	digitCount = pos - p.offset
	if digitCount == 0 {
		return nil, p.unexpected("[0-9]")
	}
	if p.data[p.offset] == '0' {
		if digitCount != 1 {
			p.offset++
			return nil, p.unexpected("[.eE]")
		}
	}
	p.offset = pos
	if p.offset < p.size && p.data[p.offset] == '.' {
		p.offset++
	LOOP_NUM_DEC:
		for pos = p.offset; pos < p.size; pos++ {
			switch p.data[pos] {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				continue
			default:
				break LOOP_NUM_DEC
			}
		}
		if pos == p.offset {
			return nil, p.unexpected("[0-9]")
		}
		decimalPart = p.data[p.offset:pos]
		p.offset = pos
	}
	if p.offset < p.size && (p.data[p.offset] == 'e' || p.data[p.offset] == 'E') {
		p.offset++
		if p.offset == p.size {
			return nil, p.unexpected("[0-9]")
		}
		if p.data[p.offset] == '-' || p.data[p.offset] == '+' {
			p.offset++
		}

	LOOP_NUM_EXP:
		for pos = p.offset; pos < p.size; pos++ {
			switch p.data[pos] {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				continue
			default:
				break LOOP_NUM_EXP
			}
		}
		if pos == p.offset {
			return nil, p.unexpected("[0-9]")
		}
		exponentPart = p.data[p.offset:pos]
		p.offset = pos
	}
	if decimalPart == nil && exponentPart == nil {
		result, err := strconv.ParseInt(string(p.data[numStart:p.offset]), 10, 64)
		return result, err
	}
	result, err := strconv.ParseFloat(string(p.data[numStart:p.offset]), 64)
	return result, err
}

func (p *parser) parseLiteral(expected []byte, value interface{}) error {
	expectedLen := len(expected)
	if p.size < p.offset+expectedLen {
		expectByte := expected[p.size-p.offset]
		return p.unexpectedAt(string(expectByte), p.size)
	}
	if bytes.Equal(p.data[p.offset:p.offset+expectedLen], expected) {
		p.offset += expectedLen
		p.curr.value.value = value
		p.curr = p.curr.parent
		return nil
	}
	// Build the full literal name for error message
	var name string
	switch {
	case &expected[0] == &bytesTrue[0]:
		name = "true"
	case &expected[0] == &bytesFalse[0]:
		name = "false"
	default:
		name = "null"
	}
	return p.unexpected(name)
}

func (p *parser) parse() error {
	for {
	LOOP_WHITESPACE:
		for ; p.offset < p.size; p.offset++ {
			switch p.data[p.offset] {
			case '\t', '\r', '\n', ' ':
				continue
			default:
				break LOOP_WHITESPACE
			}
		}
		if p.curr == nil {
			if p.offset != p.size {
				return p.unexpected("EOF")
			}
			return nil
		}
		switch p.curr.state {
		case stateArrayValueOrEnd:
			if p.offset == p.size {
				return p.unexpected("value or ]")
			}
			switch p.data[p.offset] {
			case ']':
				p.curr = p.curr.parent
				p.offset++
			default:
				p.curr.state = stateArrayEndOrComma
				p.curr = &state{p.curr.value.AddElement(), stateNone, p.curr}
			}
			continue
		case stateArrayEndOrComma:
			if p.offset == p.size {
				return p.unexpected(", or ]")
			}
			switch p.data[p.offset] {
			case ']':
				p.offset++
				p.curr = p.curr.parent
			case ',':
				p.offset++
				p.curr.state = stateArrayEndOrComma
				p.curr = &state{p.curr.value.AddElement(), stateNone, p.curr}
			default:
				return p.unexpected(", or ]")
			}
			continue
		case stateObjectColon:
			if p.offset == p.size {
				return p.unexpected(":")
			}
			p.offset++
			p.curr.state = stateNone
			continue
		case stateObjectEndOrComma:
			if p.offset == p.size {
				return p.unexpected(", or }")
			}
			switch p.data[p.offset] {
			case ',':
				p.curr.state = stateObjectKey
			case '}':
				p.curr = p.curr.parent
			default:
				return p.unexpected(", or }")
			}
			p.offset++
			continue
		case stateObjectKeyOrEnd:
			if p.offset == p.size {
				return p.unexpected("\" or }")
			}
			if p.data[p.offset] == '}' {
				p.offset++
				p.curr = p.curr.parent
				continue
			}
			fallthrough
		case stateObjectKey:
			if p.offset == p.size {
				return p.unexpected("\"")
			}
			if p.data[p.offset] != '"' {
				return p.unexpected("\"")
			}
			fallthrough
		case stateString:
			str, err := p.parseString()
			if err != nil {
				return err
			}
			if p.curr.state == stateString {
				p.curr.value.value = str
				p.curr = p.curr.parent
			} else {
				p.curr.state = stateObjectEndOrComma
				p.curr = &state{p.curr.value.AddField(str), stateObjectColon, p.curr}
			}
			continue
		default:
			if p.offset == p.size {
				return p.unexpected("value")
			}
			switch p.data[p.offset] {
			case '{':
				p.curr.state = stateObjectKeyOrEnd
				p.curr.value.value = map[string]*Value{}
				p.offset++
				continue
			case '[':
				p.curr.state = stateArrayValueOrEnd
				p.curr.value.value = []*Value{}
				p.offset++
				continue
			case '"':
				p.curr.state = stateString
				continue
			case '0', '1', '2', '3', '4',
				'5', '6', '7', '8', '9', '-':
				result, err := p.parseNumber()
				if err != nil {
					return err
				}
				p.curr.value.value = result
				p.curr = p.curr.parent
				continue
			case 'n':
				p.offset++
				if err := p.parseLiteral(bytesNull, NULL); err != nil {
					return err
				}
				continue
			case 't':
				p.offset++
				if err := p.parseLiteral(bytesTrue, true); err != nil {
					return err
				}
				continue
			case 'f':
				p.offset++
				if err := p.parseLiteral(bytesFalse, false); err != nil {
					return err
				}
				continue
			default:
				return p.unexpected(valueStart)
			}
		}
	}
}
