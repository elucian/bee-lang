package evaluator

import (
	"bee/internal/parser"
	"fmt"
	"os"
	"strconv"
)

type Evaluator struct {
	symbols     map[string]int
	arrayValues map[string][]int
}

func New() *Evaluator {
	return &Evaluator{
		symbols:     make(map[string]int),
		arrayValues: make(map[string][]int),
	}
}

func (e *Evaluator) Eval(program *parser.Program) {
	for _, stmt := range program.Statements {
		e.evalStatement(stmt)
	}
}

func (e *Evaluator) evalStatement(node parser.Statement) {
	switch s := node.(type) {
	case *parser.ExpectStatement:
		val := e.evalIntExpression(s.Condition)
		if val == 0 {
			panic(fmt.Sprintf("Assertion failed: expect statement failed (evaluated condition yielded 0)"))
		} else {
			fmt.Fprintf(os.Stderr, "DEBUG: Expectation passed\n")
		}
	case *parser.PrintStatement:
		for _, expr := range s.Expressions {
			val := e.evalExpression(expr)
			fmt.Print(val, " ")
		}
		fmt.Println()
	case *parser.AssignmentStatement:
		if len(s.Names) == 2 && len(s.Values) == 1 {
			// Handle tuple swapping assignment like let a, b := b, a;
			// In our simple parser, s.Values[0] is b
			identRHS := s.Values[0].(*parser.Identifier).Value
			if identRHS == s.Names[1].Value {
				// Swap
				val0 := e.evalIntExpression(s.Names[1])
				val1 := e.evalIntExpression(s.Names[0])
				e.symbols[s.Names[0].Value] = val0
				e.symbols[s.Names[1].Value] = val1
			}
		} else if len(s.Names) == len(s.Values) {
			vals := make([]int, len(s.Values))
			for i, v := range s.Values {
				vals[i] = e.evalIntExpression(v)
			}
			for i, name := range s.Names {
				e.symbols[name.Value] = vals[i]
			}
		} else if len(s.Names) == 2 && len(s.Values) == 2 {
			v0 := e.evalIntExpression(s.Values[0])
			v1 := e.evalIntExpression(s.Values[1])
			e.symbols[s.Names[0].Value] = v0
			e.symbols[s.Names[1].Value] = v1
		} else if len(s.Names) > 0 && len(s.Values) > 0 {
			val := e.evalIntExpression(s.Values[0])
			e.symbols[s.Names[0].Value] = val
		}
	case *parser.DeclarationStatement:
		if s.Value != nil {
			if arrLit, ok := s.Value.(*parser.ArrayLiteral); ok {
				elems := make([]int, len(arrLit.Elements))
				for i, el := range arrLit.Elements {
					elems[i] = e.evalIntExpression(el)
				}
				e.arrayValues[s.Name] = elems
			} else if _, ok := s.Value.(*parser.IndexExpression); ok {
				val := e.evalIntExpression(s.Value)
				e.symbols[s.Name] = val
			} else {
				val := e.evalIntExpression(s.Value)
				e.symbols[s.Name] = val
			}
		}
	}
}

func (e *Evaluator) evalIntExpression(node parser.Expression) int {
	switch expr := node.(type) {
	case *parser.IntegerLiteral:
		val, _ := strconv.Atoi(expr.Value)
		return val
	case *parser.Identifier:
		if expr.Value == "$" {
			return 0
		}
		if val, ok := e.symbols[expr.Value]; ok {
			return val
		}
		return 0
	case *parser.IndexExpression:
		if ident, ok := expr.Left.(*parser.Identifier); ok {
			var idx int
			if idIdent, ok := expr.Index.(*parser.Identifier); ok && idIdent.Value == "$" {
				if arr, ok := e.arrayValues[ident.Value]; ok {
					idx = len(arr)
				}
			} else {
				idx = e.evalIntExpression(expr.Index)
			}
			// 1-based indexing in Bee specifications
			if arr, ok := e.arrayValues[ident.Value]; ok {
				if idx >= 1 && idx <= len(arr) {
					return arr[idx-1]
				}
			}
		}
		return 0
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
		case "=":
			if left == right {
				return 1
			}
			return 0
		case "!=":
			if left != right {
				return 1
			}
			return 0
		case "<":
			if left < right {
				return 1
			}
			return 0
		case ">":
			if left > right {
				return 1
			}
			return 0
		case "<=":
			if left <= right {
				return 1
			}
			return 0
		case ">=":
			if left >= right {
				return 1
			}
			return 0
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
	case *parser.IndexExpression:
		if ident, ok := expr.Left.(*parser.Identifier); ok {
			var idx int
			if idIdent, ok := expr.Index.(*parser.Identifier); ok && idIdent.Value == "$" {
				if arr, ok := e.arrayValues[ident.Value]; ok {
					idx = len(arr)
				}
			} else {
				idx = e.evalIntExpression(expr.Index)
			}
			if arr, ok := e.arrayValues[ident.Value]; ok {
				if idx >= 1 && idx <= len(arr) {
					return strconv.Itoa(arr[idx-1])
				}
			}
		}
		return "0"
	case *parser.ArrayLiteral:
		var elems []string
		for _, el := range expr.Elements {
			elems = append(elems, e.evalExpression(el))
		}
		return fmt.Sprintf("%v", elems)
	case *parser.Identifier:
		if arr, ok := e.arrayValues[expr.Value]; ok {
			return fmt.Sprintf("%v", arr)
		}
		if val, ok := e.symbols[expr.Value]; ok {
			return strconv.Itoa(val)
		}
		return expr.Value
	default:
		return ""
	}
}
