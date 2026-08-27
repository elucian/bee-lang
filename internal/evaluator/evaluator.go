package evaluator

import (
	"bee/internal/parser"
	"bee/internal/token"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
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

func (e *Evaluator) DumpContext() {
	fmt.Fprintln(os.Stderr, "=== VARIABLE CONTEXT ===")
	for name, val := range e.symbols {
		fmt.Fprintf(os.Stderr, "  %s : int = %d\n", name, val)
	}
	for name, arr := range e.arrayValues {
		fmt.Fprintf(os.Stderr, "  %s : []int = %v\n", name, arr)
	}
	fmt.Fprintln(os.Stderr, "========================")
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
			e.DumpContext()
			panic(fmt.Sprintf("expect failed at line %d", int(s.Token.Pos)))
		} else {
			fmt.Fprintf(os.Stderr, "DEBUG: Expectation passed in line %d\n", int(s.Token.Pos))
		}
	case *parser.PrintStatement:
		for _, expr := range s.Expressions {
			val := e.evalExpression(expr)
			fmt.Print(val, " ")
		}
		fmt.Println()
	case *parser.AssignmentStatement:
		for i, name := range s.Names {
			if i < len(s.Values) {
				rightVal := e.evalIntExpression(s.Values[i])
				curVal := e.symbols[name.Value]
				lit := s.Token.Literal
				switch lit {
				case ":=", "=":
					e.symbols[name.Value] = rightVal
				case "+=":
					e.symbols[name.Value] = evalArithmetic("+", curVal, rightVal, "")
				case "-=":
					e.symbols[name.Value] = evalArithmetic("-", curVal, rightVal, "")
				case "*=":
					e.symbols[name.Value] = evalArithmetic("*", curVal, rightVal, "")
				case "/=":
					e.symbols[name.Value] = evalArithmetic("/", curVal, rightVal, "")
				case "%=":
					e.symbols[name.Value] = evalArithmetic("%", curVal, rightVal, "")
				case "^=":
					e.symbols[name.Value] = evalArithmetic("^", curVal, rightVal, "")
				case "√=":
					e.symbols[name.Value] = evalArithmetic("√", curVal, rightVal, "")
				default:
					e.symbols[name.Value] = rightVal
				}
			}
		}
	case *parser.DeclarationStatement:
		if s.Value != nil {
			if arrLit, ok := s.Value.(*parser.ArrayLiteral); ok {
				elems := make([]int, len(arrLit.Elements))
				for i, el := range arrLit.Elements {
					elems[i] = e.evalIntExpression(el)
				}
				e.arrayValues[s.Name] = elems
			} else {
				val := e.evalIntExpression(s.Value)
				e.symbols[s.Name] = val
			}
		} else {
			e.symbols[s.Name] = 0
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
			if arr, ok := e.arrayValues[ident.Value]; ok {
				if idx >= 1 && idx <= len(arr) {
					return arr[idx-1]
				}
			}
		}
		return 0
	case *parser.BinaryExpression:
		lit := expr.Token.Literal
		if lit == "=" || expr.Token.Type == token.EQ || lit == "==" {
			leftVal := e.evalIntExpression(expr.Left)
			rightVal := e.evalIntExpression(expr.Right)
			return evalComparison(lit, leftVal, rightVal, expr.Token.Type, lit)
		}
		if expr.Token.Type == token.NEQ_UNICODE || lit == "≠" || strings.Contains(lit, "≠") {
			leftVal := e.evalIntExpression(expr.Left)
			rightVal := e.evalIntExpression(expr.Right)
			return evalComparison(lit, leftVal, rightVal, expr.Token.Type, lit)
		}

		if strings.HasSuffix(lit, "√") || strings.Contains(lit, "√") || expr.Token.Type == token.SQRT || lit == "√" {
			deg := 2
			orderStr := strings.TrimSuffix(lit, "√")
			if orderStr != "" {
				d := parseSuperscriptInt(orderStr)
				if d > 0 {
					deg = d
				}
			}
			val := e.evalIntExpression(expr.Right)
			res := math.Round(math.Pow(float64(val), 1.0/float64(deg)))
			return int(res)
		}

		left := e.evalIntExpression(expr.Left)
		right := e.evalIntExpression(expr.Right)

		switch lit {
		case "+", "-", "*", "/", "%", "^", "√":
			return evalArithmetic(lit, left, right, lit)
		case "and", "or", "xor", "¬":
			return evalLogical(lit, left, right)
		case "<", ">", "<=", ">=":
			return evalComparison(lit, left, right, expr.Token.Type, lit)
		default:
			if strings.HasSuffix(lit, "√") || strings.Contains(lit, "√") {
				return evalArithmetic("√", left, right, lit)
			}
			return evalArithmetic(lit, left, right, lit)
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
	case *parser.Identifier:
		if val, ok := e.symbols[expr.Value]; ok {
			return strconv.Itoa(val)
		}
		return expr.Value
	}
	return ""
}
