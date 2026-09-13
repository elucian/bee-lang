package evaluator

import (
	"bee/internal/parser"
	"bee/internal/token"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// BoxedCell is a heap-allocated, mutable storage cell for closure state per
// spec/03-rules.md §5.4. A variable is boxed when declared with the `[...]`
// operator (`set .count := [start];`), which makes it persist across calls
// of the enclosing rule's closures. Unboxed `set` variables stay immutable
// in the flat symbol table.
type BoxedCell struct {
	Value int
}

// closureObject is a heap-allocated closure instance returned by a state
// generator rule (spec/03 §5.4). It owns the generator's boxed cells (shared
// by reference with every nested method rule) and the nested rule table that
// methods like `.next` resolve against. Because Nested maps directly to the
// parser's nested RuleStatement nodes and Cells holds *BoxedCell pointers,
// mutations performed inside a method call are visible to every subsequent
// method call on the same object.
type closureObject struct {
	Cells  map[string]*BoxedCell            // boxed state fields (`.count`)
	Nested map[string]*parser.RuleStatement // member rules keyed by folded name (`.next`)
}

type Evaluator struct {
	symbols       map[string]int
	stringSymbols map[string]string
	arrayValues   map[string][]int
	// listValues, setValues, mapValues store collection literals bound to
	// identifiers (spec/10 §1). All values are stored as []int for the
	// bootstrap evaluator; sets are deduplicated and sorted.
	listValues map[string][]int
	setValues  map[string][]int
	mapValues  map[string]map[string]int
	// steppedRanges stores SteppedRangeExpression ASTs by identifier name
	// so that subsequent `(ident)[i]` indexing can materialise the i-th
	// element. Decision 13 (D13, 2026-09-13).
	steppedRanges map[string]*parser.SteppedRangeExpression
	identities    map[string]int
	nextID        int
	debug         bool
	exitingRule   bool
	// ruleRegistry indexes every *parser.RuleStatement by name so that
	// CallExpression can resolve rule invocations (spec/03 §3.1).
	ruleRegistry map[string]*parser.RuleStatement
	// boxedCells holds the *current* rule frame's boxed state cells
	// (spec/03 §5.4). It is empty at the top level and swapped in/out by
	// callRule so nested method invocations share the closure object's cells.
	boxedCells map[string]*BoxedCell
	// closureObjects binds variable names to heap-allocated closure objects so
	// `c.next()` can resolve the object `c` and dispatch to its `.next` method.
	closureObjects map[string]*closureObject
	// lambdaValues binds names to pure lambda expressions (spec/07 §2.1) —
	// the first-class `L` values that CallExpression resolves before rules.
	lambdaValues map[string]*parser.LambdaExpression
	// inLambda tracks lambda invocation depth so the evaluator can enforce
	// the spec/07 §3 purity invariants (E0701: a lambda cannot call a rule;
	// E0704: a lambda cannot reference outer variables).
	inLambda int
	// dollarLen / hasDollar bind the `$` end-anchor (Decision 1) to the
	// length of the collection currently being indexed. When hasDollar is
	// true, evaluating the identifier `$` yields dollarLen (1-based last
	// index), enabling arithmetic anchors like `a[$-1]` (spec/10 §3.2).
	dollarLen int
	hasDollar bool
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
		symbols:        make(map[string]int),
		stringSymbols:  make(map[string]string),
		arrayValues:    make(map[string][]int),
		listValues:     make(map[string][]int),
		setValues:      make(map[string][]int),
		mapValues:      make(map[string]map[string]int),
		steppedRanges:  make(map[string]*parser.SteppedRangeExpression),
		identities:     make(map[string]int),
		nextID:         1,
		ruleRegistry:   make(map[string]*parser.RuleStatement),
		boxedCells:     make(map[string]*BoxedCell),
		closureObjects: make(map[string]*closureObject),
		lambdaValues:   make(map[string]*parser.LambdaExpression),
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
	// Pass 1: register all rule definitions so CallExpression can resolve
	// invocations regardless of definition order (spec/03 §5.2).
	for _, stmt := range program.Statements {
		if rs, ok := stmt.(*parser.RuleStatement); ok {
			e.ruleRegistry[rs.Name] = rs
		}
	}
	// Pass 2: execute top-level statements (rule bodies are only entered
	// via CallExpression or the implicit `main` entry point).
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
		// Only `main` is executed at the top level (spec/03 §1). Other rules
		// are invoked exclusively via CallExpression → callRule.
		if s.Name != "main" {
			return
		}
		e.debugLog("EVALUATOR DEBUG: entering rule main\n")
		prevExiting := e.exitingRule
		e.exitingRule = false
		for _, stmt := range s.Body.Statements {
			if e.exitingRule {
				break
			}
			e.evalStatement(stmt)
		}
		e.exitingRule = prevExiting
		e.debugLog("EVALUATOR DEBUG: exited rule main\n")
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
		// spec/10 §3.2 / §3.5: the LHS targets are full expressions. Index
		// targets (`let a[i] := …`, `let map["k"] := …`) mutate an element of a
		// stored collection with 1-based bounds enforcement (E1001 / E1006).
		// Identifier targets follow the legacy scalar path. Targets is
		// authoritative and positionally aligned with Values; Names mirrors
		// only the identifier / dotted targets for legacy consumers.
		for i, target := range s.Targets {
			if idxExpr, isIndex := target.(*parser.IndexExpression); isIndex {
				ident, ok := idxExpr.Left.(*parser.Identifier)
				if !ok {
					continue
				}
				if m, hasMap := e.mapValues[ident.Value]; hasMap {
					key := e.evalExpression(idxExpr.Index)
					if i < len(pendingInts) {
						m[key] = pendingInts[i]
					}
					continue
				}
				var idx int
				if arr, hasArr := e.arrayValues[ident.Value]; hasArr {
					idx = e.evalIndexExpr(idxExpr.Index, len(arr))
				}
				if arr, hasArr := e.arrayValues[ident.Value]; hasArr {
					if idx == 0 {
						fmt.Fprintf(os.Stderr, "[ERROR] E1006 ZeroBasedIndexAttempt: index 0 at line %d\n", int(idxExpr.Token.Pos))
						continue
					}
					if idx < 0 {
						fmt.Fprintf(os.Stderr, "[ERROR] E1001 NegativeIndex: index %d is negative; use a[$-n] for a relative end-anchor at line %d\n", idx, int(idxExpr.Token.Pos))
						continue
					}
					if idx < 1 || idx > len(arr) {
						fmt.Fprintf(os.Stderr, "[ERROR] E1001 IndexOutOfBounds: index %d out of range 1..%d at line %d\n", idx, len(arr), int(idxExpr.Token.Pos))
						continue
					}
					if i < len(pendingInts) {
						arr[idx-1] = pendingInts[i]
					}
				}
				continue
			}

			name, isIdent := target.(*parser.Identifier)
			if !isIdent {
				if me, isMember := target.(*parser.MemberExpression); isMember && me.Base == nil && len(me.Parts) == 1 {
					name = &parser.Identifier{Token: me.Token, Value: me.Parts[0]}
				} else {
					continue
				}
			}
			e.debugLog("EVALUATOR DEBUG: Assigning to %s\n", name.Value)
			// Boxed closure state mutation (spec/03 §5.4): `let .field op= expr;`
			// mutates the heap-allocated cell on the current closure frame.
			if strings.HasPrefix(name.Value, ".") {
				cell, hasCell := e.boxedCells[name.Value]
				if !hasCell {
					cell = &BoxedCell{}
					e.boxedCells[name.Value] = cell
				}
				rhs := pendingInts[i]
				switch s.Token.Literal {
				case "+=":
					cell.Value += rhs
				case "-=":
					cell.Value -= rhs
				case "*=":
					cell.Value *= rhs
				case "/=":
					if rhs != 0 {
						cell.Value /= rhs
					}
				case "%=":
					if rhs != 0 {
						cell.Value %= rhs
					}
				default: // ":=", "="
					cell.Value = rhs
				}
				continue
			}
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
		// Multi-result deconstruction (spec/03 §2.3): when a single
		// CallExpression is assigned to multiple names, invoke once and
		// distribute results positionally.
		if len(names) > 1 && len(s.Values) == 1 {
			if call, ok := s.Values[0].(*parser.CallExpression); ok {
				results := e.callRule(call, nil)
				for i, name := range names {
					if i < len(results) {
						e.bindResultValue(name, results[i])
					} else {
						e.symbols[name] = 0
						e.ensureIdentity(name)
					}
				}
				break
			}
		}
		// Single-result rule call (spec/03 §3.1): `new r := rule_call(...)`
		// where the CallExpression is the sole value for a single name.
		if len(names) == 1 && len(s.Values) == 1 {
			if call, ok := s.Values[0].(*parser.CallExpression); ok {
				results := e.callRule(call, nil)
				if len(results) > 0 {
					e.bindResultValue(names[0], results[0])
				} else {
					e.symbols[names[0]] = 0
					e.ensureIdentity(names[0])
				}
				break
			}
		}
		for i, name := range names {
			e.debugLog("EVALUATOR DEBUG: Declaring %s\n", name)
			// Boxed closure state (spec/03 §5.4): `set .field := [expr];` boxes the
			// value into a heap-allocated, mutable cell on the current closure
			// frame so nested method rules can mutate it persistently.
			if strings.HasPrefix(name, ".") {
				if i < len(s.Values) && s.Values[i] != nil {
					if arrLit, isArr := s.Values[i].(*parser.ArrayLiteral); isArr && len(arrLit.Elements) > 0 {
						e.boxedCells[name] = &BoxedCell{Value: e.evalIntExpression(arrLit.Elements[0])}
						continue
					}
					e.boxedCells[name] = &BoxedCell{Value: e.evalIntExpression(s.Values[i])}
					continue
				}
				e.boxedCells[name] = &BoxedCell{}
				continue
			}
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
				} else if listLit, ok := s.Values[i].(*parser.ListLiteral); ok {
					elems := make([]int, len(listLit.Elements))
					for j, el := range listLit.Elements {
						elems[j] = e.evalIntExpression(el)
					}
					e.listValues[name] = elems
					e.ensureIdentity(name)
				} else if setLit, ok := s.Values[i].(*parser.SetLiteral); ok {
					elems := make([]int, len(setLit.Elements))
					for j, el := range setLit.Elements {
						elems[j] = e.evalIntExpression(el)
					}
					e.setValues[name] = dedupSortInts(elems)
					e.ensureIdentity(name)
				} else if mapLit, ok := s.Values[i].(*parser.MapLiteral); ok {
					m := make(map[string]int)
					for _, pair := range mapLit.Pairs {
						key := e.evalExpression(pair.Key)
						m[key] = e.evalIntExpression(pair.Value)
					}
					e.mapValues[name] = m
					e.ensureIdentity(name)
				} else if lambdaExpr, ok := s.Values[i].(*parser.LambdaExpression); ok {
					// First-class lambda value (spec/07 §2.3): bind in the lambda
					// registry so CallExpression dispatch can find it.
					e.lambdaValues[name] = lambdaExpr
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
				} else if listLit, ok := s.Value.(*parser.ListLiteral); ok {
					elems := make([]int, len(listLit.Elements))
					for j, el := range listLit.Elements {
						elems[j] = e.evalIntExpression(el)
					}
					e.listValues[name] = elems
					e.ensureIdentity(name)
				} else if setLit, ok := s.Value.(*parser.SetLiteral); ok {
					elems := make([]int, len(setLit.Elements))
					for j, el := range setLit.Elements {
						elems[j] = e.evalIntExpression(el)
					}
					e.setValues[name] = dedupSortInts(elems)
					e.ensureIdentity(name)
				} else if mapLit, ok := s.Value.(*parser.MapLiteral); ok {
					m := make(map[string]int)
					for _, pair := range mapLit.Pairs {
						key := e.evalExpression(pair.Key)
						m[key] = e.evalIntExpression(pair.Value)
					}
					e.mapValues[name] = m
					e.ensureIdentity(name)
				} else if lambdaExpr, ok := s.Value.(*parser.LambdaExpression); ok {
					e.lambdaValues[name] = lambdaExpr
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
			// Within an index expression the `$` end-anchor resolves to the
			// current collection's length (1-based last index, Decision 1),
			// enabling arithmetic anchors like `a[$-1]`. Outside that
			// context it degenerates to 0 (legacy behaviour).
			if e.hasDollar {
				return e.dollarLen, e.allocID()
			}
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
		// E0704 UnboundLambdaVariable (spec/07 §3 / §7): inside a pure lambda
		// frame only parameters are in scope — anything else is unbound.
		if e.inLambda > 0 {
			fmt.Fprintf(os.Stderr, "[ERROR] E0704 UnboundLambdaVariable: %q at line %d\n", expr.Value, int(expr.Token.Pos))
		}
		return 0, e.allocID()
	case *parser.CallExpression:
		// Lambda invocation (spec/07 §6 lambda_call): first-class `L` values
		// take dispatch precedence — a rule CAN call a lambda (§3 invariant 4).
		if lambdaExpr, isLambda := e.lambdaValues[expr.Name]; isLambda {
			return e.callLambda(lambdaExpr, expr), e.allocID()
		}
		// E0701 RuleCallInLambda (spec/07 §3 invariant 4 / §7): a lambda
		// CANNOT call a stateful rule.
		if e.inLambda > 0 {
			if _, isRule := e.ruleRegistry[expr.Name]; isRule {
				fmt.Fprintf(os.Stderr, "[ERROR] E0701 RuleCallInLambda: %q at line %d\n", expr.Name, int(expr.Token.Pos))
				return 0, e.allocID()
			}
		}
		// Rule invocation (spec/03 §3.1): resolve the rule in the registry,
		// bind arguments in a scoped call frame, execute the body, and
		// return the first declared result (single-result capture).
		results := e.callRule(expr, nil)
		if len(results) > 0 {
			if iv, isInt := results[0].(int); isInt {
				return iv, e.allocID()
			}
			return 0, e.allocID()
		}
		return 0, e.allocID()
	case *parser.MemberExpression:
		// Member access (spec/03 §5.4). Leading-dot reads a boxed cell on the
		// current closure frame; `obj.member()` invokes a closure method.
		if expr.Base == nil && len(expr.Parts) == 1 {
			if cell, okCell := e.boxedCells[expr.Parts[0]]; okCell {
				return cell.Value, e.allocID()
			}
			return 0, e.allocID()
		}
		if expr.IsCall {
			if val, okVal := e.callMember(expr); okVal {
				if iv, isInt := val.(int); isInt {
					return iv, e.allocID()
				}
			}
			return 0, e.allocID()
		}
		// List member access: `.head` returns first element, `.tail` returns
		// count of remaining elements (spec/10 §3.1 list operations).
		if ident, isIdent := expr.Base.(*parser.Identifier); isIdent && len(expr.Parts) == 1 {
			if lst, ok := e.listValues[ident.Value]; ok {
				switch expr.Parts[0] {
				case "head":
					if len(lst) > 0 {
						return lst[0], e.ensureIdentity(ident.Value)
					}
					return 0, e.allocID()
				case "tail":
					if len(lst) > 1 {
						return len(lst) - 1, e.ensureIdentity(ident.Value)
					}
					return 0, e.allocID()
				}
			}
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
			if arr, ok := e.arrayValues[ident.Value]; ok {
				idx = e.evalIndexExpr(expr.Index, len(arr))
			}
			if arr, ok := e.arrayValues[ident.Value]; ok {
				if idx == 0 {
					fmt.Fprintf(os.Stderr, "[ERROR] E1006 ZeroBasedIndexAttempt: index 0 at line %d\n", int(expr.Token.Pos))
					return 0, e.allocID()
				}
				if idx < 0 {
					fmt.Fprintf(os.Stderr, "[ERROR] E1001 NegativeIndex: index %d is negative; use a[$-n] for a relative end-anchor at line %d\n", idx, int(expr.Token.Pos))
					return 0, e.allocID()
				}
				if idx < 1 || idx > len(arr) {
					fmt.Fprintf(os.Stderr, "[ERROR] E1001 IndexOutOfBounds: index %d out of range 1..%d at line %d\n", idx, len(arr), int(expr.Token.Pos))
					return 0, e.allocID()
				}
				return arr[idx-1], e.ensureIdentity(ident.Value)
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
		case "<", ">", "<=", ">=", "≤", "≥", "≠", "!=", "<>":
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
			if arr, ok := e.arrayValues[ident.Value]; ok {
				idx = e.evalIndexExpr(expr.Index, len(arr))
			}
			if arr, ok := e.arrayValues[ident.Value]; ok {
				if idx == 0 {
					fmt.Fprintf(os.Stderr, "[ERROR] E1006 ZeroBasedIndexAttempt: index 0 at line %d\n", int(expr.Token.Pos))
					return "0"
				}
				if idx < 0 {
					fmt.Fprintf(os.Stderr, "[ERROR] E1001 NegativeIndex: index %d is negative; use a[$-n] for a relative end-anchor at line %d\n", idx, int(expr.Token.Pos))
					return "0"
				}
				if idx < 1 || idx > len(arr) {
					fmt.Fprintf(os.Stderr, "[ERROR] E1001 IndexOutOfBounds: index %d out of range 1..%d at line %d\n", idx, len(arr), int(expr.Token.Pos))
					return "0"
				}
				return strconv.Itoa(arr[idx-1])
			}
		}
		return "0"
	case *parser.Identifier:
		if strVal, ok := e.stringSymbols[expr.Value]; ok {
			return strVal
		}
		if arr, ok := e.arrayValues[expr.Value]; ok {
			parts := make([]string, len(arr))
			for i, v := range arr {
				parts[i] = strconv.Itoa(v)
			}
			return "[" + strings.Join(parts, ", ") + "]"
		}
		if lst, ok := e.listValues[expr.Value]; ok {
			parts := make([]string, len(lst))
			for i, v := range lst {
				parts[i] = strconv.Itoa(v)
			}
			return "(" + strings.Join(parts, ", ") + ")"
		}
		if st, ok := e.setValues[expr.Value]; ok {
			parts := make([]string, len(st))
			for i, v := range st {
				parts[i] = strconv.Itoa(v)
			}
			return "{" + strings.Join(parts, ", ") + "}"
		}
		if m, ok := e.mapValues[expr.Value]; ok {
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			parts := make([]string, 0, len(m))
			for _, k := range keys {
				parts = append(parts, k+": "+strconv.Itoa(m[k]))
			}
			return "{" + strings.Join(parts, ", ") + "}"
		}
		if val, ok := e.symbols[expr.Value]; ok {
			return strconv.Itoa(val)
		}
		return expr.Value
	case *parser.SetLiteral:
		elems := make([]int, len(expr.Elements))
		for i, el := range expr.Elements {
			elems[i] = e.evalIntExpression(el)
		}
		sorted := dedupSortInts(elems)
		parts := make([]string, len(sorted))
		for i, v := range sorted {
			parts[i] = strconv.Itoa(v)
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *parser.ListLiteral:
		parts := make([]string, len(expr.Elements))
		for i, el := range expr.Elements {
			parts[i] = strconv.Itoa(e.evalIntExpression(el))
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case *parser.MapLiteral:
		keys := make([]string, 0, len(expr.Pairs))
		vals := make(map[string]int)
		for _, pair := range expr.Pairs {
			k := e.evalExpression(pair.Key)
			vals[k] = e.evalIntExpression(pair.Value)
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+": "+strconv.Itoa(vals[k]))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *parser.BinaryExpression:
		return strconv.Itoa(e.evalIntExpression(expr))
	case *parser.MemberExpression:
		// Member access (spec/03 §5.4): leading-dot boxed field read, or an
		// object method call `c.next()` whose int result is printed.
		// Also handles list `.head` / `.tail` (spec/10 §3.1).
		if ident, isIdent := expr.Base.(*parser.Identifier); isIdent && len(expr.Parts) == 1 && !expr.IsCall {
			if lst, ok := e.listValues[ident.Value]; ok {
				switch expr.Parts[0] {
				case "head":
					if len(lst) > 0 {
						return strconv.Itoa(lst[0])
					}
					return "0"
				case "tail":
					if len(lst) > 1 {
						return strconv.Itoa(len(lst) - 1)
					}
					return "0"
				}
			}
		}
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

// evalIndexExpr evaluates an index expression with the `$` end-anchor
// (Decision 1) bound to the collection's length. This makes arithmetic
// anchors like `a[$-1]` (spec/10 §3.2) resolve to the last-but-one element:
// `$` lowers to len(c) and the surrounding binary expression is evaluated in
// the native 1-based domain. The binding is scoped to this evaluation and
// restored afterwards so nested lookups are unaffected.
func (e *Evaluator) evalIndexExpr(idxExpr parser.Expression, n int) int {
	savedLen, savedHas := e.dollarLen, e.hasDollar
	e.dollarLen, e.hasDollar = n, true
	defer func() { e.dollarLen, e.hasDollar = savedLen, savedHas }()
	return e.evalIntExpression(idxExpr)
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

// callRule executes a rule invocation per spec/03-rules.md §3.1. It creates
// a scoped call frame: the caller's symbols are saved, the rule's declared
// parameters are bound to the evaluated argument values, the body is
// executed, and the declared result variables are captured back. The
// caller's environment is restored on return so recursion and nested calls
// do not leak state.
//
// Returns a slice of result values in declaration order. Each element is
// either an int (ordinary result) or a *closureObject (spec/03 §5.4 state
// generator result). For a single-result rule the slice has length 1; for a
// multi-result rule it matches len(Results).
//
// When `bound` is non-nil the call is a closure *method* invocation
// (`c.next()`): the object's boxed cells and nested-rule table are installed
// so the body reads/mutates persistent state. When `bound` is nil the call
// is a top-level rule invocation and a fresh closure object is allocated if
// the rule is a state generator (it declares boxed cells or nested rules).
func (e *Evaluator) callRule(call *parser.CallExpression, bound *closureObject) []interface{} {
	rule, ok := e.ruleRegistry[call.Name]
	var obj *closureObject
	if !ok {
		// Not a top-level rule — resolve as a nested closure method. The folded
		// name is `.`+call.Name (e.g. `c.next()` resolves member `.next`).
		if bound != nil {
			if nr, isNested := bound.Nested["."+call.Name]; isNested {
				rule = nr
				obj = bound
				ok = true
			}
		}
		if !ok {
			fmt.Fprintf(os.Stderr, "[ERROR] E0301 UndeclaredRule: %q at line %d\n", call.Name, int(call.Token.Pos))
			return nil
		}
	} else {
		obj = bound
	}
	if rule.ForwardDecl || rule.Body == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] E0301 UndeclaredRule: %q at line %d\n", call.Name, int(call.Token.Pos))
		return nil
	}

	// A state generator (declares boxed cells or nested member rules) allocates
	// a fresh closure object per invocation (spec/03 §5.4).
	if obj == nil && isStateGenerator(rule) {
		obj = newClosureObject(rule)
	}

	// Evaluate arguments in the caller's frame before switching.
	argVals := make([]int, len(call.Args))
	for i, arg := range call.Args {
		argVals[i] = e.evalIntExpression(arg)
	}

	// Save the caller's frame.
	savedSymbols := make(map[string]int, len(e.symbols))
	for k, v := range e.symbols {
		savedSymbols[k] = v
	}
	savedStrings := make(map[string]string, len(e.stringSymbols))
	for k, v := range e.stringSymbols {
		savedStrings[k] = v
	}
	savedArrays := make(map[string][]int, len(e.arrayValues))
	for k, v := range e.arrayValues {
		savedArrays[k] = v
	}
	savedExiting := e.exitingRule
	savedBoxed := e.boxedCells

	// Fresh frame: bind parameters.
	e.symbols = make(map[string]int)
	e.stringSymbols = make(map[string]string)
	e.arrayValues = make(map[string][]int)
	e.exitingRule = false
	// Install the closure's boxed cells so member access resolves to the
	// shared, persistent state (a fresh generator starts with its own new
	// cells; a method call reuses the object's existing cells).
	if obj != nil {
		e.boxedCells = obj.Cells
	} else {
		e.boxedCells = make(map[string]*BoxedCell)
	}
	for i, param := range rule.Params {
		if i < len(argVals) {
			e.symbols[param] = argVals[i]
		}
	}
	// Initialise result variables to zero (spec/03 §2.3 default init).
	for _, res := range rule.Results {
		e.symbols[res] = 0
	}

	// Execute the body.
	for _, stmt := range rule.Body.Statements {
		if e.exitingRule {
			break
		}
		e.evalStatement(stmt)
	}

	// Capture results. A declared result that matches a nested member rule
	// (e.g. `=> (next ∈ Rule)` with a `rule .next` body member) yields the
	// closure object itself, so `new c := counter_generator(0)` binds `c` to
	// the persistent closure (spec/03 §5.4).
	results := make([]interface{}, len(rule.Results))
	for i, res := range rule.Results {
		if obj != nil {
			if _, isMember := obj.Nested["."+res]; isMember {
				results[i] = obj
				continue
			}
		}
		results[i] = e.symbols[res]
	}

	// Restore the caller's frame.
	e.symbols = savedSymbols
	e.stringSymbols = savedStrings
	e.arrayValues = savedArrays
	e.exitingRule = savedExiting
	e.boxedCells = savedBoxed

	return results
}

// callLambda invokes a first-class lambda value (spec/07-functions.md §2).
// Lambdas are pure: the body is evaluated in a fresh frame containing only
// the bound parameters — outer variables, boxed rule state, and closure
// objects are invisible by construction (§3 invariants, E0704). Rules
// cannot be invoked from this frame (E0701, enforced at the CallExpression
// dispatch). The caller's frame is restored on return, mirroring callRule's
// savedSymbols/savedStrings discipline.
func (e *Evaluator) callLambda(le *parser.LambdaExpression, call *parser.CallExpression) int {
	// Evaluate arguments in the caller's frame before switching.
	argVals := make([]int, len(call.Args))
	for i, arg := range call.Args {
		argVals[i] = e.evalIntExpression(arg)
	}

	// Fresh pure frame: parameters only.
	savedSymbols := e.symbols
	savedStrings := e.stringSymbols
	savedArrays := e.arrayValues
	savedBoxed := e.boxedCells
	e.symbols = make(map[string]int)
	e.stringSymbols = make(map[string]string)
	e.arrayValues = make(map[string][]int)
	e.boxedCells = make(map[string]*BoxedCell)
	for i, param := range le.Params {
		if i < len(argVals) {
			e.symbols[param] = argVals[i]
		}
	}

	e.inLambda++
	result := e.evalIntExpression(le.Body)
	e.inLambda--

	// Restore the caller's frame.
	e.symbols = savedSymbols
	e.stringSymbols = savedStrings
	e.arrayValues = savedArrays
	e.boxedCells = savedBoxed
	return result
}

// isStateGenerator reports whether a rule declares boxed state cells
// (`set .field := [...]`) or nested member rules anywhere in its body — the
// two markers of a spec/03 §5.4 closure generator.
func isStateGenerator(rule *parser.RuleStatement) bool {
	if rule.Body == nil {
		return false
	}
	for _, st := range rule.Body.Statements {
		switch s := st.(type) {
		case *parser.RuleStatement:
			if strings.HasPrefix(s.Name, ".") {
				return true
			}
		case *parser.DeclarationStatement:
			for _, n := range s.Names {
				if strings.HasPrefix(n, ".") {
					return true
				}
			}
			if strings.HasPrefix(s.Name, ".") {
				return true
			}
		}
	}
	return false
}

// newClosureObject allocates a closure instance for a state-generator rule,
// collecting its nested member rules into the dispatch table and starting
// with an empty boxed-cell map (populated as `set .field := [...]` runs).
func newClosureObject(rule *parser.RuleStatement) *closureObject {
	obj := &closureObject{
		Cells:  make(map[string]*BoxedCell),
		Nested: make(map[string]*parser.RuleStatement),
	}
	for _, st := range rule.Body.Statements {
		if rs, isRule := st.(*parser.RuleStatement); isRule && strings.HasPrefix(rs.Name, ".") {
			obj.Nested[rs.Name] = rs
		}
	}
	return obj
}

// bindResultValue stores a single rule-call result under `name`. Closure
// objects are recorded in closureObjects (so `c.next()` can dispatch); ints
// land in the flat symbol table. Identity is ensured either way.
func (e *Evaluator) bindResultValue(name string, val interface{}) {
	if co, isClosure := val.(*closureObject); isClosure {
		e.closureObjects[name] = co
		delete(e.symbols, name)
		e.ensureIdentity(name)
		return
	}
	if iv, isInt := val.(int); isInt {
		e.symbols[name] = iv
	} else {
		e.symbols[name] = 0
	}
	e.ensureIdentity(name)
}

// callMember dispatches an object member call `obj.method(args...)` per
// spec/03 §5.4. It resolves the receiver object, then invokes the named
// nested rule with the object's boxed cells installed so the method shares
// the closure's persistent state. Returns the first result value.
// dedupSortInts returns a sorted copy of the input with duplicates removed.
// Used for set semantics (spec/10 §1: sets are unique and unordered; the
// bootstrap evaluator stores them sorted for deterministic output).
func dedupSortInts(elems []int) []int {
	seen := make(map[int]bool, len(elems))
	var out []int
	for _, v := range elems {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Ints(out)
	return out
}

func (e *Evaluator) callMember(me *parser.MemberExpression) (interface{}, bool) {
	ident, isIdent := me.Base.(*parser.Identifier)
	if !isIdent || len(me.Parts) == 0 {
		return 0, false
	}
	obj, ok := e.closureObjects[ident.Value]
	if !ok {
		return 0, false
	}
	// Reuse callRule's frame machinery by resolving the nested rule directly.
	// The member name folds to `.name` in the generator's Nested table; pass
	// the bare member name as the call name and bind the object.
	memberCall := &parser.CallExpression{Token: me.Token, Name: me.Parts[0], Args: me.Args}
	results := e.callRule(memberCall, obj)
	if len(results) > 0 {
		return results[0], true
	}
	return 0, true
}
