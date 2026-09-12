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
		token.RAISE, token.RESUME, token.RETRY, token.FAIL:
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
		// e.g. new a: 1, b: 2 ∈ Z;
		_ = p.parseExpression() // consume initial val
		for p.l.PeekToken().Type != token.SEMICOLON && p.l.PeekToken().Type != token.EOF {
			p.l.NextToken()
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

func isBinaryOp(tok token.Token) bool {
	switch tok.Type {
	case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.PERCENT, token.CARET,
		token.SQRT, token.MULT, token.DIV, token.EQ, token.NOT_EQ, token.NEQ_UNICODE,
		token.LT, token.LTE, token.LTE_UNICODE, token.GT, token.GTE, token.GTE_UNICODE,
		token.APPROX_EQ, token.EQUIV, token.LOGICAL_AND, token.LOGICAL_OR, token.XOR_PLUS,
		token.XOR_MINUS, token.IN_OP, token.NOT_IN, token.SET_INTERSECT, token.SET_UNION,
		token.SUBSET, token.SUPERSET, token.SYM_DIFF, token.RANGE_INCL, token.AND, token.OR,
		token.XOR, token.IN_KEYWORD:
		return true
	}
	switch tok.Literal {
	case "+", "-", "*", "/", `\`, "%", "^", "×", "÷", "√", "=", "==", "!=", "≠", "<>",
		"<", ">", "<=", ">=", "≤", "≥", "≈", "≡", "and", "or", "xor", "∧", "∨", "⊕", "⊖",
		"in", "∈", "!∈", "∩", "∪", "⊂", "⊃", "Δ", "..":
		return true
	}
	if strings.HasSuffix(tok.Literal, "√") || strings.Contains(tok.Literal, "√") {
		return true
	}
	return false
}

func (p *Parser) parseExpression() Expression {
	tok := p.l.NextToken()
	p.debugLog("DEBUG: Parsing expr, token: %q type: %v\n", tok.Literal, tok.Type)
	if tok.Type == token.EOF {
		return nil
	}
	var left Expression
	if tok.Type == token.SQRT || tok.Literal == "²√" || tok.Literal == "³√" || tok.Literal == "⁴√" || tok.Literal == "⁵√" || tok.Literal == "⁶√" || tok.Literal == "⁷√" || tok.Literal == "⁸√" || tok.Literal == "⁹√" || tok.Literal == "¹⁰√" || tok.Literal == "√" {
		right := p.parseExpression()
		return &BinaryExpression{Token: tok, Left: &Identifier{Token: token.Token{Literal: "0"}, Value: "0"}, Right: right}
	} else if tok.Type == token.LPAREN {
		left = p.parseExpression()
		// Consume ')'
		if p.l.PeekToken().Type == token.RPAREN {
			p.l.NextToken()
		}
	} else if tok.Type == token.LOGICAL_NOT {
		right := p.parseExpression()
		return &BinaryExpression{Token: tok, Left: &IntegerLiteral{Token: token.Token{Literal: "0"}, Value: "0"}, Right: right}
	} else if tok.Type == token.STRING {
		left = &StringLiteral{Token: tok, Value: tok.Literal}
	} else if tok.Type == token.INT || tok.Type == token.REAL {
		left = &IntegerLiteral{Token: tok, Value: tok.Literal}
	} else if tok.Type == token.LBRACKET {
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
		left = arrLit
	} else {
		left = &Identifier{Token: tok, Value: tok.Literal}
		// Check if token is IDENT and might be evaluated as symbol
		if tok.Type == token.IDENT {
			// Check for index expression e.g. lista[1] or lista[x]
			if p.l.PeekChar() == '[' {
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
	}

	// Simple binary expression check (including √)
	for {
		peekTok := p.l.PeekToken()
		p.debugLog("DEBUG: In binary loop, peekTok type: %v, literal: %q\n", peekTok.Type, peekTok.Literal)
		if isBinaryOp(peekTok) {
			p.l.NextToken() // Consume operator
			right := p.parseExpression()
			left = &BinaryExpression{Token: peekTok, Left: left, Right: right}
			continue
		}
		break
	}
	return left
}
