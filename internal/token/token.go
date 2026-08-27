// internal/token/token.go
// Package token defines constants, structures, and keyword mapping for lexical tokens in the Bee Programming Language.
package token

type Type string

type Pos int

type Token struct {
	Type    Type
	Literal string
	Pos     Pos
}

const (
	// Special tokens
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers & Literals
	IDENT        = "IDENT"
	INT          = "INT"
	REAL         = "REAL"
	STRING       = "STRING"
	RAW_STRING   = "RAW_STRING"
	MARKUP_BLOCK = "MARKUP_BLOCK"
	ANGLE        = "ANGLE"

	// Assignment operators
	ASSIGN       = ":="
	CLONE_ASSIGN = "::"
	PLUS_ASSIGN  = "+="
	MINUS_ASSIGN = "-="
	MUL_ASSIGN   = "*="
	DIV_ASSIGN   = "/="
	MOD_ASSIGN   = "%="
	POW_ASSIGN   = "^="
	SQRT_ASSIGN  = "√="

	// Arithmetic operators
	PLUS       = "+"
	MINUS      = "-"
	ASTERISK   = "*"
	SLASH      = "/"
	BACKSLASH  = `\`
	PERCENT    = "%"
	CARET      = "^"
	SQRT       = "√"
	PLUS_MINUS = "±"
	MULT       = "×"
	DIV        = "÷"

	// Relational / Equality operators
	EQ          = "="
	NOT_EQ      = "!="
	NEQ_UNICODE = "≠"
	LT          = "<"
	LTE         = "<="
	LTE_UNICODE = "≤"
	GT          = ">"
	GTE         = ">="
	GTE_UNICODE = "≥"
	APPROX_EQ   = "≈"
	EQUIV       = "≡"

	// Subtyping & Type Casting
	SUBTYPE   = "<:"
	TYPE_CAST = ":>"

	// Logical & Set Algebra operators
	IN            = "∈"
	NOT_IN        = "!∈"
	SET_INTERSECT = "∩"
	SET_UNION     = "∪"
	SUBSET        = "⊂"
	SUPERSET      = "⊃"
	SYM_DIFF      = "Δ"
	LOGICAL_AND   = "∧"
	LOGICAL_OR    = "∨"
	LOGICAL_NOT   = "¬"
	XOR_PLUS      = "⊕"
	XOR_MINUS     = "⊖"

	// Quantifiers
	FORALL = "∀"
	EXISTS = "∃"

	// Functional & Concurrency operators
	LAMBDA         = "λ"
	FAT_ARROW      = "=>"
	THIN_ARROW     = "->"
	PIPE_RIGHT     = ">>"
	PIPE_LEFT      = "<<"
	REDUCE_CHANNEL = "+>"

	// Range operators
	RANGE_INCL     = ".."
	RANGE_LEFT_INC = ".!"
	RANGE_RGHT_INC = "!."
	RANGE_EXCL     = "!!"

	// Delimiters & Punctuation
	LPAREN     = "("
	RPAREN     = ")"
	LBRACKET   = "["
	RBRACKET   = "]"
	LBRACE     = "{"
	RBRACE     = "}"
	COMMA      = ","
	SEMICOLON  = ";"
	COLON      = ":"
	DOT        = "."
	QUESTION   = "?"
	BANG       = "!"
	SIGIL_SYS  = "$"
	HASH       = "#"
	AT         = "@"
	AMPERSAND  = "&"
	UNDERSCORE = "_"

	// Keywords
	RULE  = "RULE"
	NEW   = "NEW"
	LET   = "LET"
	ALTER = "ALTER"
	CONST = "CONST"
	TYPE  = "TYPE"
	ZAP   = "ZAP"
	SELF  = "SELF"
	SUPER = "SUPER"
	APPLY = "APPLY"
	WITH  = "WITH"

	REQUIRE = "REQUIRE"
	ENSURE  = "ENSURE"

	IF    = "IF"
	THEN  = "THEN"
	ELSE  = "ELSE"
	MATCH = "MATCH"
	WHEN  = "WHEN"
	DO    = "DO"
	OTHER = "OTHER"
	FIRST = "FIRST"
	EVERY = "EVERY"
	TOTAL = "TOTAL"

	CYCLE  = "CYCLE"
	REPEAT = "REPEAT"
	WHILE  = "WHILE"
	FOR    = "FOR"
	IN_KW  = "IN"

	RETURN = "RETURN"
	STOP   = "STOP"
	NEXT   = "NEXT"
	YIELD  = "YIELD"
	RAISE  = "RAISE"
	RETRY  = "RETRY"

	TRIAL = "TRIAL"
	TRY   = "TRY"
	CASE  = "CASE"
	MISS  = "MISS"
	FINAL = "FINAL"
	DONE  = "DONE"

	BEGIN = "BEGIN"
	WAIT  = "WAIT"

	USE   = "USE"
	AS    = "AS"
	ALIAS = "ALIAS"

	PRINT  = "PRINT"
	WRITE  = "WRITE"
	EXPECT = "EXPECT"

	FORALL_KW = "FORALL"
	EXISTS_KW = "EXISTS"
	AND_KW    = "AND"
	OR_KW     = "OR"
	NOT_KW    = "NOT"
	XOR_KW    = "XOR"
)

var keywords = map[string]Type{
	"rule":  RULE,
	"new":   NEW,
	"let":   LET,
	"alter": ALTER,
	"const": CONST,
	"type":  TYPE,
	"zap":   ZAP,
	"self":  SELF,
	"super": SUPER,
	"apply": APPLY,
	"with":  WITH,

	"require": REQUIRE,
	"ensure":  ENSURE,

	"if":    IF,
	"then":  THEN,
	"else":  ELSE,
	"match": MATCH,
	"when":  WHEN,
	"do":    DO,
	"other": OTHER,
	"first": FIRST,
	"every": EVERY,
	"total": TOTAL,

	"cycle":  CYCLE,
	"repeat": REPEAT,
	"while":  WHILE,
	"for":    FOR,
	"in":     IN_KW,

	"return": RETURN,
	"stop":   STOP,
	"next":   NEXT,
	"yield":  YIELD,
	"raise":  RAISE,
	"retry":  RETRY,

	"trial": TRIAL,
	"try":   TRY,
	"case":  CASE,
	"miss":  MISS,
	"final": FINAL,
	"done":  DONE,

	"begin": BEGIN,
	"wait":  WAIT,

	"use":   USE,
	"as":    AS,
	"alias": ALIAS,

	"print":  PRINT,
	"write":  WRITE,
	"expect": EXPECT,

	"forall": FORALL_KW,
	"exists": EXISTS_KW,
	"and":    AND_KW,
	"or":     OR_KW,
	"not":    NOT_KW,
	"xor":    XOR_KW,
}

func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
