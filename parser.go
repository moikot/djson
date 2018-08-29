package djson

import (
	"errors"
	"fmt"
	"strconv"
)

// MergeValue deserializes the input string and merges result to the map provided.
func MergeValue(m map[string]interface{}, str string) error {
	parser := &parser{
		lex: newLex(str),
	}
	parser.rightValueReader = parser.readRightValue
	return parser.append(m, str)
}

// MergeString deserializes the input string and merges result to the map provided.
func MergeString(m map[string]interface{}, str string) error {
	parser := &parser{
		lex: newLex(str),
	}
	parser.rightValueReader = parser.readRightString
	return parser.append(m, str)
}

type parser struct {
	lex              lexer
	rightValueReader func() (interface{}, error)
}

func (p *parser) append(m map[string]interface{}, str string) error {
	builder := newRootBuilder(m)
	// Expecting a map at the top level
	err := p.readMap(builder)
	if err != nil {
		p.lex.drain()
		p.lex = nil
		return err
	}
	p.lex = nil
	return nil
}

func (p *parser) nextToken() token {
	return p.lex.nextToken()
}

func (p *parser) readMap(b mapBuilderFactory) error {
	var key string
	switch tok := p.nextToken(); tok.TokenType {
	case tokenMapKey:
		key = tok.value
	default:
		return tokenToError(tok)
	}
	return p.readLeftValue(b.newMapBuilder(key))
}

func (p *parser) readLeftValue(b builder) error {
	switch tok := p.nextToken(); tok.TokenType {
	case tokenMapKeySeparator:
		return p.readMap(b)
	case tokenArrayIndexStart:
		return p.readArray(b)
	case tokenAssignment:
		val, err := p.rightValueReader()
		if err != nil {
			return err
		}
		b.set(val)
	default:
		return tokenToError(tok)
	}
	return nil
}

func (p *parser) readArray(b builder) (err error) {
	var index int
	switch tok := p.nextToken(); tok.TokenType {
	case tokenArrayIndex:
		index, err = strconv.Atoi(tok.value)
		if err != nil {
			return err
		}
	default:
		return tokenToError(tok)
	}

	switch tok := p.nextToken(); tok.TokenType {
	case tokenArrayIndexFinish:
	default:
		return tokenToError(tok)
	}

	return p.readLeftValue(b.newArrayBuilder(index))
}

func (p *parser) readRightValue() (interface{}, error) {
	switch tok := p.nextToken(); tok.TokenType {
	case tokenEnd:
		return "", nil
	case tokenValue:
		return tryParse(tok.value), nil
	default:
		return nil, tokenToError(tok)
	}
}

func (p *parser) readRightString() (interface{}, error) {
	switch tok := p.nextToken(); tok.TokenType {
	case tokenEnd:
		return "", nil
	case tokenValue:
		return tok.value, nil
	default:
		return nil, tokenToError(tok)
	}
}

func tryParse(val string) interface{} {
	b, err := strconv.ParseBool(val)
	if err == nil {
		return b
	}
	i, err := strconv.ParseInt(val, 10, 64)
	if err == nil {
		return i
	}
	f, err := strconv.ParseFloat(val, 64)
	if err == nil {
		return f
	}
	if val == "null" || val == "NULL" || val == "Null" {
		return nil
	}
	return val
}

func tokenToError(tok token) error {
	if tok.TokenType == tokenError {
		return errors.New(tok.value)
	}
	return fmt.Errorf("unexpected \"%s\"", tok.value)
}
