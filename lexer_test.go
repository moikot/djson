package djson

import (
	"reflect"
	"testing"
)

type lexTestCase struct {
	desc     string  // Description
	input    string  // Input string
	expected []token // The expected expected
}

func newTestCase(desc, input string, expectd []token) lexTestCase {
	return lexTestCase{
		desc:     desc,
		input:    input,
		expected: expectd,
	}
}

func Test_Lex_Succeeds(t *testing.T) {
	testCases := []lexTestCase{
		newTestCase("a whitespace key", " =",
			[]token{
				newToken(tokenMapKey, 0, " "),
				newToken(tokenAssignment, 1, "="),
				newToken(tokenEnd, 2, ""),
			}),
		newTestCase("a simple key", "key=",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenEnd, 4, ""),
			}),
		newTestCase("a key after key", "key1.key2=",
			[]token{
				newToken(tokenMapKey, 0, "key1"),
				newToken(tokenMapKeySeparator, 4, "."),
				newToken(tokenMapKey, 5, "key2"),
				newToken(tokenAssignment, 9, "="),
				newToken(tokenEnd, 10, ""),
			}),
		newTestCase("a key with whitespace", "part1 part2=",
			[]token{
				newToken(tokenMapKey, 0, "part1 part2"),
				newToken(tokenAssignment, 11, "="),
				newToken(tokenEnd, 12, ""),
			}),
		newTestCase("escaping keys separator .", "part1\\.part2=",
			[]token{
				newToken(tokenMapKey, 0, "part1.part2"),
				newToken(tokenAssignment, 12, "="),
				newToken(tokenEnd, 13, ""),
			}),
		newTestCase("escaping assignment operator =", "part1\\=part2=",
			[]token{
				newToken(tokenMapKey, 0, "part1=part2"),
				newToken(tokenAssignment, 12, "="),
				newToken(tokenEnd, 13, ""),
			}),
		newTestCase("escaping open square bracket [", "part1\\[part2=",
			[]token{
				newToken(tokenMapKey, 0, "part1[part2"),
				newToken(tokenAssignment, 12, "="),
				newToken(tokenEnd, 13, ""),
			}),
		newTestCase("escaping unescapable in a key", "part1\\-part2=",
			[]token{
				newToken(tokenMapKey, 0, "part1\\-part2"),
				newToken(tokenAssignment, 12, "="),
				newToken(tokenEnd, 13, ""),
			}),
		newTestCase("a simple key value assignment", "key=value",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValue, 4, "value"),
				newToken(tokenEnd, 9, ""),
			}),
		newTestCase("a simple key string assignment", "key='value'",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenString, 4, "'value'"),
				newToken(tokenEnd, 11, ""),
			}),
		newTestCase("a string can contain a comma", "key='val1,val2'",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenString, 4, "'val1,val2'"),
				newToken(tokenEnd, 15, ""),
			}),
		newTestCase("a string containing escaped single quote", "key='val1\\'val2'",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenString, 4, "'val1'val2'"),
				newToken(tokenEnd, 16, ""),
			}),
		newTestCase("escaping unescapable in a string", "key='val1\\-val2'",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenString, 4, "'val1\\-val2'"),
				newToken(tokenEnd, 16, ""),
			}),
		newTestCase("escaping comma in a value", "key=part1\\,part2",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValue, 4, "part1,part2"),
				newToken(tokenEnd, 16, ""),
			}),
		newTestCase("escaping open curly bracket in a value", "key=part1\\{part2",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValue, 4, "part1{part2"),
				newToken(tokenEnd, 16, ""),
			}),
		newTestCase("escaping close curly bracket in a value", "key=part1\\}part2",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValue, 4, "part1}part2"),
				newToken(tokenEnd, 16, ""),
			}),
		newTestCase("escaping unescapable in a value", "key=part1\\-part2",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValue, 4, "part1\\-part2"),
				newToken(tokenEnd, 16, ""),
			}),
		newTestCase("verbatim string with single quotes", "key=@'val'",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenVerbatimString, 4, "'val'"),
				newToken(tokenEnd, 10, ""),
			}),
		newTestCase("verbatim string with commas", "key=@,val,",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenVerbatimString, 4, ",val,"),
				newToken(tokenEnd, 10, ""),
			}),
		newTestCase("verbatim string with curly brackets", "key=@{val}",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenVerbatimString, 4, "{val}"),
				newToken(tokenEnd, 10, ""),
			}),
		newTestCase("two key-value pairs", "key1=val1,key2=val2",
			[]token{
				newToken(tokenMapKey, 0, "key1"),
				newToken(tokenAssignment, 4, "="),
				newToken(tokenValue, 5, "val1"),
				newToken(tokenNextKey, 9, ","),
				newToken(tokenMapKey, 10, "key2"),
				newToken(tokenAssignment, 14, "="),
				newToken(tokenValue, 15, "val2"),
				newToken(tokenEnd, 19, ""),
			}),
		newTestCase("an empty array value", "key={}",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValueArrayStart, 4, "{"),
				newToken(tokenValueArrayFinish, 5, "}"),
				newToken(tokenEnd, 6, ""),
			}),
		newTestCase("an array with two empty elements", "key={,}",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValueArrayStart, 4, "{"),
				newToken(tokenNextValue, 5, ","),
				newToken(tokenValueArrayFinish, 6, "}"),
				newToken(tokenEnd, 7, ""),
			}),
		newTestCase("an array value with one element", "key={v1}",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValueArrayStart, 4, "{"),
				newToken(tokenValue, 5, "v1"),
				newToken(tokenValueArrayFinish, 7, "}"),
				newToken(tokenEnd, 8, ""),
			}),
		newTestCase("an array value with two elements", "key={v1,v2}",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValueArrayStart, 4, "{"),
				newToken(tokenValue, 5, "v1"),
				newToken(tokenNextValue, 7, ","),
				newToken(tokenValue, 8, "v2"),
				newToken(tokenValueArrayFinish, 10, "}"),
				newToken(tokenEnd, 11, ""),
			}),
		newTestCase("an array index", "key[10]=v",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenArrayIndexStart, 3, "["),
				newToken(tokenArrayIndex, 4, "10"),
				newToken(tokenArrayIndexFinish, 6, "]"),
				newToken(tokenAssignment, 7, "="),
				newToken(tokenValue, 8, "v"),
				newToken(tokenEnd, 9, ""),
			}),
		newTestCase("two array indexes", "key[0][1]=v",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenArrayIndexStart, 3, "["),
				newToken(tokenArrayIndex, 4, "0"),
				newToken(tokenArrayIndexFinish, 5, "]"),
				newToken(tokenArrayIndexStart, 6, "["),
				newToken(tokenArrayIndex, 7, "1"),
				newToken(tokenArrayIndexFinish, 8, "]"),
				newToken(tokenAssignment, 9, "="),
				newToken(tokenValue, 10, "v"),
				newToken(tokenEnd, 11, ""),
			}),
		newTestCase("a key after an array index", "key1[0].key2=v",
			[]token{
				newToken(tokenMapKey, 0, "key1"),
				newToken(tokenArrayIndexStart, 4, "["),
				newToken(tokenArrayIndex, 5, "0"),
				newToken(tokenArrayIndexFinish, 6, "]"),
				newToken(tokenMapKeySeparator, 7, "."),
				newToken(tokenMapKey, 8, "key2"),
				newToken(tokenAssignment, 12, "="),
				newToken(tokenValue, 13, "v"),
				newToken(tokenEnd, 14, ""),
			}),
	}

	for _, test := range testCases {
		result := testParse(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("\nIn the case of %s \"%s\"\nexpected:\n\t%+v\ngot:\n\t%+v",
				test.desc, test.input, test.expected, result)
		}
	}
}

func Test_Lex_Fails(t *testing.T) {
	testCases := []lexTestCase{
		newTestCase("an empty string", "",
			[]token{
				newToken(tokenError, 0, "unexpected end, expecting a map key"),
			}),
		newTestCase("a map key with no value", "key",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenError, 3, "unexpected end, expecting '.', '=' or '['"),
			}),
		newTestCase("an unterminated string", "key='",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenError, 4, "unterminated string, expected ''', got end"),
			}),
		newTestCase("an empty value followed by a missing key", "key=,",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenNextKey, 4, ","),
				newToken(tokenError, 5, "unexpected end, expecting a map key"),
			}),
		newTestCase("a value followed by a missing key", "key=v,",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenAssignment, 3, "="),
				newToken(tokenValue, 4, "v"),
				newToken(tokenNextKey, 5, ","),
				newToken(tokenError, 6, "unexpected end, expecting a map key"),
			}),
		newTestCase("an unexpected key separator", ".",
			[]token{
				newToken(tokenError, 0, "unexpected character: U+002E '.', expecting a map key"),
			}),
		newTestCase("an unexpected assignment operator", "=",
			[]token{
				newToken(tokenError, 0, "unexpected character: U+003D '=', expecting a map key"),
			}),
		newTestCase("a value followed by an unexpected assignment", "k=v,=",
			[]token{
				newToken(tokenMapKey, 0, "k"),
				newToken(tokenAssignment, 1, "="),
				newToken(tokenValue, 2, "v"),
				newToken(tokenNextKey, 3, ","),
				newToken(tokenError, 4, "unexpected character: U+003D '=', expecting a map key"),
			}),
		newTestCase("an unexpected end of array index", "k[",
			[]token{
				newToken(tokenMapKey, 0, "k"),
				newToken(tokenArrayIndexStart, 1, "["),
				newToken(tokenError, 2, "unexpected end, expecting an array index"),
			}),
		newTestCase("an unexpected open square bracket", "k[[",
			[]token{
				newToken(tokenMapKey, 0, "k"),
				newToken(tokenArrayIndexStart, 1, "["),
				newToken(tokenError, 2, "unexpected character: U+005B '[', expecting an array index"),
			}),
		newTestCase("an incomplete index", "key[0",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenArrayIndexStart, 3, "["),
				newToken(tokenArrayIndex, 4, "0"),
				newToken(tokenError, 5, "unexpected end, expecting ']'"),
			}),
		newTestCase("an array index with no value", "key[0]",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenArrayIndexStart, 3, "["),
				newToken(tokenArrayIndex, 4, "0"),
				newToken(tokenArrayIndexFinish, 5, "]"),
				newToken(tokenError, 6, "unexpected end, expecting '.', '=' or '['"),
			}),
		newTestCase("an unexpected key", "k1[0]k2",
			[]token{
				newToken(tokenMapKey, 0, "k1"),
				newToken(tokenArrayIndexStart, 2, "["),
				newToken(tokenArrayIndex, 3, "0"),
				newToken(tokenArrayIndexFinish, 4, "]"),
				newToken(tokenError, 5, "unexpected character: U+006B 'k', expecting '.', '=' or '['"),
			}),
		newTestCase("an unexpected open curly bracket", "k={{",
			[]token{
				newToken(tokenMapKey, 0, "k"),
				newToken(tokenAssignment, 1, "="),
				newToken(tokenValueArrayStart, 2, "{"),
				newToken(tokenError, 3, "unexpected character: U+007B '{', expecting '}', ',' or a value"),
			}),
		newTestCase("an unexpected close curly bracket at a value end", "k={v}}",
			[]token{
				newToken(tokenMapKey, 0, "k"),
				newToken(tokenAssignment, 1, "="),
				newToken(tokenValueArrayStart, 2, "{"),
				newToken(tokenValue, 3, "v"),
				newToken(tokenValueArrayFinish, 4, "}"),
				newToken(tokenError, 5, "unexpected character: U+007D '}', expecting ',' or the end"),
			}),
		newTestCase("an unexpected close curly bracket at a value start", "k=}",
			[]token{
				newToken(tokenMapKey, 0, "k"),
				newToken(tokenAssignment, 1, "="),
				newToken(tokenError, 2, "unexpected character: U+007D '}', expecting '{', ',', a value or the end"),
			}),
		newTestCase("an unexpected open curly bracket after an array value", "k={v{",
			[]token{
				newToken(tokenMapKey, 0, "k"),
				newToken(tokenAssignment, 1, "="),
				newToken(tokenValueArrayStart, 2, "{"),
				newToken(tokenValue, 3, "v"),
				newToken(tokenError, 4, "unexpected character: U+007B '{', expecting ',' or '}'"),
			}),
	}

	for _, test := range testCases {
		result := testParse(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("\nIn the case of %s \"%s\"\nexpected:\n\t%+v\ngot:\n\t%+v",
				test.desc, test.input, test.expected, result)
		}
	}
}

func testParse(input string) (tokens []token) {
	lex := newLex(input)
	for {
		tok := <-lex.tokens
		tokens = append(tokens, tok)
		if tok.TokenType == tokenEnd || tok.TokenType == tokenError {
			break
		}
	}
	return
}

func Test_Lex_Drain(t *testing.T) {
	lex := newLex("foo=bar")
	lex.drain()
	for range lex.tokens {
		t.Errorf("lex.drain does not drain the tokens")
		break
	}
}

// Every token type should have a string representation.
// It is needed for producing readable error messages
// in case of a test failure.
func Test_tokenType_String(t *testing.T) {
	i := 0
	for tType := tokenType(i); i < int(tokenUnknown); i++ {
		str := tType.String()
		if str == tokenStrings[tokenUnknown] && tokenType(i) != tokenUnknown {
			t.Errorf("String conversion for token #%d is not defined", i+1)
		}
	}

	const tokenWithoutString tokenType = -1
	str := tokenWithoutString.String()
	if str != tokenStrings[tokenUnknown] {
		t.Errorf("Expected string %s for a token with no string, got %s", tokenStrings[tokenUnknown], str)
	}
}
