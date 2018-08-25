package djson

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type tokenType int

type token struct {
	TokenType tokenType // Token type
	position  int       // Starting position
	value     string    // Token value
}

func newToken(tokenType tokenType, position int, value string) token {
	return token{
		TokenType: tokenType,
		position:  position,
		value:     value,
	}
}

type lex struct {
	input    string     // The input string
	position int        // Current position in the input
	start    int        // Starting position of the current token
	width    int        // Width of the last rune read
	buffer   []rune     // Token buffer
	tokens   chan token // Channel of parsed tokens
}

type stateFunction func(*lex) stateFunction

type strRune rune

func (r strRune) String() string {
	if r == end {
		return "end"
	}
	return fmt.Sprintf("character: %#U", r)
}

var end = strRune(0)

const (
	tokenEnd              tokenType = iota // The end of a string
	tokenError                             // An error occurred
	tokenMapKey                            // A map key
	tokenMapKeySeparator                   // A map key separator '.'
	tokenArrayIndexStart                   // An index start '['
	tokenArrayIndexFinish                  // An index finish ']'
	tokenArrayIndex                        // An index
	tokenAssignment                        // Assignment operator '='
	tokenValue                             // A value
	tokenString                            // A string value
	tokenNextKey                           // A next key token ','
	tokenValueArrayStart                   // An array start '{'
	tokenValueArrayFinish                  // An array finish '}'
	tokenNextValue                         // A next value ','
	tokenUnknown                           // An unknown token, should be the last one
)

var (
	tokenStrings = map[tokenType]string{
		tokenEnd:              "tokenEnd",
		tokenError:            "tokenError",
		tokenMapKey:           "tokenMapKey",
		tokenMapKeySeparator:  "tokenMapKeySeparator",
		tokenArrayIndexStart:  "tokenArrayIndexStart",
		tokenArrayIndexFinish: "tokenArrayIndexFinish",
		tokenArrayIndex:       "tokenArrayIndex",
		tokenAssignment:       "tokenAssignment",
		tokenValue:            "tokenValue",
		tokenString:           "tokenString",
		tokenNextKey:          "tokenNextKey",
		tokenValueArrayStart:  "tokenValueArrayStart",
		tokenValueArrayFinish: "tokenValueArrayFinish",
		tokenNextValue:        "tokenNextValue",
		tokenUnknown:          "tokenUnknown",
	}
)

func (t tokenType) String() string {
	if str, ok := tokenStrings[t]; ok {
		return str
	}
	return tokenStrings[tokenUnknown]
}

func newLex(input string) *lex {
	l := &lex{
		input:  input,
		tokens: make(chan token),
	}
	go l.run()
	return l
}

func (l *lex) drain() {
	for range l.tokens {
	}
}

func (l *lex) read() strRune {
	if int(l.position) >= len(l.input) {
		l.width = 0
		return end
	}
	r, width := utf8.DecodeRuneInString(l.input[l.position:])
	l.width = width
	l.position += l.width
	l.buffer = append(l.buffer, r)
	return strRune(r)
}

func (l *lex) unread() {
	l.position -= l.width
	if l.width != 0 {
		// Remove from the buffer if not end
		l.buffer = l.buffer[:len(l.buffer)-1]
	}
}

func (l *lex) skipLast() {
	l.buffer = l.buffer[:len(l.buffer)-1]
}

func (l *lex) peek() strRune {
	r := l.read()
	l.unread()
	return r
}

func (l *lex) emit(tokenType tokenType) {
	l.tokens <- token{
		TokenType: tokenType,
		position:  l.start,
		value:     string(l.buffer),
	}
	l.start = l.position
	l.buffer = l.buffer[:0]
}

// Emmit an error token, value is the error text
func (l *lex) error(format string, args ...interface{}) stateFunction {
	l.tokens <- token{
		TokenType: tokenError,
		position:  l.start,
		value:     fmt.Sprintf(format, args...)}
	return nil
}

// The main lexing loop.
func (l *lex) run() {
	for state := lexMapKey; state != nil; {
		state = state(l)
	}
	close(l.tokens)
}

func lexMapKey(l *lex) stateFunction {
	switch r := l.read(); {
	case r == end:
		return l.error("unexpected %v, expecting a map key", r)
	case isLeftValueChar(r):
		l.unread()
	default:
		return l.error("unexpected %v, expecting a map key", r)
	}
	l.scanMapKey()
	l.emit(tokenMapKey)
	return lexLeftValue
}

func lexLeftValue(l *lex) stateFunction {
	switch ch := l.read(); {
	case ch == '.':
		l.emit(tokenMapKeySeparator)
		return lexMapKey
	case ch == '[':
		l.emit(tokenArrayIndexStart)
		return lexArrayIndex
	case ch == '=':
		l.emit(tokenAssignment)
		return lexRightValue
	default:
		return l.error("unexpected %v, expecting '.', '=' or '['", ch)
	}
}

func lexArrayIndex(l *lex) stateFunction {
	switch ch := l.read(); {
	case isMapKeyIndexChar(ch):
	default:
		return l.error("unexpected %v, expecting an array index", ch)
	}
	l.scanArrayIndex()
	l.emit(tokenArrayIndex)
	switch ch := l.read(); ch {
	case ']':
		l.emit(tokenArrayIndexFinish)
		return lexLeftValue
	default:
		return l.error("unexpected %v, expecting ']'", ch)
	}
}

func lexRightValue(l *lex) stateFunction {
	switch ch := l.read(); {
	case ch == end:
		l.emit(tokenEnd)
		return nil
	case ch == '{':
		l.emit(tokenValueArrayStart)
		return lexArrayValue
	case ch == ',':
		l.emit(tokenNextKey)
		return lexMapKey
	case isRightValueChar(ch) || ch == '\'':
		l.unread()
		return l.lexValue(lexNextMapKey)
	default:
		return l.error("unexpected %v, expecting '{', ',', a value or the end", ch)
	}
}

func lexArrayValue(l *lex) stateFunction {
	switch ch := l.read(); {
	case ch == '}':
		l.emit(tokenValueArrayFinish)
		return lexNextMapKey
	case ch == ',':
		l.emit(tokenNextValue)
		return lexArrayValue
	case isRightValueChar(ch) || ch == '\'':
		l.unread()
		return l.lexValue(lexNextArrayValue)
	default:
		return l.error("unexpected %v, expecting '}', ',' or a value", ch)
	}
}

func lexNextArrayValue(l *lex) stateFunction {
	switch ch := l.read(); {
	case ch == ',':
		l.emit(tokenNextValue)
		return lexArrayValue
	case ch == '}':
		l.emit(tokenValueArrayFinish)
		return lexNextMapKey
	default:
		return l.error("unexpected %v, expecting ',' or '}'", ch)
	}
}

func (l *lex) lexValue(nextLex stateFunction) stateFunction {
	switch ch := l.read(); {
	case ch == '\'':
		l.scanValue()
	case isRightValueChar(ch):
		l.scanValue()
		l.emit(tokenValue)
		return nextLex
	default:
		return l.error("unexpected %v, expecting a value or '''", ch)
	}

	switch ch := l.read(); {
	case ch == '\'':
		l.emit(tokenString)
		return nextLex
	default:
		return l.error("unterminated string, expected ''', got %v", ch)
	}
}

func lexNextMapKey(l *lex) stateFunction {
	switch ch := l.read(); {
	case ch == end:
		l.emit(tokenEnd)
		return nil
	case ch == ',':
		l.emit(tokenNextKey)
		return lexMapKey
	default:
		return l.error("unexpected %v, expecting ',' or the end", ch)
	}
}

func (l *lex) scanMapKey() {
Loop:
	for {
		switch r := l.read(); {
		case r == end:
			break Loop
		case r == '\\':
			switch ch := l.peek(); {
			case !isLeftValueChar(ch):
				l.skipLast()
				l.read()
			default:
				l.read()
			}
		case isLeftValueChar(r):
		default:
			break Loop
		}
	}
	l.unread()
}

func (l *lex) scanArrayIndex() {
Loop:
	for {
		switch r := l.read(); {
		case isMapKeyIndexChar(r):
		default:
			break Loop
		}
	}
	l.unread()
}

func (l *lex) scanValue() {
Loop:
	for {
		switch r := l.read(); {
		case r == end:
			break Loop
		case r == '\\':
			switch ch := l.peek(); {
			case !isRightValueChar(ch):
				l.skipLast()
				l.read()
			default:
				l.read()
			}
		case isRightValueChar(r):
		default:
			break Loop
		}
	}
	l.unread()
}

var (
	notLeftValueChars = map[strRune]bool{
		'=': true,
		'.': true,
		'[': true,
		']': true,
	}

	notRightValueChars = map[strRune]bool{
		',':  true,
		'{':  true,
		'}':  true,
		'\'': true,
	}
)

func isLeftValueChar(r strRune) bool {
	_, ok := notLeftValueChars[r]
	return !ok
}

func isRightValueChar(r strRune) bool {
	_, ok := notRightValueChars[r]
	return !ok
}

func isMapKeyIndexChar(r strRune) bool {
	return unicode.IsNumber(rune(r))
}
