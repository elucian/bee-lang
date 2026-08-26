// internal/parser/parser.go
package parser

import (
	"bee/internal/lexer"
	"bee/internal/token"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case token.RULE:
		return p.parseRuleStatement()
	case token.PRINT:
		return p.parsePrintStatement()
	case token.NEW:
		return p.parseDeclaration()
	case token.LET:
		return p.parseAssignment()
	default:
		return nil
	}
}

// Minimal parse implementations for T0101/T0202
func (p *Parser) parseRuleStatement() Statement { return nil }
func (p *Parser) parseDeclaration() Statement   { return nil }
func (p *Parser) parseAssignment() Statement    { return nil }

func (p *Parser) parsePrintStatement() *PrintStatement {
	stmt := &PrintStatement{Token: p.curToken}
	p.nextToken() // Skip PRINT
	stmt.Expression = p.parseExpression()
	return stmt
}

func (p *Parser) parseExpression() Expression {
	if p.curToken.Type == token.STRING {
		return &StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
	}
	return &Identifier{Token: p.curToken, Value: p.curToken.Literal}
}
