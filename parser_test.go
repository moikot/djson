package djson

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type parserTestCase struct {
	desc     string                 // Description
	input    string                 // Input string
	expected map[string]interface{} // The expected result
}

func newParserTestCase(desc, input string, expected map[string]interface{}) parserTestCase {
	return parserTestCase{
		desc:     desc,
		input:    input,
		expected: expected,
	}
}

var (
	commonParserTestCases = []parserTestCase{
		newParserTestCase(
			"an simple key value pair", "key=val",
			map[string]interface{}{
				"key": "val",
			},
		),
		newParserTestCase(
			"an nested key", "key1.key2=val",
			map[string]interface{}{
				"key1": map[string]interface{}{
					"key2": "val",
				},
			},
		),
		newParserTestCase(
			"a simple list", "foo[0]=bar",
			map[string]interface{}{
				"foo": []interface{}{"bar"},
			},
		),
		newParserTestCase(
			"a list with an empty value", "foo[0]=",
			map[string]interface{}{
				"foo": []interface{}{""},
			},
		),
		newParserTestCase(
			"a map in an array", "foo[0].key=bar",
			map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"key": "bar",
					},
				},
			},
		),
		newParserTestCase(
			"a nested array with missing values", "foo[1][1]=bar",
			map[string]interface{}{
				"foo": []interface{}{
					nil,
					[]interface{}{
						nil,
						"bar",
					},
				},
			},
		),
	}
)

func Test_Value_Parser_Succeeds(t *testing.T) {
	testCases := []parserTestCase{
		newParserTestCase(
			"an boolean value true", "key=true",
			map[string]interface{}{
				"key": true,
			},
		),
		newParserTestCase(
			"an boolean value false", "key=false",
			map[string]interface{}{
				"key": false,
			},
		),
		newParserTestCase(
			"an integer value 1000", "key=1000",
			map[string]interface{}{
				"key": int64(1000),
			},
		),
		newParserTestCase(
			"a floating point value 10.01", "key=10.01",
			map[string]interface{}{
				"key": float64(10.01),
			},
		),
		newParserTestCase(
			"a floating point value 0.01", "key=0.01",
			map[string]interface{}{
				"key": float64(0.01),
			},
		),
	}
	testCases = append(testCases, commonParserTestCases...)
	for _, test := range testCases {
		m := map[string]interface{}{}
		err := MergeValue(m, test.input)
		assertNoError(t, err, test, m)
	}
}

func Test_String_Parser_Succeeds(t *testing.T) {
	testCases := []parserTestCase{
		newParserTestCase(
			"an boolean value true", "key=true",
			map[string]interface{}{
				"key": "true",
			},
		),
		newParserTestCase(
			"an boolean value false", "key=false",
			map[string]interface{}{
				"key": "false",
			},
		),
		newParserTestCase(
			"an integer value 1000", "key=1000",
			map[string]interface{}{
				"key": "1000",
			},
		),
		newParserTestCase(
			"a floating point value 10.01", "key=10.01",
			map[string]interface{}{
				"key": "10.01",
			},
		),
		newParserTestCase(
			"a floating point value 0.01", "key=0.01",
			map[string]interface{}{
				"key": "0.01",
			},
		),
	}
	testCases = append(testCases, commonParserTestCases...)
	for _, test := range testCases {
		m := map[string]interface{}{}
		err := MergeString(m, test.input)
		assertNoError(t, err, test, m)
	}
}

func assertNoError(t *testing.T, err error, test parserTestCase, m map[string]interface{}) {
	if err != nil {
		t.Errorf("\nIn the case of %s \"%s\"\nexpected:\n\tsuccess\ngot:\n\t%+v",
			test.desc, test.input, err.Error())
	} else {
		if !reflect.DeepEqual(m, test.expected) {
			t.Errorf("\nIn the case of %s \"%s\"\nexpected:\n\t%+v\ngot:\n\t%+v",
				test.desc, test.input, test.expected, m)
		}
	}
}

func Test_Parser_Accumulates_In_Input_Map(t *testing.T) {
	testCases := []parserTestCase{
		newParserTestCase(
			"merging two root keys", "key1=val1,key2=val2",
			map[string]interface{}{
				"key1": "val1",
				"key2": "val2",
			},
		),
		newParserTestCase(
			"merging with a shared root key", "foo.key1=val1,foo.key2=val2",
			map[string]interface{}{
				"foo": map[string]interface{}{
					"key1": "val1",
					"key2": "val2",
				},
			},
		),
		newParserTestCase(
			"merging with jagged keys", "foo.key1=val1,foo.bar.key2=val2",
			map[string]interface{}{
				"foo": map[string]interface{}{
					"key1": "val1",
					"bar": map[string]interface{}{
						"key2": "val2",
					},
				},
			},
		),
		newParserTestCase(
			"merging with collided keys", "foo=val1,foo=val2",
			map[string]interface{}{
				"foo": "val2",
			},
		),
		newParserTestCase(
			"merging arrays with extending", "foo[0]=val1,foo[1]=val2",
			map[string]interface{}{
				"foo": []interface{}{
					"val1",
					"val2",
				},
			},
		),
		newParserTestCase(
			"merging arrays with filling", "foo[1]=val2,foo[0]=val1",
			map[string]interface{}{
				"foo": []interface{}{
					"val1",
					"val2",
				},
			},
		),
		newParserTestCase(
			"merging arrays with rewrite", "foo[0]=val1,foo[0]=val2",
			map[string]interface{}{
				"foo": []interface{}{
					"val2",
				},
			},
		),
		newParserTestCase(
			"merging nested arrays with rewrite", "foo[0][0]=val1,foo[0][0]=val2",
			map[string]interface{}{
				"foo": []interface{}{
					[]interface{}{
						"val2",
					},
				},
			},
		),
		newParserTestCase(
			"merging a map in an array", "foo[0].key1=val1,foo[0].key2=val2",
			map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"key1": "val1",
						"key2": "val2",
					},
				},
			},
		),
		newParserTestCase(
			"a value overridden by a null value", "foo[0]=val,foo[0]=null",
			map[string]interface{}{
				"foo": []interface{}{
					nil,
				},
			},
		),
	}

	for _, test := range testCases {
		m := map[string]interface{}{}
		parts := strings.Split(test.input, ",")
		for _, input := range parts {
			err := MergeValue(m, input)
			if err != nil {
				t.Errorf("\nIn the case of %s \"%s\"\nexpected:\n\tsuccess\ngot:\n\t%+v",
					test.desc, test.input, err.Error())
			}
		}
		if !reflect.DeepEqual(m, test.expected) {
			t.Errorf("\nIn the case of %s \"%s\"\nexpected:\n\t%+v\ngot:\n\t%+v",
				test.desc, test.input, test.expected, m)
		}
	}
}

type parserErrorTestCase struct {
	desc     string // Description
	input    string // Input string
	expected string // The expected error message
}

func newParserErrorTestCase(desc, input, expected string) parserErrorTestCase {
	return parserErrorTestCase{
		desc:     desc,
		input:    input,
		expected: expected,
	}
}

func Test_Parser_Fails(t *testing.T) {
	testCases := []parserErrorTestCase{
		newParserErrorTestCase(
			"an empty string", "",
			"unable to parse \"\", unexpected end, expecting a map key",
		),
		newParserErrorTestCase(
			"key with no value", "foo",
			"unable to parse \"foo\", unexpected end, expecting '.', '=' or '['",
		),
		newParserErrorTestCase(
			"an array with no value", "foo[0]",
			"unable to parse \"foo[0]\", unexpected end, expecting '.', '=' or '['",
		),
		newParserErrorTestCase(
			"an array index is not defined", "foo[",
			"unable to parse \"foo[\", unexpected end, expecting an array index",
		),
		newParserErrorTestCase(
			"an array index is not complete", "foo[0",
			"unable to parse \"foo[0\", unexpected end, expecting ']'",
		),
		newParserErrorTestCase(
			"an array index is out of range", "foo[99999999999999999999]",
			"unable to parse \"foo[99999999999999999999]\", strconv.Atoi: parsing \"99999999999999999999\": value out of range",
		),
	}

	for _, test := range testCases {
		m := map[string]interface{}{}
		err := MergeValue(m, test.input)
		if err == nil {
			t.Errorf("\nIn the case of %s \"%s\"\nexpected error:\n\t%+v\ngot:\n\tsuccess",
				test.desc, test.input, test.expected)
		} else {
			if err.Error() != test.expected {
				t.Errorf("\nIn the case of %s \"%s\"\nexpected:\n\t%+v\ngot:\n\t%+v",
					test.desc, test.input, test.expected, err)
			}
		}
	}
}

func Test_tokenToError(t *testing.T) {
	i := 0
	for tType := tokenType(i); i < int(tokenUnknown); i++ {
		tok := token{
			TokenType: tType,
			value:     tokenType(i).String(),
		}
		err := tokenToError(tok)
		if tok.TokenType == tokenError {
			if err.Error() != tokenError.String() {
				t.Errorf("Expected \"%s\", got \"%s\"", tokenError.String(), err.Error())
			}
		} else {
			msg := fmt.Sprintf("unexpected \"%s\"", tok.value)
			if err.Error() != msg {
				t.Errorf("Expected \"%s\", got \"%s\"", msg, err.Error())
			}
		}
	}
}

type fakeLexer struct {
	index  int
	tokens []token
}

func (l *fakeLexer) drain() {
}

func (l *fakeLexer) nextToken() token {
	t := l.tokens[l.index]
	l.index++
	return t
}

func (l *fakeLexer) reset() {
	l.index = 0
}

func Test_readLeftValue(t *testing.T) {
	lexer := &fakeLexer{}
	lexer.tokens = append(lexer.tokens, token{TokenType: tokenAssignment})
	lexer.tokens = append(lexer.tokens, token{TokenType: tokenError, value: "error"})

	parser := &parser{
		lex: lexer,
	}

	// When readRightValue fails readLeftValue should return an error.
	parser.rightValueReader = parser.readRightValue

	err := parser.readLeftValue(nil)
	if err.Error() != "error" {
		t.Errorf("Expected \"%s\", got \"%s\"", "error", err.Error())
	}

	// When readRightString fails readLeftValue should return an error.
	lexer.reset()
	parser.rightValueReader = parser.readRightString

	err = parser.readLeftValue(nil)
	if err.Error() != "error" {
		t.Errorf("Expected \"%s\", got \"%s\"", "error", err.Error())
	}
}
