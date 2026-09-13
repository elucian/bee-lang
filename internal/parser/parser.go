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
	case token.RETURN, token.STOP, token.REDO, token.NEXT, token.PASS,
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
	stmt.Name = p.l.NextToken().Literal // rule identifier

	// Optional primary parameter list.
	if p.l.PeekToken().Type == token.LPAREN {
		p.l.NextToken() // consume '('
		if p.l.PeekToken().Type != token.RPAREN {
			for {
				if p.l.PeekToken().Type == token.RPAREN || p.l.PeekToken().Type == token.EOF {
					break
				}
				p.l.NextToken() // bind param name
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
					p.l.NextToken()
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
		if peek.Type == token.EOF || peek.Type == token.RETURN {
			break
		}
		// Stop at the aligned return terminator (0 relative indentation
		// see spec/02 §6.3); only top-level return — i.e. immediately after
		// the body — closes the rule.
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

// parseMatchStatement is a grammar-mapped stub for spec/02-statements.md §3.3
// `match`. Consumes to the next ';' so the trailing tokens are not silently
// dropped and emits a soft W0901 until the full grammar is implemented.
func (p *Parser) parseMatchStatement(tok token.Token) Statement {
	stmt := &MatchStatement{Token: tok}
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
	p.warnings = append(p.warnings, fmt.Sprintf(
		"W0901 UnrecognizedStatement: line=%d type=MATCH (grammar stub)", tok.Pos,
	))
	return stmt
}

// parseScopeStatement handles `start` / `with` blocks per spec/02-statements.md §3.1.
func (p *Parser) parseScopeStatement(tok token.Token) Statement {
	stmt := &ScopeStatement{Token: tok, Keyword: tok.Literal}
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

// parseCycleStatement handles `cycle`, `for`, and `while` per spec/02-statements.md §3.4.
func (p *Parser) parseCycleStatement(tok token.Token) Statement {
	stmt := &CycleStatement{Token: tok, Keyword: tok.Literal}
	for {
		peek := p.l.PeekToken()
		if peek.Type == token.SEMICOLON {
			p.l.NextToken()
			break
		}
		if peek.Type == token.EOF || peek.Type == token.DONE || peek.Literal == "repeat" {
			break
		}
		p.l.NextToken()
	}
	return stmt
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

// parseTransferStatement handles `return`, `stop`, `redo`, `next`, `pass`,
// `raise`, `resume`, `retry`, `fail` per spec/02-statements.md §5 transfer_stmt.
func (p *Parser) parseTransferStatement(tok token.Token) Statement {
	stmt := &TransferStatement{Token: tok, Keyword: tok.Literal}
	// Optional expression payload (e.g., `raise <expr>`, `fail <expr>`).
	if tok.Type == token.RAISE || tok.Type == token.FAIL || tok.Type == token.RETURN {
		if p.l.PeekToken().Type != token.SEMICOLON && p.l.PeekToken().Type != token.EOF {
			stmt.Value = p.parseExpression()
		}
	}
	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
	}
	return stmt
}

func (p *Parser) parseDeclaration(tok token.Token) Statement {
	ds := &DeclarationStatement{Token: tok}
	identTok := p.l.NextToken() // first ident
	ds.Name = identTok.Literal
	ds.Names = append(ds.Names, identTok.Literal)

	// Check for comma-separated identifier list (e.g. new a, b, c ∈ Z;)
	for p.l.PeekToken().Type == token.COMMA {
		p.l.NextToken() // consume ','
		nextIdent := p.l.NextToken()
		ds.Names = append(ds.Names, nextIdent.Literal)
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
	stmt.Names = append(stmt.Names, &Identifier{Token: tokIdent, Value: tokIdent.Literal})

	// Check for more comma-separated variables
	for p.l.PeekChar() == ',' {
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

	for p.l.PeekChar() == ',' {
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
	if tok.Type == token.LPAREN {
		left := p.parseExpressionClimb(precLogic)
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
	if tok.Type == token.LBRACKET {
		arrLit := &ArrayLiteral{Token: tok}
		for {
			elTok := p.l.NextToken()
			if elTok.Type == token.RBRACKET || elTok.Type == token.EOF {
				break
			}
			if elTok.Type == token.INT {
				arrLit.Elements = append(arrLit.Elements, &IntegerLiteral{Token: elTok, Value: elTok.Literal})
			} else if elTok.Type == token.IDENT {
				arrLit.Elements = append(arrLit.Elements, &Identifier{Token: elTok, Value: elTok.Literal})
			}
			commaOrBracket := p.l.NextToken()
			if commaOrBracket.Type == token.RBRACKET || commaOrBracket.Type == token.EOF {
				break
			}
		}
		return arrLit
	}

	var left Expression
	left = &Identifier{Token: tok, Value: tok.Literal}
	if tok.Type == token.IDENT {
		if p.l.PeekToken().Type == token.LBRACKET {
			p.l.NextToken() // consume '['
			bracketTok := p.l.NextToken()
			var idxExpr Expression
			if bracketTok.Literal == "$" {
				idxExpr = &Identifier{Token: bracketTok, Value: "$"}
			} else if bracketTok.Type == token.IDENT {
				idxExpr = &Identifier{Token: bracketTok, Value: bracketTok.Literal}
			} else {
				idxExpr = &IntegerLiteral{Token: bracketTok, Value: bracketTok.Literal}
			}
			p.l.NextToken() // consume ']'
			left = &IndexExpression{Token: bracketTok, Left: left, Index: idxExpr}
		}
	}
	return left
}
