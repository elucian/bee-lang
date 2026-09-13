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
	// steppedRanges stores SteppedRangeExpression ASTs by identifier name
	// so that subsequent `(ident)[i]` indexing can materialise the i-th
	// element. Decision 13 (D13, 2026-09-13).
	steppedRanges map[string]*parser.SteppedRangeExpression
	identities    map[string]int
	nextID        int
	debug         bool
	exitingRule   bool
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
		steppedRanges: make(map[string]*parser.SteppedRangeExpression),
		identities:    make(map[string]int),
		nextID:        1,
	}
}

func (e *Evaluator) allocID() int {
	id := e.nextID
	e.nextID++
	return id
}

// ensureIdentity mints a stable identity ID for a named binding (variable or
// parameter). The ID is preserved across reads so `a is a` returns true;
// reassignments reuse the same ID so `let a := 5; let a := 6` keeps the cell
// stable. This is the runtime hook for Decision 2 / 7.
func (e *Evaluator) ensureIdentity(name string) int {
	if id, ok := e.identities[name]; ok {
		return id
	}
	id := e.allocID()
	e.identities[name] = id
	return id
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
		prevExiting := e.exitingRule
		e.exitingRule = false
		for _, stmt := range s.Body.Statements {
			if e.exitingRule {
				break
			}
			e.evalStatement(stmt)
		}
		e.exitingRule = prevExiting
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
					e.ensureIdentity(name.Value)
					e.debugLog("EVALUATOR DEBUG: Assigned string %s = %q\n", name.Value, strLit.Value)
					continue
				} else if strIdent, ok := s.Values[i].(*parser.Identifier); ok {
					if strVal, hasStr := e.stringSymbols[strIdent.Value]; hasStr {
						e.stringSymbols[name.Value] = strVal
						delete(e.symbols, name.Value)
						e.ensureIdentity(name.Value)
						e.debugLog("EVALUATOR DEBUG: Assigned string %s = %q\n", name.Value, strVal)
						continue
					}
				}

				// Decision 13 (D13, 2026-09-13): when the value expression
				// is a SteppedRangeExpression, bind it under the variable
				// name so subsequent `(name)[i]` indexing materialises the
				// i-th element. Native int domain only — the step is
				// evaluated to its int form.
				if sre, ok := s.Values[i].(*parser.SteppedRangeExpression); ok {
					e.steppedRanges[name.Value] = sre
					delete(e.symbols, name.Value)
					e.ensureIdentity(name.Value)
					e.debugLog("EVALUATOR DEBUG: Bound stepped range to %s (start=%d, end=%d, step=%d)\n",
						name.Value,
						e.evalIntExpression(sre.Range.(*parser.BinaryExpression).Left),
						e.evalIntExpression(sre.Range.(*parser.BinaryExpression).Right),
						e.evalIntExpression(sre.Step),
					)
					continue
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
					e.ensureIdentity(name.Value)
					e.debugLog("EVALUATOR DEBUG: Result %s = %d\n", name.Value, e.symbols[name.Value])
				case "-=":
					e.symbols[name.Value] -= rightVal
					e.ensureIdentity(name.Value)
				case "*=":
					e.symbols[name.Value] *= rightVal
					e.ensureIdentity(name.Value)
				case "/=":
					if rightVal != 0 {
						e.symbols[name.Value] /= rightVal
					}
					e.ensureIdentity(name.Value)
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
					e.ensureIdentity(name.Value)
				case "^=":
					e.symbols[name.Value] = int(math.Pow(float64(e.symbols[name.Value]), float64(rightVal)))
					e.ensureIdentity(name.Value)
				case "√=":
					deg := rightVal
					if deg <= 0 {
						deg = 2
					}
					e.symbols[name.Value] = int(math.Round(math.Pow(float64(e.symbols[name.Value]), 1.0/float64(deg))))
					e.ensureIdentity(name.Value)
				case ":=", "=":
					e.symbols[name.Value] = rightVal
					e.ensureIdentity(name.Value)
				default:
					e.symbols[name.Value] = rightVal
					e.ensureIdentity(name.Value)
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
					e.ensureIdentity(name)
				} else if arrLit, ok := s.Values[i].(*parser.ArrayLiteral); ok {
					elems := make([]int, len(arrLit.Elements))
					for j, el := range arrLit.Elements {
						elems[j] = e.evalIntExpression(el)
					}
					e.arrayValues[name] = elems
					e.ensureIdentity(name)
				} else {
					val := e.evalIntExpression(s.Values[i])
					e.symbols[name] = val
					e.ensureIdentity(name)
				}
			} else if s.Value != nil && len(names) == 1 {
				if strLit, ok := s.Value.(*parser.StringLiteral); ok {
					e.stringSymbols[name] = strLit.Value
					e.ensureIdentity(name)
				} else if arrLit, ok := s.Value.(*parser.ArrayLiteral); ok {
					elems := make([]int, len(arrLit.Elements))
					for j, el := range arrLit.Elements {
						elems[j] = e.evalIntExpression(el)
					}
					e.arrayValues[name] = elems
					e.ensureIdentity(name)
				} else {
					val := e.evalIntExpression(s.Value)
					e.symbols[name] = val
					e.ensureIdentity(name)
				}
			} else {
				e.symbols[name] = 0
				e.ensureIdentity(name)
			}
		}
	case *parser.TransferStatement:
		// Decision: `over` and other early-exit transfer statements set the
		// exitingRule flag so the canonical §16.2 semantics apply — the
		// enclosing rule's body short-circuits immediately.
		if s.Keyword == "over" || s.Keyword == "stop" || s.Keyword == "exit" || s.Keyword == "abort" || s.Keyword == "panic" {
			e.exitingRule = true
		}
	}
}

func (e *Evaluator) evalIntExpression(node parser.Expression) int {
	val, _ := e.evalIntExpressionWithID(node)
	return val
}

// evalIntExpressionWithID evaluates an integer-typed expression and also returns
// the identity ID stamp (Decision 2 / Decision 7 source-of-truth). The ID is
// stable across symbol reads so `a is a` returns true; each IntegerLiteral
// node contributes a fresh ephemeral ID so literals never share identity
// (matching the spec's pointer-identity semantics).
func (e *Evaluator) evalIntExpressionWithID(node parser.Expression) (int, int) {
	switch expr := node.(type) {
	case *parser.IntegerLiteral:
		val, _ := strconv.Atoi(expr.Value)
		return val, e.allocID()
	case *parser.Identifier:
		if expr.Value == "$" {
			return 0, e.allocID()
		}
		if expr.Value == "True" || expr.Value == "true" {
			return 1, e.allocID()
		}
		if expr.Value == "False" || expr.Value == "false" {
			return 0, e.allocID()
		}
		if val, ok := e.symbols[expr.Value]; ok {
			return val, e.ensureIdentity(expr.Value)
		}
		return 0, e.allocID()
	case *parser.IndexExpression:
		// Decision 13 (D13, 2026-09-13): if the indexed expression is a
		// SteppedRangeExpression directly, or an identifier bound to a
		// stepped range, materialise the i-th element of the stepped
		// sequence (1-based, Decision 1). Otherwise fall back to the
		// existing identifier-array indexing semantics.
		if sre, ok := expr.Left.(*parser.SteppedRangeExpression); ok {
			idx := e.evalIntExpression(expr.Index)
			val := e.steppedRangeValue(sre, idx)
			return val, e.allocID()
		}
		if ident, ok := expr.Left.(*parser.Identifier); ok {
			if sre, hasSRE := e.steppedRanges[ident.Value]; hasSRE {
				idx := e.evalIntExpression(expr.Index)
				val := e.steppedRangeValue(sre, idx)
				return val, e.ensureIdentity(ident.Value)
			}
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
					return arr[idx-1], e.ensureIdentity(ident.Value)
				}
			}
		}
		return 0, e.allocID()
	case *parser.PrefixExpression:
		// D10 radical dispatch (`²√`, `³√`, …, `√`) and Decision 7 / D7
		// logical-not dispatch (`¬`). The PrefixExpression node replaces the
		// previous BinaryExpression-with-"0"-Left hack that masked the
		// precedence bug in issue 17-radical-precedence.md.
		opLit := expr.Operator
		if strings.HasSuffix(opLit, "√") || strings.Contains(opLit, "√") || opLit == "√" {
			rightVal, _ := e.evalIntExpressionWithID(expr.Right)
			if ident, ok := expr.Right.(*parser.Identifier); ok {
				if symVal, okSym := e.symbols[ident.Value]; okSym {
					rightVal = symVal
				}
			}
			return evalSqrt(opLit, rightVal), e.allocID()
		}
		if opLit == "!" || opLit == "not" || expr.Token.Type == token.LOGICAL_NOT || expr.Token.Type == token.NOT {
			rightVal, _ := e.evalIntExpressionWithID(expr.Right)
			if rightVal == 0 {
				return 1, e.allocID()
			}
			return 0, e.allocID()
		}
		// Unknown prefix — fall back to the underlying value so a malformed
		// expression reports a value rather than panicking.
		rightVal, _ := e.evalIntExpressionWithID(expr.Right)
		return rightVal, e.allocID()
	case *parser.BinaryExpression:
		lit := expr.Token.Literal
		// Identity operators (Decision 2 — pointer identity)
		if expr.Token.Type == token.IS || expr.Token.Type == token.IS_NOT || lit == "is" || lit == "is not" {
			_, leftID := e.evalIntExpressionWithID(expr.Left)
			_, rightID := e.evalIntExpressionWithID(expr.Right)
			if expr.Token.Type == token.IS_NOT || lit == "is not" {
				if leftID != rightID {
					return 1, 0
				}
				return 0, 0
			}
			if leftID == rightID {
				return 1, 0
			}
			return 0, 0
		}
		if lit == "=" || expr.Token.Type == token.EQ || lit == "==" {
			leftVal, _ := e.evalIntExpressionWithID(expr.Left)
			rightVal, _ := e.evalIntExpressionWithID(expr.Right)
			return evalComparison(lit, leftVal, rightVal, expr.Token.Type, lit), 0
		}
		if expr.Token.Type == token.NEQ || expr.Token.Type == token.NOT_EQ || lit == "¬" || lit == "!=" || lit == "<>" || expr.Token.Type == token.NEQ_UNICODE || lit == "≠" || strings.Contains(lit, "≠") {
			leftVal, _ := e.evalIntExpressionWithID(expr.Left)
			rightVal, _ := e.evalIntExpressionWithID(expr.Right)
			return evalComparison(lit, leftVal, rightVal, expr.Token.Type, lit), 0
		}
		if lit == "in" || lit == "∈" || expr.Token.Type == token.IN_OP || expr.Token.Type == token.IN_KEYWORD {
			leftVal, _ := e.evalIntExpressionWithID(expr.Left)
			if binRight, ok := expr.Right.(*parser.BinaryExpression); ok && isRangeSeparatorToken(binRight.Token) {
				start, _ := e.evalIntExpressionWithID(binRight.Left)
				end, _ := e.evalIntExpressionWithID(binRight.Right)
				if inRangePerOp(binRight.Token, leftVal, start, end) {
					return 1, 0
				}
				return 0, 0
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
				val, _ = e.evalIntExpressionWithID(expr.Right)
			}
			res := math.Round(math.Pow(float64(val), 1.0/float64(deg)))
			return int(res), 0
		}

		if lit == "^" || (lit >= "⁰" && lit <= "⁹") || strings.Contains(lit, "¹") || strings.Contains(lit, "²") || strings.Contains(lit, "³") || strings.Contains(lit, "⁴") || strings.Contains(lit, "⁵") || strings.Contains(lit, "⁶") || strings.Contains(lit, "⁷") || strings.Contains(lit, "⁸") || strings.Contains(lit, "⁹") {
			left, _ := e.evalIntExpressionWithID(expr.Left)
			right, _ := e.evalIntExpressionWithID(expr.Right)
			return evalArithmetic("^", left, right, lit), 0
		}

		left, _ := e.evalIntExpressionWithID(expr.Left)
		right, _ := e.evalIntExpressionWithID(expr.Right)

		switch lit {
		case "+", "-", "*", "/", "%", "^", "√":
			return evalArithmetic(lit, left, right, lit), 0
		case "and", "or", "xor", "¬", "∧", "∨", "⊕":
			return evalLogical(lit, left, right), 0
		case "<", ">", "<=", ">=", "≠", "!=", "<>":
			return evalComparison(lit, left, right, expr.Token.Type, lit), 0
		default:
			if strings.HasSuffix(lit, "√") || strings.Contains(lit, "√") {
				return evalArithmetic("√", left, right, lit), 0
			}
			if lit == "≠" || strings.Contains(lit, "≠") {
				return evalComparison(lit, left, right, expr.Token.Type, lit), 0
			}
			return evalArithmetic(lit, left, right, lit), 0
		}
	}
	return 0, 0
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

// isRangeSeparatorToken is true for tokens that act as range separators:
// RANGE_INCL (`..`), RANGE_LEFT_INC (`..<`), RANGE_RGHT_INC (`>..`),
// RANGE_EXCL (`>..<`). Created for Decision 13 (D13, 2026-09-13).
func isRangeSeparatorToken(tok token.Token) bool {
	switch tok.Type {
	case token.RANGE_INCL, token.RANGE_LEFT_INC, token.RANGE_RGHT_INC, token.RANGE_EXCL:
		return true
	}
	return false
}

// inRangePerOp returns true when v belongs to the range [start, end] under
// the endpoint-inclusion semantics of the given range separator token.
//   - RANGE_INCL     `..`   : left-inclusive, right-inclusive  [a, b]
//   - RANGE_LEFT_INC `..<`  : left-inclusive, right-exclusive  [a, b)
//   - RANGE_RGHT_INC `>..`  : left-exclusive, right-inclusive  (a, b]
//   - RANGE_EXCL     `>..<` : fully exclusive                  (a, b)
//
// Decision 13 (D13, 2026-09-13) canonicalised this dispatch table.
func inRangePerOp(tok token.Token, v, start, end int) bool {
	switch tok.Type {
	case token.RANGE_LEFT_INC:
		return v >= start && v < end
	case token.RANGE_RGHT_INC:
		return v > start && v <= end
	case token.RANGE_EXCL:
		return v > start && v < end
	default:
		// RANGE_INCL plus any literal-form fallback for older binaries.
		return v >= start && v <= end
	}
}

// rangeBounds extracts (start, end, op) from a binary expression whose
// operator is a range separator. Returns ok=false when `node` is not a
// range binary expression. Start/end are evaluated against the
// evaluator's int domain. Decision 13 (D13, 2026-09-13).
func (e *Evaluator) rangeBounds(node parser.Expression) (start, end int, op token.Token, ok bool) {
	be, isBin := node.(*parser.BinaryExpression)
	if !isBin || !isRangeSeparatorToken(be.Token) {
		return 0, 0, token.Token{}, false
	}
	start = e.evalIntExpression(be.Left)
	end = e.evalIntExpression(be.Right)
	return start, end, be.Token, true
}

// steppedRangeValue returns the i-th (1-based) element of sre. The
// starting offset is `step` when the range's left endpoint is exclusive
// (operators `>..` and `>..<`); otherwise the offset is 0 (operators `..`
// and `..<`). Then the i-th element is `start + offset + (i-1)*step`,
// clipped to the endpoint inclusivity of the underlying range operator.
//
// Returns 0 (with a debug log) when i falls outside the valid index
// range or the computed value sits outside the operator's bundle.
// Decision 13 (D13, 2026-09-13).
func (e *Evaluator) steppedRangeValue(sre *parser.SteppedRangeExpression, i int) int {
	be := sre.Range.(*parser.BinaryExpression)
	start := e.evalIntExpression(be.Left)
	end := e.evalIntExpression(be.Right)
	step := e.evalIntExpression(sre.Step)
	if step == 0 {
		step = 1
	}
	if i <= 0 {
		e.debugLog("EVALUATOR DEBUG: steppedRangeValue index %d out of bounds (must be ≥ 1, Decision 1)\n", i)
		return 0
	}
	// Left-exclusive ranges (`>..`, `>..<`) shift the starting offset by
	// one full step so the sequence's first valid element is the smallest
	// value strictly greater than `start`.
	leftOffset := 0
	switch be.Token.Type {
	case token.RANGE_RGHT_INC, token.RANGE_EXCL:
		leftOffset = step
	}
	val := start + leftOffset + (i-1)*step
	switch be.Token.Type {
	case token.RANGE_INCL:
		if val < start || val > end {
			return 0
		}
	case token.RANGE_LEFT_INC:
		if val < start || val >= end {
			return 0
		}
	case token.RANGE_RGHT_INC:
		if val <= start || val > end {
			return 0
		}
	case token.RANGE_EXCL:
		if val <= start || val >= end {
			return 0
		}
	default:
		if val < start || val > end {
			return 0
		}
	}
	return val
}
