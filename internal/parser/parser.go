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
	// Simplified: "new id ∈ Type;"
	p.l.NextToken() // ident
	p.l.NextToken() // ∈
	p.l.NextToken() // type
	p.l.NextToken() // ;
	return &DeclarationStatement{Token: tok}
}

func (p *Parser) parseAssignment(tok token.Token) Statement {
	stmt := &AssignmentStatement{Token: tok}
	tokIdent := p.l.NextToken()
	stmt.Names = append(stmt.Names, &Identifier{Token: tokIdent, Value: tokIdent.Literal})
	p.l.NextToken() // Skip :=
	stmt.Values = append(stmt.Values, p.parseExpression())
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
	p.l.NextToken() // Skip ;
	return stmt
}

func (p *Parser) parseExpression() Expression {
	tok := p.l.NextToken()
	if tok.Type == token.STRING {
		return &StringLiteral{Token: tok, Value: tok.Literal}
	}
	if tok.Type == token.INT {
		return &IntegerLiteral{Token: tok, Value: tok.Literal}
	}
	return &Identifier{Token: tok, Value: tok.Literal}
}
