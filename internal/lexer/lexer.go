// internal/lexer/lexer.go
// Package lexer implements the lexical analyzer for the Bee Programming Language,
// supporting UTF-8 decoding, Maximal Munch disambiguation, string interpolations,
// raw backtick strings, embedded markup blocks, and nested block comments.
package lexer

import (
	"bee/internal/token"
	"strings"
	"unicode/utf8"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           rune
	debug        bool
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	// Check for UTF-8 BOM (E0101)
	if l.ch == 0xFEFF {
		l.readChar()
	}
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		r, size := utf8.DecodeRuneInString(l.input[l.readPosition:])
		l.ch = r
		l.position = l.readPosition
		l.readPosition += size
	}
}

func (l *Lexer) PeekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.readPosition:])
	return r
}

func (l *Lexer) PeekN(n int) string {
	if l.readPosition+n > len(l.input) {
		return l.input[l.readPosition:]
	}
	return l.input[l.readPosition : l.readPosition+n]
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	var tok token.Token

	switch l.ch {
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '(':
		if l.PeekChar() == ':' {
			l.readChar()
			l.skipExprComment()
			return l.NextToken()
		}
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '?':
		tok = newToken(token.QUESTION, l.ch)
	case '$':
		tok = newToken(token.SIGIL_SYS, l.ch)
	case '#':
		tok = newToken(token.HASH, l.ch)
	case '@':
		tok = newToken(token.AT, l.ch)
	case '*':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MUL_ASSIGN, Literal: "*="}
		} else {
			tok = newToken(token.ASTERISK, l.ch)
		}
	case '_':
		tok = newToken(token.UNDERSCORE, l.ch)
	case 'λ':
		tok = newToken(token.LAMBDA, l.ch)
	case '∈':
		tok = newToken(token.IN, l.ch)
	case '∩':
		tok = newToken(token.SET_INTERSECT, l.ch)
	case '∪':
		tok = newToken(token.SET_UNION, l.ch)
	case '⊂':
		tok = newToken(token.SUBSET, l.ch)
	case '⊃':
		tok = newToken(token.SUPERSET, l.ch)
	case 'Δ':
		tok = newToken(token.SYM_DIFF, l.ch)
	case '∧':
		tok = newToken(token.LOGICAL_AND, l.ch)
	case '∨':
		tok = newToken(token.LOGICAL_OR, l.ch)
	case '¬':
		tok = newToken(token.LOGICAL_NOT, l.ch)
	case '⊕':
		tok = newToken(token.XOR_PLUS, l.ch)
	case '⊖':
		tok = newToken(token.XOR_MINUS, l.ch)
	case '∀':
		tok = newToken(token.FORALL, l.ch)
	case '∃':
		tok = newToken(token.EXISTS, l.ch)
	case '√':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.SQRT_ASSIGN, Literal: "√="}
		} else {
			tok = newToken(token.SQRT, l.ch)
		}
	case '±':
		tok = newToken(token.PLUS_MINUS, l.ch)
	case '×':
		tok = newToken(token.MULT, l.ch)
	case '÷':
		tok = newToken(token.DIV, l.ch)
	case '≈':
		tok = newToken(token.APPROX_EQ, l.ch)
	case '≡':
		tok = newToken(token.EQUIV, l.ch)
	case '≠':
		tok = newToken(token.NEQ_UNICODE, l.ch)
	case '≤':
		tok = newToken(token.LTE_UNICODE, l.ch)
	case '≥':
		tok = newToken(token.GTE_UNICODE, l.ch)
	case '=':
		if l.PeekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.FAT_ARROW, Literal: "=>"}
		} else {
			tok = newToken(token.EQ, l.ch)
		}
	case ':':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.ASSIGN, Literal: ":="}
		} else if l.PeekChar() == ':' {
			l.readChar()
			tok = token.Token{Type: token.CLONE_ASSIGN, Literal: "::"}
		} else {
			tok = newToken(token.COLON, l.ch)
		}
	case '.':
		if l.PeekChar() == '.' {
			l.readChar()
			tok = token.Token{Type: token.RANGE_INCL, Literal: ".."}
		} else if l.PeekChar() == '!' {
			l.readChar()
			tok = token.Token{Type: token.RANGE_LEFT_INC, Literal: ".!"}
		} else {
			tok = newToken(token.DOT, l.ch)
		}
	case '!':
		if l.PeekChar() == '.' {
			l.readChar()
			tok = token.Token{Type: token.RANGE_RGHT_INC, Literal: "!."}
		} else if l.PeekChar() == '!' {
			l.readChar()
			tok = token.Token{Type: token.RANGE_EXCL, Literal: "!!"}
		} else if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "!="}
		} else if l.PeekChar() == '∈' {
			l.readChar()
			tok = token.Token{Type: token.NOT_IN, Literal: "!∈"}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '+':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.PLUS_ASSIGN, Literal: "+="}
		} else if l.PeekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.REDUCE_CHANNEL, Literal: "+>"}
		} else if l.PeekChar() == '-' {
			l.readChar() // consume '-'
			l.skipBlockComment()
			return l.NextToken()
		} else {
			tok = newToken(token.PLUS, l.ch)
		}
	case '-':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MINUS_ASSIGN, Literal: "-="}
		} else if l.PeekChar() == '-' {
			l.skipSingleComment()
			return l.NextToken()
		} else if l.PeekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.THIN_ARROW, Literal: "->"}
		} else {
			tok = newToken(token.MINUS, l.ch)
		}
	case '/':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.DIV_ASSIGN, Literal: "/="}
		} else {
			tok = newToken(token.SLASH, l.ch)
		}
	case '\\':
		tok = newToken(token.BACKSLASH, l.ch)
	case '%':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MOD_ASSIGN, Literal: "%="}
		} else {
			tok = newToken(token.PERCENT, l.ch)
		}
	case '^':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.POW_ASSIGN, Literal: "^="}
		} else {
			tok = newToken(token.CARET, l.ch)
		}
	case '<':
		if l.PeekChar() == '-' {
			l.readChar()
			tok = token.Token{Type: token.PIPE_LEFT, Literal: "<<"}
		} else if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: "<="}
		} else if l.PeekChar() == ':' {
			l.readChar()
			tok = token.Token{Type: token.SUBTYPE, Literal: "<:"}
		} else if isLetter(l.PeekChar()) {
			// Check if markup block like <sql>, <html>, etc.
			if markupTok, ok := l.tryReadMarkupBlock(); ok {
				return markupTok
			}
			tok = newToken(token.LT, l.ch)
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: ">="}
		} else if l.PeekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.PIPE_RIGHT, Literal: ">>"}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		return tok
	case '`':
		tok.Type = token.RAW_STRING
		tok.Literal = l.readRawString()
		return tok
	default:
		// Check for block comment header `+-`
		if l.ch == '+' && l.PeekChar() == '-' {
			l.readChar()
			l.skipBlockComment()
			return l.NextToken()
		}

		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			return l.readNumberLiteral()
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

func (l *Lexer) skipSingleComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	l.readChar()
}

func (l *Lexer) skipExprComment() {
	l.readChar() // consume ':'
	depth := 1
	for depth > 0 && l.ch != 0 {
		if l.ch == '(' && l.PeekChar() == ':' {
			l.readChar()
			l.readChar()
			depth++
		} else if l.ch == ':' && l.PeekChar() == ')' {
			l.readChar()
			l.readChar()
			depth--
		} else {
			l.readChar()
		}
	}
}

func (l *Lexer) skipBlockComment() {
	// l.ch is already '-' after '+'
	l.readChar() // consume '-'
	depth := 1
	for depth > 0 && l.ch != 0 {
		if l.ch == '+' && l.PeekChar() == '-' {
			l.readChar()
			l.readChar()
			depth++
		} else if l.ch == '-' && l.PeekChar() == '+' {
			l.readChar()
			l.readChar()
			depth--
		} else {
			l.readChar()
		}
	}
	l.readChar() // consume past final '+'
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || (l.ch >= '₀' && l.ch <= '₉') {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumberLiteral() token.Token {
	position := l.position
	isReal := false

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.PeekChar()) {
		// Verify not part of range ..
		// Peek ahead past the digit to see if another dot follows
		isReal = true
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Exponent part e.g. 1e+10
	if l.ch == 'e' || l.ch == 'E' {
		isReal = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[position:l.position]
	if isReal {
		return token.Token{Type: token.REAL, Literal: literal}
	}
	return token.Token{Type: token.INT, Literal: literal}
}

func (l *Lexer) readString() string {
	var sb strings.Builder
	l.readChar() // consume opening quote

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteRune('\n')
			case 'r':
				sb.WriteRune('\r')
			case 't':
				sb.WriteRune('\t')
			case '\\':
				sb.WriteRune('\\')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			case '0':
				sb.WriteRune(0)
			default:
				sb.WriteRune(l.ch)
			}
		} else if l.ch == '#' && l.PeekChar() == '(' {
			// String interpolation #(expr)
			sb.WriteString("#(")
			l.readChar() // consume '('
			l.readChar()
			// Read until matching close paren
			parenDepth := 1
			for l.ch != 0 && parenDepth > 0 {
				if l.ch == '(' {
					parenDepth++
				} else if l.ch == ')' {
					parenDepth--
					if parenDepth == 0 {
						sb.WriteRune(')')
						l.readChar()
						continue
					}
				}
				sb.WriteRune(l.ch)
				l.readChar()
			}
			continue
		} else {
			sb.WriteRune(l.ch)
		}
		l.readChar()
	}
	l.readChar() // consume closing quote
	return sb.String()
}

func (l *Lexer) readRawString() string {
	position := l.position + 1
	l.readChar() // consume opening backtick
	for l.ch != '`' && l.ch != 0 {
		l.readChar()
	}
	res := l.input[position:l.position]
	l.readChar() // consume closing backtick
	return res
}

func (l *Lexer) tryReadMarkupBlock() (token.Token, bool) {
	// Look for <tag> ... </tag>
	// Save state to rollback if not a valid markup block
	savedPos := l.position
	savedReadPos := l.readPosition
	savedCh := l.ch

	l.readChar() // consume '<'
	tagName := l.readIdentifier()
	if tagName == "" {
		// Rollback
		l.position = savedPos
		l.readPosition = savedReadPos
		l.ch = savedCh
		return token.Token{}, false
	}

	// Skip attributes until '>'
	for l.ch != '>' && l.ch != 0 {
		l.readChar()
	}
	if l.ch != '>' {
		l.position = savedPos
		l.readPosition = savedReadPos
		l.ch = savedCh
		return token.Token{}, false
	}
	l.readChar() // consume '>'

	payloadStart := l.position
	closeTag := "</" + tagName + ">"

	// Find closing tag
	idx := strings.Index(l.input[l.position:], closeTag)
	if idx == -1 {
		l.position = savedPos
		l.readPosition = savedReadPos
		l.ch = savedCh
		return token.Token{}, false
	}

	payload := l.input[payloadStart : l.position+idx]
	// Advance lexer past closing tag
	consumedLen := (l.position + idx + len(closeTag)) - savedReadPos
	for i := 0; i < consumedLen; i++ {
		l.readChar()
	}

	return token.Token{
		Type:    token.MARKUP_BLOCK,
		Literal: "<" + tagName + ">" + payload + closeTag,
	}, true
}

func isLetter(ch rune) bool {
	return ('a' <= ch && ch <= 'z') ||
		('A' <= ch && ch <= 'Z') ||
		ch == '_' ||
		('α' <= ch && ch <= 'ω') ||
		('Α' <= ch && ch <= 'Ω') ||
		('а' <= ch && ch <= 'я') ||
		('А' <= ch && ch <= 'Я') ||
		ch > 127
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
