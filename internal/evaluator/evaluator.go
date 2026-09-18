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

// MatrixValue stores an N-dimensional tensor in row-major order
// (spec/10 §3.3). Elements are 1-based indexed along each dimension:
// M[i1, i2, ..., in] → Data[matrixOffset(Dims, {i1, ..., in})]. A 2D tensor
// is a matrix; higher-dimension tensors (3D cubes, 4D+ blocks) use the same
// flattened, stride-based representation.
type MatrixValue struct {
	Data []int
	Dims []int // per-dimension lengths, from most-significant to least
}

// matrixTotalSize returns the flattened element count of a tensor shaped by
// dims, i.e. the product of all dimension lengths.
func matrixTotalSize(dims []int) int {
	size := 1
	for _, d := range dims {
		size *= d
	}
	return size
}

// matrixOffset computes the 0-based linear offset for row-major storage of
// coordinates coords (1-based values, same length as dims). offset follows
// the iterative Horner form: (((i1-1)*d2 + (i2-1))*d3 + (i3-1))*… + (in-1).
func matrixOffset(dims, coords []int) int {
	offset := 0
	for i := range dims {
		offset = offset*dims[i] + (coords[i] - 1)
	}
	return offset
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
	// objectValues stores anonymous JSON-literal objects (spec/06-objects.md
	// §2.2) by name, retaining the parsed MapLiteral AST so heterogeneous
	// members (string + int) can be resolved lazily on member/index access.
	// An object is distinguished from an int-only map literal by the presence
	// of a string-valued member. This is the foundation of ``type()``/``.type()``
	// introspection returning "O" (the universal root Object).
	objectValues map[string]*ObjectValue
	// currentSelf, when non-nil, marks that the evaluator is currently executing
	// the body of an object constructor rule (spec/06 §2/§3). Bare `new` locals
	// declared during that window are captured as private instance members so
	// methods can read them by name without exporting them publicly.
	currentSelf  *ObjectValue
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
	// typeRegistry indexes every *parser.TypeDeclaration by name so that
	// `new obj := T(...)` resolves T to its default-constructor object template
	// (spec/06-objects.md §2 / §6 object_type).
	typeRegistry map[string]*parser.TypeDeclaration
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
		objectValues:   make(map[string]*ObjectValue),
		matrixValues:   make(map[string]MatrixValue),
		steppedRanges:  make(map[string]*parser.SteppedRangeExpression),
		identities:     make(map[string]int),
		nextID:         1,
		ruleRegistry:   make(map[string]*parser.RuleStatement),
		typeRegistry:   make(map[string]*parser.TypeDeclaration),
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
	// invocations regardless of definition order (spec/03 §5.2), and all
	// custom-type declarations so `new obj := T(...)` resolves T to its
	// default-constructor template (spec/06 §2 / §6).
	for _, stmt := range program.Statements {
		switch ts := stmt.(type) {
		case *parser.RuleStatement:
			e.ruleRegistry[ts.Name] = ts
		case *parser.TypeDeclaration:
			e.typeRegistry[ts.Name] = ts
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
	case *parser.TypeDeclaration:
		// Declarations are registered in Eval pass 1; executing the statement
		// has no runtime effect.
		return
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
		if s.Target != nil {
			// spec/06 §3 member dispatch: `apply obj.method` invokes the public
			// method on the instance for its side effects; the result is discarded.
			if me, isMember := s.Target.(*parser.MemberExpression); isMember {
				if ident, isIdent := me.Base.(*parser.Identifier); isIdent && len(me.Parts) == 1 {
					if obj, hasObj := e.objectValues[ident.Value]; hasObj {
						if method, hasMethod := obj.methods[me.Parts[0]]; hasMethod {
							e.callMethod(method, obj)
						}
					}
				}
			}
			break
		}
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
	case *parser.ZapStatement:
		// Object/map member removal (spec/06 §2.3): `zap obj.field;` / `zap obj["key"];`
		// removes one member. Plain `zap ident;` invalidates a whole binding
		// (spec/02 §2.3); the dotted/index wording resolves the member first.
		if s.Target != nil {
			if me, isMember := s.Target.(*parser.MemberExpression); isMember {
				if id, ok := me.Base.(*parser.Identifier); ok {
					key := me.Parts[0]
					if obj, hasObj := e.objectValues[id.Value]; hasObj {
						obj.remove(key)
						break
					}
					if m, hasMap := e.mapValues[id.Value]; hasMap {
						delete(m, key)
						break
					}
				}
			} else if idx, isIndex := s.Target.(*parser.IndexExpression); isIndex {
				if id, ok := idx.Left.(*parser.Identifier); ok {
					key := e.evalExpression(idx.Index)
					if obj, hasObj := e.objectValues[id.Value]; hasObj {
						obj.remove(key)
						break
					}
					if m, hasMap := e.mapValues[id.Value]; hasMap {
						delete(m, key)
						break
					}
				}
			}
		} else if s.Name != "" {
			e.clearBinding(s.Name)
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
				// `let M[*,c] := v` fills column c; `let M[r,c] := v` sets one cell;
				// an N-d tensor `let T[i1, ..., in] := v` sets one cell along n dims.
				if mat, hasMat := e.matrixValues[ident.Value]; hasMat {
					if i < len(pendingInts) {
						val := pendingInts[i]
						isWildcard := func(expr parser.Expression) bool {
							if id, isID := expr.(*parser.Identifier); isID {
								return id.Value == "*"
							}
							return false
						}
						dims := len(mat.Dims)
						// N-d full-coordinate single-cell write. Arity (one index per
						// dimension) must match the tensor rank and none of the
						// coordinates may be a wildcard. 2D falls through to the
						// specialised row/column logic below for unchanged behaviour.
						if dims >= 3 && len(idxExpr.ExtraIndices)+1 == dims {
							allCoord := true
							coords := make([]parser.Expression, 0, dims)
							coords = append(coords, idxExpr.Index)
							coords = append(coords, idxExpr.ExtraIndices...)
							for _, c := range coords {
								if isWildcard(c) {
									allCoord = false
									break
								}
							}
							if allCoord {
								inBounds := true
								offs := make([]int, dims)
								for k, ce := range coords {
									v := e.evalIntExpression(ce)
									offs[k] = v
									if v < 1 || v > mat.Dims[k] {
										inBounds = false
									}
								}
								if inBounds {
									mat.Data[matrixOffset(mat.Dims, offs)] = val
								}
								e.matrixValues[ident.Value] = mat
								continue
							}
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
								for r := 1; r <= mat.Dims[0]; r++ {
									mat.Data[(r-1)*mat.Dims[1]+(col-1)] = val
								}
							} else if colW {
								// M[r,*] — fill row r
								row := e.evalIntExpression(idxExpr.Index)
								for c := 1; c <= mat.Dims[1]; c++ {
									mat.Data[(row-1)*mat.Dims[1]+(c-1)] = val
								}
							} else {
								// M[r,c] — single cell
								row := e.evalIntExpression(idxExpr.Index)
								col := e.evalIntExpression(idxExpr.ExtraIndices[0])
								if row >= 1 && row <= mat.Dims[0] && col >= 1 && col <= mat.Dims[1] {
									mat.Data[(row-1)*mat.Dims[1]+(col-1)] = val
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
		// Object attribute-overlay write (spec/06 §2.3): `new obj.member := val`
		// or `new obj["key"] := val`. The base object already bound; this adds or
		// updates a single member on it rather than declaring a new identifier.
		if s.ObjectWrite != nil {
			if s.Value != nil {
				e.evalObjectWrite(s.ObjectWrite.Base, s.ObjectWrite.Key, s.Value)
			}
			return
		}
		names := s.Names
		if len(names) == 0 && s.Name != "" {
			names = []string{s.Name}
		}
		// Typed tensor declaration (spec/10 §3.3): `new M ∈ [Z](d1, d2, ..., dn)`
		// zero-initialises an n-dimensional tensor in row-major order. A 2D dim
		// list is a matrix; 3D+ is a tensor.
		if s.MatrixDims != nil {
			dims := make([]int, len(s.MatrixDims))
			copy(dims, s.MatrixDims)
			for _, name := range names {
				e.matrixValues[name] = MatrixValue{
					Data: make([]int, matrixTotalSize(dims)),
					Dims: dims,
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
				// Custom-type default constructor (spec/06 §2): `new obj := T(...)`
				// where T is a declared type. Build the object from its default
				// attributes, then bind it; no rule invocation involved.
				if e.tryTypeConstructorBinding(names[0], call) {
					break
				}
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
					// An empty `{}` is an empty anonymous object (spec/06 §2.3): it can
					// be instantiated empty and later enhanced with members via
					// attribute overlay (`new obj.member := v`). A non-empty set keeps
					// the mathematical-set semantics (spec/10 §3.4).
					if len(setLit.Elements) == 0 {
						e.objectValues[name] = newObjectValue()
						e.ensureIdentity(name)
						continue
					}
					elems := make([]int, len(setLit.Elements))
					for j, el := range setLit.Elements {
						elems[j] = e.evalIntExpression(el)
					}
					e.setValues[name] = dedupSortInts(elems)
					e.ensureIdentity(name)
				} else if mapLit, ok := s.Values[i].(*parser.MapLiteral); ok {
					// Anonymous object vs int-only map (spec/06 §2.2): a literal
					// with a string member is an object; otherwise it is a map.
					if mapLitIsObject(mapLit) {
						e.objectValues[name] = e.mapLiteralToObject(mapLit)
					} else {
						m := make(map[string]int)
						for _, pair := range mapLit.Pairs {
							key := e.evalExpression(pair.Key)
							m[key] = e.evalIntExpression(pair.Value)
						}
						e.mapValues[name] = m
					}
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
					// An empty `{}` is an empty anonymous object (spec/06 §2.3): it can
					// be instantiated empty and later enhanced with members via
					// attribute overlay (`new obj.member := v`). A non-empty set keeps
					// the mathematical-set semantics (spec/10 §3.4).
					if len(setLit.Elements) == 0 {
						e.objectValues[name] = newObjectValue()
						e.ensureIdentity(name)
					} else {
						elems := make([]int, len(setLit.Elements))
						for j, el := range setLit.Elements {
							elems[j] = e.evalIntExpression(el)
						}
						e.setValues[name] = dedupSortInts(elems)
						e.ensureIdentity(name)
					}
				} else if mapLit, ok := s.Value.(*parser.MapLiteral); ok {
					// Anonymous object vs int-only map (spec/06 §2.2): a literal
					// with a string member is an object; otherwise it is a map.
					if mapLitIsObject(mapLit) {
						e.objectValues[name] = e.mapLiteralToObject(mapLit)
					} else {
						m := make(map[string]int)
						for _, pair := range mapLit.Pairs {
							key := e.evalExpression(pair.Key)
							m[key] = e.evalIntExpression(pair.Value)
						}
						e.mapValues[name] = m
					}
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
	case *parser.ImmediatelyInvokedLambda:
		// Spec/07 §6.3 IIFE (immediate_call): evaluate the argument list in the
		// caller's frame, then run the inline lambda body in a fresh pure frame.
		// Yields the body's result value, not a stored L reference.
		return e.callImmediatelyInvokedLambda(expr), e.allocID()
	case *parser.CallExpression:
		// `type(...)` is an introspection intrinsic, not a rule call: it is
		// resolved through the string path when compared, and yields 0 in int
		// context so it never falls through to callRule (which would treat the
		// name "type" as an unregistered rule).
		if expr.Name == "type" {
			return 0, e.allocID()
		}
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
		// Anonymous object attribute (spec/06 §2.2): `obj.age` in int context
		// resolves a numeric member; string members degenerate to 0.
		if ident, isIdent := expr.Base.(*parser.Identifier); isIdent && !expr.IsCall && len(expr.Parts) == 1 {
			if obj, hasObj := e.objectValues[ident.Value]; hasObj {
				isStr, _, iVal, found := obj.fieldValue(expr.Parts[0])
				if found && !isStr {
					return iVal, e.ensureIdentity(ident.Value)
				}
				if found {
					return 0, e.allocID()
				}
				// Public method dispatch (spec/06 §3): a bare `obj.method` in an
				// expression invokes the exported method and yields its int result.
				if method, hasMethod := obj.methods[expr.Parts[0]]; hasMethod {
					if v, ok := e.callMethod(method, obj); ok {
						return v, e.allocID()
					}
				}
			}
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
			// Anonymous object attribute overlay (spec/06 §2.3): `obj["age"]`.
			// Numeric members yield their value; string members degenerate to 0
			// in int context (the string path handles string fields).
			if obj, hasObj := e.objectValues[ident.Value]; hasObj {
				key := e.evalExpression(expr.Index)
				isStr, _, iVal, found := obj.fieldValue(key)
				if found && !isStr {
					return iVal, e.ensureIdentity(ident.Value)
				}
				if found {
					return 0, e.allocID()
				}
			}
			// Map key indexing (spec/10 §3.5): M[k] looks up the value for key k.
			if m, hasMap := e.mapValues[ident.Value]; hasMap {
				key := e.evalExpression(expr.Index)
				return m[key], e.ensureIdentity(ident.Value)
			}
			// Tensor indexing (spec/10 §3.3): M[i1, ..., in] with 1-based indices
			// along each dimension. A 2D tensor is a matrix; 3D+ is a tensor.
			if mat, hasMat := e.matrixValues[ident.Value]; hasMat {
				if len(expr.ExtraIndices)+1 == len(mat.Dims) {
					coords := make([]int, 0, len(mat.Dims))
					coords = append(coords, e.evalIntExpression(expr.Index))
					for _, ei := range expr.ExtraIndices {
						coords = append(coords, e.evalIntExpression(ei))
					}
					inBounds := true
					for i, c := range coords {
						if c < 1 || c > mat.Dims[i] {
							inBounds = false
						}
					}
					if inBounds {
						return mat.Data[matrixOffset(mat.Dims, coords)], e.ensureIdentity(ident.Value)
					}
					return 0, e.allocID()
				}
				// Single-index tensor access: flatten to 1-based linear index.
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
			// Versatile `is` (spec/06 §1): when the right operand is a non-bound
			// name that denotes a type, this is a type-membership test instead of
			// a pointer-identity comparison — e.g. `obj is Object` returns true
			// when obj's reported type is the universal root Object.
			if rid, isRIdent := expr.Right.(*parser.Identifier); isRIdent && !e.isBoundValueName(rid.Value) {
				if canon, isType := e.typeNameRef(rid.Value); isType {
					result := e.typeName(expr.Left) == canon
					if expr.Token.Type == token.IS_NOT || lit == "is not" {
						result = !result
					}
					if result {
						return 1, 0
					}
					return 0, 0
				}
			}
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
			// Object/map structural equality (spec/06 §2.2): when either operand
			// denotes an object or map value, compare key sets and per-key values
			// instead of forcing both through the int domain.
			if e.isObjectExpr(expr.Left) || e.isObjectExpr(expr.Right) {
				return e.structuralObjectEq(expr.Left, expr.Right), 0
			}
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
// objField is one heterogeneous member of an ObjectValue: a string-valued
// member carries sVal; an int-valued member carries iVal (spec/06 §2.2, where
// anonymous objects mix string and numeric members).
type objField struct {
	isStr bool
	sVal  string
	iVal  int
}

// ObjectValue is a *mutable* anonymous-object / JSON-literal value (spec/06
// §2.2). Unlike the frozen AST it replaces, an ObjectValue is a heap-backed
// dictionary that supports dynamic member insertion (`new obj.k := v`),
// update, and removal (`zap obj.k` / `zap obj["k"]`) after instantiation
// (spec/06 §2.3 attribute overlay). Dot-notation and index-notation address
// the same slot, so an object created empty can be enhanced with members and
// compared structurally to a literal.
type ObjectValue struct {
	fields map[string]objField
	// typeName is the Bee type name reported by type()/obj.type(). Anonymous
	// JSON objects resolve to "Object" (the renamed universal root, formerly
	// "O"); instances built by a constructor rule carry the rule's name (e.g.
	// `Foo(...)` yields "Foo"). Empty means anonymous → "Object" (spec/06 §1).
	typeName string
	// methods holds the public methods exported on this instance (spec/06 §3):
	// nested `rule .sum(self ∈ Type) => (...)` definitions declared inside the
	// constructor body are keyed here by bare member name ("sum") so dot-notation
	// `obj.sum` and `apply obj.sum` dispatch to a callMethod frame. Methods are
	// per-instance closures over the object's own dictionary.
	methods map[string]*parser.RuleStatement
	// privates captures the constructor's private-local variables (spec/06 §3):
	// bare, non-`self.`-prefixed `new` bindings declared in the constructor body
	// (e.g. `new sentinel := 99`). They are NOT exported as public members, so
	// `c.sentinel` from outside is unbound; they are only injected into the frame
	// of methods defined on this instance so a method body may read them directly
	// by bare name (encapsulation without a prefix).
	privates map[string]objField
}

func newObjectValue() *ObjectValue {
	return &ObjectValue{fields: make(map[string]objField), methods: make(map[string]*parser.RuleStatement), privates: make(map[string]objField)}
}

// objString serializes an ObjectValue into a canonical, deterministic,
// JSON-like string with member keys sorted alphabetically (spec/06 §2.4).
// String-valued members are quoted; int-valued members are unquoted. The
// ordering is sorted so identical objects always render identically, which is
// what makes autonomous @EXPECT verification of printed objects reliable.
func objString(o *ObjectValue) string {
	if o == nil {
		return "{}"
	}
	keys := make([]string, 0, len(o.fields))
	for k := range o.fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		f := o.fields[k]
		if f.isStr {
			parts = append(parts, k+": \""+f.sVal+"\"")
		} else {
			parts = append(parts, k+": "+strconv.Itoa(f.iVal))
		}
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// typeNameOf returns the Bee type name the object reports via type()/obj.type().
// Anonymous objects (no explicit constructor type) resolve to the universal
// root "Object" (formerly "O") (spec/06 §1).
func (o *ObjectValue) typeNameOf() string {
	if o == nil || o.typeName == "" {
		return "Object"
	}
	return o.typeName
}

// fieldValue resolves a single named member, returning either its string value
// or its int value. ok is false when the key is absent. Alias of the map slot
// regardless of how it was addressed (attribute overlay, spec/06 §2.3).
func (o *ObjectValue) fieldValue(key string) (isString bool, sVal string, iVal int, ok bool) {
	f, found := o.fields[key]
	return f.isStr, f.sVal, f.iVal, found
}

func (o *ObjectValue) set(key string, f objField) {
	o.fields[key] = f
}

func (o *ObjectValue) remove(key string) {
	delete(o.fields, key)
}

func (o *ObjectValue) fieldCount() int {
	return len(o.fields)
}

// mapLitIsObject reports whether a `{...}` literal should bind as an anonymous
// object (spec/06 §2.2) rather than an int-only map (spec/10 §4). The two are
// lexically identical; the presence of a string-valued member selects the
// object/record interpretation so heterogeneous objects (e.g.
// `{name: "Cleopatra", age: 15}`) keep their string fields instead of forcing
// every value through the int domain. An empty literal `{}` is treated as an
// empty anonymous object (spec/06 §2.3: an object can be instantiated empty and
// later enhanced with members).
func mapLitIsObject(ml *parser.MapLiteral) bool {
	if len(ml.Pairs) == 0 {
		return true
	}
	for _, pair := range ml.Pairs {
		if _, isStr := pair.Value.(*parser.StringLiteral); isStr {
			return true
		}
	}
	return false
}

// mapLiteralToObject materialises a `{...}` literal into a mutable ObjectValue
// (spec/06 §2.2), resolving each pair's member key and heterogeneous value.
func (e *Evaluator) mapLiteralToObject(ml *parser.MapLiteral) *ObjectValue {
	o := newObjectValue()
	for _, pair := range ml.Pairs {
		key := e.evalExpression(pair.Key)
		if strLit, isStr := pair.Value.(*parser.StringLiteral); isStr {
			o.set(key, objField{isStr: true, sVal: strLit.Value})
		} else {
			o.set(key, objField{iVal: e.evalIntExpression(pair.Value)})
		}
	}
	return o
}

func (e *Evaluator) evalObjectWrite(base string, key string, v parser.Expression) bool {
	obj, hasObj := e.objectValues[base]
	if !hasObj {
		// A map-bound identifier can also take overlay writes, but the object
		// store is authoritative here; an unbound base is a no-op.
		return false
	}
	if strLit, isStr := v.(*parser.StringLiteral); isStr {
		obj.set(key, objField{isStr: true, sVal: strLit.Value})
		return true
	}
	if strIdent, isStr := v.(*parser.Identifier); isStr {
		if sv, hasStr := e.stringSymbols[strIdent.Value]; hasStr {
			obj.set(key, objField{isStr: true, sVal: sv})
			return true
		}
	}
	obj.set(key, objField{iVal: e.evalIntExpression(v)})
	return true
}

// setObjectField writes a heterogeneous field value onto an ObjectValue,
// mirroring evalObjectWrite: a string literal/alias is stored as a string
// field, anything else as an int field (spec/06 §2.2).
func (e *Evaluator) setObjectField(obj *ObjectValue, key string, v parser.Expression) {
	if strLit, isStr := v.(*parser.StringLiteral); isStr {
		obj.set(key, objField{isStr: true, sVal: strLit.Value})
		return
	}
	if strIdent, isStr := v.(*parser.Identifier); isStr {
		if sv, hasStr := e.stringSymbols[strIdent.Value]; hasStr {
			obj.set(key, objField{isStr: true, sVal: sv})
			return
		}
	}
	obj.set(key, objField{iVal: e.evalIntExpression(v)})
}

// tryTypeConstructorBinding resolves `new obj := T(...)` when T is a declared
// custom type with a default-constructor template (spec/06 §2 / §6): it builds
// an ObjectValue seeded with the type's default attributes, overwrites any
// member supplied positionally (in declaration order) or by name (`T(a2: 0)`),
// binds it under name, and reports whether it handled the call. It returns
// false when T is not a declared type, so the caller falls through to a rule
// invocation.
func (e *Evaluator) tryTypeConstructorBinding(name string, call *parser.CallExpression) bool {
	td, ok := e.typeRegistry[call.Name]
	if !ok {
		return false
	}
	obj := newObjectValue()
	for i, prop := range td.Props {
		val := prop.Default
		// Positional arguments override in declaration order: the i-th arg maps
		// to the i-th declared attribute.
		if i < len(call.Args) {
			val = call.Args[i]
		}
		e.setObjectField(obj, prop.Key, val)
	}
	for key, val := range call.NamedArgs {
		e.setObjectField(obj, key, val)
	}
	e.objectValues[name] = obj
	e.ensureIdentity(name)
	delete(e.symbols, name) // an object is not an int binding
	return true
}

// isConstructorRule reports whether rs is an object constructor rule
// (spec/06 §2 / §6 constructor_def): a rule whose single declared result is
// bound to `self`, e.g. `rule Foo(a, b ∈ N) => (self ∈ Foo):`. The instance
// type name is the rule's own name.
func isConstructorRule(rs *parser.RuleStatement) bool {
	if rs == nil || rs.ForwardDecl || rs.Body == nil {
		return false
	}
	return len(rs.Results) == 1 && rs.Results[0] == "self"
}

// runConstructor executes an object constructor rule (spec/06 §2): it creates
// a fresh ObjectValue whose reported type name is the rule name, binds it as
// `self` inside a private constructor frame, binds the positional parameters,
// runs the constructor body (so `new self.a := a` and nested method rules take
// effect), then restores the caller's value stores and returns the instance.
// The returned ObjectValue is bound by the caller via bindResultValue.
func (e *Evaluator) runConstructor(rs *parser.RuleStatement, call *parser.CallExpression, argVals []int) []interface{} {
	instance := newObjectValue()
	instance.typeName = rs.Name

	// Save the caller's full value-store frames.
	savedObjects := e.objectValues
	savedMaps := e.mapValues
	savedSymbols := e.symbols
	savedStrings := e.stringSymbols
	savedArrays := e.arrayValues
	savedLists := e.listValues
	savedSets := e.setValues
	savedMatrix := e.matrixValues
	savedBoxed := e.boxedCells
	savedExiting := e.exitingRule
	savedContinue := e.continueCycle
	savedTarget := e.continueTargetDepth
	savedLoopStack := e.loopLabelStack

	// Fresh private frame: only `self` and the parameters are visible.
	e.symbols = make(map[string]int)
	e.stringSymbols = make(map[string]string)
	e.arrayValues = make(map[string][]int)
	e.listValues = make(map[string][]int)
	e.setValues = make(map[string][]int)
	e.matrixValues = make(map[string]MatrixValue)
	e.mapValues = make(map[string]map[string]int)
	e.objectValues = make(map[string]*ObjectValue)
	e.boxedCells = make(map[string]*BoxedCell)
	e.exitingRule = false
	e.continueCycle = false
	e.continueTargetDepth = 0
	e.loopLabelStack = nil

	// Bind the instance to `self` so `new self.a := a` writes attributes.
	e.objectValues["self"] = instance

	// Bind positional parameters.
	for i, param := range rs.Params {
		if i < len(argVals) {
			e.symbols[param] = argVals[i]
		}
	}

	// Register public method definitions (spec/06 §3): nested `rule .sum(self
	// ∈ Type) => (...)` declarations in the constructor body are exported on the
	// instance. They are not executed for their own side effects (evalStatement
	// ignores non-`main` rules); instead they are captured here so later
	// dot-notation `obj.sum` / `apply obj.sum` can dispatch to them.
	for _, stmt := range rs.Body.Statements {
		if mst, isRule := stmt.(*parser.RuleStatement); isRule && strings.HasPrefix(mst.Name, ".") && !mst.ForwardDecl && mst.Body != nil {
			instance.methods[mst.Name[1:]] = mst
		}
	}

	// Execute the constructor body.
	for _, stmt := range rs.Body.Statements {
		if e.exitingRule {
			break
		}
		e.evalStatement(stmt)
	}

	// Restore the caller's frame.
	e.objectValues = savedObjects
	e.mapValues = savedMaps
	e.symbols = savedSymbols
	e.stringSymbols = savedStrings
	e.arrayValues = savedArrays
	e.listValues = savedLists
	e.setValues = savedSets
	e.matrixValues = savedMatrix
	e.boxedCells = savedBoxed
	e.exitingRule = savedExiting
	e.continueCycle = savedContinue
	e.continueTargetDepth = savedTarget
	e.loopLabelStack = savedLoopStack

	return []interface{}{instance}
}

// callMethod executes a public object method (spec/06 §3): a nested `rule
// .name(self ∈ Type) => (result ∈ R)` registered on an instance. It creates a
// private frame in which the instance is bound as `self` (so the body's
// `self.a` member reads resolve against the object both as the typed `self`
// parameter and via the object store), runs the body, and returns the declared
// result value. This mirrors runConstructor's frame discipline.
func (e *Evaluator) callMethod(rs *parser.RuleStatement, instance *ObjectValue) (int, bool) {
	if rs == nil || rs.ForwardDecl || rs.Body == nil || instance == nil {
		return 0, false
	}

	savedObjects := e.objectValues
	savedMaps := e.mapValues
	savedSymbols := e.symbols
	savedStrings := e.stringSymbols
	savedArrays := e.arrayValues
	savedLists := e.listValues
	savedSets := e.setValues
	savedMatrix := e.matrixValues
	savedBoxed := e.boxedCells
	savedExiting := e.exitingRule
	savedContinue := e.continueCycle
	savedTarget := e.continueTargetDepth
	savedLoopStack := e.loopLabelStack

	// Fresh private frame: only `self` and the parameters are visible.
	e.symbols = make(map[string]int)
	e.stringSymbols = make(map[string]string)
	e.arrayValues = make(map[string][]int)
	e.listValues = make(map[string][]int)
	e.setValues = make(map[string][]int)
	e.matrixValues = make(map[string]MatrixValue)
	e.mapValues = make(map[string]map[string]int)
	e.objectValues = make(map[string]*ObjectValue)
	e.boxedCells = make(map[string]*BoxedCell)
	e.exitingRule = false
	e.continueCycle = false
	e.continueTargetDepth = 0
	e.loopLabelStack = nil

	// Bind the receiver instance as `self` so `self.a` member reads and writes
	// resolve against the object. Parameter-less methods need no other binding.
	e.objectValues["self"] = instance

	// Initialise declared results to zero (spec/03 §2.3 default init). The
	// `self` formal param is satisfied by the receiver, so it is skipped.
	for _, res := range rs.Results {
		if res != "self" {
			e.symbols[res] = 0
		}
	}

	// Execute the method body.
	for _, stmt := range rs.Body.Statements {
		if e.exitingRule {
			break
		}
		e.evalStatement(stmt)
	}

	// Capture the first declared result.
	result := 0
	for _, res := range rs.Results {
		if res != "self" {
			if v, ok := e.symbols[res]; ok {
				result = v
			}
			break
		}
	}

	// Restore the caller's frame.
	e.objectValues = savedObjects
	e.mapValues = savedMaps
	e.symbols = savedSymbols
	e.stringSymbols = savedStrings
	e.arrayValues = savedArrays
	e.listValues = savedLists
	e.setValues = savedSets
	e.matrixValues = savedMatrix
	e.boxedCells = savedBoxed
	e.exitingRule = savedExiting
	e.continueCycle = savedContinue
	e.continueTargetDepth = savedTarget
	e.loopLabelStack = savedLoopStack

	return result, true
}

// isObjectExpr reports whether an expression denotes an object/map value (a
// stored identifier or a literal), used to route equality to structural
// comparison (spec/06 §2.2) instead of the integer path.
func (e *Evaluator) isObjectExpr(expr parser.Expression) bool {
	switch x := expr.(type) {
	case *parser.Identifier:
		if _, ok := e.objectValues[x.Value]; ok {
			return true
		}
		_, ok := e.mapValues[x.Value]
		return ok
	case *parser.MapLiteral:
		return true
	}
	return false
}

// objectFieldMap normalises an object/map expression to a key→field map so two
// such values can be compared structurally (same key set, per-key value
// equality).
func (e *Evaluator) objectFieldMap(expr parser.Expression) (map[string]objField, bool) {
	switch x := expr.(type) {
	case *parser.Identifier:
		if obj, ok := e.objectValues[x.Value]; ok {
			m := make(map[string]objField, obj.fieldCount())
			for k, f := range obj.fields {
				m[k] = f
			}
			return m, true
		}
		if mp, ok := e.mapValues[x.Value]; ok {
			m := make(map[string]objField, len(mp))
			for k, v := range mp {
				m[k] = objField{iVal: v}
			}
			return m, true
		}
		return nil, false
	case *parser.MapLiteral:
		m := make(map[string]objField, len(x.Pairs))
		for _, pair := range x.Pairs {
			k := e.evalExpression(pair.Key)
			if strLit, isStr := pair.Value.(*parser.StringLiteral); isStr {
				m[k] = objField{isStr: true, sVal: strLit.Value}
			} else {
				m[k] = objField{iVal: e.evalIntExpression(pair.Value)}
			}
		}
		return m, true
	}
	return nil, false
}

// structuralObjectEq compares two object/map expressions structurally (spec/06
// §2.2): identical key sets and per-key equal values. Returns 1 when equal, 0
// otherwise.
func (e *Evaluator) structuralObjectEq(a, b parser.Expression) int {
	fa, okA := e.objectFieldMap(a)
	fb, okB := e.objectFieldMap(b)
	if !okA || !okB {
		return 0
	}
	if len(fa) != len(fb) {
		return 0
	}
	for k, va := range fa {
		vb, ok := fb[k]
		if !ok {
			return 0
		}
		if va.isStr != vb.isStr {
			return 0
		}
		if va.isStr {
			if va.sVal != vb.sVal {
				return 0
			}
		} else if va.iVal != vb.iVal {
			return 0
		}
	}
	return 1
}

// typeName introspects an expression and returns its Bee type-name string
// (spec/06 §1 universal entity model). The anonymous object resolves to "O"
// (the universal root Object); native ints return "Z", strings "S", reals
// "R", booleans "B", collections "A"/"L"/"E"/"M". Not all branches are
// reachable from every expression context, but the mapping is total and
// deterministic for the bootstrap evaluator.
func (e *Evaluator) typeName(expr parser.Expression) string {
	switch x := expr.(type) {
	case *parser.Identifier:
		if obj, ok := e.objectValues[x.Value]; ok {
			return obj.typeNameOf()
		}
		if _, ok := e.stringSymbols[x.Value]; ok {
			return "S"
		}
		if _, ok := e.arrayValues[x.Value]; ok {
			return "A"
		}
		if _, ok := e.listValues[x.Value]; ok {
			return "L"
		}
		if _, ok := e.setValues[x.Value]; ok {
			return "E"
		}
		if _, ok := e.mapValues[x.Value]; ok {
			return "M"
		}
		if _, ok := e.matrixValues[x.Value]; ok {
			return "Z"
		}
		if _, ok := e.symbols[x.Value]; ok {
			return "Z"
		}
		if x.Value == "True" || x.Value == "true" || x.Value == "False" || x.Value == "false" {
			return "B"
		}
		return "?"
	case *parser.IntegerLiteral:
		return "Z"
	case *parser.StringLiteral:
		return "S"
	case *parser.ArrayLiteral:
		return "A"
	case *parser.ListLiteral:
		return "L"
	case *parser.SetLiteral:
		return "E"
	case *parser.MapLiteral:
		return "M"
	}
	return "?"
}

// isBoundValueName reports whether name is currently bound to a runtime value
// in any of the evaluator's value stores. It is used to disambiguate the
// versatile `is` operator (spec/06 §1 + Decision 14): when the right operand
// is a *non-bound* name that denotes a type, `is` performs a type-membership
// check; when both operands are bound values it stays a pointer-identity
// comparison. A name that is a type but also bound as a value keeps identity
// semantics.
func (e *Evaluator) isBoundValueName(name string) bool {
	if _, ok := e.symbols[name]; ok {
		return true
	}
	if _, ok := e.stringSymbols[name]; ok {
		return true
	}
	if _, ok := e.objectValues[name]; ok {
		return true
	}
	if _, ok := e.arrayValues[name]; ok {
		return true
	}
	if _, ok := e.listValues[name]; ok {
		return true
	}
	if _, ok := e.setValues[name]; ok {
		return true
	}
	if _, ok := e.mapValues[name]; ok {
		return true
	}
	if _, ok := e.matrixValues[name]; ok {
		return true
	}
	if _, ok := e.boxedCells[name]; ok {
		return true
	}
	if _, ok := e.closureObjects[name]; ok {
		return true
	}
	return false
}

// typeNameRef canonicalises a type-name spelling to the canonical type-name
// string produced by typeName, and reports whether name denotes a TYPE at all.
// Builtin Bee type names (Object, N/Z/R, Str/S, Bool/B, and the collection
// types) plus every registered custom type qualify. This powers the versatile
// `is` operator: `x is TypeName` becomes a type-membership test (spec/06 §1).
func (e *Evaluator) typeNameRef(name string) (string, bool) {
	switch name {
	case "Object":
		return "Object", true
	case "N", "Z", "Int", "Integer":
		return "Z", true
	case "R", "Real", "Float":
		return "R", true
	case "S", "Str", "String":
		return "S", true
	case "B", "Bool", "Boolean":
		return "B", true
	case "A", "Array":
		return "A", true
	case "L", "List":
		return "L", true
	case "E", "Set":
		return "E", true
	case "M", "Map", "Dict", "Dictionary":
		return "M", true
	}
	if _, ok := e.typeRegistry[name]; ok {
		return name, true
	}
	return "", false
}

func (e *Evaluator) isStringExpr(node parser.Expression) bool {
	switch expr := node.(type) {
	case *parser.StringLiteral:
		return true
	case *parser.Identifier:
		_, ok := e.stringSymbols[expr.Value]
		return ok
	case *parser.CallExpression:
		// `type(...)` yields a type-name string (spec/06 §1).
		return expr.Name == "type"
	case *parser.MemberExpression:
		// `expr.type()` introspection call yields a type-name string.
		if expr.IsCall && len(expr.Parts) == 1 && expr.Parts[0] == "type" {
			return true
		}
		// `obj.field` where field is a string-valued object member.
		if ident, isIdent := expr.Base.(*parser.Identifier); isIdent && len(expr.Parts) == 1 {
			if obj, hasObj := e.objectValues[ident.Value]; hasObj {
				isStr, _, _, found := obj.fieldValue(expr.Parts[0])
				return found && isStr
			}
		}
	case *parser.IndexExpression:
		// `obj["field"]` where field is a string-valued object member.
		if ident, isIdent := expr.Left.(*parser.Identifier); isIdent {
			if obj, hasObj := e.objectValues[ident.Value]; hasObj {
				key := e.evalExpression(expr.Index)
				isStr, _, _, found := obj.fieldValue(key)
				return found && isStr
			}
		}
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
			// Object attribute overlay (spec/06 §2.3): `obj["field"]`.
			if obj, hasObj := e.objectValues[ident.Value]; hasObj {
				key := e.evalExpression(expr.Index)
				isStr, sVal, iVal, found := obj.fieldValue(key)
				if found {
					if isStr {
						return sVal
					}
					return strconv.Itoa(iVal)
				}
			}
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
		if obj, hasObj := e.objectValues[expr.Value]; hasObj {
			return objString(obj)
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
	case *parser.CallExpression:
		// `type(expr)` introspection (spec/06 §1): return the type-name string.
		if expr.Name == "type" && len(expr.Args) == 1 {
			return e.typeName(expr.Args[0])
		}
		return strconv.Itoa(e.evalIntExpression(expr))
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
		// `expr.type()` introspection (spec/06 §1) returns the type-name string.
		if expr.IsCall && len(expr.Parts) == 1 && expr.Parts[0] == "type" {
			return e.typeName(expr.Base)
		}
		if ident, isIdent := expr.Base.(*parser.Identifier); isIdent && len(expr.Parts) == 1 && !expr.IsCall {
			// Anonymous object attribute (spec/06 §2.2): `obj.field`.
			if obj, hasObj := e.objectValues[ident.Value]; hasObj {
				isStr, sVal, iVal, found := obj.fieldValue(expr.Parts[0])
				if found {
					if isStr {
						return sVal
					}
					return strconv.Itoa(iVal)
				}
				// Public method dispatch (spec/06 §3): a bare `obj.method` in any
				// eager/evalExpression context invokes the method and yields its
				// int result as a string.
				if method, hasMethod := obj.methods[expr.Parts[0]]; hasMethod {
					if v, ok := e.callMethod(method, obj); ok {
						return strconv.Itoa(v)
					}
				}
			}
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

	// Constructor rule (spec/06 §2): a rule whose single result is bound to
	// `self` (e.g. `rule Foo(a, b ∈ N) => (self ∈ Foo):`) constructs an object
	// instance. It runs the body with `self` bound to a fresh ObjectValue, then
	// returns that instance so the caller can bind it (`new x := Foo(...)`).
	// This is checked here (after argument evaluation, before any caller frame
	// is saved) because the constructor manages its own value-store frames.
	if isConstructorRule(rule) {
		return e.runConstructor(rule, call, argVals)
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

// callImmediatelyInvokedLambda evaluates an IIFE (spec/07 §6.3 immediate_call):
// the argument list is evaluated in the caller's frame, then the inline lambda
// body runs in a fresh pure frame bound to those arguments. It shares
// evalLambdaBody's purity discipline, so the body cannot see outer state or
// call a rule, and yields a value rather than a stored L reference.
func (e *Evaluator) callImmediatelyInvokedLambda(iie *parser.ImmediatelyInvokedLambda) int {
	argVals := make([]int, len(iie.Args))
	for i, arg := range iie.Args {
		argVals[i] = e.evalIntExpression(arg)
	}
	return e.evalLambdaBody(iie.Lambda, argVals)
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
	if ov, isObj := val.(*ObjectValue); isObj {
		e.objectValues[name] = ov
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
	hasObject   bool
	objectVal   *ObjectValue
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
		s.hasMap || s.hasObject || s.hasMatrix || s.hasStepped || s.hasLambda || s.hasPartial
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
	if v, ok := e.objectValues[name]; ok {
		sn.hasObject, sn.objectVal = true, v
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
	delete(e.objectValues, name)
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
	if sn.hasObject {
		e.objectValues[name] = sn.objectVal
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
