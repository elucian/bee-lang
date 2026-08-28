// internal/parser/parser.go
package parser

import (
	"bee/internal/lexer"
	"bee/internal/token"
	"fmt"
	"os"
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
	}
	return nil
}

func (p *Parser) parseDeclaration(tok token.Token) Statement {
	ds := &DeclarationStatement{Token: tok}
	identTok := p.l.NextToken() // ident
	ds.Name = identTok.Literal
	nextTok := p.l.NextToken() // ∈ or :=
	if nextTok.Literal == ":=" {
		ds.Value = p.parseExpression()
	} else {
		// e.g. new lista := [10, 20, 30];
		// Let's check next token
		t2 := p.l.NextToken()
		if t2.Literal == ":=" {
			ds.Value = p.parseExpression()
		}
	}
	p.l.NextToken() // Skip ;
	return ds
}

func (p *Parser) parseAssignment(tok token.Token) Statement {
	stmt := &AssignmentStatement{Token: tok}
	tokIdent := p.l.NextToken()
	fmt.Fprintf(os.Stderr, "PARSER DEBUG: ident tok=%q\n", tokIdent.Literal)
	stmt.Names = append(stmt.Names, &Identifier{Token: tokIdent, Value: tokIdent.Literal})

	// Check for more comma-separated variables
	for p.l.PeekChar() == ',' {
		p.l.NextToken() // comma
		tokIdent2 := p.l.NextToken()
		stmt.Names = append(stmt.Names, &Identifier{Token: tokIdent2, Value: tokIdent2.Literal})
	}

	// Look ahead to check if next token is an assignment operator
	opTok := p.l.NextToken()
	fmt.Fprintf(os.Stderr, "PARSER DEBUG: opTok literal=%q, type=%v\n", opTok.Literal, opTok.Type)

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

	fmt.Fprintf(os.Stderr, "PARSER DEBUG: final stmt.Token literal=%q, type=%v\n", opTok.Literal, opTok.Type)
	stmt.Token = opTok
	stmt.Values = append(stmt.Values, p.parseExpression())

	for p.l.PeekChar() == ',' {
		p.l.NextToken() // comma
		stmt.Values = append(stmt.Values, p.parseExpression())
	}
	p.l.NextToken() // Skip ;
	return stmt
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

	// Consume optional '('
	hasParens := false
	if p.l.PeekToken().Type == token.LPAREN {
		hasParens = true
		p.l.NextToken()
	}

	// Parse arguments
	fmt.Fprintf(os.Stderr, "DEBUG: parsing args...\n")
	for {
		expr := p.parseExpression()
		if expr != nil {
			stmt.Expressions = append(stmt.Expressions, expr)
			fmt.Fprintf(os.Stderr, "DEBUG: Added expr, len: %d\n", len(stmt.Expressions))
		}
		peek := p.l.PeekToken()
		fmt.Fprintf(os.Stderr, "DEBUG: peeked: type=%v, lit=%q\n", peek.Type, peek.Literal)
		if peek.Type == token.COMMA {
			p.l.NextToken() // consume ','
		} else {
			break
		}
	}
	fmt.Fprintf(os.Stderr, "DEBUG: Done parsing args, count: %d\n", len(stmt.Expressions))
	fmt.Fprintf(os.Stderr, "DEBUG: Done parsing args, count: %d\n", len(stmt.Expressions))

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

	if p.l.PeekChar() == ';' {
		p.l.NextToken()
	}
	return stmt
}

func (p *Parser) parseExpression() Expression {
	tok := p.l.NextToken()
	fmt.Fprintf(os.Stderr, "DEBUG: Parsing expr, token: %q type: %v\n", tok.Literal, tok.Type)
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
		nextTok := p.l.NextToken()
		if nextTok.Type == token.CARET || nextTok.Literal == "^" || nextTok.Type == token.INT || (nextTok.Literal >= "⁰" && nextTok.Literal <= "⁹") {
			var right Expression
			if nextTok.Type == token.INT {
				right = &IntegerLiteral{Token: nextTok, Value: nextTok.Literal}
			} else {
				right = p.parseExpression()
			}
			return &BinaryExpression{Token: token.Token{Type: token.CARET, Literal: "^"}, Left: left, Right: right}
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
		fmt.Fprintf(os.Stderr, "DEBUG: In binary loop, peekTok type: %v, literal: %q\n", peekTok.Type, peekTok.Literal)
		if peekTok.Type == token.PLUS || peekTok.Type == token.MINUS || peekTok.Type == token.ASTERISK || peekTok.Type == token.SLASH || peekTok.Literal == "/" || peekTok.Type == token.PERCENT || peekTok.Literal == "%" || peekTok.Type == token.CARET || peekTok.Literal == "^" || peekTok.Type == token.SQRT || peekTok.Literal == "√" || peekTok.Literal == "²√" || peekTok.Literal == "³√" || peekTok.Literal == "⁴√" || peekTok.Literal == "⁵√" || peekTok.Literal == "⁶√" || peekTok.Literal == "⁷√" || peekTok.Literal == "⁸√" || peekTok.Literal == "⁹√" || peekTok.Literal == "¹⁰√" || peekTok.Type == token.EQ || peekTok.Literal == "=" || peekTok.Literal == "==" || peekTok.Type == token.NEQ_UNICODE || peekTok.Literal == "≠" || peekTok.Literal == "â‰ " || peekTok.Type == token.NOT_EQ || peekTok.Type == token.GT || peekTok.Type == token.GTE || peekTok.Type == token.LT || peekTok.Type == token.LTE || peekTok.Type == token.AND || peekTok.Type == token.OR || peekTok.Type == token.XOR || peekTok.Literal == ">" || peekTok.Literal == "<" || peekTok.Literal == ">=" || peekTok.Literal == "<=" || peekTok.Literal == "!=" {
			p.l.NextToken() // Consume operator
			right := p.parseExpression()
			left = &BinaryExpression{Token: peekTok, Left: left, Right: right}
			continue
		}
		break
	}
	return left
}
