// internal/parser/parser.go
package parser

import (
	"bee/internal/lexer"
	"bee/internal/token"
	"fmt"
)

type Parser struct {
	l *lexer.Lexer
}

func New(l *lexer.Lexer) *Parser {
	return &Parser{l: l}
}

// ParseProgram iterates through the lexer token stream and dispatches
// to the appropriate statement parsing methods. It serves as the primary
// entry point for building the program AST.
func (p *Parser) ParseProgram() *Program {
	program := &Program{}
	for {
		tok := p.l.NextToken()
		if tok.Type == token.EOF {
			break
		}

		// DEBUG: Log token stream to visualize the parse state
		fmt.Printf("DEBUG: Parsing Token: %s, Literal: %s\n", tok.Type, tok.Literal)

		if tok.Type == token.RULE || tok.Type == token.NEW || tok.Type == token.LET || tok.Type == token.RETURN {
			continue
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
	case token.PRINT:
		return p.parsePrintStatement(tok)
	case token.IDENT:
		// Handle potential identifier usage in statement
		return nil
	default:
		return nil
	}
}

func (p *Parser) parsePrintStatement(tok token.Token) *PrintStatement {
	stmt := &PrintStatement{Token: tok}
	stmt.Expression = p.parseExpression()
	return stmt
}

func (p *Parser) parseExpression() Expression {
	tok := p.l.NextToken()
	fmt.Printf("DEBUG: Token Type: %s, Literal: %s\n", tok.Type, tok.Literal)
	if tok.Type == token.STRING {
		return &StringLiteral{Token: tok, Value: tok.Literal}
	}
	return &Identifier{Token: tok, Value: tok.Literal}
}
