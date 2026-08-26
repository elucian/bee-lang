// internal/parser/ast_nodes.go
package parser

import (
	"bee/internal/token"
)

type Node interface {
	Pos() token.Pos
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

type PrintStatement struct {
	Token      token.Token
	Expression Expression
}

func (ps *PrintStatement) statementNode() {}
func (ps *PrintStatement) Pos() token.Pos { return ps.Token.Pos }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) Pos() token.Pos  { return i.Token.Pos }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) Pos() token.Pos  { return sl.Token.Pos }
