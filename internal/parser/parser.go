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
	l      *lexer.Lexer
	errors []string
	debug  bool
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

func New(l *lexer.Lexer) *Parser {
	return &Parser{l: l}
}

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
		}
	}
	return program
}

func (p *Parser) parseStatement(tok token.Token) Statement {
	switch tok.Type {
	case token.NEW:
		return p.parseDeclaration(tok)
	case token.LET:
		return p.parseAssignment(tok)
	case token.PRINT:
		return p.parsePrintStatement(tok)
	case token.ASSERT:
		return p.parseAssertStatement(tok)
	case token.EXPECT:
		return p.parseExpectStatement(tok)
	case token.IF:
		return p.parseIfStatement(tok)
	}
	return nil
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

func (p *Parser) parsePrintStatement(tok token.Token) *PrintStatement {
	stmt := &PrintStatement{Token: tok}

	if p.l.PeekToken().Type == token.SEMICOLON {
		p.l.NextToken()
		return stmt
	}

	// Consume optional '('
	hasParens := false
	if p.l.PeekToken().Type == token.LPAREN {
		hasParens = true
		p.l.NextToken()
	}

	// Parse arguments
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

	// Consume optional ')'
	if hasParens && p.l.PeekToken().Type == token.RPAREN {
		p.l.NextToken()
	}

	// Parse "using" separator
	peek := p.l.PeekToken()
	if peek.Literal == "using" || peek.Type == token.USING {
		p.l.NextToken() // Consume "using"
		if p.l.PeekToken().Type == token.COLON {
			p.l.NextToken() // Consume ':'
		}
		stmt.Separator = p.parseExpression()
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
