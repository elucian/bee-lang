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

// MatrixValue stores a 2D matrix in row-major order (spec/10 §3.3).
// Elements are 1-based indexed: M[r,c] → Data[(r-1)*Cols + (c-1)].
type MatrixValue struct {
	Data []int
	Rows int
	Cols int
}

type Evaluator struct {
	symbols       map[string]int
	stringSymbols map[string]string
	arrayValues   map[string][]int
	// listValues, setValues, mapValues store collection literals bound to
	// identifiers (spec/10 §1). All values are stored as []int for the
	// bootstrap evaluator; sets are deduplicated and sorted.
	listValues   map[string][]int
	setValues    map[string][]int
	mapValues    map[string]map[string]int
	matrixValues map[string]MatrixValue
	// steppedRanges stores SteppedRangeExpression ASTs by identifier name
	// so that subsequent `(ident)[i]` indexing can materialise the i-th
	// element. Decision 13 (D13, 2026-09-13).
	steppedRanges map[string]*parser.SteppedRangeExpression
	identities    map[string]int
	nextID        int
	debug         bool
	exitingRule   bool
	continueCycle bool
	// loopLabelStack records the labels (empty string for anonymous) of the
	// cycles currently being evaluated, innermost last. It lets `next <label>`
	// resolve *which* enclosing cycle a labeled jump targets so a cross-label
	// jump can unwind the intervening loops (spec/02-statements.md §3.4).
	loopLabelStack []string
	// continueTargetDepth is the index (into loopLabelStack, 0 = outermost) of
	// the cycle that a pending `next` should continue. A loop whose depth equals
	// continueTargetDepth consumes the jump and advances its iteration; a loop
	// deeper than the target unwinds entirely (skipping its `then`) without
	// consuming it. Unlabeled `next` targets the innermost enclosing cycle.
	continueTargetDepth int
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
	// partialValues binds names to partial applications (spec/07-functions.md §6):
	// the wrapped lambda plus captured arguments left over from a call such as
	// add(5, ?). A later invocation add5(3) supplies the open placeholder.
	partialValues map[string]*partialApplication
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
	// scopes is the lexical block-scope stack (spec/02 §blocks). Each frame
	// records names (re)declared with `new` inside a `do...done` block so a
	// shadowed outer variable is restored when the scope exits.
	scopes []*scopeFrame
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
		matrixValues:   make(map[string]MatrixValue),
		steppedRanges:  make(map[string]*parser.SteppedRangeExpression),
		identities:     make(map[string]int),
		nextID:         1,
		ruleRegistry:   make(map[string]*parser.RuleStatement),
		boxedCells:     make(map[string]*BoxedCell),
		closureObjects: make(map[string]*closureObject),
		lambdaValues:   make(map[string]*parser.LambdaExpression),
		partialValues:  make(map[string]*partialApplication),
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
		prevContinue := e.continueCycle
		prevTarget := e.continueTargetDepth
		e.exitingRule = false
		e.continueCycle = false
		for _, stmt := range s.Body.Statements {
			if e.exitingRule {
				break
			}
			e.evalStatement(stmt)
		}
		e.exitingRule = prevExiting
		e.continueCycle = prevContinue
		e.continueTargetDepth = prevTarget
		e.debugLog("EVALUATOR DEBUG: exited rule main\n")
	case *parser.BlockStatement:
		// A `do...done` block opens a lexical scope (spec/02 §blocks): `new`
		// bindings inside shadow outer names and are restored on exit.
		e.enterScope()
		for _, stmt := range s.Statements {
			e.evalStatement(stmt)
		}
		e.leaveScope()
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
			// spec/03-rules.md §4/§7: a failed expect raises E0303
			// ExpectationFailed, a fatal runtime error (failure exit status)
			// — never a Go panic (config/AGENTS.md §2: no panic in error flows).
			fmt.Fprintf(os.Stderr, "[ERROR] E0303 ExpectationFailed: expect failed at line %d\n", int(s.Token.Pos))
			os.Exit(1)
		} else {
			e.debugLog("DEBUG: Expectation passed in line %d\n", int(s.Token.Pos))
		}
	case *parser.ApplyStatement:
		// spec/03 §3.1: execute the rule for side effects, discarding results.
		// By-reference @args are written back inside callRule before the
		// caller's frame is restored.
		e.callRule(s.Call, nil)
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
	case *parser.WriteStatement:
		// spec/02 §5: `write expr;` prints without trailing newline.
		if s.Value != nil {
			fmt.Print(e.evalExpression(s.Value))
		}
	case *parser.CutStatement:
		// `cut target[index];` removes the element at the given 1-based index
		// from a list/array (shifting remaining elements) or the key from a map.
		if idxExpr, isIndex := s.Target.(*parser.IndexExpression); isIndex {
			if ident, ok := idxExpr.Left.(*parser.Identifier); ok {
				// Map key removal
				if m, hasMap := e.mapValues[ident.Value]; hasMap {
					key := e.evalExpression(idxExpr.Index)
					delete(m, key)
					break
				}
				// List/array element removal (1-based)
				for _, store := range []map[string][]int{e.listValues, e.arrayValues} {
					if coll, found := store[ident.Value]; found {
						idx := e.evalIndexExpr(idxExpr.Index, len(coll))
						if idx >= 1 && idx <= len(coll) {
							store[ident.Value] = append(coll[:idx-1], coll[idx:]...)
						}
						break
					}
				}
			}
		}
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
				// Matrix broadcast/slice mutation (spec/10 §3.3 + spec/11 §5.2):
				// `let M[*] := v` fills all elements; `let M[r,*] := v` fills row r;
				// `let M[*,c] := v` fills column c; `let M[r,c] := v` sets one cell.
				if mat, hasMat := e.matrixValues[ident.Value]; hasMat {
					if i < len(pendingInts) {
						val := pendingInts[i]
						isWildcard := func(expr parser.Expression) bool {
							if id, isID := expr.(*parser.Identifier); isID {
								return id.Value == "*"
							}
							return false
						}
						if len(idxExpr.ExtraIndices) == 1 {
							rowW := isWildcard(idxExpr.Index)
							colW := isWildcard(idxExpr.ExtraIndices[0])
							if rowW && colW {
								// M[*,*] — fill all
								for j := range mat.Data {
									mat.Data[j] = val
								}
							} else if rowW {
								// M[*,c] — fill column c
								col := e.evalIntExpression(idxExpr.ExtraIndices[0])
								for r := 1; r <= mat.Rows; r++ {
									mat.Data[(r-1)*mat.Cols+(col-1)] = val
								}
							} else if colW {
								// M[r,*] — fill row r
								row := e.evalIntExpression(idxExpr.Index)
								for c := 1; c <= mat.Cols; c++ {
									mat.Data[(row-1)*mat.Cols+(c-1)] = val
								}
							} else {
								// M[r,c] — single cell
								row := e.evalIntExpression(idxExpr.Index)
								col := e.evalIntExpression(idxExpr.ExtraIndices[0])
								if row >= 1 && row <= mat.Rows && col >= 1 && col <= mat.Cols {
									mat.Data[(row-1)*mat.Cols+(col-1)] = val
								}
							}
						} else if isWildcard(idxExpr.Index) {
							// M[*] — fill all elements (broadcast)
							for j := range mat.Data {
								mat.Data[j] = val
							}
						}
						e.matrixValues[ident.Value] = mat
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
						// spec/10 §2.1 + tracking/solutions/06-indexing-strategy.md: raw negative indexing is a hard
						// error (use `a[$-n]`). Halt with a non-zero exit so @NEGATIVE
						// tests observe the graceful failure, mirroring expect.
						fmt.Fprintf(os.Stderr, "[ERROR] E1001 NegativeIndex: index %d is negative; use a[$-n] for a relative end-anchor at line %d\n", idx, int(idxExpr.Token.Pos))
						panic(fmt.Sprintf("negative index not allowed at line %d", int(idxExpr.Token.Pos)))
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
			// Collection append operators (spec/01-lexical-structure.md
			// coll_op): `let l <+ x;` appends x at the end (List Append),
			// `let l +> x;` prepends x at the beginning (Pipeline Append,
			// Map-Reduce pattern). Operates on the stored slice in place;
			// lists and arrays share the []int materialisation. The rebind
			// replaces the stored slice header, so callers observing the
			// same variable see the grown collection.
			if s.Token.Literal == "<+" || s.Token.Literal == "+>" {
				rhs := pendingInts[i]
				if lst, ok := e.listValues[name.Value]; ok {
					if s.Token.Literal == "<+" {
						e.listValues[name.Value] = append(lst, rhs)
					} else {
						e.listValues[name.Value] = append([]int{rhs}, lst...)
					}
					continue
				}
				if arr, ok := e.arrayValues[name.Value]; ok {
					if s.Token.Literal == "<+" {
						e.arrayValues[name.Value] = append(arr, rhs)
					} else {
						e.arrayValues[name.Value] = append([]int{rhs}, arr...)
					}
					continue
				}
				continue
			}
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
				// String compound concatenation (spec/02 §2.2): `let s += e`
				// where s and e are strings rebinds s to the concatenation
				// s+e. Strings are immutable (spec/05 §2), so this must be
				// detected BEFORE the plain string assignment below — the
				// literal/alias path would otherwise overwrite s instead of
				// concatenating. The target keeps its own identity (create or
				// reuse) so reference stability is preserved.
				if s.Token.Literal == "+=" {
					if cur, curIsString := e.stringSymbols[name.Value]; curIsString {
						if strLit, ok := s.Values[i].(*parser.StringLiteral); ok {
							e.stringSymbols[name.Value] = cur + strLit.Value
							delete(e.symbols, name.Value)
							e.ensureIdentity(name.Value)
							e.debugLog("EVALUATOR DEBUG: Concatenated string %s = %q\n", name.Value, e.stringSymbols[name.Value])
							continue
						} else if strIdent, ok := s.Values[i].(*parser.Identifier); ok {
							if strVal, hasStr := e.stringSymbols[strIdent.Value]; hasStr {
								e.stringSymbols[name.Value] = cur + strVal
								delete(e.symbols, name.Value)
								e.ensureIdentity(name.Value)
								e.debugLog("EVALUATOR DEBUG: Concatenated string %s = %q\n", name.Value, e.stringSymbols[name.Value])
								continue
							}
						}
					}
				}
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

				// Set element add/remove (spec/10 §3.4): `let s += x;` adds
				// x to the set; `let s -= x;` removes x from the set.
				lit := s.Token.Literal
				e.debugLog("EVALUATOR DEBUG: SetCheck name=%s lit=%s setKeys=%v\n", name.Value, lit, func() []string {
					keys := make([]string, 0, len(e.setValues))
					for k := range e.setValues {
						keys = append(keys, k)
					}
					return keys
				}())
				if lit == "+=" || lit == "-=" {
					if st, isSet := e.setValues[name.Value]; isSet {
						if lit == "+=" {
							e.setValues[name.Value] = dedupSortInts(append(st, rightVal))
						} else {
							filtered := make([]int, 0, len(st))
							for _, v := range st {
								if v != rightVal {
									filtered = append(filtered, v)
								}
							}
							e.setValues[name.Value] = filtered
						}
						e.ensureIdentity(name.Value)
						continue
					}
				}

				// Assignment handles both initial assignment and mutation (compound op)
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
		// Typed matrix declaration (spec/10 §3.3): `new M ∈ [Z](r, c)`
		// zero-initialises an r×c matrix in row-major order.
		if s.MatrixDims != nil {
			rows, cols := s.MatrixDims[0], s.MatrixDims[1]
			for _, name := range names {
				e.matrixValues[name] = MatrixValue{
					Data: make([]int, rows*cols),
					Rows: rows,
					Cols: cols,
				}
				e.ensureIdentity(name)
			}
			// If a value is also provided, fall through to the value binding
			// below so `new M ∈ [Z](2,3) := [[1,2,3],[4,5,6]];` initialises.
			if len(s.Values) == 0 && s.Value == nil {
				break
			}
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
			// A partial-application call (new f := fn(a, ?)) must NOT be routed
			// through the single-result rule-call shortcut below: the main loop
			// registers it as a curried value (spec/07-functions.md §6).
			if call, ok := s.Values[0].(*parser.CallExpression); ok && !e.hasPlaceholderArg(call.Args) {
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
		// Deconstruction with spread (spec/11 §5.1):
		// `new x, y, *tail := [1,2,3,4,5];` — head elements bound positionally,
		// remaining elements collected into the spread target as a collection.
		if s.SpreadIndex >= 0 && len(s.Values) == 1 {
			if elems, kind, ok := e.resolveCollection(s.Values[0]); ok {
				headCount := len(names) - 1 // all names except the spread target
				for i := 0; i < headCount && i < len(elems); i++ {
					e.symbols[names[i]] = elems[i]
					e.ensureIdentity(names[i])
				}
				if s.SpreadIndex < len(names) && headCount <= len(elems) {
					tail := elems[headCount:]
					switch kind {
					case kindArray:
						e.arrayValues[names[s.SpreadIndex]] = tail
					case kindList:
						e.listValues[names[s.SpreadIndex]] = tail
					default:
						e.arrayValues[names[s.SpreadIndex]] = tail
					}
					e.ensureIdentity(names[s.SpreadIndex])
				}
				break
			}
		}
		for i, name := range names {
			// `new` declares in the current lexical scope, shadowing any outer
			// binding (spec/02 §blocks) so it can be restored on block exit.
			e.declareBinding(name)
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
				} else if sre, ok := s.Values[i].(*parser.SteppedRangeExpression); ok {
					// Decision 13 (D13): bind stepped range so `(name)[i]`
					// indexing materialises the i-th element (1-based, D1).
					e.steppedRanges[name] = sre
					delete(e.symbols, name)
					e.ensureIdentity(name)
				} else if builder, ok := s.Values[i].(*parser.BuilderExpression); ok {
					// Collection builder (spec/11 §5): materialise the builder
					// into the appropriate collection store.
					e.evalBuilderBinding(name, builder)
				} else if call, ok := s.Values[i].(*parser.CallExpression); ok {
					// Partial application (spec/07-functions.md §6): a lambda
					// call containing a ? placeholder yields a curried value.
					if e.registerPartialFromCall(name, call) {
						continue
					}
					val := e.evalIntExpression(call)
					e.symbols[name] = val
					e.ensureIdentity(name)
				} else {
					// Deep clone binding: `new X :: expr;` (spec/10 §2.4).
					if s.Clone {
						if e.tryCloneBinding(name, s.Values[i]) {
							continue
						}
					}
					// Slice view binding: `new w := a[2..$-1];` (spec/10 §2.3).
					if idxExpr, isIdx := s.Values[i].(*parser.IndexExpression); isIdx {
						if ident, ok := idxExpr.Left.(*parser.Identifier); ok {
							if coll, kind, okColl := e.resolveCollection(ident); okColl {
								if slice := e.evalSliceView(idxExpr, coll); slice != nil {
									switch kind {
									case kindArray:
										e.arrayValues[name] = slice
									case kindList:
										e.listValues[name] = slice
									case kindSet:
										e.setValues[name] = dedupSortInts(slice)
									}
									e.ensureIdentity(name)
									continue
								}
							}
						}
					}
					// Set algebra: `new u := s1 ∪ s2;` (spec/10 §3.4).
					if e.trySetAlgebra(name, s.Values[i]) {
						continue
					}
					// Collection concatenation: `new c := [1,2] + [3];` (spec/10).
					if e.tryCollectionConcat(name, s.Values[i]) {
						continue
					}
					// Reference binding: `new alias := collection;` (spec/10 §2.4).
					if e.tryRefBinding(name, s.Values[i]) {
						continue
					}
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
				} else if call, ok := s.Value.(*parser.CallExpression); ok {
					// Partial application (spec/07-functions.md §6): a lambda
					// call containing a ? placeholder yields a curried value.
					if e.registerPartialFromCall(name, call) {
						continue
					}
					val := e.evalIntExpression(call)
					e.symbols[name] = val
					e.ensureIdentity(name)
				} else {
					// Deep clone binding: `new X :: expr;` (spec/10 §2.4).
					if s.Clone {
						if e.tryCloneBinding(name, s.Value) {
							continue
						}
					}
					// Slice view binding: `new w := a[2..$-1];` (spec/10 §2.3).
					if idxExpr, isIdx := s.Value.(*parser.IndexExpression); isIdx {
						if ident, ok := idxExpr.Left.(*parser.Identifier); ok {
							if coll, kind, okColl := e.resolveCollection(ident); okColl {
								if slice := e.evalSliceView(idxExpr, coll); slice != nil {
									switch kind {
									case kindArray:
										e.arrayValues[name] = slice
									case kindList:
										e.listValues[name] = slice
									case kindSet:
										e.setValues[name] = dedupSortInts(slice)
									}
									e.ensureIdentity(name)
									continue
								}
							}
						}
					}
					// Set algebra: `new u := s1 ∪ s2;` (spec/10 §3.4).
					if e.trySetAlgebra(name, s.Value) {
						continue
					}
					// Collection concatenation: `new cm := l + m;` (spec/10).
					if e.tryCollectionConcat(name, s.Value) {
						continue
					}
					// Reference binding: `new alias := collection;` (spec/10 §2.4).
					if e.tryRefBinding(name, s.Value) {
						continue
					}
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
		// D14 extends the jump grammar with an optional `if <cond>` guard; a
		// guarded transfer only acts when the guard is truthy. The guard serves
		// both the early-exit forms below and D15's `next` loop-jump.
		take := true
		if s.Condition != nil {
			take = e.evalIntExpression(s.Condition) != 0
		}
		if take {
			// Early-exit transfers set the exitingRule flag so the canonical
			// §16.2 semantics apply — the enclosing rule's body short-circuits
			// immediately.
			if s.Keyword == "over" || s.Keyword == "stop" || s.Keyword == "exit" || s.Keyword == "abort" || s.Keyword == "panic" {
				e.exitingRule = true
			}
			// D15: `next` is the canonical loop-jump (continue) transfer. It sets
			// the continueCycle flag so the enclosing cycle skips the remainder
			// of the current iteration's body and advances to the next increment/
			// evaluation phase (spec/02-statements.md §3.4). A labeled `next
			// <label>` targets that named cycle directly: it jumps to the next
			// iteration of the outer labeled loop, unwinding every inner loop in
			// between (cross-label jump, spec/02-statements.md §3.4).
			if s.Keyword == "next" {
				e.continueCycle = true
				// Unlabeled `next` continues the innermost enclosing cycle.
				e.continueTargetDepth = len(e.loopLabelStack) - 1
				// Labeled `next <label>` resolves the named cycle's depth in the
				// active loop stack. If the label is not an enclosing cycle, it is
				// treated as the innermost (a copy of the unlabeled form).
				if s.Label != nil {
					for i := len(e.loopLabelStack) - 1; i >= 0; i-- {
						if e.loopLabelStack[i] == s.Label.Value {
							e.continueTargetDepth = i
							break
						}
					}
				}
			}
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
// jumpKind enumerates how a pending `next` jump should affect the cycle that
// is currently testing it.
const (
	jumpNone     = iota // no pending `next`
	jumpContinue        // this cycle is the target: continue its next iteration
	jumpUnwind          // this cycle sits between the `next` and its target: break out entirely
	jumpClear           // target is a cycle already inside this one (stale): consume and proceed
)

// jumpMode decides what a cycle currently executing at loopLabelStack's top
// must do with a pending `next` jump. Positions are compared against
// continueTargetDepth (0 = outermost). A cycle at the target depth consumes
// the jump and moves to its next iteration; a deeper cycle unwinds without
// consuming; a shallower cycle means the jump was already consumed inside
// (stale flag), so it just clears and proceeds.
func (e *Evaluator) jumpMode() int {
	if !e.continueCycle {
		return jumpNone
	}
	top := len(e.loopLabelStack) - 1
	switch {
	case top == e.continueTargetDepth:
		return jumpContinue
	case top > e.continueTargetDepth:
		return jumpUnwind
	default:
		return jumpClear
	}
}

// cycleLabel returns the label of a cycle statement, or "" for an anonymous
// cycle (used to populate the evaluator's loopLabelStack).
func cycleLabel(s *parser.CycleStatement) string {
	if s.Label != nil {
		return s.Label.Value
	}
	return ""
}

func (e *Evaluator) evalCycleStatement(s *parser.CycleStatement) {
	// Push this cycle's label (empty string for anonymous cycles) so labeled
	// `next <label>` / `stop <label>` / `redo <label>` jumps can resolve which
	// enclosing cycle they target. The defer guarantees the stack is unwound on
	// every return path, including an early transfer-induced exit.
	e.loopLabelStack = append(e.loopLabelStack, cycleLabel(s))
	defer func() { e.loopLabelStack = e.loopLabelStack[:len(e.loopLabelStack)-1] }()

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
					if m := e.jumpMode(); m != jumpNone {
						if m == jumpContinue || m == jumpClear {
							e.continueCycle = false
							break
						}
						// jumpUnwind: a labeled `next` targets an outer cycle; break
						// this loop entirely without consuming the jump.
						breakOut = true
						break
					}
					e.evalStatement(stmt)
				}
			}
			// A `next` placed as the last body statement is not seen by the
			// per-statement check above; act on any pending jump here.
			if m := e.jumpMode(); m == jumpUnwind {
				breakOut = true
			} else if m == jumpClear || m == jumpContinue {
				e.continueCycle = false
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
					if m := e.jumpMode(); m != jumpNone {
						if m == jumpContinue || m == jumpClear {
							e.continueCycle = false
							break
						}
						breakOut = true
						break
					}
					e.evalStatement(stmt)
				}
			}
			if m := e.jumpMode(); m == jumpUnwind {
				breakOut = true
			} else if m == jumpClear || m == jumpContinue {
				e.continueCycle = false
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
							if m := e.jumpMode(); m != jumpNone {
								if m == jumpContinue || m == jumpClear {
									e.continueCycle = false
									break
								}
								// jumpUnwind: a labeled `next` targets an outer cycle;
								// exit this loop (skipping its `then`) without
								// consuming the jump, so the outer loop handles it.
								deferRestore()
								return
							}
							e.evalStatement(stmt)
						}
						if m := e.jumpMode(); m == jumpUnwind {
							deferRestore()
							return
						} else if m == jumpClear || m == jumpContinue {
							e.continueCycle = false
						}
					}
				}
			case *parser.BinaryExpression:
				// Range expression like `0..n-2` (Decision 13). Evaluate
				// the inclusive endpoints and iterate.
				leftVal := e.evalIntExpression(rng.Left)
				rightVal := e.evalIntExpression(rng.Right)
				step := 1
				// `(a..b)` is an ascending inclusive range; per inRangePerOp it
				// contains no members when a > b, so a degenerate range like
				// `(1..n-1)` with n=1 yields zero iterations (must not descend or
				// touch invalid index 0). Descending domains use step forms.
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
								if m := e.jumpMode(); m != jumpNone {
									if m == jumpContinue || m == jumpClear {
										e.continueCycle = false
										break
									}
									deferRestore()
									return
								}
								e.evalStatement(stmt)
							}
							if m := e.jumpMode(); m == jumpUnwind {
								deferRestore()
								return
							} else if m == jumpClear || m == jumpContinue {
								e.continueCycle = false
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
							if m := e.jumpMode(); m != jumpNone {
								if m == jumpContinue || m == jumpClear {
									e.continueCycle = false
									break
								}
								deferRestore()
								return
							}
							e.evalStatement(stmt)
						}
						if m := e.jumpMode(); m == jumpUnwind {
							deferRestore()
							return
						} else if m == jumpClear || m == jumpContinue {
							e.continueCycle = false
						}
					}
				}
			case *parser.Identifier:
				// Map iteration with key,value: `for k, v ∈ m do` binds
				// k to each key and v to each value (spec/10 §3.5).
				if s.Value != nil {
					if m, isMap := e.mapValues[rng.Value]; isMap {
						valueName := s.Value.Value
						prevValV, hadPrevV := e.symbols[valueName]
						prevIDV, hadPrevIDV := e.identities[valueName]
						defer func() {
							if hadPrevV {
								e.symbols[valueName] = prevValV
							} else {
								delete(e.symbols, valueName)
							}
							if hadPrevIDV {
								e.identities[valueName] = prevIDV
							} else {
								delete(e.identities, valueName)
							}
						}()
						// Iterate in sorted-key order for determinism.
						keys := make([]string, 0, len(m))
						for k := range m {
							keys = append(keys, k)
						}
						sort.Strings(keys)
						for _, k := range keys {
							e.symbols[indexName] = 0 // keys are strings in mapValues
							e.stringSymbols[indexName] = k
							e.symbols[valueName] = m[k]
							e.ensureIdentity(indexName)
							e.ensureIdentity(valueName)
							if s.Body != nil {
								for _, stmt := range s.Body.Statements {
									if e.exitingRule {
										deferRestore()
										return
									}
									if m := e.jumpMode(); m != jumpNone {
										if m == jumpContinue || m == jumpClear {
											e.continueCycle = false
											break
										}
										deferRestore()
										return
									}
									e.evalStatement(stmt)
								}
								if m := e.jumpMode(); m == jumpUnwind {
									deferRestore()
									return
								} else if m == jumpClear || m == jumpContinue {
									e.continueCycle = false
								}
							}
						}
						break
					}
				}
				// Foreach over a stored collection (spec/10 §3.1 lists /
				// §3.2 arrays / spec/11 pipelines): `for e ∈ l do` binds the
				// loop variable to each element in collection order. Arrays,
				// lists, and sets all materialise as []int in this evaluator.
				var elems []int
				if arr, ok := e.arrayValues[rng.Value]; ok {
					elems = arr
				} else if lst, ok := e.listValues[rng.Value]; ok {
					elems = lst
				} else if st, ok := e.setValues[rng.Value]; ok {
					elems = st
				}
				for _, v := range elems {
					e.symbols[indexName] = v
					e.ensureIdentity(indexName)
					if s.Body != nil {
						for _, stmt := range s.Body.Statements {
							if e.exitingRule {
								deferRestore()
								return
							}
							if m := e.jumpMode(); m != jumpNone {
								if m == jumpContinue || m == jumpClear {
									e.continueCycle = false
									break
								}
								deferRestore()
								return
							}
							e.evalStatement(stmt)
						}
						if m := e.jumpMode(); m == jumpUnwind {
							deferRestore()
							return
						} else if m == jumpClear || m == jumpContinue {
							e.continueCycle = false
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
						if m := e.jumpMode(); m != jumpNone {
							if m == jumpContinue || m == jumpClear {
								e.continueCycle = false
								break
							}
							deferRestore()
							return
						}
						e.evalStatement(stmt)
					}
					if m := e.jumpMode(); m == jumpUnwind {
						deferRestore()
						return
					} else if m == jumpClear || m == jumpContinue {
						e.continueCycle = false
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
		// Partial-application invocation (spec/07-functions.md §6): a name
		// bound to a partial application supplies the open placeholders here.
		if pa, isPa := e.partialValues[expr.Name]; isPa {
			return e.callPartial(pa, expr), e.allocID()
		}
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
			idx := e.evalIndexExpr(expr.Index, e.steppedRangeLength(sre))
			val := e.steppedRangeValue(sre, idx)
			return val, e.allocID()
		}
		if ident, ok := expr.Left.(*parser.Identifier); ok {
			if sre, hasSRE := e.steppedRanges[ident.Value]; hasSRE {
				idx := e.evalIndexExpr(expr.Index, e.steppedRangeLength(sre))
				val := e.steppedRangeValue(sre, idx)
				return val, e.ensureIdentity(ident.Value)
			}
			// Map key indexing (spec/10 §3.5): M[k] looks up the value for key k.
			if m, hasMap := e.mapValues[ident.Value]; hasMap {
				key := e.evalExpression(expr.Index)
				return m[key], e.ensureIdentity(ident.Value)
			}
			// Matrix 2D indexing (spec/10 §3.3): M[r, c] with 1-based indices.
			if mat, hasMat := e.matrixValues[ident.Value]; hasMat {
				if len(expr.ExtraIndices) == 1 {
					row := e.evalIntExpression(expr.Index)
					col := e.evalIntExpression(expr.ExtraIndices[0])
					if row >= 1 && row <= mat.Rows && col >= 1 && col <= mat.Cols {
						return mat.Data[(row-1)*mat.Cols+(col-1)], e.ensureIdentity(ident.Value)
					}
					return 0, e.allocID()
				}
				// Single-index matrix access: flatten to 1-based linear index.
				idx := e.evalIndexExpr(expr.Index, len(mat.Data))
				if idx >= 1 && idx <= len(mat.Data) {
					return mat.Data[idx-1], e.ensureIdentity(ident.Value)
				}
				return 0, e.allocID()
			}
			var idx int
			// Resolve the collection store in spec/10 order: arrays, then
			// lists, then sets. All materialise as []int with 1-based indexing
			// (Decision 1); the bounds/anchor rules are identical across kinds.
			var coll []int
			if arr, ok := e.arrayValues[ident.Value]; ok {
				coll = arr
			} else if lst, ok := e.listValues[ident.Value]; ok {
				coll = lst
			} else if st, ok := e.setValues[ident.Value]; ok {
				coll = st
			}
			if coll != nil {
				idx = e.evalIndexExpr(expr.Index, len(coll))
				if idx == 0 {
					fmt.Fprintf(os.Stderr, "[ERROR] E1006 ZeroBasedIndexAttempt: index 0 at line %d\n", int(expr.Token.Pos))
					return 0, e.allocID()
				}
				if idx < 0 {
					// spec/10 §2.1 + tracking/solutions/06-indexing-strategy.md: raw negative indexing is a hard
					// error (use `a[$-n]`). Halt with a non-zero exit so @NEGATIVE
					// tests observe the graceful failure, mirroring expect.
					fmt.Fprintf(os.Stderr, "[ERROR] E1001 NegativeIndex: index %d is negative; use a[$-n] for a relative end-anchor at line %d\n", idx, int(expr.Token.Pos))
					panic(fmt.Sprintf("negative index not allowed at line %d", int(expr.Token.Pos)))
				}
				if idx < 1 || idx > len(coll) {
					fmt.Fprintf(os.Stderr, "[ERROR] E1001 IndexOutOfBounds: index %d out of range 1..%d at line %d\n", idx, len(coll), int(expr.Token.Pos))
					return 0, e.allocID()
				}
				return coll[idx-1], e.ensureIdentity(ident.Value)
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
	case *parser.QuantifierExpression:
		// spec/11 §3: (∀ x ∈ S : P(x)) returns 1 iff every element satisfies
		// the predicate; (∃ x ∈ S : P(x)) returns 1 iff at least one does.
		// The loop variable is bound per-iteration and restored afterwards.
		elems, _, ok := e.resolveCollection(expr.Domain)
		if !ok {
			return 0, e.allocID()
		}
		varName := expr.Variable
		prevVal, hadPrev := e.symbols[varName]
		prevID, hadPrevID := e.identities[varName]
		defer func() {
			if hadPrev {
				e.symbols[varName] = prevVal
			} else {
				delete(e.symbols, varName)
			}
			if hadPrevID {
				e.identities[varName] = prevID
			} else {
				delete(e.identities, varName)
			}
		}()
		isForall := expr.Token.Type == token.FORALL || expr.Token.Literal == "∀"
		for _, elem := range elems {
			e.symbols[varName] = elem
			e.ensureIdentity(varName)
			result := e.evalIntExpression(expr.Condition)
			if isForall && result == 0 {
				return 0, e.allocID() // universal: any false → false
			}
			if !isForall && result != 0 {
				return 1, e.allocID() // existential: any true → true
			}
		}
		if isForall {
			return 1, e.allocID() // all passed
		}
		return 0, e.allocID() // none found
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
			// String-aware equality (spec/02 §2.2 / spec/05 §2): when either
			// operand is a string literal or a string-bound identifier, compare
			// the actual string values. The int path below would otherwise force
			// both operands to 0 (strings carry no int value), silently making
			// mismatched strings compare equal.
			if e.isStringExpr(expr.Left) || e.isStringExpr(expr.Right) {
				if e.evalExpression(expr.Left) == e.evalExpression(expr.Right) {
					return 1, 0
				}
				return 0, 0
			}
			leftVal, _ := e.evalIntExpressionWithID(expr.Left)
			rightVal, _ := e.evalIntExpressionWithID(expr.Right)
			return evalComparison(lit, leftVal, rightVal, expr.Token.Type, lit), 0
		}
		if expr.Token.Type == token.NEQ || expr.Token.Type == token.NOT_EQ || lit == "¬" || lit == "!=" || lit == "<>" || expr.Token.Type == token.NEQ_UNICODE || lit == "≠" || strings.Contains(lit, "≠") {
			if e.isStringExpr(expr.Left) || e.isStringExpr(expr.Right) {
				if e.evalExpression(expr.Left) != e.evalExpression(expr.Right) {
					return 1, 0
				}
				return 0, 0
			}
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
			// Collection membership (spec/10 §3): x ∈ coll for arrays, lists,
			// and sets bound by name. The right-hand side is an identifier;
			// linear scan over the value store (bootstrap evaluator).
			if id, ok := expr.Right.(*parser.Identifier); ok {
				if arr, isArr := e.arrayValues[id.Value]; isArr {
					for _, v := range arr {
						if v == leftVal {
							return 1, 0
						}
					}
					return 0, 0
				}
				if lst, isLst := e.listValues[id.Value]; isLst {
					for _, v := range lst {
						if v == leftVal {
							return 1, 0
						}
					}
					return 0, 0
				}
				if st, isSet := e.setValues[id.Value]; isSet {
					for _, v := range st {
						if v == leftVal {
							return 1, 0
						}
					}
					return 0, 0
				}
			}
			return 0, 0
		}
		// Not-in (membership negation, spec/10 §3.4): x !∈ coll.
		if lit == "!∈" || expr.Token.Type == token.NOT_IN {
			leftVal, _ := e.evalIntExpressionWithID(expr.Left)
			if id, ok := expr.Right.(*parser.Identifier); ok {
				for _, store := range []map[string][]int{e.arrayValues, e.listValues, e.setValues} {
					if coll, found := store[id.Value]; found {
						for _, v := range coll {
							if v == leftVal {
								return 0, 0 // found → !∈ is false
							}
						}
						return 1, 0 // not found → !∈ is true
					}
				}
			}
			return 1, 0 // not a collection → vacuously true
		}
		// Set algebra binary operators (spec/10 §3.4): ∩ ∪ Δ ⊂ ⊃.
		// When these appear in expression context (not declaration binding),
		// evaluate them as boolean predicates for ⊂/⊃, or materialise a
		// temporary set and check membership for ∩/∪/Δ.
		if expr.Token.Type == token.SUBSET || expr.Token.Type == token.SUPERSET || lit == "⊂" || lit == "⊃" {
			leftElems, _, okL := e.resolveCollection(expr.Left)
			rightElems, _, okR := e.resolveCollection(expr.Right)
			if okL && okR {
				ls := toIntSet(leftElems)
				rs := toIntSet(rightElems)
				if lit == "⊂" || expr.Token.Type == token.SUBSET {
					if isSubset(ls, rs) {
						return 1, 0
					}
					return 0, 0
				}
				// ⊃
				if isSubset(rs, ls) {
					return 1, 0
				}
				return 0, 0
			}
			return 0, 0
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

// isStringExpr reports whether the expression yields a string value: a string
// literal, or an identifier currently bound in the stringSymbols table. It is
// used to steer equality/inequality (spec/02 §2.2, spec/05 §2) and friends to
// a string-aware comparison rather than through the integer path, where a
// string operand degrades to 0.
func (e *Evaluator) isStringExpr(node parser.Expression) bool {
	switch expr := node.(type) {
	case *parser.StringLiteral:
		return true
	case *parser.Identifier:
		_, ok := e.stringSymbols[expr.Value]
		return ok
	}
	return false
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
					// spec/10 §2.1 + tracking/solutions/06-indexing-strategy.md: raw negative indexing is a hard
					// error (use `a[$-n]`). Halt with a non-zero exit so @NEGATIVE
					// tests observe the graceful failure, mirroring expect.
					fmt.Fprintf(os.Stderr, "[ERROR] E1001 NegativeIndex: index %d is negative; use a[$-n] for a relative end-anchor at line %d\n", idx, int(expr.Token.Pos))
					panic(fmt.Sprintf("negative index not allowed at line %d", int(expr.Token.Pos)))
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
		// String concatenation: when either operand is a string literal or
		// a string-bound identifier, `+` concatenates in string context.
		if expr.Token.Literal == "+" {
			leftStr := e.evalExpression(expr.Left)
			rightStr := e.evalExpression(expr.Right)
			// Heuristic: if either side is non-numeric, treat as string concat.
			if !isNumericString(leftStr) || !isNumericString(rightStr) {
				return leftStr + rightStr
			}
		}
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

// steppedRangeLength returns the number of elements in the stepped range
// sequence, used to bind the `$` end-anchor for index expressions (D1+D13).
func (e *Evaluator) steppedRangeLength(sre *parser.SteppedRangeExpression) int {
	be := sre.Range.(*parser.BinaryExpression)
	start := e.evalIntExpression(be.Left)
	end := e.evalIntExpression(be.Right)
	step := e.evalIntExpression(sre.Step)
	if step == 0 {
		step = 1
	}
	leftOffset := 0
	switch be.Token.Type {
	case token.RANGE_RGHT_INC, token.RANGE_EXCL:
		leftOffset = step
	}
	first := start + leftOffset
	// Compute count by iterating from first until the bundle excludes the value.
	count := 0
	for i := 1; ; i++ {
		val := start + leftOffset + (i-1)*step
		inRange := false
		switch be.Token.Type {
		case token.RANGE_INCL:
			inRange = val >= start && val <= end
		case token.RANGE_LEFT_INC:
			inRange = val >= start && val < end
		case token.RANGE_RGHT_INC:
			inRange = val > start && val <= end
		case token.RANGE_EXCL:
			inRange = val > start && val < end
		default:
			inRange = val >= start && val <= end
		}
		if !inRange {
			break
		}
		count++
	}
	_ = first
	return count
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
		// By-reference argument (@ident): copy the caller cell's *value* in —
		// evaluating the PrefixExpression would yield the D12 identity ID,
		// not the variable's value. Write-back happens after the body runs.
		if prefix, isRef := arg.(*parser.PrefixExpression); isRef && prefix.Operator == "@" {
			if ident, isIdent := prefix.Right.(*parser.Identifier); isIdent {
				argVals[i] = e.symbols[ident.Value]
				continue
			}
		} else {
			argVals[i] = e.evalIntExpression(arg)
		}
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
	savedContinue := e.continueCycle
	savedTarget := e.continueTargetDepth
	savedLoopStack := e.loopLabelStack
	savedBoxed := e.boxedCells

	// Fresh frame: bind parameters.
	e.symbols = make(map[string]int)
	e.stringSymbols = make(map[string]string)
	e.arrayValues = make(map[string][]int)
	e.exitingRule = false
	e.continueCycle = false
	e.continueTargetDepth = 0
	// A labelled `next <label>` may only target a cycle within the same rule;
	// the callee starts with an empty loop stack so caller labels stay out of
	// scope (spec/02-statements.md §3.4).
	e.loopLabelStack = nil
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

			// Composite by-share binding (spec/03 §2.2): when the argument
			// is an identifier (optionally @-prefixed) that names a stored
			// collection, bind the parameter to the *same* slice so element
			// mutations inside the rule are visible to the caller. The
			// caller's frame is saved above, so aliasing is safe here.
			argExpr := call.Args[i]
			if prefix, isRef := argExpr.(*parser.PrefixExpression); isRef && prefix.Operator == "@" {
				argExpr = prefix.Right
			}
			if ident, isIdent := argExpr.(*parser.Identifier); isIdent {
				if arr, ok := savedArrays[ident.Value]; ok {
					e.arrayValues[param] = arr
					delete(e.symbols, param)
				}
			}
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

	// By-reference write-back (spec/03 §3.1 apply-directive semantics):
	// an argument spelled `@ident` at the call site binds the callee's
	// parameter to the caller's cell. After the body completes, copy the
	// parameter's final value back into the caller's saved frame so the
	// mutation is visible to the caller (copy-in/copy-out).
	for i, arg := range call.Args {
		prefix, isRef := arg.(*parser.PrefixExpression)
		if !isRef || prefix.Operator != "@" {
			continue
		}
		ident, isIdent := prefix.Right.(*parser.Identifier)
		if !isIdent || i >= len(rule.Params) {
			continue
		}
		param := rule.Params[i]
		if val, ok := e.symbols[param]; ok {
			savedSymbols[ident.Value] = val
		}
		if arr, ok := e.arrayValues[param]; ok {
			savedArrays[ident.Value] = arr
		}
		e.ensureIdentity(ident.Value)
	}

	// Restore the caller's frame.
	e.symbols = savedSymbols
	e.stringSymbols = savedStrings
	e.arrayValues = savedArrays
	e.exitingRule = savedExiting
	e.continueCycle = savedContinue
	e.continueTargetDepth = savedTarget
	e.loopLabelStack = savedLoopStack
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
	return e.evalLambdaBody(le, argVals)
}

// evalLambdaBody binds paramVals to le.Params in a fresh pure frame and
// evaluates the lambda body, restoring the caller's frame afterwards. It is
// shared by full lambda calls and partial-application invocations.
func (e *Evaluator) evalLambdaBody(le *parser.LambdaExpression, paramVals []int) int {
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
		if i < len(paramVals) {
			e.symbols[param] = paramVals[i]
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

// partialArgument records one parameter slot of a partial application: either
// a captured (bound) value or a still-open ? placeholder awaiting a later arg.
type partialArgument struct {
	hasBound bool
	val      int
}

// partialApplication wraps a pure lambda together with the argument values
// captured at partial-application time (spec/07-functions.md §6). Slots with
// hasBound==false are open placeholders filled in call order when the partial
// is later invoked.
type partialApplication struct {
	lambda *parser.LambdaExpression
	args   []partialArgument
}

// buildPartialApplication derives a partial application from a lambda call
// whose argument list contains a ? placeholder: the provided (non-?) args are
// captured positionally onto the new partial. Placeholder slots are left open
// for a subsequent invocation to fill.
func (e *Evaluator) buildPartialApplication(le *parser.LambdaExpression, call *parser.CallExpression) *partialApplication {
	pa := &partialApplication{lambda: le}
	for range le.Params {
		pa.args = append(pa.args, partialArgument{})
	}
	for i, arg := range call.Args {
		if _, isPh := arg.(*parser.PlaceholderExpression); isPh {
			continue // leave the slot open (unbound placeholder)
		}
		if i < len(pa.args) {
			pa.args[i].hasBound = true
			pa.args[i].val = e.evalIntExpression(arg)
		}
	}
	return pa
}

// hasPlaceholderArg reports whether any of the call arguments is a ?.
func (e *Evaluator) hasPlaceholderArg(args []parser.Expression) bool {
	for _, a := range args {
		if _, ok := a.(*parser.PlaceholderExpression); ok {
			return true
		}
	}
	return false
}

// callPartial invokes a partial application, filling its open placeholder
// slots with the supplied arguments in order, then evaluating the wrapped
// lambda with the fully-bound parameter values (spec/07-functions.md §6).
func (e *Evaluator) callPartial(pa *partialApplication, call *parser.CallExpression) int {
	paramVals := make([]int, len(pa.args))
	for i, a := range pa.args {
		if a.hasBound {
			paramVals[i] = a.val
		}
	}
	argVals := make([]int, len(call.Args))
	for i, arg := range call.Args {
		argVals[i] = e.evalIntExpression(arg)
	}
	next := 0
	for _, av := range argVals {
		for next < len(pa.args) && pa.args[next].hasBound {
			next++
		}
		if next >= len(pa.args) {
			break // extra arguments beyond the lambda arity are dropped
		}
		paramVals[next] = av
		next++
	}
	return e.evalLambdaBody(pa.lambda, paramVals)
}

// registerPartialFromCall binds name to a partial application when call
// targets a lambda and contains a ? placeholder (spec/07-functions.md §6).
// Returns true if a partial application was registered; false means the call
// is a normal rule/lambda invocation and should be evaluated as usual.
func (e *Evaluator) registerPartialFromCall(name string, call *parser.CallExpression) bool {
	lambdaExpr, isLambda := e.lambdaValues[call.Name]
	if !isLambda {
		return false
	}
	if !e.hasPlaceholderArg(call.Args) {
		return false
	}
	pa := e.buildPartialApplication(lambdaExpr, call)
	e.partialValues[name] = pa
	e.ensureIdentity(name)
	return true
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

// collectionKind tags which evaluator store a collection binding belongs to
// so concatenation preserves the source kind (arrays stay arrays, lists stay
// lists, sets stay dedup-sorted sets) per spec/10-collections.md.
type collectionKind int

const (
	kindNone collectionKind = iota
	kindArray
	kindList
	kindSet
)

// resolveCollection evaluates an expression to a stored collection slice.
// Identifier operands resolve through the three collection stores; literal
// operands materialise directly. Returns ok=false when the expression does
// not denote a collection.
func (e *Evaluator) resolveCollection(expr parser.Expression) ([]int, collectionKind, bool) {

	switch v := expr.(type) {
	case *parser.Identifier:
		if arr, ok := e.arrayValues[v.Value]; ok {
			return arr, kindArray, true
		}
		if lst, ok := e.listValues[v.Value]; ok {
			return lst, kindList, true
		}
		if st, ok := e.setValues[v.Value]; ok {
			return st, kindSet, true
		}
	case *parser.ArrayLiteral:
		elems := make([]int, len(v.Elements))
		for j, el := range v.Elements {
			elems[j] = e.evalIntExpression(el)
		}
		return elems, kindArray, true
	case *parser.ListLiteral:
		elems := make([]int, len(v.Elements))
		for j, el := range v.Elements {
			elems[j] = e.evalIntExpression(el)
		}
		return elems, kindList, true
	case *parser.SetLiteral:
		elems := make([]int, len(v.Elements))
		for j, el := range v.Elements {
			elems[j] = e.evalIntExpression(el)
		}
		return dedupSortInts(elems), kindSet, true
	case *parser.BinaryExpression:
		// Range expression as domain (e.g. `1..6`): materialise the sequence.

		if isRangeSepToken(v.Token.Type) {
			return e.materialiseRange(v), kindArray, true
		}
	case *parser.SteppedRangeExpression:
		// Stepped range as domain (e.g. `(1..10)(2)`): materialise.
		return e.materialiseSteppedRange(v), kindArray, true
	}
	return nil, kindNone, false
}

// isRangeSepToken reports whether tok is one of the four D13 range separators.
func isRangeSepToken(t token.Type) bool {
	switch t {
	case token.RANGE_INCL, token.RANGE_LEFT_INC, token.RANGE_RGHT_INC, token.RANGE_EXCL:
		return true
	}
	return false
}

// materialiseRange expands a range binary expression (e.g. `1..6`) into an
// ordered []int honouring the endpoint inclusivity of the range operator (D13).
func (e *Evaluator) materialiseRange(be *parser.BinaryExpression) []int {
	start := e.evalIntExpression(be.Left)
	end := e.evalIntExpression(be.Right)
	lo, hi := start, end
	switch be.Token.Type {
	case token.RANGE_LEFT_INC: // ..<
		hi = end - 1
	case token.RANGE_RGHT_INC: // >..
		lo = start + 1
	case token.RANGE_EXCL: // >..<
		lo = start + 1
		hi = end - 1
	}
	var out []int
	for i := lo; i <= hi; i++ {
		out = append(out, i)
	}
	return out
}

// materialiseSteppedRange expands a SteppedRangeExpression into an ordered
// []int honouring endpoint inclusivity (D13+D1).
func (e *Evaluator) materialiseSteppedRange(sre *parser.SteppedRangeExpression) []int {
	be := sre.Range.(*parser.BinaryExpression)
	start := e.evalIntExpression(be.Left)
	end := e.evalIntExpression(be.Right)
	step := e.evalIntExpression(sre.Step)
	if step == 0 {
		step = 1
	}
	leftOffset := 0
	switch be.Token.Type {
	case token.RANGE_RGHT_INC, token.RANGE_EXCL:
		leftOffset = step
	}
	var out []int
	for i := 1; ; i++ {
		val := start + leftOffset + (i-1)*step
		inRange := false
		switch be.Token.Type {
		case token.RANGE_INCL:
			inRange = val >= start && val <= end
		case token.RANGE_LEFT_INC:
			inRange = val >= start && val < end
		case token.RANGE_RGHT_INC:
			inRange = val > start && val <= end
		case token.RANGE_EXCL:
			inRange = val > start && val < end
		default:
			inRange = val >= start && val <= end
		}
		if !inRange {
			break
		}
		out = append(out, val)
	}
	return out
}

// evalBuilderBinding evaluates a BuilderExpression and binds the result to
// the named collection store (set, array, or map). spec/11 §5.
func (e *Evaluator) evalBuilderBinding(name string, b *parser.BuilderExpression) {
	domain, _, ok := e.resolveCollection(b.Domain)
	if !ok {
		return
	}
	varName := b.Variable
	prevVal, hadPrev := e.symbols[varName]
	prevID, hadPrevID := e.identities[varName]
	defer func() {
		if hadPrev {
			e.symbols[varName] = prevVal
		} else {
			delete(e.symbols, varName)
		}
		if hadPrevID {
			e.identities[varName] = prevID
		} else {
			delete(e.identities, varName)
		}
	}()

	if b.IsMap {
		// Map builder: { (k:v) | x ∈ domain [∧ cond] }
		pair, isPair := b.MapExpr.(*parser.MapPairExpression)
		if !isPair {
			return
		}
		m := make(map[string]int)
		for _, elem := range domain {
			e.symbols[varName] = elem
			e.ensureIdentity(varName)
			if b.Condition != nil && e.evalIntExpression(b.Condition) == 0 {
				continue
			}
			key := e.evalExpression(pair.Key)
			val := e.evalIntExpression(pair.Value)
			m[key] = val
		}
		e.mapValues[name] = m
		e.ensureIdentity(name)
		return
	}

	// Set or array builder.
	var results []int
	for _, elem := range domain {
		e.symbols[varName] = elem
		e.ensureIdentity(varName)
		if b.Condition != nil && e.evalIntExpression(b.Condition) == 0 {
			continue
		}
		results = append(results, e.evalIntExpression(b.MapExpr))
	}
	if b.Token.Literal == "[" {
		e.arrayValues[name] = results
	} else {
		e.setValues[name] = dedupSortInts(results)
	}
	e.ensureIdentity(name)
}

// trySetAlgebra evaluates a binary set-algebra expression (`∩`, `∪`, `Δ`, `⊂`, `⊃`)
// and binds the result under `name` in the set store. Returns true when handled.
// spec/10-collections.md §3.4.
func (e *Evaluator) trySetAlgebra(name string, expr parser.Expression) bool {
	be, ok := expr.(*parser.BinaryExpression)
	if !ok {
		return false
	}
	op := be.Token.Literal
	if op != "∩" && op != "∪" && op != "Δ" && op != "⊂" && op != "⊃" {
		return false
	}
	leftElems, _, okL := e.resolveCollection(be.Left)
	rightElems, _, okR := e.resolveCollection(be.Right)
	if !okL || !okR {
		return false
	}
	leftSet := toIntSet(leftElems)
	rightSet := toIntSet(rightElems)
	switch op {
	case "∩":
		e.setValues[name] = dedupSortInts(setIntersect(leftSet, rightSet))
	case "∪":
		e.setValues[name] = dedupSortInts(setUnion(leftSet, rightSet))
	case "Δ":
		e.setValues[name] = dedupSortInts(setSymDiff(leftSet, rightSet))
	case "⊂":
		if isSubset(leftSet, rightSet) {
			e.symbols[name] = 1
		} else {
			e.symbols[name] = 0
		}
	case "⊃":
		if isSubset(rightSet, leftSet) {
			e.symbols[name] = 1
		} else {
			e.symbols[name] = 0
		}
	}
	e.ensureIdentity(name)
	return true
}

// tryCloneBinding implements the `::` deep-copy binding (spec/10 §2.4).
// It resolves the RHS expression to a collection and stores an independent
// deep copy under `name`. Returns true when the binding was performed.
func (e *Evaluator) tryCloneBinding(name string, expr parser.Expression) bool {
	// Resolve the RHS to a collection.
	switch v := expr.(type) {
	case *parser.Identifier:
		if arr, ok := e.arrayValues[v.Value]; ok {
			cp := make([]int, len(arr))
			copy(cp, arr)
			e.arrayValues[name] = cp
			e.ensureIdentity(name)
			return true
		}
		if lst, ok := e.listValues[v.Value]; ok {
			cp := make([]int, len(lst))
			copy(cp, lst)
			e.listValues[name] = cp
			e.ensureIdentity(name)
			return true
		}
		if st, ok := e.setValues[v.Value]; ok {
			cp := make([]int, len(st))
			copy(cp, st)
			e.setValues[name] = cp
			e.ensureIdentity(name)
			return true
		}
		if m, ok := e.mapValues[v.Value]; ok {
			cp := make(map[string]int, len(m))
			for k, val := range m {
				cp[k] = val
			}
			e.mapValues[name] = cp
			e.ensureIdentity(name)
			return true
		}
	case *parser.IndexExpression:
		// Clone of a slice view: `new frozen :: list[2..3];`
		// Materialise the view into a real sub-collection.
		if ident, ok := v.Left.(*parser.Identifier); ok {
			if coll, _, okColl := e.resolveCollection(ident); okColl {
				slice := e.evalSliceView(v, coll)
				if slice != nil {
					e.arrayValues[name] = slice
					e.ensureIdentity(name)
					return true
				}
			}
		}
	}
	return false
}

// tryRefBinding implements the `:=` reference binding for identifiers that
// resolve to collections (spec/10 §2.4). Both names share the same storage.
// Returns true when the binding was performed.
func (e *Evaluator) tryRefBinding(name string, expr parser.Expression) bool {
	ident, ok := expr.(*parser.Identifier)
	if !ok {
		return false
	}
	if arr, ok := e.arrayValues[ident.Value]; ok {
		e.arrayValues[name] = arr // shared reference
		e.ensureIdentity(name)
		return true
	}
	if lst, ok := e.listValues[ident.Value]; ok {
		e.listValues[name] = lst
		e.ensureIdentity(name)
		return true
	}
	if st, ok := e.setValues[ident.Value]; ok {
		e.setValues[name] = st
		e.ensureIdentity(name)
		return true
	}
	if m, ok := e.mapValues[ident.Value]; ok {
		e.mapValues[name] = m
		e.ensureIdentity(name)
		return true
	}
	return false
}

// evalSliceView extracts a sub-slice from a collection using a range index
// expression (e.g. `a[2..$-1]`). Returns nil when the index is not a range.
// The returned slice is a NEW copy (materialised view), not a reference.
func (e *Evaluator) evalSliceView(idxExpr *parser.IndexExpression, coll []int) []int {
	be, ok := idxExpr.Index.(*parser.BinaryExpression)
	if !ok || !isRangeSeparatorToken(be.Token) {
		return nil
	}
	savedLen, savedHas := e.dollarLen, e.hasDollar
	e.dollarLen, e.hasDollar = len(coll), true
	defer func() { e.dollarLen, e.hasDollar = savedLen, savedHas }()
	start := e.evalIntExpression(be.Left)
	end := e.evalIntExpression(be.Right)
	// 1-based inclusive range: clamp to [1, len]
	if start < 1 {
		start = 1
	}
	if end > len(coll) {
		end = len(coll)
	}
	if start > end {
		return []int{}
	}
	result := make([]int, end-start+1)
	copy(result, coll[start-1:end])
	return result
}

// --- Set helper functions ---

func isNumericString(s string) bool {
	if s == "" {
		return true
	}
	_, err := strconv.Atoi(s)
	return err == nil
}

func toIntSet(elems []int) map[int]bool {
	s := make(map[int]bool, len(elems))
	for _, v := range elems {
		s[v] = true
	}
	return s
}

func setIntersect(a, b map[int]bool) []int {
	var result []int
	for v := range a {
		if b[v] {
			result = append(result, v)
		}
	}
	return result
}

func setUnion(a, b map[int]bool) []int {
	var result []int
	for v := range a {
		result = append(result, v)
	}
	for v := range b {
		if !a[v] {
			result = append(result, v)
		}
	}
	return result
}

func setSymDiff(a, b map[int]bool) []int {
	var result []int
	for v := range a {
		if !b[v] {
			result = append(result, v)
		}
	}
	for v := range b {
		if !a[v] {
			result = append(result, v)
		}
	}
	return result
}

func isSubset(a, b map[int]bool) bool {
	for v := range a {
		if !b[v] {
			return false
		}
	}
	return true
}

// tryCollectionConcat evaluates `left + right` as collection concatenation
// (spec/10). When both operands resolve to collections, the concatenated
// slice is bound under `name` in the store matching the left operand's kind
// (sets re-dedup). Returns true when the binding was performed.
func (e *Evaluator) tryCollectionConcat(name string, expr parser.Expression) bool {
	be, ok := expr.(*parser.BinaryExpression)
	if !ok || be.Token.Literal != "+" {
		return false
	}
	leftElems, leftKind, okL := e.resolveCollection(be.Left)
	rightElems, _, okR := e.resolveCollection(be.Right)
	if !okL || !okR {
		return false
	}
	joined := append(append([]int{}, leftElems...), rightElems...)
	switch leftKind {
	case kindArray:
		e.arrayValues[name] = joined
	case kindList:
		e.listValues[name] = joined
	case kindSet:
		e.setValues[name] = dedupSortInts(joined)
	}
	e.ensureIdentity(name)
	return true
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

// scopedBindingSnapshot captures an identifier's full binding across the
// evaluator's typed store maps so lexical block shadowing (spec/02 §blocks)
// can restore the outer value when the block scope exits.
type scopedBindingSnapshot struct {
	hasInt      bool
	intVal      int
	hasString   bool
	stringVal   string
	hasArray    bool
	arrayVal    []int
	hasList     bool
	listVal     []int
	hasSet      bool
	setVal      []int
	hasMap      bool
	mapVal      map[string]int
	hasMatrix   bool
	matrixVal   MatrixValue
	hasStepped  bool
	steppedVal  *parser.SteppedRangeExpression
	hasLambda   bool
	lambdaVal   *parser.LambdaExpression
	hasPartial  bool
	partialVal  *partialApplication
	hasIdentity bool
	identityVal int
}

func (s *scopedBindingSnapshot) any() bool {
	return s.hasInt || s.hasString || s.hasArray || s.hasList || s.hasSet ||
		s.hasMap || s.hasMatrix || s.hasStepped || s.hasLambda || s.hasPartial
}

// scopeFrame is one lexical scope. It records which names were (re)declared in
// this scope with `new`, together with a snapshot of each shadowed outer
// binding so leaveScope can restore it on block exit.
type scopeFrame struct {
	saved    map[string]*scopedBindingSnapshot
	declared map[string]bool
}

func (e *Evaluator) currentScope() *scopeFrame {
	if len(e.scopes) == 0 {
		return nil
	}
	return e.scopes[len(e.scopes)-1]
}

func (e *Evaluator) enterScope() {
	e.scopes = append(e.scopes, &scopeFrame{
		saved:    make(map[string]*scopedBindingSnapshot),
		declared: make(map[string]bool),
	})
}

func (e *Evaluator) leaveScope() {
	if len(e.scopes) == 0 {
		return
	}
	frame := e.scopes[len(e.scopes)-1]
	e.scopes = e.scopes[:len(e.scopes)-1]
	for name := range frame.declared {
		e.clearBinding(name)
		if sn, ok := frame.saved[name]; ok {
			e.restoreBinding(name, sn)
		}
	}
}

// declareBinding records a `new` binding in the innermost scope frame. If the
// name already binds to an outer variable, that outer binding is snapshotted so
// it can be restored when the scope exits (block-scoped shadowing, spec/02 §blocks).
func (e *Evaluator) declareBinding(name string) {
	if strings.HasPrefix(name, ".") {
		return
	}
	frame := e.currentScope()
	if frame == nil {
		return
	}
	if frame.declared[name] {
		return
	}
	if sn := e.snapshotBinding(name); sn.any() {
		frame.saved[name] = sn
	}
	frame.declared[name] = true
}

// snapshotBinding captures the current visible binding of name across all typed
// stores. It is called before a shadowing `new` overwrites the name.
func (e *Evaluator) snapshotBinding(name string) *scopedBindingSnapshot {
	sn := &scopedBindingSnapshot{}
	if v, ok := e.symbols[name]; ok {
		sn.hasInt, sn.intVal = true, v
	}
	if v, ok := e.stringSymbols[name]; ok {
		sn.hasString, sn.stringVal = true, v
	}
	if v, ok := e.arrayValues[name]; ok {
		sn.hasArray, sn.arrayVal = true, v
	}
	if v, ok := e.listValues[name]; ok {
		sn.hasList, sn.listVal = true, v
	}
	if v, ok := e.setValues[name]; ok {
		sn.hasSet, sn.setVal = true, v
	}
	if v, ok := e.mapValues[name]; ok {
		sn.hasMap, sn.mapVal = true, v
	}
	if v, ok := e.matrixValues[name]; ok {
		sn.hasMatrix, sn.matrixVal = true, v
	}
	if v, ok := e.steppedRanges[name]; ok {
		sn.hasStepped, sn.steppedVal = true, v
	}
	if v, ok := e.lambdaValues[name]; ok {
		sn.hasLambda, sn.lambdaVal = true, v
	}
	if v, ok := e.partialValues[name]; ok {
		sn.hasPartial, sn.partialVal = true, v
	}
	if v, ok := e.identities[name]; ok {
		sn.hasIdentity, sn.identityVal = true, v
	}
	return sn
}

// clearBinding removes every typed binding for name. Called on scope exit for
// names (re)declared within that scope.
func (e *Evaluator) clearBinding(name string) {
	delete(e.symbols, name)
	delete(e.stringSymbols, name)
	delete(e.arrayValues, name)
	delete(e.listValues, name)
	delete(e.setValues, name)
	delete(e.mapValues, name)
	delete(e.matrixValues, name)
	delete(e.steppedRanges, name)
	delete(e.lambdaValues, name)
	delete(e.partialValues, name)
	delete(e.identities, name)
}

// restoreBinding reinstates a snapshotted outer binding after a shadowed scope
// exits. Only the stores recorded in the snapshot are restored.
func (e *Evaluator) restoreBinding(name string, sn *scopedBindingSnapshot) {
	if sn.hasInt {
		e.symbols[name] = sn.intVal
	}
	if sn.hasString {
		e.stringSymbols[name] = sn.stringVal
	}
	if sn.hasArray {
		e.arrayValues[name] = sn.arrayVal
	}
	if sn.hasList {
		e.listValues[name] = sn.listVal
	}
	if sn.hasSet {
		e.setValues[name] = sn.setVal
	}
	if sn.hasMap {
		e.mapValues[name] = sn.mapVal
	}
	if sn.hasMatrix {
		e.matrixValues[name] = sn.matrixVal
	}
	if sn.hasStepped {
		e.steppedRanges[name] = sn.steppedVal
	}
	if sn.hasLambda {
		e.lambdaValues[name] = sn.lambdaVal
	}
	if sn.hasPartial {
		e.partialValues[name] = sn.partialVal
	}
	if sn.hasIdentity {
		e.identities[name] = sn.identityVal
	}
}