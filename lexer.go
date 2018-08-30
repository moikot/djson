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
	tokenError                             // An error
	tokenMapKey                            // A map key
	tokenMapKeySeparator                   // A map key separator '.'
	tokenArrayIndexStart                   // An array index start '['
	tokenArrayIndexFinish                  // An array index finish ']'
	tokenArrayIndex                        // An array index
	tokenAssignment                        // Assignment operator '='
	tokenValue                             // A value
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
		tokenUnknown:          "tokenUnknown",
	}
)

func (t tokenType) String() string {
	if str, ok := tokenStrings[t]; ok {
		return str
	}
	return tokenStrings[tokenUnknown]
}

type lexer interface {
	drain()
	nextToken() token
}

func newLex(input string) lexer {
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

func (l *lex) nextToken() token {
	return <-l.tokens
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
		value:     fmt.Sprintf("in position %d got %s", l.position, fmt.Sprintf(format, args...))}
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
	case !isStopChar(r, stopLeftValueChars):
		l.unread()
	default:
		return l.error("unexpected %v, expecting a map key", r)
	}
	err := l.scan(stopLeftValueChars)
	if err != nil {
		return l.error(err.Error())
	}
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
		return lexValue
	default:
		return l.error("unexpected %v, expecting '.', '=' or '['", ch)
	}
}

func lexArrayIndex(l *lex) stateFunction {
	switch ch := l.read(); {
	case isArrayIndexChar(ch):
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

func lexValue(l *lex) stateFunction {
	var valueLength = 0
	for r := l.read(); r != end; r = l.read() {
		valueLength++
	}
	if valueLength > 0 {
		l.unread()
		l.emit(tokenValue)
		l.read()
	}
	l.emit(tokenEnd)
	return nil
}

func (l *lex) scan(stopCharSet map[strRune]bool) error {
Loop:
	for {
		switch r := l.read(); {
		case r == end:
			break Loop
		case r == '\\':
			switch ch := l.peek(); {
			case isStopChar(ch, stopCharSet) || ch == '\\':
				l.skipLast()
				l.read()
			default:
				l.read()
				return fmt.Errorf("unknown escape sequence: %v", ch)
			}
		case !isStopChar(r, stopCharSet):
		default:
			break Loop
		}
	}
	l.unread()
	return nil
}

func (l *lex) scanArrayIndex() {
Loop:
	for {
		switch r := l.read(); {
		case isArrayIndexChar(r):
		default:
			break Loop
		}
	}
	l.unread()
}

var (
	stopLeftValueChars = map[strRune]bool{
		'=': true,
		'.': true,
		'[': true,
	}
)

func isStopChar(r strRune, stopCharSet map[strRune]bool) bool {
	_, ok := stopCharSet[r]
	return ok
}

func isArrayIndexChar(r strRune) bool {
	return unicode.IsNumber(rune(r))
}
