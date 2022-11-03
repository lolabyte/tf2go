package lexer

import (
	"testing"

	"github.com/lolabyte/tf2go/terraform/token"
)

func TestNextToken(t *testing.T) {
	type tok struct {
		expectedType    token.TokenType
		expectedLiteral string
	}

	testCases := []struct {
		title  string
		input  string
		tokens []tok
	}{
		{
			title: "Scalar keyword bool",
			input: "bool",
			tokens: []tok{
				{token.BOOL_TYPE, "bool"},
			},
		},
		{
			title: "Scalar keyword number",
			input: "number",
			tokens: []tok{
				{token.NUMBER_TYPE, "number"},
			},
		},
		{
			title: "Scalar keyword string",
			input: "string",
			tokens: []tok{
				{token.STRING_TYPE, "string"},
			},
		},
		{
			title: "Bool literal: true",
			input: "true",
			tokens: []tok{
				{token.TRUE, "true"},
			},
		},
		{
			title: "Bool literal: false",
			input: "false",
			tokens: []tok{
				{token.FALSE, "false"},
			},
		},
		{
			title: "Number literal",
			input: "99",
			tokens: []tok{
				{token.NUMBER, "99"},
			},
		},
		{
			input: `list(object({
				name    = string
				enabled = optional(bool, true)
				website = optional(
					object({
						index_document = optional(string, "index.html")
					}), {})
				}))`,
			tokens: []tok{
				{token.LIST_TYPE, "list"},
				{token.LEFT_PAREN, "("},
				{token.OBJECT_TYPE, "object"},
				{token.LEFT_PAREN, "("},
				{token.LEFT_CURLY_BRACE, "{"},
				{token.IDENT, "name"},
				{token.ASSIGN, "="},
				{token.STRING_TYPE, "string"},
				{token.IDENT, "enabled"},
				{token.ASSIGN, "="},
				{token.OPTIONAL, "optional"},
				{token.LEFT_PAREN, "("},
				{token.BOOL_TYPE, "bool"},
				{token.COMMA, ","},
				{token.TRUE, "true"},
				{token.RIGHT_PAREN, ")"},
				{token.IDENT, "website"},
				{token.ASSIGN, "="},
				{token.OPTIONAL, "optional"},
				{token.LEFT_PAREN, "("},
				{token.OBJECT_TYPE, "object"},
				{token.LEFT_PAREN, "("},
				{token.LEFT_CURLY_BRACE, "{"},
				{token.IDENT, "index_document"},
				{token.ASSIGN, "="},
				{token.OPTIONAL, "optional"},
				{token.LEFT_PAREN, "("},
				{token.STRING_TYPE, "string"},
				{token.COMMA, ","},
				{token.STRING, "index.html"},
				{token.RIGHT_PAREN, ")"},
				{token.RIGHT_CURLY_BRACE, "}"},
				{token.RIGHT_PAREN, ")"},
				{token.COMMA, ","},
				{token.LEFT_CURLY_BRACE, "{"},
				{token.RIGHT_CURLY_BRACE, "}"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.title, func(t *testing.T) {
			l := New(testCase.input)
			for i, expected := range testCase.tokens {
				tkn := l.NextToken()
				if tkn.Type != expected.expectedType {
					t.Fatalf("token #%d has wrong token.Type, expected=%q, got=%q", i, expected.expectedType, tkn.Type)
				}
			}
		})
	}
}
