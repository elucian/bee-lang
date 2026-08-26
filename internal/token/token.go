package token

type Type string

type Token struct {
	Type    Type
	Literal string
	Pos     int
}

const (
	ILLEGAL   = "ILLEGAL"
	EOF       = "EOF"
	IDENT     = "IDENT"
	INT       = "INT"
	ASSIGN    = ":="
	EQ        = "="
	SEMICOLON = ";"
	RULE      = "RULE"
	PRINT     = "PRINT"
	RETURN    = "RETURN"
	NEW       = "NEW"
	LET       = "LET"
	EXPECT    = "EXPECT"
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
