package evaluator

import (
	"bee/internal/parser"
	"fmt"
	"os"
	"strconv"
)

type Evaluator struct {
	symbols map[string]int
}

func New() *Evaluator {
	return &Evaluator{symbols: make(map[string]int)}
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
			val := e.evalExpression(expr)
			fmt.Print(val, " ")
		}
		fmt.Println()
	case *parser.AssignmentStatement:
		val := e.evalIntExpression(s.Values[0])
		e.symbols[s.Names[0].Value] = val
		fmt.Fprintf(os.Stderr, "DEBUG: Assignment %s = %d\n", s.Names[0].Value, val)
	}
}

func (e *Evaluator) evalIntExpression(node parser.Expression) int {
	switch expr := node.(type) {
	case *parser.IntegerLiteral:
		val, _ := strconv.Atoi(expr.Value)
		return val
	case *parser.Identifier:
		return e.symbols[expr.Value]
	case *parser.BinaryExpression:
		left := e.evalIntExpression(expr.Left)
		right := e.evalIntExpression(expr.Right)
		switch expr.Token.Literal {
		case "+":
			return left + right
		case "-":
			return left - right
		case "*":
			return left * right
		}
	}
	return 0
}

func (e *Evaluator) evalExpression(node parser.Expression) string {
	switch expr := node.(type) {
	case *parser.StringLiteral:
		return expr.Value
	case *parser.IntegerLiteral:
		return expr.Value
	case *parser.Identifier:
		if val, ok := e.symbols[expr.Value]; ok {
			return strconv.Itoa(val)
		}
		return expr.Value
	default:
		return ""
	}
}
