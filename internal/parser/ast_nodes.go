// internal/parser/ast_nodes.go
package parser

import (
	"bee/internal/token"
)

type Node interface {
	Pos() token.Pos
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}
func (bs *BlockStatement) Pos() token.Pos { return bs.Token.Pos }

type IfStatement struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode() {}
func (is *IfStatement) Pos() token.Pos { return is.Token.Pos }

type PrintStatement struct {
	Token       token.Token
	Expressions []Expression
	Separator   Expression
	// NamedArgs preserves any curried `(name: value, ...)` postfix on the
	// call site. Map keys are the identifiers; values are the bound
	// expressions. Reserved by the named-parameter slot introduced in
	// spec/03-rules.md §2.4 / §3.1 and spec/02-statements.md §5 io_stmt
	// (Decision 6, 2026-09-12). For the level0 print rule's signature the
	// canonical name is `sep`; other names are forwarded to the evaluator
	// but currently unused at the bootstrap level.
	NamedArgs map[string]Expression
}

func (ps *PrintStatement) statementNode() {}
func (ps *PrintStatement) Pos() token.Pos { return ps.Token.Pos }

type AssignmentStatement struct {
	Token  token.Token
	Names  []*Identifier
	Values []Expression
}

func (as *AssignmentStatement) statementNode() {}
func (as *AssignmentStatement) Pos() token.Pos { return as.Token.Pos }

type AssertStatement struct {
	Token     token.Token
	Condition Expression
}

func (as *AssertStatement) statementNode() {}
func (as *AssertStatement) Pos() token.Pos { return as.Token.Pos }

type ExpectStatement struct {
	Token     token.Token
	Condition Expression
}

func (es *ExpectStatement) statementNode() {}
func (es *ExpectStatement) Pos() token.Pos { return es.Token.Pos }

type DeclarationStatement struct {
	Token  token.Token
	Name   string
	Names  []string
	Value  Expression
	Values []Expression
}

func (ds *DeclarationStatement) statementNode() {}
func (ds *DeclarationStatement) Pos() token.Pos { return ds.Token.Pos }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) Pos() token.Pos  { return i.Token.Pos }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) Pos() token.Pos  { return sl.Token.Pos }

type IntegerLiteral struct {
	Token token.Token
	Value string
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) Pos() token.Pos  { return il.Token.Pos }

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}
func (al *ArrayLiteral) Pos() token.Pos  { return al.Token.Pos }

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode() {}
func (ie *IndexExpression) Pos() token.Pos  { return ie.Token.Pos }

type BinaryExpression struct {
	Token token.Token
	Left  Expression
	Right Expression
}

func (be *BinaryExpression) expressionNode() {}
func (be *BinaryExpression) Pos() token.Pos  { return be.Token.Pos }

// PrefixExpression implements a nullary/unary prefix operator node.
// It is the canonical AST form for radicals (√, ²√, ³√, …) per
// D10 (Ratified 2026-09-12) and for logical NOT (¬, "not").
// Operator carries the literal form so the evaluator can dispatch on
// the operator family without losing the degree-n information encoded
// in the leading superscript of a radical.
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}
func (pe *PrefixExpression) Pos() token.Pos  { return pe.Token.Pos }

// TernaryExpression implements the parenthesised conditional expression
// selector per spec/02-statements.md §3.2:
//
//	( expr_true if condition else expr_false )
//
// It is always parenthesised in Bee, so the parser folds it in parsePrimary
// at the LPAREN case once it sees the `if` keyword after the first operand.
// Precedence is therefore syntactic (the surrounding parentheses), not part
// of the climbing ladder.
type TernaryExpression struct {
	Token     token.Token
	Condition Expression // the `if` test expression
	Then      Expression // expr_true — evaluated when Condition is truthy (≠ 0)
	Else      Expression // expr_false — evaluated when Condition is falsy (== 0)
}

func (te *TernaryExpression) expressionNode() {}
func (te *TernaryExpression) Pos() token.Pos  { return te.Token.Pos }

// Issue 14 — new AST nodes to back the rigid-syntax-gate dispatch table.
// Each block maps directly to a section in spec/02-statements.md §5 EBNF / §1.

type RuleStatement struct {
	Token token.Token
	Name  string
	// Params holds the primary positional parameter names (spec/03 §2.1).
	Params []string
	// Results holds the declared result variable names (spec/03 §2.3).
	Results []string
	// Body holds the indented statements that follow the rule header,
	// terminated by the aligned `return;`. See spec/03-rules.md §2.1.
	// Nil for forward declarations (signature ending in `;`).
	Body *BlockStatement
	// ForwardDecl is true when the rule was declared as a signature-only
	// forward declaration (no body). See spec/03-rules.md §5.2.
	ForwardDecl bool
	// IsStub emitted W0901 if the signature grammar is partial (Decision 6
	// Phase 7.2 gate). The compiler still emits the body but flags the gap.
	SignatureGrammarPartial bool
}

func (rs *RuleStatement) statementNode() {}
func (rs *RuleStatement) Pos() token.Pos { return rs.Token.Pos }

type ZapStatement struct {
	Token token.Token
	Name  string
}

func (zs *ZapStatement) statementNode() {}
func (zs *ZapStatement) Pos() token.Pos { return zs.Token.Pos }

type WriteStatement struct {
	Token token.Token
	Value Expression
}

func (ws *WriteStatement) statementNode() {}
func (ws *WriteStatement) Pos() token.Pos { return ws.Token.Pos }

type ReadStatement struct {
	Token  token.Token
	Target *Identifier
}

func (rs *ReadStatement) statementNode() {}
func (rs *ReadStatement) Pos() token.Pos { return rs.Token.Pos }

// MatchCase is a single `when targets do block` arm of a match statement.
type MatchCase struct {
	Token   token.Token  // the `when` keyword token
	Targets []Expression // expression ( "," expression )* — the values/ranges to match
	Body    *BlockStatement
}

// MatchStatement implements the `match` enumerable selector per
// spec/02-statements.md §3.3 and §5 EBNF:
//
//	match_stmt ::= "match" expression [ "all" | "one" ] ":" [ block ]
//	                ( "when" match_targets "do" block )+ [ "other" block ]
//	                "done" ;
//
// Mode is "one" (first match wins) or "all" (evaluate every matching arm);
// an omitted mode defaults to "one".
type MatchStatement struct {
	Token    token.Token
	Subject  Expression      // the value being matched against the `when` targets
	Mode     string          // "one", "all", or "one" when omitted
	Prologue *BlockStatement // optional declaration block between `:` and the first `when`
	Cases    []MatchCase
	Other    *BlockStatement // optional `other` default fallback block
}

func (ms *MatchStatement) statementNode() {}
func (ms *MatchStatement) Pos() token.Pos { return ms.Token.Pos }

// ScopeStatement implements the local-scope and qualifier-suppression blocks
// per spec/02-statements.md §3.1 and §5 EBNF:
//
//	scope_stmt ::= "start" [ label ] ":" [ block ] "do" block "done" [ label ]
//	             | "with" expression "do" block "done" ;
//
// For `start`, Label/Prologue capture the named scope opening; for `with`,
// Qualifier carries the suppressed module/object prefix expression.
type ScopeStatement struct {
	Token     token.Token
	Keyword   string          // "start" or "with"
	Label     *Identifier     // optional label on `start`
	Qualifier Expression      // the suppressed qualifier expression on `with`
	Prologue  *BlockStatement // optional `start` prologue block (run before `do`)
	Body      *BlockStatement // the `do` block
	DoneLabel *Identifier     // optional label on `done`
}

func (ss *ScopeStatement) statementNode() {}
func (ss *ScopeStatement) Pos() token.Pos { return ss.Token.Pos }

// CycleStatement implements the `cycle` repetition statement per
// spec/02-statements.md §3.4 and the Decision 14 (D14, 2026-09-13) EBNF:
//
//	cycle_stmt ::= "cycle" [ label ] [ ":" decl_block ]
//	               ( "do" | "while" expression "do"
//	               | "for" [ "∀" ] identifier ( "∈" | "in" ) expression "do" )
//	               block
//	               [ "then" block ]
//	               "done" [ label ] ";"
//	             | "for" [ "∀" ] identifier ( "∈" | "in" ) expression "do"
//	               block "done" ";" ;
//
// is partitioned into:
//   - Label: optional cycle label. Per D14 the label is independent of the
//     `:` scope marker; `cycle name do` is a labeled jump target only,
//     `cycle name:` opens a labeled prologue, `cycle:` opens an anonymous
//     prologue, and `cycle do` is fully anonymous with no outer scope.
//   - Prologue: optional stable outer scope. Per D14, the colon (`:`) — not
//     the label — is the sole scope marker. Created once before the first
//     iteration and shared across all iterations when present.
//   - Header: BodyHeader ∈ {"do", "while", "for"} together with Condition /
//     Index / Range domain — captures the body-header form.
//   - Body: the volatile block (re-created on every iteration).
//   - ThenBlock: optional `then` post-loop epilogue block (D14).
//   - DoneLabel: the optional closing label on `done`.
type CycleStatement struct {
	Token      token.Token
	Keyword    string // "cycle", "for", or "while"
	Label      *Identifier
	Prologue   *BlockStatement // stable outer scope (nil when no `:` marker — D14)
	BodyHeader string          // "do" | "while" | "for"
	Condition  Expression      // for `while expr do` (nil otherwise)
	Index      *Identifier     // for `for i ∈ expr do` (nil otherwise)
	Range      Expression      // for `for i ∈ expr do` (nil otherwise)
	IsForall   bool            // true if `for ∀ i ∈ …` (D11 quantifier)
	Body       *BlockStatement // volatile block
	ThenBlock  *BlockStatement // optional `then` post-loop epilogue (D14)
	DoneLabel  *Identifier     // optional label on `done` (D14)
}

func (cs *CycleStatement) statementNode() {}
func (cs *CycleStatement) Pos() token.Pos { return cs.Token.Pos }

type TrialStatement struct {
	Token   token.Token
	Keyword string
}

func (ts *TrialStatement) statementNode() {}
func (ts *TrialStatement) Pos() token.Pos { return ts.Token.Pos }

type TransferStatement struct {
	Token     token.Token
	Keyword   string
	Value     Expression
	Label     *Identifier // optional target label for `stop label;` / `repeat label;` / `redo label;` (D14)
	Condition Expression  // optional `if cond` clause on D14 jump statements
}

func (ts *TransferStatement) statementNode() {}
func (ts *TransferStatement) Pos() token.Pos { return ts.Token.Pos }

// SteppedRangeExpression implements a postfix step on a range expression
// per Decision 13 (D13, 2026-09-13). The grammar is
//
//	stepped_range ::= range_expr "(" step_expr ")"
//
// where `range_expr` is a binary expression over one of the four
// RANGE_INCL / RANGE_LEFT_INC / RANGE_RGHT_INC / RANGE_EXCL separators.
// The evaluator materialises the implicit 1-based sequence
// (start, start+step, start+2*step, …) up to but not crossing the endpoint,
// and treats `[i]` indexing as the i-th element (1-based, Decision 1).
type SteppedRangeExpression struct {
	Token token.Token
	Range Expression // a BinaryExpression whose operator is a RANGE_*
	Step  Expression // an IntegerLiteral (or broader Expr in future)
}

func (sre *SteppedRangeExpression) expressionNode() {}
func (sre *SteppedRangeExpression) Pos() token.Pos  { return sre.Token.Pos }

// CallExpression implements the rule invocation expression per
// spec/03-rules.md §3.1 and §6 EBNF:
//
//	rule_call ::= identifier "(" [ arg_list ] ")" [ "(" [ named_arg_list ] ")" ] ;
//
// The parser produces this node when an IDENT primary is immediately
// followed by `(`. The evaluator resolves Name against the rule registry
// and binds Args to the rule's declared parameters in a scoped call frame.
type CallExpression struct {
	Token token.Token
	Name  string
	Args  []Expression
}

func (ce *CallExpression) expressionNode() {}
func (ce *CallExpression) Pos() token.Pos  { return ce.Token.Pos }

// MemberExpression implements a dotted member-access path per spec/03-rules.md
// §5.4 (Closures & State Generators). It is produced in two surface forms:
//
//	set .count := [start];   -- leading-dot: a field on the enclosing rule's
//	                         -- own closure frame (Parts = [".count"], Base = nil)
//	print c.next();          -- object-dot: a method/field access on a bound
//	                         -- closure object (Base = c, Parts = ["next"])
//
// `Parts` is the dotted path with the leading dot folded in, so `.count`
// yields Parts=[".count"] and `c.next` yields Parts=["next"] with Base=c.
// `IsCall` distinguishes `c.next()` (invoke the closure method) from a bare
// `c.next` reference. Args holds the call argument list when IsCall is set.
//
// The evaluator resolves the leading-dot form against the *current* rule
// frame's boxed cells and the object-dot form against a heap-allocated
// closure object, keeping closure state out of the flat symbol table.
type MemberExpression struct {
	Token  token.Token
	Base   Expression   // receiver object for `obj.member`; nil for leading-dot `.member`
	Parts  []string     // dotted member path (leading `.member` folds the dot into Parts[0])
	IsCall bool         // true when followed by `(...)` — invoke as a closure method
	Args   []Expression // call arguments (only when IsCall)
}

func (me *MemberExpression) expressionNode() {}
func (me *MemberExpression) Pos() token.Pos  { return me.Token.Pos }
