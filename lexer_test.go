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

func newTestCase(desc, input string, expected []token) lexTestCase {
	return lexTestCase{
		desc:     desc,
		input:    input,
		expected: expected,
	}
}

func Test_Lex_Value_Succeeds(t *testing.T) {
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
		newTestCase("escaping a keys separator . in a key", "part1\\.part2=",
			[]token{
				newToken(tokenMapKey, 0, "part1.part2"),
				newToken(tokenAssignment, 12, "="),
				newToken(tokenEnd, 13, ""),
			}),
		newTestCase("escaping an assignment operator = in a key", "part1\\=part2=",
			[]token{
				newToken(tokenMapKey, 0, "part1=part2"),
				newToken(tokenAssignment, 12, "="),
				newToken(tokenEnd, 13, ""),
			}),
		newTestCase("escaping an open square bracket [ in a key", "part1\\[part2=",
			[]token{
				newToken(tokenMapKey, 0, "part1[part2"),
				newToken(tokenAssignment, 12, "="),
				newToken(tokenEnd, 13, ""),
			}),
		newTestCase("escaping a backslash \\ in a key", "part1\\\\part2=",
			[]token{
				newToken(tokenMapKey, 0, "part1\\part2"),
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
		result := testLex(test.input)
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
				newToken(tokenError, 0, "in position 0 got unexpected end, expecting a map key"),
			}),
		newTestCase("a map key with no value", "key",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenError, 3, "in position 3 got unexpected end, expecting '.', '=' or '['"),
			}),
		newTestCase("an unexpected key separator", ".",
			[]token{
				newToken(tokenError, 0, "in position 1 got unexpected character: U+002E '.', expecting a map key"),
			}),
		newTestCase("an unexpected assignment operator", "=",
			[]token{
				newToken(tokenError, 0, "in position 1 got unexpected character: U+003D '=', expecting a map key"),
			}),
		newTestCase("an unexpected end of array index", "k[",
			[]token{
				newToken(tokenMapKey, 0, "k"),
				newToken(tokenArrayIndexStart, 1, "["),
				newToken(tokenError, 2, "in position 2 got unexpected end, expecting an array index"),
			}),
		newTestCase("an unexpected open square bracket", "k[[",
			[]token{
				newToken(tokenMapKey, 0, "k"),
				newToken(tokenArrayIndexStart, 1, "["),
				newToken(tokenError, 2, "in position 3 got unexpected character: U+005B '[', expecting an array index"),
			}),
		newTestCase("an incomplete index", "key[0",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenArrayIndexStart, 3, "["),
				newToken(tokenArrayIndex, 4, "0"),
				newToken(tokenError, 5, "in position 5 got unexpected end, expecting ']'"),
			}),
		newTestCase("an array index with no value", "key[0]",
			[]token{
				newToken(tokenMapKey, 0, "key"),
				newToken(tokenArrayIndexStart, 3, "["),
				newToken(tokenArrayIndex, 4, "0"),
				newToken(tokenArrayIndexFinish, 5, "]"),
				newToken(tokenError, 6, "in position 6 got unexpected end, expecting '.', '=' or '['"),
			}),
		newTestCase("an unexpected key", "k1[0]k2",
			[]token{
				newToken(tokenMapKey, 0, "k1"),
				newToken(tokenArrayIndexStart, 2, "["),
				newToken(tokenArrayIndex, 3, "0"),
				newToken(tokenArrayIndexFinish, 4, "]"),
				newToken(tokenError, 5, "in position 6 got unexpected character: U+006B 'k', expecting '.', '=' or '['"),
			}),
		newTestCase("escaping unescapable in a key", "part1\\-part2=",
			[]token{
				newToken(tokenError, 0, "in position 7 got unknown escape sequence: character: U+002D '-'"),
			}),
	}

	for _, test := range testCases {
		result := testLex(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("\nIn the case of %s \"%s\"\nexpected:\n\t%+v\ngot:\n\t%+v",
				test.desc, test.input, test.expected, result)
		}
	}
}

func testLex(input string) (tokens []token) {
	lex := newLex(input)
	for {
		tok := lex.nextToken()
		tokens = append(tokens, tok)
		if tok.TokenType == tokenEnd || tok.TokenType == tokenError {
			break
		}
	}
	return
}

func Test_Lex_Drain(t *testing.T) {
	lex := &lex{
		input:  "foo=bar",
		tokens: make(chan token),
	}
	go lex.run()
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
