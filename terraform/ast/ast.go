package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/lolabyte/tf2go/terraform/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Type struct {
	Token      token.Token // token.BOOL | token.NUMBER | token.STRING | token.LIST | token.TUPLE | token.MAP | token.OBJECT
	Statements []Statement
}

func (ts *Type) TokenLiteral() string { return ts.Token.Literal }
func (ts *Type) String() string {
	var out bytes.Buffer

	for _, s := range ts.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type Bool struct {
	Token token.Token // token.BOOL
	Value bool
}

func (nl *Bool) expressionNode()      {}
func (nl *Bool) TokenLiteral() string { return nl.Token.Literal }
func (nl *Bool) String() string       { return nl.Token.Literal }

type NumberLiteral struct {
	Token token.Token // token.NUMBER
	Value int64
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NumberLiteral) String() string       { return nl.Token.Literal }

type StringLiteral struct {
	Token token.Token // token.STR
	Value string
}

func (nl *StringLiteral) expressionNode()      {}
func (nl *StringLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *StringLiteral) String() string       { return nl.Token.Literal }

type ListLiteral struct {
	Token    token.Token // token.RIGHT_SQUARE_BRACE
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal }
func (ll *ListLiteral) String() string {
	var out bytes.Buffer

	elements := make([]string, 0, len(ll.Elements))
	for i, el := range ll.Elements {
		elements[i] = el.String()
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type TupleLiteral struct {
	ListLiteral
}

func (tl *TupleLiteral) String() string {
	var out bytes.Buffer

	elements := make([]string, 0, len(tl.Elements))
	for i, exp := range tl.Elements {
		elements[i] = exp.String()
	}

	out.WriteString("(")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString(")")

	return out.String()
}

type ObjectLiteral struct {
	Token   token.Token // token.LEFT_CURLY_BRACE
	KVPairs map[Expression]Expression
}

func (ol *ObjectLiteral) expressionNode()      {}
func (ol *ObjectLiteral) TokenLiteral() string { return ol.Token.Literal }
func (ol *ObjectLiteral) String() string {
	var out bytes.Buffer

	elements := make([]string, 0, len(ol.KVPairs))
	for k, v := range ol.KVPairs {
		elements = append(elements, fmt.Sprintf("%s = %s", k.String(), v.String()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("}")

	return out.String()
}

type NullLiteral struct {
	Token token.Token // token.NULL
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return "null" }

type KeyValueStatement struct {
	Token token.Token // token.ASSIGN
	Name  *Identifier
	Value Expression
}

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

func (kv *KeyValueStatement) statementNode()       {}
func (kv *KeyValueStatement) TokenLiteral() string { return kv.Token.Literal }
func (kv *KeyValueStatement) String() string {
	var out bytes.Buffer
	out.WriteString(kv.Name.String() + " = " + kv.Value.String())
	return out.String()
}

type OptionalExpression struct {
	Token        token.Token // token.OPTIONAL
	TypeToken    token.Token // may be any Type Keyword token (e.g. token.LIST, token.NUMBER)
	DefaultValue Expression
}

func (os *OptionalExpression) expressionNode()      {}
func (os *OptionalExpression) TokenLiteral() string { return os.Token.Literal }
func (os *OptionalExpression) String() string {
	var out bytes.Buffer

	out.WriteString(os.TokenLiteral())
	out.WriteString("(")
	out.WriteString(os.TypeToken.Literal)

	if os.DefaultValue != nil {
		out.WriteString(os.DefaultValue.String())
	}

	out.WriteString(")")

	return out.String()
}

type BoolTypeLiteral struct {
	Token token.Token // token.BOOL
}

func (bt *BoolTypeLiteral) expressionNode()      {}
func (bt *BoolTypeLiteral) TokenLiteral() string { return bt.Token.Literal }
func (bt *BoolTypeLiteral) String() string       { return bt.Token.Literal }

type NumberTypeLiteral struct {
	Token token.Token // token.NUMBER
}

func (nt *NumberTypeLiteral) expressionNode()      {}
func (nt *NumberTypeLiteral) TokenLiteral() string { return nt.Token.Literal }
func (nt *NumberTypeLiteral) String() string       { return nt.Token.Literal }

type StringTypeLiteral struct {
	Token token.Token // token.STRING
}

func (st *StringTypeLiteral) expressionNode()      {}
func (st *StringTypeLiteral) TokenLiteral() string { return st.Token.Literal }
func (st *StringTypeLiteral) String() string       { return st.Token.Literal }

type ListTypeLiteral struct {
	Token          token.Token // token.LIST
	TypeExpression Expression
}

func (lt *ListTypeLiteral) expressionNode()      {}
func (lt *ListTypeLiteral) TokenLiteral() string { return lt.Token.Literal }
func (lt *ListTypeLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(lt.TokenLiteral())
	out.WriteString("(")
	out.WriteString(lt.TypeExpression.String())
	out.WriteString(")")

	return out.String()
}

type ObjectTypeLiteral struct {
	Token      token.Token // token.OBJECT
	ObjectSpec Expression
}

func (ot *ObjectTypeLiteral) expressionNode()      {}
func (ot *ObjectTypeLiteral) TokenLiteral() string { return ot.Token.Literal }
func (ot *ObjectTypeLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(ot.TokenLiteral())
	out.WriteString("(")
	out.WriteString(ot.ObjectSpec.String())
	out.WriteString(")")

	return out.String()
}

type MapTypeLiteral struct {
	Token          token.Token
	TypeExpression Expression
}

func (mt *MapTypeLiteral) expressionNode()      {}
func (mt *MapTypeLiteral) TokenLiteral() string { return mt.Token.Literal }
func (mt *MapTypeLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(mt.TokenLiteral())
	out.WriteString("(")
	out.WriteString(mt.TypeExpression.String())
	out.WriteString(")")

	return out.String()
}
