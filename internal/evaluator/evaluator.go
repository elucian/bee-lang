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
		// Pre-evaluate all RHS integer values before mutating any symbol, so a
		// parallel assignment (`let a, b := b, a`) reads the pre-swap values.
		// String literals/aliases and stepped ranges live in separate maps and
		// are resolved in the loop below, so this pre-pass only snapshots the
		// integer `symbols` map.
		pendingInts := make([]int, len(s.Values))
		for i, val := range s.Values {
			pendingInts[i] = e.evalIntExpression(val)
		}
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
						// String assignment is a reference alias (immutable strings):
						// `let n := s` shares the source cell's identity so `s is n`
						// and `@s = @n` are true (Decision 12 reference-of).
						if srcID, hasID := e.identities[strIdent.Value]; hasID {
							e.identities[name.Value] = srcID
						} else {
							e.identities[name.Value] = e.ensureIdentity(strIdent.Value)
						}
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

				rightVal := pendingInts[i]
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
	case *parser.CycleStatement:
		e.evalCycleStatement(s)
	case *parser.MatchStatement:
		e.evalMatchStatement(s)
	case *parser.ScopeStatement:
		e.evalScopeStatement(s)
	}
}

// evalCycleStatement dispatches a cycle_stmt (spec/02-statements.md §3.4) over
// the four body-header forms and enforces the volatile-body / stable-prologue
// partition documented in §3.4.
//
// Note: the current Evaluator uses a flat symbol table (no lexical scoping).
// The §3.4 invariant "Volatile body scope — its locals are re-created on every
// iteration" is therefore enforced by convention here (prologue runs once,
// body re-runs each pass) rather than by partitioning the environment. Full
// scope partitioning is deferred to the Phase 5 typechecker (Task 5.3).
func (e *Evaluator) evalCycleStatement(s *parser.CycleStatement) {
	// 1. Prologue runs exactly once for labeled cycles (stable outer scope).
	//    Anonymous cycles have no prologue — skip.
	if s.Prologue != nil {
		for _, stmt := range s.Prologue.Statements {
			if e.exitingRule {
				return
			}
			e.evalStatement(stmt)
		}
	}

	// 2. Iterate the volatile body. The dispatch differs by body-header:
	//    - "do"     : infinite loop until a transfer statement breaks out.
	//    - "while"  : condition-checked loop. Body runs only while Condition
	//                 evaluates non-zero.
	//    - "for"    : indexed domain loop over Range. The §3.4 model iterates
	//                 i through the domain Expression and re-runs Body per
	//                 pass. The volatile body scope means the index binding
	//                 is reset each pass.
	//
	// Per D14 (2026-09-13), inline `repeat`, `stop`, and `redo` are
	// TransferStatements dispatched by evalStatement. `repeat` and `stop`
	// set exitingRule to break out of the loop (jump-to-exit) the same
	// way; `redo` rewinds to the loop-header without advancing the
	// iterator (we approximate this by running the body again without
	// letting the outer for/while advance the count).
	breakOut := false
	switch s.BodyHeader {
	case "do":
		for {
			if e.exitingRule {
				breakOut = true
				break
			}
			if s.Body != nil {
				for _, stmt := range s.Body.Statements {
					if e.exitingRule {
						breakOut = true
						break
					}
					e.evalStatement(stmt)
				}
			}
			if breakOut {
				break
			}
		}
	case "while":
		for {
			if s.Condition != nil && e.evalIntExpression(s.Condition) == 0 {
				break
			}
			if e.exitingRule {
				breakOut = true
				break
			}
			if s.Body != nil {
				for _, stmt := range s.Body.Statements {
					if e.exitingRule {
						breakOut = true
						break
					}
					e.evalStatement(stmt)
				}
			}
			if breakOut {
				break
			}
		}
	case "for":
		// For `for i ∈ expr do`, drive a bounded sequence over the Range
		// expression. The integer-domain case is evaluated directly here;
		// collection / stepped-range domains fall back to set semantics
		// (handled by the IndexExpression evaluator at lookup time).
		if s.Index != nil && s.Range != nil {
			indexName := s.Index.Value
			// Snapshot previous binding (if any) so the volatile scope
			// behaves as documented. Phase 5 typechecker will replace
			// this with a proper scope push/pop.
			prevVal, hadPrev := e.symbols[indexName]
			prevID, hadPrevID := e.identities[indexName]
			deferRestore := func() {
				if hadPrev {
					e.symbols[indexName] = prevVal
				} else {
					delete(e.symbols, indexName)
				}
				if hadPrevID {
					e.identities[indexName] = prevID
				} else {
					delete(e.identities, indexName)
				}
			}
			switch rng := s.Range.(type) {
			case *parser.IntegerLiteral:
				val, _ := strconv.Atoi(rng.Value)
				for iter := 1; iter <= val; iter++ {
					e.symbols[indexName] = iter
					e.ensureIdentity(indexName)
					if s.Body != nil {
						for _, stmt := range s.Body.Statements {
							if e.exitingRule {
								deferRestore()
								return
							}
							e.evalStatement(stmt)
						}
					}
				}
			case *parser.BinaryExpression:
				// Range expression like `0..n-2` (Decision 13). Evaluate
				// the inclusive endpoints and iterate.
				leftVal := e.evalIntExpression(rng.Left)
				rightVal := e.evalIntExpression(rng.Right)
				step := 1
				if leftVal <= rightVal {
					for iter := leftVal; iter <= rightVal; iter += step {
						e.symbols[indexName] = iter
						e.ensureIdentity(indexName)
						if s.Body != nil {
							for _, stmt := range s.Body.Statements {
								if e.exitingRule {
									deferRestore()
									return
								}
								e.evalStatement(stmt)
							}
						}
					}
				} else {
					for iter := leftVal; iter >= rightVal; iter -= step {
						e.symbols[indexName] = iter
						e.ensureIdentity(indexName)
						if s.Body != nil {
							for _, stmt := range s.Body.Statements {
								if e.exitingRule {
									deferRestore()
									return
								}
								e.evalStatement(stmt)
							}
						}
					}
				}
			case *parser.SteppedRangeExpression:
				// Stepped domain `(start..end)(step)` (Decision 13). Materialise
				// the implicit sequence start, start+step, … honouring the
				// endpoint inclusivity of the underlying range operator via
				// inRangePerOp (mirrors steppedRangeValue's bundle rules).
				be, ok := rng.Range.(*parser.BinaryExpression)
				if !ok {
					break
				}
				start := e.evalIntExpression(be.Left)
				end := e.evalIntExpression(be.Right)
				step := e.evalIntExpression(rng.Step)
				if step == 0 {
					step = 1
				}
				// Left-exclusive ranges (`>..`, `>..<`) begin one full step past
				// start so the first valid element is strictly greater.
				firstOffset := 0
				if be.Token.Type == token.RANGE_RGHT_INC || be.Token.Type == token.RANGE_EXCL {
					firstOffset = step
				}
				for v := start + firstOffset; inRangePerOp(be.Token, v, start, end); v += step {
					e.symbols[indexName] = v
					e.ensureIdentity(indexName)
					if s.Body != nil {
						for _, stmt := range s.Body.Statements {
							if e.exitingRule {
								deferRestore()
								return
							}
							e.evalStatement(stmt)
						}
					}
				}
			default:
				// Fallback: single iteration when domain is non-integer.
				if s.Body != nil {
					for _, stmt := range s.Body.Statements {
						if e.exitingRule {
							deferRestore()
							return
						}
						e.evalStatement(stmt)
					}
				}
			}
			deferRestore()
		}
	}

	// 3. `then` clause runs once when the loop terminates naturally
	//    (i.e. not via a transfer statement that set exitingRule).
	if !breakOut && !e.exitingRule && s.ThenBlock != nil {
		for _, stmt := range s.ThenBlock.Statements {
			if e.exitingRule {
				return
			}
			e.evalStatement(stmt)
		}
	}
}

// evalMatchStatement evaluates a `match` selector per spec/02-statements.md
// §3.3. It runs the optional prologue once, then walks the `when` arms in
// order. In "one" mode (default) the first matching arm executes and the
// statement returns; in "all" mode every matching arm executes. If no arm
// matches, the optional `other` fallback runs.
func (e *Evaluator) evalMatchStatement(s *parser.MatchStatement) {
	if s.Prologue != nil {
		e.evalStatement(s.Prologue)
	}

	subject := e.evalIntExpression(s.Subject)
	matched := false
	for _, c := range s.Cases {
		armMatched := false
		for _, target := range c.Targets {
			if e.matchTarget(target, subject) {
				armMatched = true
				break
			}
		}
		if !armMatched {
			continue
		}
		matched = true
		if c.Body != nil {
			e.evalStatement(c.Body)
		}
		if e.exitingRule {
			return
		}
		// "one" mode: stop after the first matching arm.
		if s.Mode != "all" {
			return
		}
	}
	if !matched && s.Other != nil {
		e.evalStatement(s.Other)
	}
}

// matchTarget reports whether a single `when` target matches the subject
// value. A target that is a range binary expression (e.g. `4..10`) matches by
// endpoint-inclusive membership; any other expression matches by value
// equality after evaluating to its integer form.
func (e *Evaluator) matchTarget(target parser.Expression, subject int) bool {
	if be, ok := target.(*parser.BinaryExpression); ok && isRangeSeparatorToken(be.Token) {
		start, end, op, ok := e.rangeBounds(target)
		if ok && inRangePerOp(op, subject, start, end) {
			return true
		}
		return false
	}
	return e.evalIntExpression(target) == subject
}

// evalScopeStatement evaluates a `start` / `with` scope block. The evaluator
// uses a flat symbol table (no lexical partitioning yet — see Task 5.3), so
// the optional prologue and the `do` body simply execute in program order.
func (e *Evaluator) evalScopeStatement(s *parser.ScopeStatement) {
	if s.Keyword == "with" {
		// Qualifier suppression (`with <module> do ... done`) requires a
		// module/object-prefix resolution system that is not implemented yet
		// (spec/02-statements.md §3.1). Surface a graceful runtime error
		// instead of silently treating the qualifier as a no-op.
		mod := "module"
		if id, ok := s.Qualifier.(*parser.Identifier); ok && id.Value != "" {
			mod = id.Value
		}
		fmt.Fprintf(os.Stderr, "runtime error: %s module not implemented yet (line %d)\n", mod, int(s.Token.Pos))
		os.Exit(1)
	}
	if s.Prologue != nil {
		e.evalStatement(s.Prologue)
	}
	if s.Body != nil {
		e.evalStatement(s.Body)
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
		// String symbols carry a stable reference identity too (Decision 12
		// reference-of `@`): an identifier bound to a string still resolves to
		// its cell identity so `@s = @n` / `s is n` compare references.
		if _, ok := e.stringSymbols[expr.Value]; ok {
			return 0, e.ensureIdentity(expr.Value)
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
	case *parser.TernaryExpression:
		// Parenthesised conditional selector (spec/02 §3.2): evaluate the
		// condition, then yield only the chosen branch — the untaken branch
		// is never evaluated (short-circuit).
		if e.evalIntExpression(expr.Condition) != 0 {
			return e.evalIntExpressionWithID(expr.Then)
		}
		return e.evalIntExpressionWithID(expr.Else)
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
		if opLit == "@" || expr.Token.Type == token.AT {
			// Reference-of (Decision 12): yield the operand's stable identity ID
			// as its value so `@a = @b` compares references (⇔ `a is b`).
			_, id := e.evalIntExpressionWithID(expr.Right)
			return id, e.allocID()
		}
		if opLit == "!" || opLit == "not" || expr.Token.Type == token.LOGICAL_NOT || expr.Token.Type == token.NOT {
			rightVal, _ := e.evalIntExpressionWithID(expr.Right)
			if rightVal == 0 {
				return 1, e.allocID()
			}
			return 0, e.allocID()
		}
		if opLit == "-" || expr.Token.Type == token.MINUS {
			// Unary negation: negate the operand's value. Each evaluation yields
			// a fresh ephemeral identity (Decision 2), matching literal semantics.
			rightVal, _ := e.evalIntExpressionWithID(expr.Right)
			return -rightVal, e.allocID()
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
