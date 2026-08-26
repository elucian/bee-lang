// internal/ast/nodes.go
package ast

import "bee/token"

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

// Declarations
type Declaration struct {
    Token token.Token
    Name  *Identifier
    Value Expression
}

type Assignment struct {
    Token token.Token
    Name  *Identifier
    Value Expression
}
