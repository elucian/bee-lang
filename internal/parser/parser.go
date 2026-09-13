// internal/parser/parser.go
package parser

import (
	"bee/internal/lexer"
	"bee/internal/token"
	"fmt"
	"os"
	"strings"
)

func parseSuperscriptIntStatic(s string) int {
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

type Parser struct {
	l        *lexer.Lexer
	errors   []string
	warnings []string
	debug    bool
}

func (p *Parser) SetDebug(debug bool) {
	p.debug = debug
}

func (p *Parser) debugLog(format string, a ...interface{}) {
	if p.debug {
		fmt.Fprintf(os.Stderr, format, a...)
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) Warnings() []string {
	return p.warnings
}

func New(l *lexer.Lexer) *Parser {
	return &Parser{l: l}
}

// memberNameKeywords are lexer keywords that spec/03-rules.md §5.4 (Closures &
// State Generators) and §2.x repurposes as ordinary identifiers in signature
// and member-access position. The canonical closure example names a parameter
// `start`, a result/member `next`, and a state field `.count` — but `start`
// and `next` are also statement keywords (START / NEXT). Rather than touch
// the lexer (which would break `start:` scope blocks and `next` cycle
// transfers), the parser contextually accepts these tokens as identifiers
// wherever a name — not a statement — is expected.
var memberNameKeywords = map[token.Type]bool{
	token.START: true, // `start` — §5.4 param name / boxed-state element
	token.NEXT:  true, // `next`  — §5.4 result name / closure method name
}

// identLiteral returns the surface identifier text for a name token,
// tolerating the memberNameKeywords that the lexer promoted to keyword types.
// The token's Literal already carries the exact source text, so this is a
// pure pass-through; the map is consulted by callers to decide whether a
// non-IDENT token is admissible as a name in the first place.
func (p *Parser) identLiteral(tok token.Token) (string, bool) {
	if tok.Type == token.IDENT {
		return tok.Literal, true
	}
	if memberNameKeywords[tok.Type] {
		return tok.Literal, true
	}
	return tok.Literal, false
}

// parseMemberName folds a leading-dot member path into a single dotted name.
// In spec/03 §5.4 the form `set .count := [start];` declares a boxed state
// field on the enclosing rule's closure frame; the lexer emits the `.` and
// `count` as two tokens (DOT then IDENT). This helper consumes the leading
// DOT and returns the folded name `.count`. When the leading token is not a
// DOT it is returned unchanged (the common single-identifier path).
func (p *Parser) parseMemberName(first token.Token) string {
	if first.Type != token.DOT && first.Literal != "." {
		name, _ := p.identLiteral(first)
		return name
	}
	// Leading-dot member access: fold `.` + identifier into `.name`.
	memberTok := p.l.NextToken()
	member, _ := p.identLiteral(memberTok)
	return "." + member
}

// ParseProgram tokenizes the full input stream and dispatches every non-EOF
// token to parseStatement. Per fix for issue 14 (parser-silent-token-drop),
// any non-EOF token whose dispatch returns nil is recorded as E0009 on
// p.errors with the token's Type, Literal, and source line. cmd/bee/main.go's
// -c and -e paths distinguish errors (hard exit) from warnings (parser
// diagnostic surface only) so the build does not exit on W0901 stubs.
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	for {
		tok := p.l.NextToken()
		if tok.Type == token.EOF {
			break
		}

		stmt := p.parseStatement(tok)
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		} else {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:UnrecognizedStatement: line=%d type=%s literal=%q",
				tok.Pos, tok.Type, tok.Literal,
			))
		}
	}
	return program
}

// parseStatement dispatches top-level tokens to their respective parsers.
// The switch covers the full executive statement taxonomy enumerated in
// spec/02-statements.md §1 and §5 EBNF: declarations (NEW, SET),
// memory directives (ZAP), mutation operators (LET), contracts (ASSERT,
// EXPECT), input/output directives (PRINT, WRITE, READ), control flow (IF,
// MATCH, START, WITH, CYCLE, FOR, WHILE), transactional error handling
// (TRIAL, TRY, CASE, MISS, FINAL), transfers (RETURN, STOP, REDO, NEXT,
// PASS, RAISE, RESUME, RETRY, FAIL), and the canonical entry-block keyword
// RULE. Any token outside this dispatch is reported as nil and bubbles up
// to ParseProgram where it is recorded as E0009.
func (p *Parser) parseStatement(tok token.Token) Statement {
	switch tok.Type {
	case token.RULE:
		return p.parseRuleEntry(tok)
	case token.NEW, token.SET:
		return p.parseDeclaration(tok)
	case token.LET:
		return p.parseAssignment(tok)
	case token.ZAP:
		return p.parseZapStatement(tok)
	case token.ASSERT:
		return p.parseAssertStatement(tok)
	case token.EXPECT:
		return p.parseExpectStatement(tok)
	case token.PRINT:
		return p.parsePrintStatement(tok)
	case token.WRITE:
		return p.parseWriteStatement(tok)
	case token.READ:
		return p.parseReadStatement(tok)
	case token.IF:
		return p.parseIfStatement(tok)
	case token.MATCH:
		return p.parseMatchStatement(tok)
	case token.START, token.WITH:
		return p.parseScopeStatement(tok)
	case token.CYCLE, token.FOR, token.WHILE:
		return p.parseCycleStatement(tok)
	case token.TRIAL, token.TRY, token.CASE, token.MISS, token.FINAL:
		return p.parseTrialStatement(tok)
	case token.RETURN, token.STOP, token.REDO, token.REPEAT, token.NEXT, token.PASS,
		token.RAISE, token.RESUME, token.RETRY, token.FAIL, token.OVER,
		token.ABORT, token.EXIT, token.PANIC, token.YIELD:
		return p.parseTransferStatement(tok)
	}
	p.debugLog("PARSER DEBUG: unrecognized top-level token type=%s literal=%q line=%d\n",
		tok.Type, tok.Literal, tok.Pos)
	return nil
}

// parseRuleEntry handles the canonical entry-block form `rule name:` from
// spec/03-rules.md §1. The current implementation is a stub that consumes
// the rule identifier and the trailing colon, then records a soft W0901
// warning so the parser still surfaces later errors but does not falsely
// PASS. Full grammar mapping is pending Phase 7.2.
// parseRuleEntry handles the canonical entry-block form
// `rule identifier [ "(" [ param_list ] ")" ] [ "(" [ named_param_list ] ")" ]
//
//	[ "=>" "(" [ result_list ] ")" ] ":" [ contract_clause ] block "return" ";"`
//
// from spec/03-rules.md §2.1 and §6. The signature grammar for
// `param_list`, `named_param_list`, and `result_list` is consumed fast-forward
// (Decision 6 gate — full signature validation deferred to Phase 7.2) so
// that the block body is parsed and emitted as a BlockStatement attached
// to RuleStatement.Body. The block terminator `return;` is consumed in
// alignment with the rule header (0 relative indentation, §2.3).
//
// Forward declaration form (`rule identifier(...)...;` with no body and
// immediate `;`) is detected by a `;` lookahead immediately after the
// signature closing tokens; in that case RuleStatement.ForwardDecl is set
// and no body is consumed. See spec/03-rules.md §5.2.
func (p *Parser) parseRuleEntry(tok token.Token) Statement {
	stmt := &RuleStatement{Token: tok}
	// Rule identifier. spec/03 §5.4 permits a leading-dot member rule inside a
	// closure generator (`rule .next() => ...`), which the lexer emits as DOT
	// then the member name. parseMemberName folds the two into `.next`.
	stmt.Name = p.parseMemberName(p.l.NextToken())

	// Optional primary parameter list.
	if p.l.PeekToken().Type == token.LPAREN {
		p.l.NextToken() // consume '('
		if p.l.PeekToken().Type != token.RPAREN {
			for {
				if p.l.PeekToken().Type == token.RPAREN || p.l.PeekToken().Type == token.EOF {
					break
				}
				paramTok := p.l.NextToken() // bind param name
				// §5.4 names a parameter `start`, which the lexer promotes to the
				// START keyword; accept admissible name tokens (identLiteral).
				paramName, _ := p.identLiteral(paramTok)
				stmt.Params = append(stmt.Params, paramName)
				if p.l.PeekToken().Literal == ":" {
					p.l.NextToken()
					if p.l.PeekToken().Type != token.RPAREN {
						p.parseExpression()
					}
				}
				if p.l.PeekToken().Type == token.IN_OP || p.l.PeekToken().Type == token.IN_KEYWORD ||
					p.l.PeekToken().Literal == "∈" || p.l.PeekToken().Literal == "in" {
					p.l.NextToken()
					p.l.NextToken() // type specifier token
				}
				if p.l.PeekToken().Type == token.COMMA {
					p.l.NextToken()
					continue
				}
				break
			}
		}
		if p.l.PeekToken().Type == token.RPAREN {
			p.l.NextToken()
		}
	}

	// Optional Decision-6 named-parameter slot `(sep: ... ∈ Str, end: ... ∈ Str)`.
	if p.l.PeekToken().Type == token.LPAREN {
		// Heuristic: a second `(` directly following the primary `)` opens
		// the named slot. (We do not require an `∈` annotation to land
		// because spec/03 §2.4 allows defaults to be omitted.)
		p.l.NextToken() // consume '('
		if p.l.PeekToken().Type != token.RPAREN {
			for {
				if p.l.PeekToken().Type == token.RPAREN || p.l.PeekToken().Type == token.EOF {
					break
				}
				p.l.NextToken() // bind named param name
				if p.l.PeekToken().Literal == ":" {
					p.l.NextToken()
					if p.l.PeekToken().Type != token.RPAREN && p.l.PeekToken().Literal != "∈" &&
						p.l.PeekToken().Literal != "in" {
						p.parseExpression()
					}
				}
				if p.l.PeekToken().Type == token.IN_OP || p.l.PeekToken().Type == token.IN_KEYWORD ||
					p.l.PeekToken().Literal == "∈" || p.l.PeekToken().Literal == "in" {
					p.l.NextToken()
					p.l.NextToken() // type specifier token
				}
				if p.l.PeekToken().Type == token.COMMA {
					p.l.NextToken()
					continue
				}
				break
			}
		}
		if p.l.PeekToken().Type == token.RPAREN {
			p.l.NextToken()
		}
	}

	// Optional result list `=> (res_list)`.
	if p.l.PeekToken().Type == token.FAT_ARROW || p.l.PeekToken().Literal == "=>" {
		p.l.NextToken()
		if p.l.PeekToken().Type == token.LPAREN {
			p.l.NextToken()
			if p.l.PeekToken().Type != token.RPAREN {
				for {
					if p.l.PeekToken().Type == token.RPAREN || p.l.PeekToken().Type == token.EOF {
						break
					}
					resTok := p.l.NextToken()
					// §5.4 names a result `next`, which the lexer promotes to NEXT.
					resName, _ := p.identLiteral(resTok)
					stmt.Results = append(stmt.Results, resName)
					if p.l.PeekToken().Literal == ":" {
						p.l.NextToken()
						if p.l.PeekToken().Type != token.RPAREN && p.l.PeekToken().Literal != "∈" &&
							p.l.PeekToken().Literal != "in" {
							p.parseExpression()
						}
					}
					if p.l.PeekToken().Type == token.IN_OP || p.l.PeekToken().Type == token.IN_KEYWORD ||
						p.l.PeekToken().Literal == "∈" || p.l.PeekToken().Literal == "in" {
						p.l.NextToken()
						p.l.NextToken()
					}
					if p.l.PeekToken().Type == token.COMMA {
						p.l.NextToken()
						continue
					}
					break
				}
			}
			if p.l.PeekToken().Type == token.RPAREN {
				p.l.NextToken()
			}
		}
	}

	// Forward declaration — signature concludes with ';'.
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken() // consume ';'
		stmt.ForwardDecl = true
		p.warnings = append(p.warnings, fmt.Sprintf(
			"W0901 SignatureGrammarPartial: line=%d rule=%q (signature decoder is heuristic; full type binding deferred to Phase 7.2)",
			tok.Pos, stmt.Name,
		))
		return stmt
	}

	// Header colon.
	if p.l.PeekToken().Type == token.COLON {
		p.l.NextToken()
	}

	// Optional contract clauses (assert/expect). Bind them into the body
	// prefix so the evaluator enforces them at the canonical §2.4 order.
	body := &BlockStatement{Token: tok}
	for {
		peek := p.l.PeekToken()
		if peek.Type == token.EOF {
			break
		}
		// Stop only at the rule's own aligned `return;` terminator (0 relative
		// indentation, spec/03 §2.3 / spec/02 §6.3). A nested rule body's
		// internal `return;` is consumed by the recursive parseRuleEntry call,
		// so when control returns here the stream is parked just past it. A
		// RETURN immediately followed by `;` is therefore this (enclosing)
		// rule's terminator, not a nested body's — break and let the trailing
		// `return;` consumer below take it.
		if peek.Type == token.RETURN {
			break
		}
		// A stray `;` can only belong to this rule's terminator.
		if peek.Type == token.SEMICOLON {
			break
		}
		nextTok := p.l.NextToken()
		if nextTok.Type == token.EOF {
			break
		}
		sub := p.parseStatement(nextTok)
		if sub != nil {
			body.Statements = append(body.Statements, sub)
		}
	}
	stmt.Body = body

	// Consume aligned `return;` terminator (spec/03 §2.3 / spec/02 §6.3).
	if p.l.PeekToken().Type == token.RETURN || p.l.PeekToken().Literal == "return" {
		retTok := p.l.NextToken()
		body.Statements = append(body.Statements,
			&TransferStatement{Token: retTok, Keyword: "return"})
		if p.l.PeekToken().Type == token.SEMICOLON {
			p.l.NextToken()
		}
	}

	p.warnings = append(p.warnings, fmt.Sprintf(
		"W0901 SignatureGrammarPartial: line=%d rule=%q (signature decoder is heuristic; full type binding deferred to Phase 7.2)",
		tok.Pos, stmt.Name,
	))
	return stmt
}

// parseZapStatement handles `zap identifier;` per spec/02-statements.md §2.3.
func (p *Parser) parseZapStatement(tok token.Token) Statement {
	stmt := &ZapStatement{Token: tok}
	identTok := p.l.NextToken()
	stmt.Name = identTok.Literal
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
	}
	return stmt
}

// parseWriteStatement handles `write expression;` per spec/02-statements.md §5.
func (p *Parser) parseWriteStatement(tok token.Token) Statement {
	stmt := &WriteStatement{Token: tok}
	stmt.Value = p.parseExpression()
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
	}
	return stmt
}

// parseReadStatement handles `read identifier;` per spec/02-statements.md §5.
func (p *Parser) parseReadStatement(tok token.Token) Statement {
	stmt := &ReadStatement{Token: tok}
	identTok := p.l.NextToken()
	stmt.Target = &Identifier{Token: identTok, Value: identTok.Literal}
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
	}
	return stmt
}

// parseMatchStatement implements spec/02-statements.md §3.3 / §5 EBNF:
//
//	match_stmt ::= "match" expression [ "all" | "one" ] ":" [ block ]
//	                ( "when" match_targets "do" block )+ [ "other" block ]
//	                "done" ;
//
// `one` / `all` are not registered keywords, so they arrive as IDENT tokens
// and are matched by literal. An omitted mode defaults to "one" (first match).
func (p *Parser) parseMatchStatement(tok token.Token) Statement {
	stmt := &MatchStatement{Token: tok, Mode: "one"}

	// Subject expression (`match <expr> ...`).
	stmt.Subject = p.parseExpression()

	// Optional mode selector: `one` | `all`.
	if peek := p.l.PeekToken(); peek.Type == token.IDENT && (peek.Literal == "one" || peek.Literal == "all") {
		stmt.Mode = peek.Literal
		p.l.NextToken()
	}

	// Mandatory colon after the subject/mode.
	if peek := p.l.PeekToken(); peek.Type == token.COLON {
		p.l.NextToken()
	} else {
		p.errors = append(p.errors, fmt.Sprintf(
			"E0009 SyntaxError:InvalidMatchHeader: line=%d (expected `:` after match subject; got type=%s literal=%q)",
			tok.Pos, peek.Type, peek.Literal,
		))
	}

	// Optional prologue block (declarations) before the first `when`.
	stmt.Prologue = p.parseMatchBlock(tok)

	// `when` clauses.
	for {
		peek := p.l.PeekToken()
		if peek.Type == token.WHEN || (peek.Type == token.IDENT && peek.Literal == "when") {
			whenTok := p.l.NextToken()
			stmt.Cases = append(stmt.Cases, p.parseMatchCase(whenTok))
			continue
		}
		break
	}

	// `other` default fallback branch.
	if peek := p.l.PeekToken(); peek.Type == token.OTHER || (peek.Type == token.IDENT && peek.Literal == "other") {
		p.l.NextToken()
		stmt.Other = p.parseMatchBlock(tok)
	}

	// `done` terminator (D14) with optional trailing `;`.
	if peek := p.l.PeekToken(); peek.Type == token.DONE || (peek.Type == token.IDENT && peek.Literal == "done") {
		p.l.NextToken()
	}
	if peek := p.l.PeekToken(); peek.Type == token.SEMICOLON {
		p.l.NextToken()
	}

	return stmt
}

// parseMatchCase parses a single `when targets do block` arm.
func (p *Parser) parseMatchCase(whenTok token.Token) MatchCase {
	c := MatchCase{Token: whenTok}
	// match_targets ::= expression ( "," expression )*  — terminated by `do`.
	for {
		c.Targets = append(c.Targets, p.parseExpression())
		if p.l.PeekToken().Type == token.COMMA {
			p.l.NextToken() // consume `,`
			continue
		}
		break
	}
	p.matchDo()
	c.Body = p.parseMatchBlock(whenTok)
	return c
}

// parseMatchBlock collects statements until a `when`, `other`, or `done`
// boundary (or EOF). It is used for both the optional prologue and the arm
// bodies, since all three share the same stop-set.
func (p *Parser) parseMatchBlock(tok token.Token) *BlockStatement {
	block := &BlockStatement{Token: tok}
	for {
		peek := p.l.PeekToken()
		if peek.Type == token.EOF {
			break
		}
		if peek.Type == token.WHEN || peek.Type == token.OTHER || peek.Type == token.DONE ||
			peek.Literal == "when" || peek.Literal == "other" || peek.Literal == "done" {
			break
		}
		if peek.Type == token.SEMICOLON {
			p.l.NextToken()
			continue
		}
		nextTok := p.l.NextToken()
		if nextTok.Type == token.EOF {
			break
		}
		sub := p.parseStatement(nextTok)
		if sub != nil {
			block.Statements = append(block.Statements, sub)
		}
	}
	return block
}

// parseScopeStatement implements spec/02-statements.md §3.1 / §5 EBNF:
//
//	scope_stmt ::= "start" [ label ] ":" [ block ] "do" block "done" [ label ]
//	             | "with" expression "do" block "done" ;
func (p *Parser) parseScopeStatement(tok token.Token) Statement {
	stmt := &ScopeStatement{Token: tok, Keyword: tok.Literal}

	if tok.Type == token.WITH {
		// `with expression do block done;`
		stmt.Qualifier = p.parseExpression()
		p.matchDo()
		stmt.Body = p.parseScopeBlock(tok)
		// done
		if peek := p.l.PeekToken(); peek.Type == token.DONE || peek.Literal == "done" {
			p.l.NextToken()
		}
		if peek := p.l.PeekToken(); peek.Type == token.SEMICOLON {
			p.l.NextToken()
		}
		return stmt
	}

	// `start [label] : [block] do block done [label];`
	peek := p.l.PeekToken()
	if peek.Type == token.IDENT && peek.Literal != "do" && peek.Literal != "done" && peek.Literal != "with" {
		labelTok := p.l.NextToken()
		stmt.Label = &Identifier{Token: labelTok, Value: labelTok.Literal}
	}
	// Mandatory colon.
	if peek := p.l.PeekToken(); peek.Type == token.COLON {
		p.l.NextToken()
	} else {
		p.errors = append(p.errors, fmt.Sprintf(
			"E0009 SyntaxError:InvalidScopeHeader: line=%d (expected `:` after `start` label/header; got type=%s literal=%q)",
			tok.Pos, peek.Type, peek.Literal,
		))
	}
	// Optional prologue block (run before the `do` body).
	stmt.Prologue = p.parseScopePrologue(tok)
	// `do` body header.
	p.matchDo()
	stmt.Body = p.parseScopeBlock(tok)
	// `done [label];`
	if peek := p.l.PeekToken(); peek.Type == token.DONE || peek.Literal == "done" {
		p.l.NextToken()
		if p.l.PeekToken().Type == token.IDENT {
			labelTok := p.l.NextToken()
			stmt.DoneLabel = &Identifier{Token: labelTok, Value: labelTok.Literal}
		}
	}
	if peek := p.l.PeekToken(); peek.Type == token.SEMICOLON {
		p.l.NextToken()
	}
	return stmt
}

// parseScopePrologue collects `start` declarations before the `do` header.
func (p *Parser) parseScopePrologue(tok token.Token) *BlockStatement {
	block := &BlockStatement{Token: tok}
	for {
		peek := p.l.PeekToken()
		if peek.Type == token.EOF {
			break
		}
		if peek.Type == token.DO || peek.Literal == "do" {
			break
		}
		if peek.Type == token.SEMICOLON {
			p.l.NextToken()
			continue
		}
		nextTok := p.l.NextToken()
		if nextTok.Type == token.EOF {
			break
		}
		sub := p.parseStatement(nextTok)
		if sub != nil {
			block.Statements = append(block.Statements, sub)
		}
	}
	return block
}

// parseScopeBlock collects the `do` body until the `done` terminator.
func (p *Parser) parseScopeBlock(tok token.Token) *BlockStatement {
	block := &BlockStatement{Token: tok}
	for {
		peek := p.l.PeekToken()
		if peek.Type == token.EOF {
			break
		}
		if peek.Type == token.DONE || peek.Literal == "done" {
			break
		}
		if peek.Type == token.SEMICOLON {
			p.l.NextToken()
			continue
		}
		nextTok := p.l.NextToken()
		if nextTok.Type == token.EOF {
			break
		}
		sub := p.parseStatement(nextTok)
		if sub != nil {
			block.Statements = append(block.Statements, sub)
		}
	}
	return block
}

// parseCycleStatement implements spec/02-statements.md §3.4 cycle_stmt grammar
// per Decision 14 (D14, ratified 2026-09-13). The D14 grammar decouples the
// optional label from the colon — they are independent options:
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
// The colon (`:`) is the **scope marker**: when present (with or without a
// label) the cycle opens a stable outer-scope prologue that runs exactly
// once before the first iteration and is shared across all iterations. A
// label without a colon (`cycle name do …`) is purely a jump target for
// `stop/repeat/redo`; it creates no scope. The block terminator is
// uniformly `done [label];` for every cycle form. Inline `repeat` /
// `redo` / `stop` statements appearing INSIDE the volatile body are
// dispatched as TransferStatements (continue/break semantics).
func (p *Parser) parseCycleStatement(tok token.Token) Statement {
	stmt := &CycleStatement{Token: tok, Keyword: tok.Literal}

	// 1. Optional label + optional colon — D14 makes them INDEPENDENT.
	//    Only the entry keyword `cycle` accepts a label; bare `for` and
	//    `while` cycles are unlabeled by construction. The dispatch is:
	//
	//      peek IDENT               → consume label
	//        peek COLON             → consume colon → parse prologue (Label set)
	//        peek do|while|for      → label as jump target only (no prologue)
	//      peek COLON              → bare `cycle:` → anonymous prologue
	//      peek do|while|for       → fully anonymous (no label, no prologue)
	if tok.Type == token.CYCLE {
		peek := p.l.PeekToken()
		if peek.Type == token.IDENT &&
			peek.Literal != "do" && peek.Literal != "while" && peek.Literal != "for" &&
			peek.Literal != "then" && peek.Literal != "done" {
			// Optional label.
			labelTok := p.l.NextToken() // consume label ident
			stmt.Label = &Identifier{Token: labelTok, Value: labelTok.Literal}
			// Optional colon — if present, opens the prologue.
			if p.l.PeekToken().Type == token.COLON {
				p.l.NextToken() // consume colon
				stmt.Prologue = p.parseCyclePrologue(tok)
			}
			// No colon: label is a pure jump target (no prologue).
		} else if peek.Type == token.COLON {
			// Bare `cycle:` (anonymous prologue, no label).
			p.l.NextToken() // consume colon
			stmt.Prologue = p.parseCyclePrologue(tok)
		}
		// Else: fully anonymous (no label, no colon, no prologue).
	}

	// 2. Body-header dispatch. Two entry shapes per spec §3.4 / EBNF:
	//
	//    - When the entry keyword is `cycle`, the body-header
	//      (`do` | `while expr do` | `for ∀? ident ∈ expr do`) follows
	//      the optional label and prologue. We dispatch on the NEXT
	//      token (peek) to choose the form.
	//
	//    - When the entry keyword is `for` or `while`, the keyword
	//      was already consumed by parseStatement, so the body-header
	//      shape is fixed by `tok.Type`. We dispatch on `tok` directly
	//      and parse the header tokens that immediately follow.
	switch {
	case tok.Type == token.FOR:
		// Entry was `for [∀] ident ∈ expr do` — kw already consumed.
		stmt.BodyHeader = "for"
		// Optional ∀ quantifier (D11).
		if p.l.PeekToken().Type == token.FORALL {
			p.l.NextToken()
			stmt.IsForall = true
		}
		// Index loop variable — required identifier.
		indexTok := p.l.NextToken()
		if indexTok.Type != token.IDENT {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:InvalidCycleHeader: line=%d (expected index identifier after `for` in cycle header; got type=%s literal=%q)",
				indexTok.Pos, indexTok.Type, indexTok.Literal,
			))
		} else {
			stmt.Index = &Identifier{Token: indexTok, Value: indexTok.Literal}
		}
		// ∈ / in domain operator.
		peekOp := p.l.PeekToken()
		if peekOp.Type != token.IN_OP && peekOp.Type != token.IN_KEYWORD {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:InvalidCycleHeader: line=%d (expected `∈` or `in` after for-index identifier; got type=%s literal=%q)",
				peekOp.Pos, peekOp.Type, peekOp.Literal,
			))
		} else {
			p.l.NextToken() // consume ∈ / in
		}
		stmt.Range = p.parseExpression()
		if !p.matchDo() {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:InvalidCycleHeader: line=%d (expected `do` after `for` domain expression in cycle header)",
				tok.Pos,
			))
		}
	case tok.Type == token.WHILE:
		// Entry was `while expr do` — kw already consumed.
		stmt.BodyHeader = "while"
		stmt.Condition = p.parseExpression()
		if !p.matchDo() {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:InvalidCycleHeader: line=%d (expected `do` after `while` condition in cycle header)",
				tok.Pos,
			))
		}
	default:
		// Entry was `cycle` — body-header keyword follows the (optional)
		// label and prologue. Dispatch on the next token's type:
		//   WHILE → `while expr do`
		//   FOR   → `for [∀] ident ∈ expr do`
		//   DO    → `do` (infinite cycle)
		//   IDENT matching `do`/`while`/`for` → same as their tokenised forms
		switch p.l.PeekToken().Type {
		case token.WHILE:
			p.l.NextToken() // consume `while`
			stmt.BodyHeader = "while"
			stmt.Condition = p.parseExpression()
			if !p.matchDo() {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:InvalidCycleHeader: line=%d (expected `do` after `while` condition in cycle header)",
					tok.Pos,
				))
			}
		case token.FOR:
			p.l.NextToken() // consume `for`
			stmt.BodyHeader = "for"
			// Optional ∀ quantifier (D11).
			if p.l.PeekToken().Type == token.FORALL {
				p.l.NextToken()
				stmt.IsForall = true
			}
			// Index loop variable — required identifier.
			indexTok := p.l.NextToken()
			if indexTok.Type != token.IDENT {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:InvalidCycleHeader: line=%d (expected index identifier after `for` in cycle header; got type=%s literal=%q)",
					indexTok.Pos, indexTok.Type, indexTok.Literal,
				))
			} else {
				stmt.Index = &Identifier{Token: indexTok, Value: indexTok.Literal}
			}
			// ∈ / in domain operator.
			peekOp := p.l.PeekToken()
			if peekOp.Type != token.IN_OP && peekOp.Type != token.IN_KEYWORD {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:InvalidCycleHeader: line=%d (expected `∈` or `in` after for-index identifier; got type=%s literal=%q)",
					peekOp.Pos, peekOp.Type, peekOp.Literal,
				))
			} else {
				p.l.NextToken() // consume ∈ / in
			}
			stmt.Range = p.parseExpression()
			if !p.matchDo() {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:InvalidCycleHeader: line=%d (expected `do` after `for` domain expression in cycle header)",
					tok.Pos,
				))
			}
		default:
			// `do` is the infinite-cycle body-header. Required for `cycle`.
			if !p.matchDo() {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:InvalidCycleHeader: line=%d (expected `do`, `while`, or `for` as cycle body-header)",
					tok.Pos,
				))
				stmt.BodyHeader = "do"
			} else {
				stmt.BodyHeader = "do"
			}
		}
	}

	// 3. Volatile body block — terminated by `repeat` (cycle terminator,
	//    spec §6.3) or EOF.
	stmt.Body = p.parseCycleBody(tok)

	// 4. Optional `then` clause (post-loop epilogue; runs once after loop
	//    exit per D14).
	if p.l.PeekToken().Literal == "then" {
		p.l.NextToken() // consume `then`
		stmt.ThenBlock = p.parseCycleBody(tok)
	}

	// 5. Required `done` terminator with optional `[label]` followed by
	//    `;` (D14 unifies the cycle terminator).
	if p.l.PeekToken().Type != token.DONE {
		p.errors = append(p.errors, fmt.Sprintf(
			"E0203 UnterminatedBlock:MissingDone: line=%d (cycle missing `done` terminator)",
			tok.Pos,
		))
		return stmt
	}
	p.l.NextToken() // consume `done`

	// Optional done label.
	if p.l.PeekToken().Type == token.IDENT &&
		p.l.PeekToken().Literal != ";" {
		labelTok := p.l.NextToken()
		stmt.DoneLabel = &Identifier{Token: labelTok, Value: labelTok.Literal}
		// Spec §3.4 / D14 / E0205: closing label must match the opening
		// label when present.
		if stmt.Label != nil && stmt.Label.Value != stmt.DoneLabel.Value {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0205 LabelMismatch: line=%d (cycle `done %s` does not match opening label `%s`)",
				labelTok.Pos, stmt.DoneLabel.Value, stmt.Label.Value,
			))
		}
	}

	// Consume trailing semicolon (cycle terminator requires `;`).
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
	}

	return stmt
}

// matchDo consumes an optional `do` keyword — either the keyword token
// (token.DO) or an identifier with the literal "do". Returns true if a
// `do` was consumed.
func (p *Parser) matchDo() bool {
	peek := p.l.PeekToken()
	if peek.Type == token.DO || (peek.Type == token.IDENT && peek.Literal == "do") {
		p.l.NextToken()
		return true
	}
	return false
}

// parseCyclePrologue collects the optional stable-outer-scope declarations
// following `cycle name:` and before the body-header. Stops at `do` /
// `while` / `for` (body-header) or EOF.
func (p *Parser) parseCyclePrologue(tok token.Token) *BlockStatement {
	prologue := &BlockStatement{Token: tok}
	for {
		peek := p.l.PeekToken()
		if peek.Type == token.EOF {
			break
		}
		if peek.Type == token.DO ||
			peek.Literal == "do" ||
			peek.Type == token.WHILE || peek.Literal == "while" ||
			peek.Type == token.FOR || peek.Literal == "for" {
			break
		}
		if peek.Type == token.SEMICOLON {
			// Stray semicolons in prologue are tolerated.
			p.l.NextToken()
			continue
		}
		nextTok := p.l.NextToken()
		if nextTok.Type == token.EOF {
			break
		}
		sub := p.parseStatement(nextTok)
		if sub != nil {
			prologue.Statements = append(prologue.Statements, sub)
		}
	}
	return prologue
}

// parseCycleBody collects the volatile body block. Stops at `then`
// (post-loop epilogue), `done` (cycle terminator), or EOF. Inline
// `repeat`/`redo`/`stop` statements are dispatched as TransferStatements
// and remain inside the body (D14, Decision 14, 2026-09-13).
func (p *Parser) parseCycleBody(tok token.Token) *BlockStatement {
	body := &BlockStatement{Token: tok}
	for {
		peek := p.l.PeekToken()
		if peek.Type == token.EOF {
			break
		}
		if peek.Literal == "then" || peek.Type == token.DONE {
			break
		}
		if peek.Type == token.SEMICOLON {
			// Stray semicolons in body are tolerated.
			p.l.NextToken()
			continue
		}
		nextTok := p.l.NextToken()
		if nextTok.Type == token.EOF {
			break
		}
		sub := p.parseStatement(nextTok)
		if sub != nil {
			body.Statements = append(body.Statements, sub)
		}
	}
	return body
}

// parseTrialStatement handles `trial`, `try`, `case`, `miss`, `final` per
// spec/02-statements.md §4 Transactional Error Handling.
func (p *Parser) parseTrialStatement(tok token.Token) Statement {
	stmt := &TrialStatement{Token: tok, Keyword: tok.Literal}
	for {
		peek := p.l.PeekToken()
		if peek.Type == token.SEMICOLON {
			p.l.NextToken()
			break
		}
		if peek.Type == token.EOF || peek.Type == token.DONE {
			break
		}
		p.l.NextToken()
	}
	return stmt
}

// parseTransferStatement handles `return`, `stop`, `redo`, `repeat` (D14),
// `next`, `pass`, `raise`, `resume`, `retry`, `fail` per spec/02-statements.md §5
// transfer_stmt. D14 (2026-09-13) extends the grammar to:
//
//	jump_stmt ::= ( "repeat" | "stop" | "redo" ) [ label ] [ "if" condition ] ";" ;
//
// so inline jump statements carry an optional target label and an optional
// `if <cond>` guard.
func (p *Parser) parseTransferStatement(tok token.Token) Statement {
	stmt := &TransferStatement{Token: tok, Keyword: tok.Literal}
	// Optional expression payload (e.g., `raise <expr>`, `fail <expr>`).
	if tok.Type == token.RAISE || tok.Type == token.FAIL || tok.Type == token.RETURN {
		if p.l.PeekToken().Type != token.SEMICOLON && p.l.PeekToken().Type != token.EOF {
			stmt.Value = p.parseExpression()
		}
	}
	// D14: jump-style transfers accept an optional target label and an
	// optional `if <cond>` guard. D15 (2026-09-13): `next` is the canonical
	// loop-jump keyword; legacy `repeat` lexes to NEXT with an E0010
	// deprecation warning, so this dispatch sees both spellings as NEXT. The
	// token.REPEAT case is retained defensively for directly constructed
	// legacy tokens.
	if tok.Type == token.STOP || tok.Type == token.REDO ||
		tok.Type == token.REPEAT || tok.Type == token.NEXT {
		// Optional label (skip reserved body-header keywords so they remain
		// grammatically distinct in case the parser is misdispatched).
		if peek := p.l.PeekToken(); peek.Type == token.IDENT &&
			peek.Literal != "if" && peek.Literal != "do" &&
			peek.Literal != "while" && peek.Literal != "for" {
			labTok := p.l.NextToken()
			stmt.Label = &Identifier{Token: labTok, Value: labTok.Literal}
		}
		// Optional `if <cond>` guard.
		if p.l.PeekToken().Type == token.IF ||
			(p.l.PeekToken().Type == token.IDENT && p.l.PeekToken().Literal == "if") {
			p.l.NextToken() // consume `if`
			stmt.Condition = p.parseExpression()
		}
	}
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
	}
	return stmt
}

func (p *Parser) parseDeclaration(tok token.Token) Statement {
	ds := &DeclarationStatement{Token: tok}

	// Decision 11 (2026-09-13): parallel parenthesised colon-initialisation.
	// When `new` is followed by `(`, enter the `(ident_list) : (expr_list) ∈ T;`
	// branch. Arity N=M is enforced statically; missing trailing type or
	// bare comma-only form is E0009.
	if p.l.PeekToken().Type == token.LPAREN {
		openParen := p.l.NextToken() // consume '('
		// Parse ident_list
		firstIdent := p.l.NextToken()
		if firstIdent.Type != token.IDENT {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:UnrecognizedStatement: line=%d type=%s literal=%q (expected identifier inside '(' in D11 parallel declaration)",
				firstIdent.Pos, firstIdent.Type, firstIdent.Literal,
			))
			return ds
		}
		ds.Name = firstIdent.Literal
		ds.Names = append(ds.Names, firstIdent.Literal)
		for p.l.PeekToken().Type == token.COMMA {
			p.l.NextToken() // consume ','
			nextIdent := p.l.NextToken()
			if nextIdent.Type != token.IDENT {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:UnrecognizedStatement: line=%d type=%s literal=%q (expected identifier after ',' in D11 parallel ident list)",
					nextIdent.Pos, nextIdent.Type, nextIdent.Literal,
				))
				return ds
			}
			ds.Names = append(ds.Names, nextIdent.Literal)
		}
		closeParen := p.l.NextToken() // expect ')'
		if closeParen.Type != token.RPAREN {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:UnrecognizedStatement: line=%d type=%s literal=%q (expected ')' to close D11 ident list)",
				closeParen.Pos, closeParen.Type, closeParen.Literal,
			))
			return ds
		}
		colonTok := p.l.NextToken() // expect ':'
		if colonTok.Literal != ":" {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:UnrecognizedStatement: line=%d type=%s literal=%q (expected ':' after ')' in D11 parallel declaration)",
				colonTok.Pos, colonTok.Type, colonTok.Literal,
			))
			return ds
		}
		openValParen := p.l.NextToken() // expect '(' for value list
		if openValParen.Type != token.LPAREN {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:UnrecognizedStatement: line=%d type=%s literal=%q (expected '(' to open D11 expr list)",
				openValParen.Pos, openValParen.Type, openValParen.Literal,
			))
			return ds
		}
		firstVal := p.parseExpression()
		ds.Value = firstVal
		ds.Values = append(ds.Values, firstVal)
		for p.l.PeekToken().Type == token.COMMA {
			p.l.NextToken() // consume ','
			ds.Values = append(ds.Values, p.parseExpression())
		}
		closeValParen := p.l.NextToken() // expect ')'
		if closeValParen.Type != token.RPAREN {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:UnrecognizedStatement: line=%d type=%s literal=%q (expected ')' to close D11 expr list)",
				closeValParen.Pos, closeValParen.Type, closeValParen.Literal,
			))
			return ds
		}
		// Static arity check N = M (Decision 11).
		if len(ds.Names) != len(ds.Values) {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:ArityMismatch: line=%d (D11 parallel declaration: ident count N=%d does not match expr count M=%d)",
				openParen.Pos, len(ds.Names), len(ds.Values),
			))
			return ds
		}
		// Trailing `∈` or `in` is mandatory (Decision 11 rule).
		typeTok := p.l.NextToken()
		if !(typeTok.Type == token.IN_OP || typeTok.Type == token.IN_KEYWORD || typeTok.Literal == "∈" || typeTok.Literal == "in") {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:MissingTypeAnnotation: line=%d type=%s literal=%q (D11 parallel form requires trailing '∈ Type'; use ':=' for inference)",
				typeTok.Pos, typeTok.Type, typeTok.Literal,
			))
			return ds
		}
		// Consume the type identifier (full type binding deferred to Phase 7.2).
		_ = p.l.NextToken()
		if p.l.PeekToken().Type == token.SEMICOLON {
			p.l.NextToken()
		}
		return ds
	}

	identTok := p.l.NextToken() // first ident
	ds.Name = p.parseMemberName(identTok)
	ds.Names = append(ds.Names, ds.Name)

	// Check for comma-separated identifier list (e.g. new a, b, c ∈ Z;)
	for p.l.PeekToken().Type == token.COMMA {
		p.l.NextToken() // consume ','
		nextIdent := p.l.NextToken()
		ds.Names = append(ds.Names, p.parseMemberName(nextIdent))
	}

	nextTok := p.l.NextToken() // ∈, in, :=, or colon
	if nextTok.Literal == ":=" {
		val := p.parseExpression()
		ds.Value = val
		ds.Values = append(ds.Values, val)
		for p.l.PeekToken().Type == token.COMMA {
			p.l.NextToken()
			ds.Values = append(ds.Values, p.parseExpression())
		}
	} else if nextTok.Type == token.IN_OP || nextTok.Type == token.IN_KEYWORD || nextTok.Literal == "∈" || nextTok.Literal == "in" {
		// e.g. new a, b, c ∈ Z [:= ...]
		_ = p.l.NextToken() // consume type (e.g. Z)
		if p.l.PeekToken().Literal == ":=" {
			p.l.NextToken() // consume :=
			val := p.parseExpression()
			ds.Value = val
			ds.Values = append(ds.Values, val)
			for p.l.PeekToken().Type == token.COMMA {
				p.l.NextToken()
				ds.Values = append(ds.Values, p.parseExpression())
			}
		} else if p.l.PeekToken().Type == token.EQ {
			// Explicit Type Initialization (spec/02 §2.1): `new xo, yo, zo ∈ Z = 10;`
			// broadcasts the single value to every declared identifier.
			p.l.NextToken() // consume '='
			val := p.parseExpression()
			ds.Value = val
			for range ds.Names {
				ds.Values = append(ds.Values, val)
			}
		}
	} else if nextTok.Literal == ":" {
		// D9 (ratified 2026-09-13): `:` is structural pair-up with NO type
		// inference. Grammar: `new a: 1, b: 2 ∈ Z;` — N parallel bindings
		// with a REQUIRED trailing `∈ Type`. Bare `new a: 1;` is E0009;
		// the user must write `new a := 1;` for type inference.
		//
		// Split rule: `∈` is an infix operator (precLogic), so for the final
		// pair parseExpression returns `BinaryExpression{Left: value, Op: ∈,
		// Right: Type}`. Split that node inline: Left becomes the pair value,
		// Right is the type annotation (discarded; binding deferred to Phase
		// 7.2). The literal check excludes `!∈` (NOT_IN), a genuine operator.
		typeSeen := false
		splitPair := func(expr Expression) Expression {
			if be, ok := expr.(*BinaryExpression); ok &&
				(be.Token.Literal == "∈" || be.Token.Literal == "in") {
				typeSeen = true
				return be.Left
			}
			return expr
		}
		val := splitPair(p.parseExpression())
		ds.Value = val
		ds.Values = append(ds.Values, val)
		for p.l.PeekToken().Type == token.COMMA {
			p.l.NextToken() // consume ','
			pairIdent := p.l.NextToken()
			if pairIdent.Type != token.IDENT {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:UnrecognizedStatement: line=%d type=%s literal=%q (expected identifier after ',' in ':' pair-up declaration)",
					pairIdent.Pos, pairIdent.Type, pairIdent.Literal,
				))
				return ds
			}
			ds.Names = append(ds.Names, pairIdent.Literal)
			pairColon := p.l.NextToken()
			if pairColon.Literal != ":" {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:UnrecognizedStatement: line=%d (expected ':' after identifier %q in pair-up declaration)",
					pairIdent.Pos, pairIdent.Literal,
				))
				return ds
			}
			ds.Values = append(ds.Values, splitPair(p.parseExpression()))
		}
		if !typeSeen {
			typeTok := p.l.NextToken()
			if typeTok.Type == token.IN_OP || typeTok.Type == token.IN_KEYWORD || typeTok.Literal == "∈" || typeTok.Literal == "in" {
				_ = p.l.NextToken() // consume type name (full type binding deferred to Phase 7.2)
			} else {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:MissingTypeAnnotation: line=%d type=%s literal=%q (':' pair-up requires explicit '∈ Type'; use ':=' for type inference)",
					typeTok.Pos, typeTok.Type, typeTok.Literal,
				))
				return ds
			}
		}
	}

	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken() // Skip ;
	}
	return ds
}

func (p *Parser) parseAssignment(tok token.Token) Statement {
	stmt := &AssignmentStatement{Token: tok}
	tokIdent := p.l.NextToken()
	p.debugLog("PARSER DEBUG: ident tok=%q\n", tokIdent.Literal)
	name := p.parseMemberName(tokIdent)
	stmt.Names = append(stmt.Names, &Identifier{Token: tokIdent, Value: name})

	// Check for more comma-separated variables
	for p.l.PeekToken().Type == token.COMMA {
		p.l.NextToken() // comma
		tokIdent2 := p.l.NextToken()
		stmt.Names = append(stmt.Names, &Identifier{Token: tokIdent2, Value: tokIdent2.Literal})
	}

	// Look ahead to check if next token is an assignment operator
	opTok := p.l.NextToken()
	p.debugLog("PARSER DEBUG: opTok literal=%q, type=%v\n", opTok.Literal, opTok.Type)

	// Handle compound assignment operators explicitly
	if opTok.Literal == "+=" || opTok.Type == token.PLUS_ASSIGN {
		opTok.Type = token.PLUS_ASSIGN
	} else if opTok.Literal == "-=" || opTok.Type == token.MINUS_ASSIGN {
		opTok.Type = token.MINUS_ASSIGN
	} else if opTok.Literal == "*=" || opTok.Type == token.MUL_ASSIGN {
		opTok.Type = token.MUL_ASSIGN
	} else if opTok.Literal == "/=" || opTok.Type == token.DIV_ASSIGN {
		opTok.Type = token.DIV_ASSIGN
	} else if opTok.Literal == "%=" || opTok.Type == token.MOD_ASSIGN {
		opTok.Type = token.MOD_ASSIGN
	} else if opTok.Literal == ":=" {
		opTok.Type = token.ASSIGN
	} else if opTok.Literal == "=" {
		opTok.Type = token.EQ
	}

	p.debugLog("PARSER DEBUG: final stmt.Token literal=%q, type=%v\n", opTok.Literal, opTok.Type)
	stmt.Token = opTok
	stmt.Values = append(stmt.Values, p.parseExpression())

	for p.l.PeekToken().Type == token.COMMA {
		p.l.NextToken() // comma
		stmt.Values = append(stmt.Values, p.parseExpression())
	}
	p.l.NextToken() // Skip ;
	return stmt
}

func (p *Parser) parseIfStatement(tok token.Token) Statement {
	ifStmt := &IfStatement{Token: tok}
	ifStmt.Condition = p.parseExpression()

	// Condition must be followed by "do"
	if p.l.PeekToken().Type == token.DO || p.l.PeekToken().Literal == "do" {
		p.l.NextToken() // Consume "do"
	}

	// Parse consequence block
	consequence := &BlockStatement{Token: tok}
	for {
		pTok := p.l.PeekToken()
		if pTok.Type == token.EOF || pTok.Type == token.ELSE || pTok.Literal == "else" || pTok.Type == token.DONE || pTok.Literal == "done" {
			break
		}
		nextTok := p.l.NextToken()
		if nextTok.Type == token.EOF {
			break
		}
		stmt := p.parseStatement(nextTok)
		if stmt != nil {
			consequence.Statements = append(consequence.Statements, stmt)
		}
	}
	ifStmt.Consequence = consequence

	// Check for "else"
	pTok := p.l.PeekToken()
	if pTok.Type == token.ELSE || pTok.Literal == "else" {
		elseTok := p.l.NextToken() // Consume "else"

		// Check for "else if"
		if p.l.PeekToken().Type == token.IF || p.l.PeekToken().Literal == "if" {
			elseIfTok := p.l.NextToken()
			elseIfStmt := p.parseIfStatement(elseIfTok)
			alternative := &BlockStatement{Token: elseIfTok, Statements: []Statement{elseIfStmt}}
			ifStmt.Alternative = alternative
			return ifStmt
		}

		alternative := &BlockStatement{Token: elseTok}
		for {
			pNext := p.l.PeekToken()
			if pNext.Type == token.EOF || pNext.Type == token.DONE || pNext.Literal == "done" {
				break
			}
			nextTok := p.l.NextToken()
			if nextTok.Type == token.EOF {
				break
			}
			stmt := p.parseStatement(nextTok)
			if stmt != nil {
				alternative.Statements = append(alternative.Statements, stmt)
			}
		}
		ifStmt.Alternative = alternative
	}

	// Consume "done"
	if p.l.PeekToken().Type == token.DONE || p.l.PeekToken().Literal == "done" {
		p.l.NextToken()
	}
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
	}

	return ifStmt
}

func (p *Parser) parseAssertStatement(tok token.Token) Statement {
	stmt := &AssertStatement{Token: tok}
	stmt.Condition = p.parseExpression()
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken() // Skip ;
	}
	return stmt
}

func (p *Parser) parseExpectStatement(tok token.Token) Statement {
	stmt := &ExpectStatement{Token: tok}
	stmt.Condition = p.parseExpression()
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken() // Skip ;
	}
	return stmt
}

// parsePrintStatement parses the canonical print directive.
// Grammar (spec/02 §5 io_stmt, Decision 6):
//
//	"print" [ "(" expression (, expression)* ")" ] [ "(" [ named_arg_list ] ")" ] ";"
//
// The trailing `(named_arg_list)` is the curried application of the
// canonical `print` rule's named-parameter slot (sep ∈ Str).
// Legacy postfix `"using" [":"] expression` is tolerated but emits
// `E0011 DeprecatedSymbol 'using'` on p.warnings (Decision 6 deprecation
// per spec/03-rules.md §3.1).
func (p *Parser) parsePrintStatement(tok token.Token) *PrintStatement {
	stmt := &PrintStatement{Token: tok}

	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
		return stmt
	}

	// Consume optional '(' for the positional argument list.
	hasParens := false
	if p.l.PeekToken().Type == token.LPAREN {
		hasParens = true
		p.l.NextToken()
	}

	// Parse positional arguments.
	p.debugLog("DEBUG: parsing args...\n")
	for {
		if p.l.PeekToken().Type == token.SEMICOLON || p.l.PeekToken().Type == token.EOF {
			break
		}
		if hasParens && p.l.PeekToken().Type == token.RPAREN {
			break
		}
		expr := p.parseExpression()
		if expr != nil {
			stmt.Expressions = append(stmt.Expressions, expr)
			p.debugLog("DEBUG: Added expr, len: %d\n", len(stmt.Expressions))
		}
		peek := p.l.PeekToken()
		p.debugLog("DEBUG: peeked: type=%v, lit=%q\n", peek.Type, peek.Literal)
		if peek.Type == token.COMMA {
			p.l.NextToken() // consume ','
		} else {
			break
		}
	}
	p.debugLog("DEBUG: Done parsing args, count: %d\n", len(stmt.Expressions))

	// Consume optional ')' of the positional list.
	if hasParens && p.l.PeekToken().Type == token.RPAREN {
		p.l.NextToken()
	}

	// Decision 6 curried named-argument postfix `(name: expr, ...)`.
	// Only entered when there is no separator/token interrupting the
	// canonical form; the legacy `using` postfix is still tolerated as a
	// soft deprecation per Decision 6.
	peek := p.l.PeekToken()
	if peek.Literal == "using" || peek.Type == token.USING {
		p.l.NextToken() // consume "using"
		if p.l.PeekToken().Type == token.COLON {
			p.l.NextToken() // consume ':'
		}
		stmt.Separator = p.parseExpression()
		p.warnings = append(p.warnings, fmt.Sprintf(
			"E0011 DeprecatedSymbol: line=%d symbol='using' (use curried '(sep: ...)'; hardening to E0009 in Phase 7.2 audit)",
			tok.Pos,
		))
	} else if peek.Type == token.LPAREN {
		// Curried Decision-6 form: ( sep: "x" [, end: "y"]* )
		p.l.NextToken() // consume '('
		stmt.NamedArgs = make(map[string]Expression)
		for {
			if p.l.PeekToken().Type == token.SEMICOLON || p.l.PeekToken().Type == token.EOF {
				break
			}
			if p.l.PeekToken().Type == token.RPAREN {
				break
			}
			nameTok := p.l.NextToken()
			if nameTok.Type == token.RPAREN || nameTok.Type == token.EOF || nameTok.Type == token.SEMICOLON {
				break
			}
			if nameTok.Type != token.IDENT {
				// Non-identifier in the named slot list — record a soft E0009
				// (parser still must surface this so authors see it in tests).
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:UnrecognizedStatement: line=%d type=%s literal=%q (expected identifier in curried named-argument list)",
					nameTok.Pos, nameTok.Type, nameTok.Literal,
				))
				break
			}
			if p.l.PeekToken().Type != token.COLON {
				p.errors = append(p.errors, fmt.Sprintf(
					"E0009 SyntaxError:UnrecognizedStatement: line=%d (expected ':' after identifier %q in named-argument list)",
					nameTok.Pos, nameTok.Literal,
				))
				break
			}
			p.l.NextToken() // consume ':'
			valExpr := p.parseExpression()
			if valExpr != nil {
				stmt.NamedArgs[nameTok.Literal] = valExpr
				if nameTok.Literal == "sep" {
					stmt.Separator = valExpr
				}
			}
			if p.l.PeekToken().Type == token.COMMA {
				p.l.NextToken()
				continue
			}
			break
		}
		if p.l.PeekToken().Type == token.RPAREN {
			p.l.NextToken()
		}
	}

	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
	}
	return stmt
}

// isRangeSeparatorToken is true when the token is one of the four
// range-separator tokens introduced by Decision 13 (D13, 2026-09-13):
// RANGE_INCL (`..`), RANGE_LEFT_INC (`..<`), RANGE_RGHT_INC (`>..`),
// or RANGE_EXCL (`>..<`). Used to detect when a parenthesised expression
// in parsePrimary is a range_expr and is therefore eligible for the
// `(step)` postfix step.
func isRangeSeparatorToken(tok token.Token) bool {
	switch tok.Type {
	case token.RANGE_INCL, token.RANGE_LEFT_INC, token.RANGE_RGHT_INC, token.RANGE_EXCL:
		return true
	}
	return false
}

func isBinaryOp(tok token.Token) bool {
	switch tok.Type {
	case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.PERCENT, token.CARET,
		token.SQRT, token.MULT, token.DIV, token.EQ, token.NEQ, token.NOT_EQ, token.NEQ_UNICODE,
		token.LT, token.LTE, token.LTE_UNICODE, token.GT, token.GTE, token.GTE_UNICODE,
		token.APPROX_EQ, token.EQUIV, token.LOGICAL_AND, token.LOGICAL_OR, token.XOR_PLUS,
		token.XOR_MINUS, token.IN_OP, token.NOT_IN, token.SET_INTERSECT, token.SET_UNION,
		token.SUBSET, token.SUPERSET, token.SYM_DIFF, token.RANGE_INCL, token.AND, token.OR,
		token.XOR, token.IN_KEYWORD, token.IS, token.IS_NOT:
		return true
	}
	switch tok.Literal {
	case "+", "-", "*", "/", `\`, "%", "^", "×", "÷", "√", "=", "==", "!=", "≠", "<>", "¬",
		"<", ">", "<=", ">=", "≤", "≥", "≈", "≡", "and", "or", "xor", "∧", "∨", "⊕", "⊖",
		"in", "∈", "!∈", "∩", "∪", "⊂", "⊃", "Δ", "..":
		return true
	}
	if strings.HasSuffix(tok.Literal, "√") || strings.Contains(tok.Literal, "√") {
		return true
	}
	return false
}

// Precedence levels for the precedence-climbing expression parser.
// Convention: HIGHER integer = TIGHTER binding (Clinger / Wirth style).
// The numbering mirrors the canonical table in `todo/DECISIONS.md` §D10
// and `spec/02-statements.md` §5 EBNF, where the spec's "Level 1 = parens"
// (highest) is mapped to the tightest binary level.
const (
	precLogic    = 2 // and, or, xor, ∨, ∧, ⊕ (left-assoc)
	precCompare  = 3 // =, ¬, <, >, <=, >=, ≈ (Decision 12: ¬ canonical; left-assoc)
	precRange    = 4 // .., .! (left-assoc)
	precAdd      = 5 // +, -, ± (left-assoc)
	precMul      = 6 // ×, ÷, *, /, %, \\ (left-assoc)
	precPower    = 7 // ^, ², ³, ⁴, … (RIGHT-associative per D10)
	precRadicals = 8 // unused; √ / ²√ / … are prefix-only, not climbed
)

// opPrecedence returns (precedence, isRightAssociative) for a *binary*
// operator token, or (-1, false) if the token is not a climb-able binary
// operator. Prefix operators (radicals, logical NOT) are NOT in this table;
// they are handled in parsePrimary. The climb loop relies on this helper
// exclusively, which means SQRT/√ is never tracked as a binary mid-climb —
// closing the early-bail loop reported in issues/17-radical-precedence.md.
func opPrecedence(tok token.Token) (int, bool) {
	// Higher number = tighter binding (Clinger convention).
	switch tok.Type {
	case token.ASTERISK, token.MUL_ASSIGN, token.SLASH, token.DIV_ASSIGN,
		token.PERCENT, token.MOD_ASSIGN, token.MULT, token.DIV:
		return precMul, false
	case token.PLUS, token.MINUS, token.PLUS_ASSIGN, token.MINUS_ASSIGN, token.PLUS_MINUS:
		return precAdd, false
	case token.RANGE_INCL, token.RANGE_LEFT_INC, token.RANGE_RGHT_INC, token.RANGE_EXCL:
		return precRange, false
	case token.EQ, token.NEQ, token.NOT_EQ, token.NEQ_UNICODE, token.LT, token.LTE,
		token.LTE_UNICODE, token.GT, token.GTE, token.GTE_UNICODE,
		token.APPROX_EQ, token.EQUIV, token.IS, token.IS_NOT:
		return precCompare, false
	case token.LOGICAL_AND, token.LOGICAL_OR, token.XOR_PLUS, token.XOR_MINUS,
		token.AND, token.OR, token.XOR, token.IN_OP, token.IN_KEYWORD,
		token.NOT_IN, token.SET_INTERSECT, token.SET_UNION, token.SUBSET,
		token.SUPERSET, token.SYM_DIFF:
		return precLogic, false
	case token.CARET:
		// Power family: ^, ², ³, ⁴ … ⁿ. Right-associative per D10.
		return precPower, true
	case token.SQRT:
		// Prefix radical — handled in parsePrimary. The climb loop should
		// never see SQRT as a binary mid-expression. Returning -1 makes the
		// climb terminate cleanly even if a stray √ peeks through.
		return -1, false
	}
	switch tok.Literal {
	case "*", "/", "\\", "%", "×", "÷":
		return precMul, false
	case "+", "-", "±":
		return precAdd, false
	case "..":
		return precRange, false
	case "=", "==", "¬", "<>", "!=", "≠", "<", ">", "<=", ">=", "≤", "≥", "≈", "≡", "is", "is not":
		return precCompare, false
	case "and", "or", "xor", "∧", "∨", "⊕", "⊖", "∪", "∩", "⊂", "⊃", "Δ",
		"in", "∈", "!∈":
		return precLogic, false
	case "^", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹":
		return precPower, true
	}
	return -1, false
}

// parseExpression is the public entry point used by every statement-level
// parser (print, expect, assert, if, let, decl, raise, etc.). It begins
// the climb at the loosest level (logic) so that the entire precedence
// grammar is honoured.
func (p *Parser) parseExpression() Expression {
	return p.parseExpressionClimb(precLogic)
}

// parseExpressionClimb implements the precedence-climbing algorithm from
// Wirth / Clinger, replacing the previous naive for+isBinaryOp loop that
// could not model level-dependent binding for radicals (issue 17).
//
// Precedence ladder (highest→lowest), mapped to integer values in this
// file: parens (always first, syntactic) > power (precPower, right-assoc)
// > radical prefix (consumed in parsePrimary at minPrec=precPower per D10)
// > mul/div (precMul) > add/sub (precAdd) > range (precRange) > compare
// (precCompare) > logic (precLogic).
//
// The loop continues while the next binary operator has precedence
// >= minPrec; the recursive call for the right operand receives minPrec
// = prec+1 for left-associative operators and minPrec = prec for
// right-associative operators (canonical power).
func (p *Parser) parseExpressionClimb(minPrec int) Expression {
	left := p.parsePrimary()
	// Postfix superscript powers (², ³, …, ⁿ) bind tightest after grouping
	// (D10 level 2). The exponent is encoded in the superscript literal
	// itself, so `x³` ≡ x^3 has no separate right operand. We synthesize the
	// exponent literal and reuse BinaryExpression so the evaluator's existing
	// power dispatch (which reads the superscript literal) handles it
	// unchanged. Bare `^` (caret) is a *binary* operator (needs a right
	// operand) and is left for the climb loop below.
	for {
		supTok := p.l.PeekToken()
		if supTok.Type != token.CARET || supTok.Literal == "^" {
			break
		}
		p.l.NextToken() // consume superscript power
		exp := parseSuperscriptIntStatic(supTok.Literal)
		right := &IntegerLiteral{Token: supTok, Value: fmt.Sprintf("%d", exp)}
		left = &BinaryExpression{Token: supTok, Left: left, Right: right}
	}
	for {
		opTok := p.l.PeekToken()
		prec, isRight := opPrecedence(opTok)
		if prec < minPrec {
			break
		}
		p.debugLog("DEBUG: climb consume op %q (prec %d, rightAssoc=%v) minPrec=%d\n",
			opTok.Literal, prec, isRight, minPrec)
		p.l.NextToken() // consume operator
		var nextMin int
		if isRight {
			nextMin = prec
		} else {
			nextMin = prec + 1
		}
		right := p.parseExpressionClimb(nextMin)
		left = &BinaryExpression{Token: opTok, Left: left, Right: right}
	}
	return left
}

// parsePrimary reads a single atomic operand or a prefix operator and
// returns the parsed Expression. It NEVER reads ahead past the operand
// it must consume, leaving the climb loop in parseExpressionClimb to
// pick up any trailing binary operators.
//
// Special cases:
//   - LPAREN recurses the full expression grammar (loosest minPrec) so
//     `(1 + 2)` parses exactly the same way as `1 + 2` — parens are
//     syntactic grouping, not a separate precedence rung.
//   - SQRT (any leading-superscript radical like ²√, ³√, …, plus the
//     bare √) is a NULLARY prefix with minPrec=precPower per D10, so
//     `³√ 2³` correctly yields prefix(radical, Binary(2, ^, 3)). This
//     is the core fix for issue 17.
//   - LOGICAL_NOT (!) recurses at minPrec=precLogic so the NOT binds
//     as loosely as possible, preserving the old hack semantics where
//     the operand consumes the entire trailing expression.
func (p *Parser) parsePrimary() Expression {
	tok := p.l.NextToken()
	p.debugLog("DEBUG: Parsing primary, token: %q type: %v\n", tok.Literal, tok.Type)
	if tok.Type == token.EOF {
		return nil
	}
	if tok.Type == token.SQRT {
		right := p.parseExpressionClimb(precPower)
		return &PrefixExpression{Token: tok, Operator: tok.Literal, Right: right}
	}
	if tok.Type == token.LOGICAL_NOT {
		right := p.parseExpressionClimb(precLogic)
		return &PrefixExpression{Token: tok, Operator: tok.Literal, Right: right}
	}
	// Decision 7 (D7, 2026-09-12): `not` keyword is a synonym for the
	// canonical `!` logical NOT prefix. Treated identically here so the
	// evaluator can dispatch on `Operator` (`"not"` vs `"!"`) or rely on the
	// type-agnostic prefix semantics.
	if tok.Type == token.NOT {
		right := p.parseExpressionClimb(precLogic)
		return &PrefixExpression{Token: tok, Operator: tok.Literal, Right: right}
	}
	// Unary minus: a leading `-` is a prefix negation operator. It binds looser
	// than power (so `-2^2` ≡ `-(2^2)`) but tighter than mul/add (so
	// `-2*3` ≡ `(-2)*3`), which the recursive climb at minPrec=precPower encodes.
	if tok.Type == token.MINUS {
		right := p.parseExpressionClimb(precPower)
		return &PrefixExpression{Token: tok, Operator: tok.Literal, Right: right}
	}
	// Decision 12 (D12, 2026-09-13): `@` is the reference-of prefix. It yields
	// the referenced cell's identity so that `@a = @b` ⇔ `a is b`. The operand
	// is a single referenceable primary (identifier, index, or paren group).
	if tok.Type == token.AT {
		right := p.parsePrimary()
		return &PrefixExpression{Token: tok, Operator: tok.Literal, Right: right}
	}
	// Lambda expression (spec/07-functions.md §2.1 / §6):
	// `λ(params) => (body) [∈ Type]`.
	if tok.Type == token.LAMBDA {
		return p.parseLambdaExpression(tok)
	}
	if tok.Type == token.LPAREN {
		// Inline callback shorthand (spec/07 §2.2 short_lambda):
		// `(x, y) => expr` is a lambda, not a parenthesised expression.
		// Speculatively probe for the `ident_list ) =>` head; on mismatch
		// rewind and parse the parenthesised expression as usual.
		snap := p.l.Snapshot()
		if params, ok := p.tryParseShortLambdaHead(); ok {
			body := p.parseExpression()
			return &LambdaExpression{Token: tok, Params: params, Body: body}
		}
		p.l.Restore(snap)
		left := p.parseExpressionClimb(precLogic)

		// Conditional expression selector (ternary) per spec/02-statements.md
		// §3.2: ( expr_true if condition else expr_false ). The `if` keyword is
		// not a climb-able binary operator, so after parsing expr_true the
		// parser naturally parks on `if`. When present, fold the full
		// parenthesised ternary into a TernaryExpression node; the surrounding
		// parentheses give it syntactic precedence, so no climbing is needed.
		if p.l.PeekToken().Type == token.IF || p.l.PeekToken().Literal == "if" {
			p.l.NextToken() // consume `if`
			cond := p.parseExpressionClimb(precLogic)
			if p.l.PeekToken().Type == token.ELSE || p.l.PeekToken().Literal == "else" {
				p.l.NextToken() // consume `else`
			}
			elseExpr := p.parseExpressionClimb(precLogic)
			if p.l.PeekToken().Type == token.RPAREN {
				p.l.NextToken() // consume ')'
			}
			return &TernaryExpression{Token: tok, Condition: cond, Then: left, Else: elseExpr}
		}

		if p.l.PeekToken().Type == token.RPAREN {
			p.l.NextToken() // consume ')'
		}
		// Decision 13 (D13, 2026-09-13): postfix step `(step)` on a range
		// expression. We only wrap when `left` is a binary expression whose
		// operator is a RANGE_ separator (RANGE_INCL, RANGE_LEFT_INC,
		// RANGE_RGHT_INC, or RANGE_EXCL). For non-range primaries, the
		// trailing `(...)` is reserved for future function-call syntax and
		// is left unparsed (the climb loop will see it as the start of a
		// new primary, which is logically a TypeError at evaluation time).
		if p.l.PeekToken().Type == token.LPAREN {
			if be, ok := left.(*BinaryExpression); ok && isRangeSeparatorToken(be.Token) {
				p.l.NextToken() // consume '('
				stepExpr := p.parseExpressionClimb(precLogic)
				if p.l.PeekToken().Type == token.RPAREN {
					p.l.NextToken() // consume ')'
				}
				left = &SteppedRangeExpression{Token: tok, Range: be, Step: stepExpr}
			}
		}
		return left
	}
	if tok.Type == token.STRING {
		return &StringLiteral{Token: tok, Value: tok.Literal}
	}
	if tok.Type == token.INT || tok.Type == token.REAL {
		return &IntegerLiteral{Token: tok, Value: tok.Literal}
	}
	// Leading-dot boxed-state access (spec/03 §5.4): `.count` reads the named
	// field on the enclosing rule's closure frame. Emitted as a
	// MemberExpression with no Base so the evaluator resolves it against the
	// current call frame's boxed cells.
	if tok.Type == token.DOT || tok.Literal == "." {
		memberTok := p.l.NextToken()
		member, _ := p.identLiteral(memberTok)
		return &MemberExpression{Token: tok, Parts: []string{"." + member}}
	}
	if tok.Type == token.LBRACE {
		// spec/10-collections.md §4: set_lit ::= "{" expression ( "," expression )* "}" ;
		//                          map_lit ::= "{" key_value_pair ( "," key_value_pair )* "}" ;
		// Distinguished per D9 (ratified 2026-09-13): `:` is the structural
		// pair-up. The climb loop treats `:` and `,` as non-operators, so the
		// first element parse parks on whichever separator follows. A top-level
		// `:` after the key expression selects MapLiteral; otherwise the
		// element list folds into a SetLiteral. Empty `{}` is an empty set.
		if p.l.PeekToken().Type == token.RBRACE {
			p.l.NextToken() // consume '}'
			return &SetLiteral{Token: tok}
		}
		first := p.parseExpression()
		if p.l.PeekToken().Type == token.COLON {
			p.l.NextToken() // consume ':'
			mapLit := &MapLiteral{Token: tok}
			mapLit.Pairs = append(mapLit.Pairs, MapPair{Key: first, Value: p.parseExpression()})
			for p.l.PeekToken().Type == token.COMMA {
				p.l.NextToken() // consume ','
				key := p.parseExpression()
				if p.l.PeekToken().Type == token.COLON {
					p.l.NextToken() // consume ':'
				}
				mapLit.Pairs = append(mapLit.Pairs, MapPair{Key: key, Value: p.parseExpression()})
			}
			if p.l.PeekToken().Type == token.RBRACE {
				p.l.NextToken() // consume '}'
			}
			return mapLit
		}
		setLit := &SetLiteral{Token: tok}
		setLit.Elements = append(setLit.Elements, first)
		for p.l.PeekToken().Type == token.COMMA {
			p.l.NextToken() // consume ','
			setLit.Elements = append(setLit.Elements, p.parseExpression())
		}
		if p.l.PeekToken().Type == token.RBRACE {
			p.l.NextToken() // consume '}'
		}
		return setLit
	}
	if tok.Type == token.LBRACKET {
		// spec/10-collections.md §4: array_lit  ::= "[" expression ( "," expression )* "]" ;
		//                           matrix_lit ::= "[" array_lit ( "," array_lit )* "]" ;
		// Elements are full expressions (comma is not a binary operator, so
		// the climb loop parks on ',' and ']'), which also covers nested
		// array literals for matrices.
		arrLit := &ArrayLiteral{Token: tok}
		for p.l.PeekToken().Type != token.RBRACKET && p.l.PeekToken().Type != token.EOF {
			arrLit.Elements = append(arrLit.Elements, p.parseExpression())
			if p.l.PeekToken().Type == token.COMMA {
				p.l.NextToken() // consume ','
				continue
			}
			break
		}
		if p.l.PeekToken().Type == token.RBRACKET {
			p.l.NextToken() // consume ']'
		}
		return arrLit
	}

	var left Expression
	left = &Identifier{Token: tok, Value: tok.Literal}
	if _, isName := p.identLiteral(tok); tok.Type == token.IDENT || isName {
		// Rule call expression (spec/03 §3.1): IDENT followed by `(` is a
		// CallExpression, not a plain identifier. Consume the full arg list.
		if p.l.PeekToken().Type == token.LPAREN {
			p.l.NextToken() // consume '('
			call := &CallExpression{Token: tok, Name: tok.Literal}
			if p.l.PeekToken().Type != token.RPAREN {
				for {
					if p.l.PeekToken().Type == token.RPAREN || p.l.PeekToken().Type == token.EOF {
						break
					}
					call.Args = append(call.Args, p.parseExpression())
					if p.l.PeekToken().Type == token.COMMA {
						p.l.NextToken() // consume ','
						continue
					}
					break
				}
			}
			if p.l.PeekToken().Type == token.RPAREN {
				p.l.NextToken() // consume ')'
			}
			return call
		}
		// Object member access (spec/03 §5.4): `c.next` or `c.next()` invokes a
		// method / reads a field on a bound closure object. Consume the dotted
		// path and the optional call argument list.
		if p.l.PeekToken().Type == token.DOT || p.l.PeekToken().Literal == "." {
			p.l.NextToken() // consume '.'
			memberTok := p.l.NextToken()
			member, _ := p.identLiteral(memberTok)
			me := &MemberExpression{Token: tok, Base: left, Parts: []string{member}}
			if p.l.PeekToken().Type == token.LPAREN {
				p.l.NextToken() // consume '('
				me.IsCall = true
				if p.l.PeekToken().Type != token.RPAREN {
					for {
						if p.l.PeekToken().Type == token.RPAREN || p.l.PeekToken().Type == token.EOF {
							break
						}
						me.Args = append(me.Args, p.parseExpression())
						if p.l.PeekToken().Type == token.COMMA {
							p.l.NextToken() // consume ','
							continue
						}
						break
					}
				}
				if p.l.PeekToken().Type == token.RPAREN {
					p.l.NextToken() // consume ')'
				}
			}
			return me
		}
		if p.l.PeekToken().Type == token.LBRACKET {
			// spec/10-collections.md §4: indexing ::= expression "[" index_expr
			// ( "," index_expr )* "]" ; with index_expr ::= expression | "$" |
			// range_expr. The index is a full expression: the climb loop parks
			// on ',' and ']', so parsing it here consumes exactly the index and
			// leaves the closing bracket under our control. Supports chained
			// indexing (`m[i][j]`) via the surrounding loop.
			for p.l.PeekToken().Type == token.LBRACKET {
				p.l.NextToken() // consume '['
				bracketTok := p.l.PeekToken()
				var idxExpr Expression
				if bracketTok.Literal == "$" {
					p.l.NextToken() // consume '$'
					idxExpr = &Identifier{Token: bracketTok, Value: "$"}
				} else {
					idxExpr = p.parseExpression()
				}
				if p.l.PeekToken().Type == token.RBRACKET {
					p.l.NextToken() // consume ']'
				}
				left = &IndexExpression{Token: bracketTok, Left: left, Index: idxExpr}
			}
		}
	}
	return left
}

// parseLambdaExpression parses an explicit lambda expression per
// spec/07-functions.md §2.1 / §6:
//
//	lambda_expr ::= ( "λ" | "\" ) "(" [ param_list ] ")" "=>" "(" expression ")"
//	                [ ( "∈" | "in" ) type_specifier ] ;
//
// The λ token has already been consumed by parsePrimary. Parameter type
// annotations and the optional result-type annotation are consumed but not
// type-checked (mirrors the rule-signature fast-forward, Phase 7.2 gate).
// The result-type annotation MUST be consumed here: trailing `∈` would
// otherwise be folded into a membership BinaryExpression by the climb loop.
func (p *Parser) parseLambdaExpression(tok token.Token) Expression {
	le := &LambdaExpression{Token: tok}
	if p.l.PeekToken().Type != token.LPAREN {
		p.errors = append(p.errors, fmt.Sprintf(
			"E0009 SyntaxError:InvalidLambdaHeader: line=%d (expected `(` after `λ` in lambda_expr; got type=%s literal=%q)",
			tok.Pos, p.l.PeekToken().Type, p.l.PeekToken().Literal,
		))
		return le
	}
	p.l.NextToken() // consume '('
	le.Params = p.parseParamNameList()
	if p.l.PeekToken().Type == token.RPAREN {
		p.l.NextToken() // consume ')'
	} else {
		p.errors = append(p.errors, fmt.Sprintf(
			"E0009 SyntaxError:InvalidLambdaHeader: line=%d (expected `)` after lambda parameter list; got type=%s literal=%q)",
			tok.Pos, p.l.PeekToken().Type, p.l.PeekToken().Literal,
		))
	}
	if p.l.PeekToken().Type == token.FAT_ARROW {
		p.l.NextToken() // consume '=>'
	} else {
		p.errors = append(p.errors, fmt.Sprintf(
			"E0009 SyntaxError:InvalidLambdaHeader: line=%d (expected `=>` after lambda parameter list; got type=%s literal=%q)",
			tok.Pos, p.l.PeekToken().Type, p.l.PeekToken().Literal,
		))
	}
	// Body — parenthesised expression per the EBNF.
	if p.l.PeekToken().Type == token.LPAREN {
		p.l.NextToken() // consume '('
		le.Body = p.parseExpression()
		if p.l.PeekToken().Type == token.RPAREN {
			p.l.NextToken() // consume ')'
		} else {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:InvalidLambdaBody: line=%d (expected `)` after lambda body expression; got type=%s literal=%q)",
				tok.Pos, p.l.PeekToken().Type, p.l.PeekToken().Literal,
			))
		}
	} else {
		p.errors = append(p.errors, fmt.Sprintf(
			"E0009 SyntaxError:InvalidLambdaBody: line=%d (expected `(` before lambda body expression; got type=%s literal=%q)",
			tok.Pos, p.l.PeekToken().Type, p.l.PeekToken().Literal,
		))
	}
	// Optional result-type annotation `∈ Type` — consumed, binding deferred
	// to Phase 7.2 (same convention as rule signatures).
	if p.l.PeekToken().Type == token.IN_OP || p.l.PeekToken().Type == token.IN_KEYWORD {
		p.l.NextToken() // consume ∈ / in
		p.l.NextToken() // consume type specifier token
	}
	return le
}

// parseParamNameList consumes a comma-separated identifier list where each
// group may carry a shared `∈ Type` annotation (mirroring the
// rule-signature fast-forward in parseRuleEntry) and returns the parameter
// names. The cursor stops on the closing `)` (not consumed).
func (p *Parser) parseParamNameList() []string {
	var params []string
	for {
		pt := p.l.PeekToken()
		if pt.Type == token.RPAREN || pt.Type == token.EOF {
			break
		}
		nameTok := p.l.NextToken()
		name, ok := p.identLiteral(nameTok)
		if !ok {
			p.errors = append(p.errors, fmt.Sprintf(
				"E0009 SyntaxError:InvalidLambdaParam: line=%d (expected parameter identifier; got type=%s literal=%q)",
				nameTok.Pos, nameTok.Type, nameTok.Literal,
			))
			break
		}
		params = append(params, name)
		// Optional group type annotation `∈ Z` — consumed, not type-checked.
		if p.l.PeekToken().Type == token.IN_OP || p.l.PeekToken().Type == token.IN_KEYWORD {
			p.l.NextToken() // consume ∈ / in
			p.l.NextToken() // consume type specifier token
		}
		if p.l.PeekToken().Type == token.COMMA {
			p.l.NextToken() // consume ','
			continue
		}
		break
	}
	return params
}

// tryParseShortLambdaHead probes for the short_lambda head
// `( ident_list ) =>` of spec/07 §2.2 / §6. On success it returns the
// parameter names with the cursor positioned just past `=>`; on mismatch it
// returns ok=false and the caller MUST Restore the lexer snapshot taken
// before the probe.
func (p *Parser) tryParseShortLambdaHead() ([]string, bool) {
	var params []string
	// Empty parameter list: `() => expr`.
	if p.l.PeekToken().Type == token.RPAREN {
		p.l.NextToken() // consume ')'
		if p.l.PeekToken().Type == token.FAT_ARROW {
			p.l.NextToken() // consume '=>'
			return params, true
		}
		return nil, false
	}
	for {
		nameTok := p.l.NextToken()
		name, ok := p.identLiteral(nameTok)
		if !ok {
			return nil, false
		}
		params = append(params, name)
		nt := p.l.NextToken()
		if nt.Type == token.COMMA {
			continue
		}
		if nt.Type == token.RPAREN {
			break
		}
		return nil, false
	}
	if p.l.PeekToken().Type != token.FAT_ARROW {
		return nil, false
	}
	p.l.NextToken() // consume '=>'
	return params, true
}
