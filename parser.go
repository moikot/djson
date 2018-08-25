package djson

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Parse parses the input string and converts it into a map.
func Parse(str string) (map[string]interface{}, error) {
	parser := newParser(str)
	m, err := parser.parse()
	if err != nil {
		parser.lex.drain()
		parser.lex = nil
		return nil, err
	}
	parser.lex = nil
	return m, nil
}

type parser struct {
	lex *lex
}

func newParser(str string) *parser {
	return &parser{
		lex: newLex(str),
	}
}

func (p *parser) nextToken() token {
	return <-p.lex.tokens
}

// The expression is: (key=[value])[,(key=[value])]
func (p *parser) parse() (map[string]interface{}, error) {
	builder := newRootBuilder()
	for {
		// Expecting a map at the top level
		err := p.readMap(builder)
		if err != nil {
			return nil, err
		}
		switch tok := p.nextToken(); tok.TokenType {
		case tokenEnd:
			return builder.get(), nil
		case tokenNextKey:
			continue
		default:
			return nil, tokenToError(tok)
		}
	}
}

func (p *parser) readMap(b builder) error {
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
		val, err := p.readRightValue()
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
	case tokenString:
		return strings.Trim(tok.value, "'"), nil
	case tokenVerbatimString:
		return tok.value, nil
	case tokenValueArrayStart:
		return p.readValuesArray()
	default:
		return nil, tokenToError(tok)
	}
}

func (p *parser) readValuesArray() ([]interface{}, error) {
	var arr []interface{}
	var valueDefined = false
	for {
		switch tok := p.nextToken(); tok.TokenType {
		case tokenValue:
			arr = append(arr, tryParse(tok.value))
			valueDefined = true
		case tokenString:
			arr = append(arr, strings.Trim(tok.value, "'"))
			valueDefined = true
		case tokenNextValue:
			if !valueDefined {
				// If value is not defined, add an empty string.
				arr = append(arr, "")
			}
			valueDefined = false
		case tokenValueArrayFinish:
			return arr, nil
		default:
			return nil, tokenToError(tok)
		}
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
