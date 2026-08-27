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

func parseSuperscriptInt(s string) int {
	val := 0
	for _, r := range s {
		digit := -1
		switch r {
		case '⁰':
			digit = 0
		case '¹':
			digit = 1
		case '²':
			digit = 2
		case '³':
			digit = 3
		case '⁴':
			digit = 4
		case '⁵':
			digit = 5
		case '⁶':
			digit = 6
		case '⁷':
			digit = 7
		case '⁸':
			digit = 8
		case '⁹':
			digit = 9
		}
		if digit >= 0 {
			val = val*10 + digit
		}
	}
	if val == 0 {
		return 2
	}
	return val
}

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
				if lit == ":=" || lit == "=" {
					e.symbols[name.Value] = rightVal
				} else if lit == "+=" {
					e.symbols[name.Value] = curVal + rightVal
				} else if lit == "-=" {
					e.symbols[name.Value] = curVal - rightVal
				} else if lit == "*=" {
					e.symbols[name.Value] = curVal * rightVal
				} else if lit == "/=" {
					if rightVal != 0 {
						e.symbols[name.Value] = curVal / rightVal
					}
				} else if lit == "%=" {
					if rightVal != 0 {
						res := curVal % rightVal
						if res < 0 {
							if rightVal > 0 {
								res += rightVal
							} else {
								res -= rightVal
							}
						}
						e.symbols[name.Value] = res
					}
				} else if lit == "^=" {
					res := 1
					base := curVal
					for k := 0; k < rightVal; k++ {
						res *= base
					}
					e.symbols[name.Value] = res
				} else if lit == "√=" {
					res := 1
					if rightVal == 2 {
						res = int(math.Round(math.Sqrt(float64(curVal))))
					} else if rightVal == 3 {
						res = int(math.Round(math.Cbrt(float64(curVal))))
					} else {
						res = int(math.Round(math.Pow(float64(curVal), 1.0/float64(rightVal))))
					}
					e.symbols[name.Value] = res
				} else {
					e.symbols[name.Value] = rightVal
				}
			}
		}
	case *parser.DeclarationStatement:
		val := 0
		if s.Value != nil {
			if arrLit, ok := s.Value.(*parser.ArrayLiteral); ok {
				elems := make([]int, len(arrLit.Elements))
				for i, el := range arrLit.Elements {
					elems[i] = e.evalIntExpression(el)
				}
				e.arrayValues[s.Name] = elems
			} else {
				val = e.evalIntExpression(s.Value)
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
			// 1-based indexing in Bee specifications
			if arr, ok := e.arrayValues[ident.Value]; ok {
				if idx >= 1 && idx <= len(arr) {
					return arr[idx-1]
				}
			}
		}
		return 0
	case *parser.BinaryExpression:
		if expr.Token.Literal == "=" || expr.Token.Type == token.EQ || expr.Token.Literal == "==" || expr.Token.Type == token.ASSIGN {
			leftVal := e.evalIntExpression(expr.Left)
			rightVal := e.evalIntExpression(expr.Right)
			if leftVal == rightVal {
				return 1
			}
			return 0
		}
		left := e.evalIntExpression(expr.Left)
		right := e.evalIntExpression(expr.Right)
		if expr.Token.Type == token.NEQ_UNICODE || expr.Token.Literal == "≠" || strings.Contains(expr.Token.Literal, "≠") || strings.Contains(expr.Token.Literal, "â‰") {
			if left != right {
				return 1
			}
			return 0
		}
		if expr.Token.Literal == "<" || expr.Token.Type == token.LT {
			leftVal := e.evalIntExpression(expr.Left)
			rightVal := e.evalIntExpression(expr.Right)
			if leftVal < rightVal {
				return 1
			}
			return 0
		}
		if expr.Token.Literal == ">" || expr.Token.Type == token.GT {
			leftVal := e.evalIntExpression(expr.Left)
			rightVal := e.evalIntExpression(expr.Right)
			if leftVal > rightVal {
				return 1
			}
			return 0
		}
		switch expr.Token.Literal {
		case "+":
			return left + right
		case "-":
			return left - right
		case "*":
			return left * right
		case "/":
			if right != 0 {
				return left / right
			}
			return 0
		case "%":
			if right != 0 {
				res := left % right
				if res < 0 {
					if right > 0 {
						res += right
					} else {
						res -= right
					}
				}
				return res
			}
			return 0
		case "and":
			if left != 0 && right != 0 {
				return 1
			}
			return 0
		case "or":
			if left != 0 || right != 0 {
				return 1
			}
			return 0
		case "xor":
			if (left != 0) != (right != 0) {
				return 1
			}
			return 0
		case "¬":
			if left == 0 && right == 0 { // unary not emulation
				if right == 0 {
					return 1
				}
				return 0
			}
			if right == 0 {
				return 1
			}
			return 0
		case "=", "==":
			if left == right {
				return 1
			}
			return 0
		case "≠", "!=":
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
		case "^":
			res := 1
			rightVal := right
			lit := expr.Token.Literal
			if lit == "³" || lit == "3" {
				rightVal = 3
			} else if lit == "²" || lit == "2" {
				rightVal = 2
			} else if lit == "⁴" || lit == "4" {
				rightVal = 4
			} else if lit == "⁵" || lit == "5" {
				rightVal = 5
			} else if lit == "⁶" || lit == "6" {
				rightVal = 6
			} else if lit == "⁷" || lit == "7" {
				rightVal = 7
			} else if lit == "⁸" || lit == "8" {
				rightVal = 8
			} else if lit == "⁹" || lit == "9" {
				rightVal = 9
			} else {
				rightVal = parseSuperscriptInt(lit)
				if rightVal == 2 && lit != "²" {
					rightVal = right
				}
			}
			for i := 0; i < rightVal; i++ {
				res *= left
			}
			return res
		case "√":
			if left == 0 {
				left = 2
			}
			res := 1
			for i := 1; i <= right; i++ {
				pow := 1
				for j := 0; j < left; j++ {
					pow *= i
				}
				if pow == right {
					return i
				}
				if pow > right {
					return i - 1
				}
			}
			return res
		default:
			if strings.HasSuffix(expr.Token.Literal, "√") {
				orderStr := strings.TrimSuffix(expr.Token.Literal, "√")
				deg := 2
				if orderStr != "" {
					deg = parseSuperscriptInt(orderStr)
				}
				left := deg
				res := 1
				for i := 1; i <= right; i++ {
					pow := 1
					for j := 0; j < left; j++ {
						pow *= i
					}
					if pow == right {
						return i
					}
					if pow > right {
						return i - 1
					}
				}
				return res
			}
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
