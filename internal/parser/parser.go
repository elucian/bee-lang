// internal/parser/parser.go
package parser

import (
	"bee/internal/lexer"
	"bee/internal/token"
)

type Parser struct {
	l *lexer.Lexer
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
	case token.EXPECT:
		return p.parseExpectStatement(tok)
	case token.PRINT:
		return p.parsePrintStatement(tok)
	default:
		return nil
	}
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
	stmt.Names = append(stmt.Names, &Identifier{Token: tokIdent, Value: tokIdent.Literal})
	for p.l.PeekChar() == ',' {
		p.l.NextToken() // comma
		tokIdent2 := p.l.NextToken()
		stmt.Names = append(stmt.Names, &Identifier{Token: tokIdent2, Value: tokIdent2.Literal})
	}
	opTok := p.l.NextToken() // Skip := or ::
	stmt.Token = opTok
	stmt.Values = append(stmt.Values, p.parseExpression())
	for p.l.PeekChar() == ',' {
		p.l.NextToken() // comma
		stmt.Values = append(stmt.Values, p.parseExpression())
	}
	p.l.NextToken() // Skip ;
	return stmt
}

func (p *Parser) parseExpectStatement(tok token.Token) Statement {
	stmt := &ExpectStatement{Token: tok}
	stmt.Condition = p.parseExpression()
	p.l.NextToken() // Skip ;
	return stmt
}

func (p *Parser) parsePrintStatement(tok token.Token) *PrintStatement {
	stmt := &PrintStatement{Token: tok}
	stmt.Expressions = append(stmt.Expressions, p.parseExpression())
	for p.l.PeekChar() == ',' {
		p.l.NextToken() // consume comma
		stmt.Expressions = append(stmt.Expressions, p.parseExpression())
	}
	p.l.NextToken() // Skip ;
	return stmt
}

func (p *Parser) parseExpression() Expression {
	tok := p.l.NextToken()
	var left Expression
	if tok.Type == token.SQRT || tok.Literal == "²√" || tok.Literal == "³√" || tok.Literal == "⁴√" || tok.Literal == "⁵√" || tok.Literal == "⁶√" || tok.Literal == "⁷√" || tok.Literal == "⁸√" || tok.Literal == "⁹√" || tok.Literal == "¹⁰√" || tok.Literal == "√" {
		right := p.parseExpression()
		return &BinaryExpression{Token: tok, Left: &IntegerLiteral{Token: token.Token{Literal: "0"}, Value: "0"}, Right: right}
	} else if tok.Type == token.LPAREN {
		left = p.parseExpression()
		p.l.NextToken() // consume ')'
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
	} else if tok.Type == token.INT {
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

	// Simple binary expression check (including √)
	peekTok := p.l.NextToken()
	if peekTok.Type == token.PLUS || peekTok.Type == token.MINUS || peekTok.Type == token.ASTERISK || peekTok.Type == token.SLASH || peekTok.Literal == "/" || peekTok.Type == token.PERCENT || peekTok.Type == token.PERCENT || peekTok.Literal == "%" || peekTok.Type == token.CARET || peekTok.Literal == "^" || peekTok.Type == token.SQRT || peekTok.Literal == "√" || peekTok.Literal == "²√" || peekTok.Literal == "³√" || peekTok.Literal == "⁴√" || peekTok.Literal == "⁵√" || peekTok.Literal == "⁶√" || peekTok.Literal == "⁷√" || peekTok.Literal == "⁸√" || peekTok.Literal == "⁹√" || peekTok.Literal == "¹⁰√" || peekTok.Type == token.EQ || peekTok.Literal == "=" || peekTok.Literal == "==" || peekTok.Type == token.NEQ_UNICODE || peekTok.Literal == "≠" || peekTok.Type == token.NOT_EQ || peekTok.Type == token.GT || peekTok.Type == token.GTE || peekTok.Type == token.LT || peekTok.Type == token.LTE || peekTok.Type == token.AND_KW || peekTok.Type == token.OR_KW || peekTok.Type == token.XOR_KW || peekTok.Literal == ">" || peekTok.Literal == "<" || peekTok.Literal == ">=" || peekTok.Literal == "<=" || peekTok.Literal == "!=" {
		right := p.parseExpression()
		return &BinaryExpression{Token: peekTok, Left: left, Right: right}
	}
	return left
}
