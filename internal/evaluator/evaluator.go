package evaluator

import (
	"bee/internal/parser"
	"fmt"
)

type Evaluator struct {
	symbols map[string]string
}

func New() *Evaluator {
	return &Evaluator{symbols: make(map[string]string)}
}

func (e *Evaluator) Eval(program *parser.Program) {
	for _, stmt := range program.Statements {
		e.evalStatement(stmt)
	}
}

func (e *Evaluator) evalStatement(node parser.Statement) {
	switch s := node.(type) {
	case *parser.PrintStatement:
		for _, expr := range s.Expressions {
			fmt.Print(e.evalExpression(expr), " ")
		}
		fmt.Println()
	case *parser.AssignmentStatement:
		val := e.evalExpression(s.Values[0])
		e.symbols[s.Names[0].Value] = val
	}
}

func (e *Evaluator) evalExpression(node parser.Expression) string {
	switch expr := node.(type) {
	case *parser.StringLiteral:
		return expr.Value
	case *parser.IntegerLiteral:
		return expr.Value
	case *parser.Identifier:
		if val, ok := e.symbols[expr.Value]; ok {
			return val
		}
		return expr.Value
	default:
		return ""
	}
}

// SetSymbol updates the internal symbol table for VM execution
func (e *Evaluator) SetSymbol(name string, value string) {
	e.symbols[name] = value
}
