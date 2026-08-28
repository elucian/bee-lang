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
	IN_OP         = "∈"
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
	SET   = "SET"
	TYPE  = "TYPE"
	ZAP   = "ZAP"
	SELF  = "SELF"
	SUPER = "SUPER"
	APPLY = "APPLY"
	WITH  = "WITH"

	BEGIN      = "BEGIN"
	ALIAS      = "ALIAS"
	AND        = "AND"
	ABORT      = "ABORT"
	OTHER      = "OTHER"
	CASE       = "CASE"
	CONTINUE   = "CONTINUE"
	DONE       = "DONE"
	DEFAULT    = "DEFAULT"
	IF         = "IF"
	IS         = "IS"
	DO         = "DO"
	ELSE       = "ELSE"
	EXIT       = "EXIT"
	FAIL       = "FAIL"
	FINAL      = "FINAL"
	MISS       = "MISS"
	PANIC      = "PANIC"
	LIKE       = "LIKE"
	LOAD       = "LOAD"
	NEXT       = "NEXT"
	JOB        = "JOB"
	MATCH      = "MATCH"
	OVER       = "OVER"
	PRINT      = "PRINT"
	PASS       = "PASS"
	VOID       = "VOID"
	RETURN     = "RETURN"
	RETRY      = "RETRY"
	NONE       = "NONE"
	SCRAP      = "SCRAP"
	READ       = "READ"
	TRIAL      = "TRIAL"
	STOP       = "STOP"
	YIELD      = "YIELD"
	RAISE      = "RAISE"
	XOR        = "XOR"
	WRITE      = "WRITE"
	WAIT       = "WAIT"
	WHEN       = "WHEN"
	OR         = "OR"
	HIDE       = "HIDE"
	CYCLE      = "CYCLE"
	WHILE      = "WHILE"
	FOR        = "FOR"
	RESUME     = "RESUME"
	PUT        = "PUT"
	POP        = "POP"
	NOT        = "NOT"
	AS         = "AS"
	IN_KEYWORD = "IN"
	START      = "START"
	TRY        = "TRY"
	EXPECT     = "EXPECT"
)

var keywords = map[string]Type{
	"rule":  RULE,
	"new":   NEW,
	"let":   LET,
	"set":   SET,
	"type":  TYPE,
	"zap":   ZAP,
	"self":  SELF,
	"super": SUPER,
	"apply": APPLY,
	"with":  WITH,

	"begin":    BEGIN,
	"alias":    ALIAS,
	"and":      AND,
	"abort":    ABORT,
	"other":    OTHER,
	"case":     CASE,
	"continue": CONTINUE,
	"done":     DONE,
	"default":  DEFAULT,
	"if":       IF,
	"is":       IS,
	"do":       DO,
	"else":     ELSE,
	"exit":     EXIT,
	"fail":     FAIL,
	"final":    FINAL,
	"miss":     MISS,
	"panic":    PANIC,
	"like":     LIKE,
	"load":     LOAD,
	"next":     NEXT,
	"job":      JOB,
	"match":    MATCH,
	"over":     OVER,
	"print":    PRINT,
	"pass":     PASS,
	"void":     VOID,
	"return":   RETURN,
	"retry":    RETRY,
	"none":     NONE,
	"scrap":    SCRAP,
	"read":     READ,
	"trial":    TRIAL,
	"stop":     STOP,
	"yield":    YIELD,
	"xor":      XOR,
	"write":    WRITE,
	"wait":     WAIT,
	"when":     WHEN,
	"or":       OR,
	"hide":     HIDE,
	"cycle":    CYCLE,
	"while":    WHILE,
	"for":      FOR,
	"resume":   RESUME,
	"put":      PUT,
	"pop":      POP,
	"raise":    RAISE,
	"not":      NOT,
	"as":       AS,
	"in":       IN_KEYWORD,
	"start":    START,
	"try":      TRY,
	"expect":   EXPECT,
}

func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
