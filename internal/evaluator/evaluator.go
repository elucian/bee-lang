package evaluator

import (
	"bee/internal/parser"
	"fmt"
)

type Evaluator struct{}

func New() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Eval(program *parser.Program) {
	for _, stmt := range program.Statements {
		e.evalStatement(stmt)
	}
}

func (e *Evaluator) evalStatement(node parser.Statement) {
	switch s := node.(type) {
	case *parser.PrintStatement:
		val := e.evalExpression(s.Expression)
		fmt.Println(val)
	}
}

func (e *Evaluator) evalExpression(node parser.Expression) string {
	switch expr := node.(type) {
	case *parser.StringLiteral:
		return expr.Value
	case *parser.Identifier:
		return expr.Value
	default:
		return ""
	}
}
