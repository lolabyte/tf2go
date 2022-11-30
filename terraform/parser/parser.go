package parser

import (
	"fmt"
	"strconv"

	"github.com/lolabyte/tf2go/terraform/ast"
	"github.com/lolabyte/tf2go/terraform/lexer"
	"github.com/lolabyte/tf2go/terraform/token"
)

type (
	prefixParseFn func() ast.Expression
)

type TypeParser struct {
	lex *lexer.Lexer

	currToken token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn

	errors []string
}

func New(l *lexer.Lexer) *TypeParser {
	p := &TypeParser{
		lex:            l,
		errors:         []string{},
		prefixParseFns: make(map[token.TokenType]prefixParseFn),
	}

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.TRUE, p.parseBool)
	p.registerPrefix(token.FALSE, p.parseBool)
	p.registerPrefix(token.NUMBER, p.parseNumberLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LEFT_SQUARE_BRACE, p.parseListLiteral)
	p.registerPrefix(token.LEFT_CURLY_BRACE, p.parseObjectLiteral)

	p.registerPrefix(token.ANY_TYPE, p.parseAnyTypeLiteral)
	p.registerPrefix(token.BOOL_TYPE, p.parseBoolTypeLiteral)
	p.registerPrefix(token.NUMBER_TYPE, p.parseNumberTypeLiteral)
	p.registerPrefix(token.STRING_TYPE, p.parseStringTypeLiteral)
	p.registerPrefix(token.LIST_TYPE, p.parseListTypeLiteral)
	p.registerPrefix(token.OBJECT_TYPE, p.parseObjectTypeLiteral)
	p.registerPrefix(token.MAP_TYPE, p.parseMapTypeLiteral)
	p.registerPrefix(token.OPTIONAL_TYPE, p.parseOptionalTypeLiteral)
	// TODO: Add token.TUPLE_TYPE. Should be very similar to LIST_TYPE

	p.nextToken()
	p.nextToken()

	return p
}

func (p *TypeParser) ParseType() *ast.Type {
	t := &ast.Type{
		Statements: []ast.Statement{},
	}

	for !p.currTokenIs(token.EOF) {
		s := p.parseStatement()
		if s != nil {
			t.Statements = append(t.Statements, s)
		}
		p.nextToken()
	}

	return t
}

func (p *TypeParser) Errors() []string {
	return p.errors
}

func (p *TypeParser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *TypeParser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.lex.NextToken()
}

func (p *TypeParser) currTokenIs(t token.TokenType) bool {
	return p.currToken.Type == t
}

func (p *TypeParser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *TypeParser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *TypeParser) parseStatement() ast.Statement {
	return p.parseExpressionStatement()
}

func (p *TypeParser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.currToken}
	stmt.Expression = p.parseExpression()

	return stmt
}

func (p *TypeParser) parseExpression() ast.Expression {
	prefix := p.prefixParseFns[p.currToken.Type]
	if prefix == nil {
		// TODO: error
		return nil
	}

	leftExp := prefix()

	return leftExp

}

func (p *TypeParser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression())

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression())
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *TypeParser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
}

func (p *TypeParser) parseBool() ast.Expression {
	boo := &ast.Bool{Token: p.currToken}

	b, err := strconv.ParseBool(p.currToken.Literal)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as bool", p.currToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	boo.Value = b

	return boo
}

func (p *TypeParser) parseNumberLiteral() ast.Expression {
	lit := &ast.NumberLiteral{Token: p.currToken}

	value, err := strconv.ParseInt(p.currToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.currToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *TypeParser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.currToken, Value: p.currToken.Literal}
}

func (p *TypeParser) parseListLiteral() ast.Expression {
	list := &ast.ListLiteral{
		Token:    p.currToken,
		Elements: p.parseExpressionList(token.RIGHT_SQUARE_BRACE),
	}

	return list
}

func (p *TypeParser) parseObjectLiteral() ast.Expression {
	obj := &ast.ObjectLiteral{
		Token: p.currToken,
	}
	obj.KVPairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RIGHT_CURLY_BRACE) {
		p.nextToken()
		key := p.parseExpression()

		if !p.expectPeek(token.ASSIGN) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression()

		if p.peekTokenIs(token.COMMA) {
			// Skip optional comma
			p.nextToken()
		}

		obj.KVPairs[key] = value
	}

	p.nextToken()
	return obj
}

func (p *TypeParser) parseAnyTypeLiteral() ast.Expression {
	return &ast.AnyTypeLiteral{Token: p.currToken}
}

func (p *TypeParser) parseBoolTypeLiteral() ast.Expression {
	return &ast.BoolTypeLiteral{Token: p.currToken}
}

func (p *TypeParser) parseNumberTypeLiteral() ast.Expression {
	return &ast.NumberTypeLiteral{Token: p.currToken}
}

func (p *TypeParser) parseStringTypeLiteral() ast.Expression {
	return &ast.StringTypeLiteral{Token: p.currToken}
}

func (p *TypeParser) parseListTypeLiteral() ast.Expression {
	list := &ast.ListTypeLiteral{Token: p.currToken}

	if !p.expectPeek(token.LEFT_PAREN) {
		return nil
	}

	p.nextToken()
	list.TypeExpression = p.parseExpression()
	p.nextToken()

	return list
}

func (p *TypeParser) parseMapTypeLiteral() ast.Expression {
	m := &ast.MapTypeLiteral{Token: p.currToken}

	if !p.expectPeek(token.LEFT_PAREN) {
		return nil
	}

	p.nextToken()
	m.TypeExpression = p.parseExpression()
	p.nextToken()

	return m
}

func (p *TypeParser) parseObjectTypeLiteral() ast.Expression {
	obj := &ast.ObjectTypeLiteral{Token: p.currToken}

	if !p.expectPeek(token.LEFT_PAREN) {
		return nil
	}

	p.nextToken()
	obj.ObjectSpec = p.parseObjectLiteral()
	p.nextToken()

	return obj
}

func (p *TypeParser) parseOptionalTypeLiteral() ast.Expression {
	opt := &ast.OptionalTypeLiteral{Token: p.currToken}

	if !p.expectPeek(token.LEFT_PAREN) {
		return nil
	}

	p.nextToken()
	opt.TypeExpression = p.parseExpression()

	if p.expectPeek(token.COMMA) {
		p.nextToken()
		opt.DefaultValue = p.parseExpression()
	}

	p.nextToken()

	return opt
}

func (p *TypeParser) peekError(t token.TokenType) {
	p.errors = append(
		p.errors,
		fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type),
	)
}
