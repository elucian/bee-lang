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

// Issue 14 — new AST nodes to back the rigid-syntax-gate dispatch table.
// Each block maps directly to a section in spec/02-statements.md §5 EBNF / §1.

type RuleStatement struct {
	Token token.Token
	Name  string
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

type MatchStatement struct {
	Token token.Token
}

func (ms *MatchStatement) statementNode() {}
func (ms *MatchStatement) Pos() token.Pos { return ms.Token.Pos }

type ScopeStatement struct {
	Token   token.Token
	Keyword string
}

func (ss *ScopeStatement) statementNode() {}
func (ss *ScopeStatement) Pos() token.Pos { return ss.Token.Pos }

type CycleStatement struct {
	Token   token.Token
	Keyword string
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
	Token   token.Token
	Keyword string
	Value   Expression
}

func (ts *TransferStatement) statementNode() {}
func (ts *TransferStatement) Pos() token.Pos { return ts.Token.Pos }
