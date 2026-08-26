package token

type Type string

type Pos int

type Token struct {
	Type    Type
	Literal string
	Pos     Pos
}

const (
	ILLEGAL   = "ILLEGAL"
	EOF       = "EOF"
	IDENT     = "IDENT"
	INT       = "INT"
	STRING    = "STRING"
	ASSIGN    = ":="
	EQ        = "="
	PLUS      = "+"
	MINUS     = "-"
	ASTERISK  = "*"
	SEMICOLON = ";"
	RULE      = "RULE"
	PRINT     = "PRINT"
	RETURN    = "RETURN"
	NEW       = "NEW"
	LET       = "LET"
	EXPECT    = "EXPECT"
	COLON     = ":"
)

var keywords = map[string]Type{
	"rule":   RULE,
	"print":  PRINT,
	"return": RETURN,
	"new":    NEW,
	"let":    LET,
	"expect": EXPECT,
}

func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
