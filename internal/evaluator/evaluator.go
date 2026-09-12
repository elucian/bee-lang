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
	symbols       map[string]int
	stringSymbols map[string]string
	arrayValues   map[string][]int
	debug         bool
}

func (e *Evaluator) SetDebug(debug bool) {
	e.debug = debug
}

func (e *Evaluator) debugLog(format string, a ...interface{}) {
	if e.debug {
		fmt.Fprintf(os.Stderr, format, a...)
	}
}

func New() *Evaluator {
	return &Evaluator{
		symbols:       make(map[string]int),
		stringSymbols: make(map[string]string),
		arrayValues:   make(map[string][]int),
	}
}

func (e *Evaluator) DumpContext() {
	fmt.Fprintln(os.Stderr, "=== VARIABLE CONTEXT ===")
	for name, val := range e.symbols {
		fmt.Fprintf(os.Stderr, "  %s : int = %d\n", name, val)
	}
	for name, s := range e.stringSymbols {
		fmt.Fprintf(os.Stderr, "  %s : string = %q\n", name, s)
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
	case *parser.RuleStatement:
		// Forward declarations carry no body — they bind the rule signature
		// for self-recursion and mutual recursion. See spec/03-rules.md §5.2.
		if s.ForwardDecl || s.Body == nil {
			return
		}
		if s.Name == "main" {
			e.debugLog("EVALUATOR DEBUG: entering rule main\n")
		}
		for _, stmt := range s.Body.Statements {
			e.evalStatement(stmt)
		}
		if s.Name == "main" {
			e.debugLog("EVALUATOR DEBUG: exited rule main\n")
		}
	case *parser.BlockStatement:
		for _, stmt := range s.Statements {
			e.evalStatement(stmt)
		}
	case *parser.IfStatement:
		condVal := e.evalIntExpression(s.Condition)
		if condVal != 0 {
			if s.Consequence != nil {
				e.evalStatement(s.Consequence)
			}
		} else {
			if s.Alternative != nil {
				e.evalStatement(s.Alternative)
			}
		}
	case *parser.AssertStatement:
		val := e.evalIntExpression(s.Condition)
		if val == 0 {
			fmt.Fprintf(os.Stderr, "[WARNING] Assertion failed at line %d\n", int(s.Token.Pos))
		}
	case *parser.ExpectStatement:
		val := e.evalIntExpression(s.Condition)
		if val == 0 {
			e.DumpContext()
			panic(fmt.Sprintf("expect failed at line %d", int(s.Token.Pos)))
		} else {
			e.debugLog("DEBUG: Expectation passed in line %d\n", int(s.Token.Pos))
		}
	case *parser.PrintStatement:
		separator := " "
		if s.Separator != nil {
			sepVal := e.evalExpression(s.Separator)
			separator = strings.Trim(sepVal, "\"")
		}

		for i, expr := range s.Expressions {
			if i > 0 {
				fmt.Print(separator)
			}
			val := e.evalExpression(expr)
			fmt.Print(val)
		}
		fmt.Println()
	case *parser.AssignmentStatement:
		e.debugLog("EVALUATOR DEBUG: AssignmentStatement start, token lit=%q, type=%v\n", s.Token.Literal, s.Token.Type)
		for i, name := range s.Names {
			e.debugLog("EVALUATOR DEBUG: Assigning to %s\n", name.Value)
			if i < len(s.Values) {
				if strLit, ok := s.Values[i].(*parser.StringLiteral); ok {
					e.stringSymbols[name.Value] = strLit.Value
					delete(e.symbols, name.Value)
					e.debugLog("EVALUATOR DEBUG: Assigned string %s = %q\n", name.Value, strLit.Value)
					continue
				} else if strIdent, ok := s.Values[i].(*parser.Identifier); ok {
					if strVal, hasStr := e.stringSymbols[strIdent.Value]; hasStr {
						e.stringSymbols[name.Value] = strVal
						delete(e.symbols, name.Value)
						e.debugLog("EVALUATOR DEBUG: Assigned string %s = %q\n", name.Value, strVal)
						continue
					}
				}

				rightVal := e.evalIntExpression(s.Values[i])
				e.debugLog("EVALUATOR DEBUG: RightVal = %d\n", rightVal)

				curVal, ok := e.symbols[name.Value]
				e.debugLog("EVALUATOR DEBUG: Name %s exists? %v, curVal = %d\n", name.Value, ok, curVal)

				// Assignment handles both initial assignment and mutation (compound op)
				lit := s.Token.Literal
				e.debugLog("EVALUATOR DEBUG: Executing mutation, lit=%q\n", lit)
				switch lit {
				case "+=":
					e.debugLog("EVALUATOR DEBUG: Apply += to %s (val %d)\n", name.Value, rightVal)
					e.symbols[name.Value] += rightVal
					e.debugLog("EVALUATOR DEBUG: Result %s = %d\n", name.Value, e.symbols[name.Value])
				case "-=":
					e.symbols[name.Value] -= rightVal
				case "*=":
					e.symbols[name.Value] *= rightVal
				case "/=":
					if rightVal != 0 {
						e.symbols[name.Value] /= rightVal
					}
				case "%=":
					if rightVal != 0 {
						res := e.symbols[name.Value] % rightVal
						if res < 0 {
							if rightVal > 0 {
								res += rightVal
							} else {
								res -= rightVal
							}
						}
						e.symbols[name.Value] = res
					}
				case "^=":
					e.symbols[name.Value] = int(math.Pow(float64(e.symbols[name.Value]), float64(rightVal)))
				case "√=":
					deg := rightVal
					if deg <= 0 {
						deg = 2
					}
					e.symbols[name.Value] = int(math.Round(math.Pow(float64(e.symbols[name.Value]), 1.0/float64(deg))))
				case ":=", "=":
					e.symbols[name.Value] = rightVal
				default:
					e.symbols[name.Value] = rightVal
				}
				e.debugLog("EVALUATOR DEBUG: After assign, %s = %d\n", name.Value, e.symbols[name.Value])
				e.debugLog("EVALUATOR DEBUG: Final value of %s = %d\n", name.Value, e.symbols[name.Value])
			}
		}
	case *parser.DeclarationStatement:
		names := s.Names
		if len(names) == 0 && s.Name != "" {
			names = []string{s.Name}
		}
		for i, name := range names {
			e.debugLog("EVALUATOR DEBUG: Declaring %s\n", name)
			if i < len(s.Values) && s.Values[i] != nil {
				if strLit, ok := s.Values[i].(*parser.StringLiteral); ok {
					e.stringSymbols[name] = strLit.Value
				} else if arrLit, ok := s.Values[i].(*parser.ArrayLiteral); ok {
					elems := make([]int, len(arrLit.Elements))
					for j, el := range arrLit.Elements {
						elems[j] = e.evalIntExpression(el)
					}
					e.arrayValues[name] = elems
				} else {
					val := e.evalIntExpression(s.Values[i])
					e.symbols[name] = val
				}
			} else if s.Value != nil && len(names) == 1 {
				if strLit, ok := s.Value.(*parser.StringLiteral); ok {
					e.stringSymbols[name] = strLit.Value
				} else if arrLit, ok := s.Value.(*parser.ArrayLiteral); ok {
					elems := make([]int, len(arrLit.Elements))
					for j, el := range arrLit.Elements {
						elems[j] = e.evalIntExpression(el)
					}
					e.arrayValues[name] = elems
				} else {
					val := e.evalIntExpression(s.Value)
					e.symbols[name] = val
				}
			} else {
				e.symbols[name] = 0
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
		if expr.Value == "True" || expr.Value == "true" {
			return 1
		}
		if expr.Value == "False" || expr.Value == "false" {
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
		if expr.Token.Type == token.NOT_EQ || lit == "!=" || lit == "<>" || expr.Token.Type == token.NEQ_UNICODE || lit == "≠" || strings.Contains(lit, "≠") {
			leftVal := e.evalIntExpression(expr.Left)
			rightVal := e.evalIntExpression(expr.Right)
			return evalComparison(lit, leftVal, rightVal, expr.Token.Type, lit)
		}
		if lit == "in" || lit == "∈" || expr.Token.Type == token.IN_OP || expr.Token.Type == token.IN_KEYWORD {
			leftVal := e.evalIntExpression(expr.Left)
			if binRight, ok := expr.Right.(*parser.BinaryExpression); ok && (binRight.Token.Literal == ".." || binRight.Token.Type == token.RANGE_INCL) {
				start := e.evalIntExpression(binRight.Left)
				end := e.evalIntExpression(binRight.Right)
				if leftVal >= start && leftVal <= end {
					return 1
				}
				return 0
			}
		}

		// <!-- EVAL: RADICAL_OPERATOR_EVALUATION -->
		if strings.HasSuffix(lit, "√") || strings.Contains(lit, "√") || expr.Token.Type == token.SQRT || lit == "√" {
			deg := 2
			orderStr := strings.TrimSuffix(lit, "√")
			if orderStr != "" {
				d := parseSuperscriptInt(orderStr)
				if d > 0 {
					deg = d
				}
			}
			var val int
			if ident, ok := expr.Right.(*parser.Identifier); ok {
				if symVal, okSym := e.symbols[ident.Value]; okSym {
					val = symVal
				}
			}
			if val == 0 {
				val = e.evalIntExpression(expr.Right)
			}
			res := math.Round(math.Pow(float64(val), 1.0/float64(deg)))
			return int(res)
		}

		if lit == "^" || (lit >= "⁰" && lit <= "⁹") || strings.Contains(lit, "¹") || strings.Contains(lit, "²") || strings.Contains(lit, "³") || strings.Contains(lit, "⁴") || strings.Contains(lit, "⁵") || strings.Contains(lit, "⁶") || strings.Contains(lit, "⁷") || strings.Contains(lit, "⁸") || strings.Contains(lit, "⁹") {
			left := e.evalIntExpression(expr.Left)
			right := e.evalIntExpression(expr.Right)
			return evalArithmetic("^", left, right, lit)
		}

		left := e.evalIntExpression(expr.Left)
		right := e.evalIntExpression(expr.Right)

		switch lit {
		case "+", "-", "*", "/", "%", "^", "√":
			return evalArithmetic(lit, left, right, lit)
		case "and", "or", "xor", "¬", "∧", "∨", "⊕":
			return evalLogical(lit, left, right)
		case "<", ">", "<=", ">=", "≠", "!=", "<>":
			return evalComparison(lit, left, right, expr.Token.Type, lit)
		default:
			if strings.HasSuffix(lit, "√") || strings.Contains(lit, "√") {
				return evalArithmetic("√", left, right, lit)
			}
			if lit == "≠" || strings.Contains(lit, "≠") {
				return evalComparison(lit, left, right, expr.Token.Type, lit)
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
		if strVal, ok := e.stringSymbols[expr.Value]; ok {
			return strVal
		}
		if val, ok := e.symbols[expr.Value]; ok {
			return strconv.Itoa(val)
		}
		return expr.Value
	case *parser.BinaryExpression:
		return strconv.Itoa(e.evalIntExpression(expr))
	}
	if il, ok := node.(*parser.IntegerLiteral); ok {
		return il.Value
	}
	return strconv.Itoa(e.evalIntExpression(node))
}
