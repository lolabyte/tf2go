package lexer

import (
	"github.com/lolabyte/tf2go/terraform/token"
)

type Lexer struct {
	input        string
	currPosition int  // current position in the input (current char)
	readPosition int  // current reading position in the input (after current char)
	ch           byte // current char
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) peek() byte {
	return l.input[l.readPosition]
}

func (l *Lexer) isAtEnd() bool {
	return l.currPosition == len(l.input)-1
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	l.currPosition = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) readChars(n int) {
	for i := 0; i < n; i++ {
		l.readChar()
	}
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case '(':
		tok = newToken(token.LEFT_PAREN, l.ch)
	case ')':
		tok = newToken(token.RIGHT_PAREN, l.ch)
	case '[':
		tok = newToken(token.LEFT_SQUARE_BRACE, l.ch)
	case ']':
		tok = newToken(token.RIGHT_SQUARE_BRACE, l.ch)
	case '{':
		tok = newToken(token.LEFT_CURLY_BRACE, l.ch)
	case '}':
		tok = newToken(token.RIGHT_CURLY_BRACE, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '"', '\'':
		tok.Type = token.STRING
		tok.Literal = l.readString(l.ch)
		return tok
	// TODO: Add support for comments
	//case '/', '#':
	//	tok.Type = token.COMMENT
	//	tok.Literal = l.readComment()
	case 0:
		tok.Type = token.EOF
		tok.Literal = ""
		return tok
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.NUMBER
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readNumber() string {
	start := l.currPosition
	// TODO: add support for fractional numbers (e.g 3.14)
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.currPosition]
}

func (l *Lexer) readString(quoteCh byte) string {
	start := l.currPosition
	for l.peek() != quoteCh && !l.isAtEnd() {
		l.readChar()
	}
	s := l.input[start+1 : l.currPosition-1]
	// advance to consume the terminating quote
	l.readChars(2)
	return s
}

func (l *Lexer) readIdentifier() string {
	start := l.currPosition
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[start:l.currPosition]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
