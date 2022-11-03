package token

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers
	IDENT = "IDENT"

	// Literals
	BOOL   = "BOOL"
	NUMBER = "NUMBER"
	STRING = "STRING"

	// Keywords
	OPTIONAL = "OPTIONAL"
	NULL     = "NULL"
	FALSE    = "FALSE"
	TRUE     = "TRUE"

	// Scalar type keywords
	BOOL_TYPE   = "BOOL_TYPE"
	NUMBER_TYPE = "NUMBER_TYPE"
	STRING_TYPE = "STRING_TYPE"

	// Collection type keywords
	LIST_TYPE   = "LIST_TYPE"
	TUPLE_TYPE  = "TUPLE_TYPE"
	MAP_TYPE    = "MAP_TYPE"
	OBJECT_TYPE = "OBJECT_TYPE"

	// Operators
	ASSIGN = "="

	// Delimiters
	COMMA = ","

	LEFT_PAREN         = "("
	RIGHT_PAREN        = ")"
	LEFT_CURLY_BRACE   = "{"
	RIGHT_CURLY_BRACE  = "}"
	LEFT_SQUARE_BRACE  = "["
	RIGHT_SQUARE_BRACE = "]"

	// Comments
	COMMENT = "COMMENT"
	HASH    = "#" // comment prefix
	SLASH   = "/" // comment prefix
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"bool":     BOOL_TYPE,
	"number":   NUMBER_TYPE,
	"string":   STRING_TYPE,
	"list":     LIST_TYPE,
	"tuple":    TUPLE_TYPE,
	"map":      MAP_TYPE,
	"object":   OBJECT_TYPE,
	"optional": OPTIONAL,
	"null":     NULL,
	"true":     TRUE,
	"false":    FALSE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
