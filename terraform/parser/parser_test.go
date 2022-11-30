package parser

import (
	"fmt"
	"testing"

	"github.com/lolabyte/tf2go/terraform/ast"
	"github.com/lolabyte/tf2go/terraform/lexer"
	"github.com/stretchr/testify/assert"
)

func TestParseListLiteral(t *testing.T) {
	input := "[1,2,3]"

	l := lexer.New(input)
	p := New(l)
	typeDef := p.ParseType()
	checkParserErrors(t, p)

	stmt, _ := typeDef.Statements[0].(*ast.ExpressionStatement)
	list, ok := stmt.Expression.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("exp not ast.ListLiteral. got=%T", stmt.Expression)
	}

	if len(list.Elements) != 3 {
		t.Fatalf("len(array.Elements) not 3. got=%d", len(list.Elements))
	}

	testIntegerLiteral(t, list.Elements[0], 1)
}

func TestParsingObjectLiteral(t *testing.T) {
	input := `{
		a_number = 1
		a_bool = true
		string_list = [1,2,3]
	}`

	l := lexer.New(input)
	p := New(l)
	typeDef := p.ParseType()
	checkParserErrors(t, p)

	stmt := typeDef.Statements[0].(*ast.ExpressionStatement)
	obj, ok := stmt.Expression.(*ast.ObjectLiteral)
	if !ok {
		t.Fatalf("exp is not ast.ObjectLiteral. got=%T", stmt.Expression)
	}

	expected := map[string]interface{}{
		"a_number":      1,
		"a_bool":        true,
		"a_string_list": []int64{1, 2, 3},
	}

	if len(obj.KVPairs) != len(expected) {
		t.Errorf("obj.KVPairs has wrong length. got=%d", len(obj.KVPairs))
	}

	for key, value := range obj.KVPairs {
		var k string
		switch key.(type) {
		case *ast.Identifier:
			k = key.String()
		case *ast.StringLiteral:
			k = key.String()
		default:
			k = ""
			t.Errorf("key is not ast.Identifier or ast.StringLiteral. got=%T", key)
			continue
		}

		expectedValue := expected[k]
		switch expected[k].(type) {
		case bool:
			testBooleanLiteral(t, value, expectedValue.(bool))
		case int64:
			testIntegerLiteral(t, value, expectedValue.(int64))
		case []int64:
			for i, n := range expectedValue.([]int64) {
				testIntegerLiteral(t, value.(*ast.ListLiteral).Elements[i], n)
			}
		}
	}
}

func TestParseListTypeLiteral(t *testing.T) {
	input := `list(string)`

	l := lexer.New(input)
	p := New(l)
	typeDef := p.ParseType()
	checkParserErrors(t, p)

	stmt := typeDef.Statements[0].(*ast.ExpressionStatement)
	list, ok := stmt.Expression.(*ast.ListTypeLiteral)
	if !ok {
		t.Fatalf("exp is not ast.ListTypeLiteral. got=%T", stmt.Expression)
	}

	assert.Equal(t, "string", list.TypeExpression.String())
}

func TestParsingObjectTypeLiteral(t *testing.T) {
	input := `object({ 
		a_number = number
		a_bool = bool
		a_string_list = list(string) 
	})`

	l := lexer.New(input)
	p := New(l)
	typeDef := p.ParseType()
	checkParserErrors(t, p)

	stmt := typeDef.Statements[0].(*ast.ExpressionStatement)
	obj, ok := stmt.Expression.(*ast.ObjectTypeLiteral)
	if !ok {
		t.Fatalf("exp is not ast.ObjectTypeLiteral. got=%T", stmt.Expression)
	}

	expectedObjSpec := map[string]string{
		"a_number":      "number",
		"a_bool":        "bool",
		"a_string_list": "list(string)",
	}

	objSpec := obj.ObjectSpec.(*ast.ObjectLiteral)
	if len(objSpec.KVPairs) != len(expectedObjSpec) {
		t.Errorf("obj.KVPairs has wrong length. got=%d", len(objSpec.KVPairs))
	}

	for key := range objSpec.KVPairs {
		assert.Equal(t, objSpec.KVPairs[key].String(), expectedObjSpec[key.String()])
	}
}

func TestParseComplexType(t *testing.T) {
	input := `
		list(
		  object({
			name                     = string
			ip_cidr_range            = string
			private_ip_google_access = bool
			purpose                  = string
			secondary_ip_ranges = list(
			  object({
				range_name    = string
				ip_cidr_range = string
			  })
			)
		  })
		)
	  `

	l := lexer.New(input)
	p := New(l)
	typeDef := p.ParseType()
	checkParserErrors(t, p)

	stmt := typeDef.Statements[0].(*ast.ExpressionStatement)
	_, ok := stmt.Expression.(*ast.ListTypeLiteral)
	if !ok {
		t.Fatalf("exp is not ast.ListTypeLiteral. got=%T", stmt.Expression)
	}

}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.NumberLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			integ.TokenLiteral())
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
			ident.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.Bool)
	if !ok {
		t.Errorf("exp not *ast.Bool. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s",
			value, bo.TokenLiteral())
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *TypeParser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
