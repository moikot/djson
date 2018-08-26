package djson

import (
	"fmt"
	"reflect"
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

func Test_Parser_Succeeds(t *testing.T) {
	testCases := []parserTestCase{
		newParserTestCase(
			"an simple key value pair", "key=val",
			map[string]interface{}{
				"key": "val",
			},
		),
		newParserTestCase(
			"an boolean value", "key=true",
			map[string]interface{}{
				"key": true,
			},
		),
		newParserTestCase(
			"a string value", "key='true'",
			map[string]interface{}{
				"key": "true",
			},
		),
		newParserTestCase(
			"a verbatim string", "key=@{'true'},",
			map[string]interface{}{
				"key": "{'true'},",
			},
		),
		newParserTestCase(
			"an integer value", "key=1000",
			map[string]interface{}{
				"key": int64(1000),
			},
		),
		newParserTestCase(
			"a floating point value", "key=10.01",
			map[string]interface{}{
				"key": float64(10.01),
			},
		),
		newParserTestCase(
			"a value array", "key={val1,true,'true',1000,10.01}",
			map[string]interface{}{
				"key": []interface{}{
					"val1",
					true,
					"true",
					int64(1000),
					float64(10.01),
				},
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
			"an array with missing elements", "key={val1,,val2}",
			map[string]interface{}{
				"key": []interface{}{
					"val1",
					"",
					"val2",
				},
			},
		),
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
			"a value overridden by a null value", "foo[0]={val1},foo[0]=null",
			map[string]interface{}{
				"foo": []interface{}{
					nil,
				},
			},
		),
	}

	for _, test := range testCases {
		m := map[string]interface{}{}
		err := Unmarshal(test.input, m)
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

func Test_Parser_Accumulates_To_Input_Map(t *testing.T) {
	m := map[string]interface{}{}
	err := Unmarshal("key1=val1", m)
	if err != nil {
		t.Error(err)
	}
	err = Unmarshal("key2=val2", m)
	if err != nil {
		t.Error(err)
	}

	expected := map[string]interface{}{
		"key1": "val1",
		"key2": "val2",
	}

	if !reflect.DeepEqual(m, expected) {
		t.Errorf("expected:\n\t%+v\ngot:\n\t%+v", expected, m)
	}
}

func Test_Parser_Fails(t *testing.T) {
	testCases := []parserErrorTestCase{
		newParserErrorTestCase(
			"an empty string", "",
			"unexpected end, expecting a map key",
		),
		newParserErrorTestCase(
			"key with no value", "foo",
			"unexpected end, expecting '.', '=' or '['",
		),
		newParserErrorTestCase(
			"an array with no value", "foo[0]",
			"unexpected end, expecting '.', '=' or '['",
		),
		newParserErrorTestCase(
			"an array index is not defined", "foo[",
			"unexpected end, expecting an array index",
		),
		newParserErrorTestCase(
			"an array index is not complete", "foo[0",
			"unexpected end, expecting ']'",
		),
		newParserErrorTestCase(
			"an array index is out of range", "foo[99999999999999999999]",
			"strconv.Atoi: parsing \"99999999999999999999\": value out of range",
		),
		newParserErrorTestCase(
			"a value is malformed", "foo=}",
			"unexpected character: U+007D '}', expecting '{', ',', a value or the end",
		),
		newParserErrorTestCase(
			"an array value is malformed", "foo={{",
			"unexpected character: U+007B '{', expecting '}', ',' or a value",
		),
		newParserErrorTestCase(
			"next key value pair is malformed", "foo={val}{",
			"unexpected character: U+007B '{', expecting ',' or the end",
		),
	}

	for _, test := range testCases {
		m := map[string]interface{}{}
		err := Unmarshal(test.input, m)
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
