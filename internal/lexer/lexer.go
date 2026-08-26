/*
Package lexer provides the lexical analysis engine for the Bee language.
Responsibility: Transforms source code into a token stream.
Strategy: Uses a single-pass scanner with Maximal Munch disambiguation for operators.
*/
package lexer

import (
	"bee/internal/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           rune
	debug        bool
}

// New initializes the Lexer with input and starts the read head.
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// readChar reads the next character from input and advances the position.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.readPosition])
	}
	l.position = l.readPosition
	l.readPosition++
}

// peekChar returns the next character without advancing the lexer.
func (l *Lexer) PeekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return rune(l.input[l.readPosition])
}

// NextToken scans the input and returns the next valid token.
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	var tok token.Token

	switch l.ch {
	case '=':
		tok = newToken(token.EQ, l.ch)
	case ':':
		if l.PeekChar() == '=' {
			tok = token.Token{Type: token.ASSIGN, Literal: ":="}
			l.readChar()
		} else {
			tok = newToken(token.COLON, l.ch)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		return tok
	case '-':
		if l.PeekChar() == '-' {
			l.skipComment()
			return l.NextToken()
		}
		tok = newToken(token.ILLEGAL, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType token.Type, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	l.readChar()
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	res := l.input[position:l.position]
	l.readChar()
	return res
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch > 127
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
