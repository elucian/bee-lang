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
	line         int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
	l.readChar()
	// Check for UTF-8 BOM (E0101)
	if l.ch == 0xFEFF {
		l.readChar()
	}
	return l
}

func (l *Lexer) readChar() {
	if l.ch == '\n' {
		l.line++
	}
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
		tok = token.Token{Type: token.SEMICOLON, Literal: ";", Pos: token.Pos(l.line)}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: ",", Pos: token.Pos(l.line)}
	case '(':
		if l.PeekChar() == ':' {
			l.readChar()
			l.skipExprComment()
			return l.NextToken()
		}
		tok = token.Token{Type: token.LPAREN, Literal: "(", Pos: token.Pos(l.line)}
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: ")", Pos: token.Pos(l.line)}
	case '[':
		tok = token.Token{Type: token.LBRACKET, Literal: "[", Pos: token.Pos(l.line)}
	case ']':
		tok = token.Token{Type: token.RBRACKET, Literal: "]", Pos: token.Pos(l.line)}
	case '{':
		tok = token.Token{Type: token.LBRACE, Literal: "{", Pos: token.Pos(l.line)}
	case '}':
		tok = token.Token{Type: token.RBRACE, Literal: "}", Pos: token.Pos(l.line)}
	case '?':
		tok = token.Token{Type: token.QUESTION, Literal: "?", Pos: token.Pos(l.line)}
	case '$':
		tok = token.Token{Type: token.SIGIL_SYS, Literal: "$", Pos: token.Pos(l.line)}
	case '#':
		tok = token.Token{Type: token.HASH, Literal: "#", Pos: token.Pos(l.line)}
	case '@':
		tok = token.Token{Type: token.AT, Literal: "@", Pos: token.Pos(l.line)}
	case '*':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MUL_ASSIGN, Literal: "*=", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.ASTERISK, Literal: "*", Pos: token.Pos(l.line)}
		}
	case '_':
		tok = token.Token{Type: token.UNDERSCORE, Literal: "_", Pos: token.Pos(l.line)}
	case 'λ':
		tok = token.Token{Type: token.LAMBDA, Literal: "λ", Pos: token.Pos(l.line)}
	case '∈':
		tok = token.Token{Type: token.IN, Literal: "∈", Pos: token.Pos(l.line)}
	case '∩':
		tok = token.Token{Type: token.SET_INTERSECT, Literal: "∩", Pos: token.Pos(l.line)}
	case '∪':
		tok = token.Token{Type: token.SET_UNION, Literal: "∪", Pos: token.Pos(l.line)}
	case '⊂':
		tok = token.Token{Type: token.SUBSET, Literal: "⊂", Pos: token.Pos(l.line)}
	case '⊃':
		tok = token.Token{Type: token.SUPERSET, Literal: "⊃", Pos: token.Pos(l.line)}
	case 'Δ':
		tok = token.Token{Type: token.SYM_DIFF, Literal: "Δ", Pos: token.Pos(l.line)}
	case '∧':
		tok = token.Token{Type: token.LOGICAL_AND, Literal: "∧", Pos: token.Pos(l.line)}
	case '∨':
		tok = token.Token{Type: token.LOGICAL_OR, Literal: "∨", Pos: token.Pos(l.line)}
	case '¬':
		tok = token.Token{Type: token.LOGICAL_NOT, Literal: "¬", Pos: token.Pos(l.line)}
	case '⊕':
		tok = token.Token{Type: token.XOR_PLUS, Literal: "⊕", Pos: token.Pos(l.line)}
	case '⊖':
		tok = token.Token{Type: token.XOR_MINUS, Literal: "⊖", Pos: token.Pos(l.line)}
	case '∀':
		tok = token.Token{Type: token.FORALL, Literal: "∀", Pos: token.Pos(l.line)}
	case '∃':
		tok = token.Token{Type: token.EXISTS, Literal: "∃", Pos: token.Pos(l.line)}
	case '⁰', '¹', '²', '³', '⁴', '⁵', '⁶', '⁷', '⁸', '⁹':
		var sb strings.Builder
		sb.WriteRune(l.ch)
		for {
			l.readChar()
			if l.ch >= '⁰' && l.ch <= '⁹' {
				sb.WriteRune(l.ch)
			} else {
				break
			}
		}
		if l.ch == '√' {
			sb.WriteRune('√')
			tok = token.Token{Type: token.SQRT, Literal: sb.String(), Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.CARET, Literal: sb.String(), Pos: token.Pos(l.line)}
			// Do NOT return here immediately so l.readChar() at the end of switch advances properly!
		}
	case '√':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.SQRT_ASSIGN, Literal: "√=", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.SQRT, Literal: "√", Pos: token.Pos(l.line)}
		}
	case '×':
		tok = token.Token{Type: token.MULT, Literal: "×", Pos: token.Pos(l.line)}
	case '÷':
		tok = token.Token{Type: token.DIV, Literal: "÷", Pos: token.Pos(l.line)}
	case '≈':
		tok = token.Token{Type: token.APPROX_EQ, Literal: "≈", Pos: token.Pos(l.line)}
	case '≡':
		tok = token.Token{Type: token.EQUIV, Literal: "≡", Pos: token.Pos(l.line)}
	case '≠':
		tok = token.Token{Type: token.NEQ_UNICODE, Literal: "≠", Pos: token.Pos(l.line)}
	case '≤':
		tok = token.Token{Type: token.LTE_UNICODE, Literal: "≤", Pos: token.Pos(l.line)}
	case '≥':
		tok = token.Token{Type: token.GTE_UNICODE, Literal: "≥", Pos: token.Pos(l.line)}
	case '=':
		if l.PeekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.FAT_ARROW, Literal: "=>", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.EQ, Literal: "=", Pos: token.Pos(l.line)}
		}
	case ':':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.ASSIGN, Literal: ":=", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == ':' {
			l.readChar()
			tok = token.Token{Type: token.CLONE_ASSIGN, Literal: "::", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.COLON, Literal: ":", Pos: token.Pos(l.line)}
		}
	case '.':
		if l.PeekChar() == '.' {
			l.readChar()
			tok = token.Token{Type: token.RANGE_INCL, Literal: "..", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '!' {
			l.readChar()
			tok = token.Token{Type: token.RANGE_LEFT_INC, Literal: ".!", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.DOT, Literal: ".", Pos: token.Pos(l.line)}
		}
	case '!':
		if l.PeekChar() == '.' {
			l.readChar()
			tok = token.Token{Type: token.RANGE_RGHT_INC, Literal: "!.", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '!' {
			l.readChar()
			tok = token.Token{Type: token.RANGE_EXCL, Literal: "!!", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "!=", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '∈' {
			l.readChar()
			tok = token.Token{Type: token.NOT_IN, Literal: "!∈", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.BANG, Literal: "!", Pos: token.Pos(l.line)}
		}
	case '+':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.PLUS_ASSIGN, Literal: "+=", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.REDUCE_CHANNEL, Literal: "+>", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '-' {
			l.readChar() // consume '-'
			l.skipBlockComment()
			return l.NextToken()
		} else {
			tok = token.Token{Type: token.PLUS, Literal: "+", Pos: token.Pos(l.line)}
		}
	case '-':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MINUS_ASSIGN, Literal: "-=", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '-' {
			l.skipSingleComment()
			return l.NextToken()
		} else if l.PeekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.THIN_ARROW, Literal: "->", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.MINUS, Literal: "-", Pos: token.Pos(l.line)}
		}
	case '/':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.DIV_ASSIGN, Literal: "/=", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.SLASH, Literal: "/", Pos: token.Pos(l.line)}
		}
	case '%':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MOD_ASSIGN, Literal: "%=", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.PERCENT, Literal: "%", Pos: token.Pos(l.line)}
		}
	case '^':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.POW_ASSIGN, Literal: "^=", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.CARET, Literal: "^", Pos: token.Pos(l.line)}
		}
	case '<':
		if l.PeekChar() == '-' {
			l.readChar()
			tok = token.Token{Type: token.THIN_ARROW, Literal: "<-", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == ':' {
			l.readChar()
			tok = token.Token{Type: token.SUBTYPE, Literal: "<:", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '<' {
			l.readChar()
			tok = token.Token{Type: token.PIPE_LEFT, Literal: "<<", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: "<=", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.LT, Literal: "<", Pos: token.Pos(l.line)}
		}
	case '>':
		if l.PeekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: ">=", Pos: token.Pos(l.line)}
		} else if l.PeekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.PIPE_RIGHT, Literal: ">>", Pos: token.Pos(l.line)}
		} else {
			tok = token.Token{Type: token.GT, Literal: ">", Pos: token.Pos(l.line)}
		}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Pos = token.Pos(l.line)
		return tok
	case '`':
		tok.Type = token.RAW_STRING
		tok.Literal = l.readRawString()
		tok.Pos = token.Pos(l.line)
		return tok
	default:
		if l.ch == '+' && l.PeekChar() == '-' {
			l.readChar()
			l.skipBlockComment()
			return l.NextToken()
		}

		if isLetter(l.ch) {
			lit := l.readIdentifier()
			tok = token.Token{Type: token.LookupIdent(lit), Literal: lit, Pos: token.Pos(l.line)}
			return tok
		} else if isDigit(l.ch) {
			numTok := l.readNumberLiteral()
			numTok.Pos = token.Pos(l.line)
			return numTok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Pos: token.Pos(l.line)}
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType token.Type, ch rune, line int) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Pos: token.Pos(line)}
}

func (l *Lexer) tok(tokenType token.Type, literal string) token.Token {
	return token.Token{Type: tokenType, Literal: literal, Pos: token.Pos(l.line)}
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
		(ch >= '⁰' && ch <= '⁹') ||
		ch > 127
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
