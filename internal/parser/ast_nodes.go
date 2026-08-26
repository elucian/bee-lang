// internal/parser/ast_nodes.go
package ast

import "bee/token"

// Node represents the base interface for all AST elements
type Node interface {
    Pos() token.Pos
}

// Statement defines a node that represents a Bee statement
type Statement interface {
    Node
    statementNode()
}

// Expression defines a node that represents a Bee expression
type Expression interface {
    Node
    expressionNode()
}

// ----------------------------------------------------------------------------
// Declarations
// ----------------------------------------------------------------------------

type Declaration struct {
    Token token.Token // "new" or "set"
    Name  *Identifier
    Type  Expression
    Value Expression
}
func (d *Declaration) statementNode() {}
func (d *Declaration) Pos() token.Pos { return d.Token.Pos }

type Identifier struct {
    Token token.Token
    Value string
}
func (i *Identifier) expressionNode() {}
func (i *Identifier) Pos() token.Pos { return i.Token.Pos }

// ----------------------------------------------------------------------------
// Control Flow
// ----------------------------------------------------------------------------

type IfStatement struct {
    Token     token.Token // "if"
    Condition Expression
    Body      *Block
    Else      *Block
}
func (i *IfStatement) statementNode() {}
func (i *IfStatement) Pos() token.Pos { return i.Token.Pos }

type Block struct {
    Statements []Statement
}

// ----------------------------------------------------------------------------
// Expressions
// ----------------------------------------------------------------------------

type BinaryExpression struct {
    Token    token.Token // operator
    Left     Expression
    Right    Expression
}
func (b *BinaryExpression) expressionNode() {}
func (b *BinaryExpression) Pos() token.Pos { return b.Token.Pos }
