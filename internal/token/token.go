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
	ASSIGN    = "="
	SEMICOLON = ";"
)

func LookupIdent(ident string) Type {
	return IDENT
}
