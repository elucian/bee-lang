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
	case token.LET:
		return p.parseAssignment(tok)
	case token.PRINT:
		return p.parsePrintStatement(tok)
	case token.NEW:
		return p.parseDeclaration(tok)
	default:
		return nil
	}
}

func (p *Parser) parsePrintStatement(tok token.Token) *PrintStatement {
	stmt := &PrintStatement{Token: tok}
	for {
		stmt.Expressions = append(stmt.Expressions, p.parseExpression())
		if p.l.PeekChar() == ';' {
			p.l.NextToken()
			break
		}
		p.l.NextToken() // skip comma
	}
	return stmt
}

func (p *Parser) parseDeclaration(tok token.Token) Statement {
	p.l.NextToken() // skip ident
	p.l.NextToken() // skip ∈
	p.l.NextToken() // skip type
	p.l.NextToken() // skip ;
	return nil      // Simplified for now
}

func (p *Parser) parseAssignment(tok token.Token) Statement {
	stmt := &AssignmentStatement{Token: tok}
	tokIdent := p.l.NextToken()
	stmt.Names = append(stmt.Names, &Identifier{Token: tokIdent, Value: tokIdent.Literal})
	p.l.NextToken() // Skip :=
	stmt.Values = append(stmt.Values, p.parseExpression())
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
