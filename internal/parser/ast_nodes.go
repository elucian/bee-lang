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

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}
func (bs *BlockStatement) Pos() token.Pos { return bs.Token.Pos }

type IfStatement struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode() {}
func (is *IfStatement) Pos() token.Pos { return is.Token.Pos }

type PrintStatement struct {
	Token       token.Token
	Expressions []Expression
	Separator   Expression
}

func (ps *PrintStatement) statementNode() {}
func (ps *PrintStatement) Pos() token.Pos { return ps.Token.Pos }

type AssignmentStatement struct {
	Token  token.Token
	Names  []*Identifier
	Values []Expression
}

func (as *AssignmentStatement) statementNode() {}
func (as *AssignmentStatement) Pos() token.Pos { return as.Token.Pos }

type AssertStatement struct {
	Token     token.Token
	Condition Expression
}

func (as *AssertStatement) statementNode() {}
func (as *AssertStatement) Pos() token.Pos { return as.Token.Pos }

type ExpectStatement struct {
	Token     token.Token
	Condition Expression
}

func (es *ExpectStatement) statementNode() {}
func (es *ExpectStatement) Pos() token.Pos { return es.Token.Pos }

type DeclarationStatement struct {
	Token token.Token
	Name  string
	Value Expression
}

func (ds *DeclarationStatement) statementNode() {}
func (ds *DeclarationStatement) Pos() token.Pos { return ds.Token.Pos }

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

type IntegerLiteral struct {
	Token token.Token
	Value string
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) Pos() token.Pos  { return il.Token.Pos }

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}
func (al *ArrayLiteral) Pos() token.Pos  { return al.Token.Pos }

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode() {}
func (ie *IndexExpression) Pos() token.Pos  { return ie.Token.Pos }

type BinaryExpression struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (be *BinaryExpression) expressionNode() {}
func (be *BinaryExpression) Pos() token.Pos  { return be.Token.Pos }
