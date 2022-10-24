package interpreter

import "fmt"

type TokenType int64

const (
	// Single character tokens
	TK_LEFT_PAREN TokenType = iota
	TK_RIGHT_PAREN
	TK_LEFT_BRACE
	TK_RIGHT_BRACE
	TK_COMMA
	TK_DOT
	TK_MINUS
	TK_PLUS
	TK_SEMICOLON
	TK_SLASH
	TK_STAR
	TK_QUESTION
	TK_COLON

	// One or two character tokens
	TK_BANG
	TK_BANG_EQUAL
	TK_EQUAL
	TK_EQUAL_EQUAL
	TK_GREATER
	TK_GREATER_EQUAL
	TK_LESS
	TK_LESS_EQUAL

	// Literals
	TK_IDENTIFIER
	TK_STRING
	TK_NUMBER

	// Keywords
	TK_AND
	TK_CLASS
	TK_ELSE
	TK_FALSE
	TK_FOR
	TK_FUN
	TK_IF
	TK_NIL
	TK_OR
	TK_PRINT
	TK_RETURN
	TK_SUPER
	TK_THIS
	TK_TRUE
	TK_VAR
	TK_WHILE

	TK_EOF
)

var TokenTypeKeywords = map[string]TokenType{
	"and":    TK_AND,
	"class":  TK_CLASS,
	"else":   TK_ELSE,
	"false":  TK_FALSE,
	"for":    TK_FOR,
	"fun":    TK_FUN,
	"if":     TK_IF,
	"nil":    TK_NIL,
	"or":     TK_OR,
	"print":  TK_PRINT,
	"return": TK_RETURN,
	"super":  TK_SUPER,
	"this":   TK_THIS,
	"true":   TK_TRUE,
	"var":    TK_VAR,
	"while":  TK_WHILE,
}

var TokenTypeNames = map[TokenType]string{
	TK_LEFT_PAREN:    "TK_LEFT_PAREN",
	TK_RIGHT_PAREN:   "TK_RIGHT_PAREN",
	TK_LEFT_BRACE:    "TK_LEFT_BRACE",
	TK_RIGHT_BRACE:   "TK_RIGHT_BRACE",
	TK_COMMA:         "TK_COMMA",
	TK_DOT:           "TK_DOT",
	TK_MINUS:         "TK_MINUS",
	TK_PLUS:          "TK_PLUS",
	TK_SEMICOLON:     "TK_SEMICOLON",
	TK_SLASH:         "TK_SLASH",
	TK_STAR:          "TK_STAR",
	TK_BANG:          "TK_BANG",
	TK_BANG_EQUAL:    "TK_BANG_EQUAL",
	TK_EQUAL:         "TK_EQUAL",
	TK_EQUAL_EQUAL:   "TK_EQUAL_EQUAL",
	TK_GREATER:       "TK_GREATER",
	TK_GREATER_EQUAL: "TK_GREATER_EQUAL",
	TK_LESS:          "TK_LESS",
	TK_LESS_EQUAL:    "TK_LESS_EQUAL",
	TK_IDENTIFIER:    "TK_IDENTIFIER",
	TK_STRING:        "TK_STRING",
	TK_NUMBER:        "TK_NUMBER",
	TK_AND:           "TK_AND",
	TK_CLASS:         "TK_CLASS",
	TK_ELSE:          "TK_ELSE",
	TK_FALSE:         "TK_FALSE",
	TK_FOR:           "TK_FOR",
	TK_FUN:           "TK_FUN",
	TK_IF:            "TK_IF",
	TK_NIL:           "TK_NIL",
	TK_OR:            "TK_OR",
	TK_PRINT:         "TK_PRINT",
	TK_RETURN:        "TK_RETURN",
	TK_SUPER:         "TK_SUPER",
	TK_THIS:          "TK_THIS",
	TK_TRUE:          "TK_TRUE",
	TK_VAR:           "TK_VAR",
	TK_WHILE:         "TK_WHILE",
	TK_EOF:           "TK_EOF",
	TK_QUESTION:      "TK_QUESTION",
	TK_COLON:         "TK_COLON",
}

type Token struct {
	TokenType TokenType
	Lexeme    string
	Literal   interface{}
	Line      int
}

func NewToken(tokenType TokenType, lexeme string, literal interface{}, line int) Token {
	return Token{tokenType, lexeme, literal, line}
}

func (token Token) String() string {
	return fmt.Sprintf("%s %s %v", TokenTypeNames[token.TokenType], token.Lexeme, token.Literal)
}

func (token Token) Equals(other Token) bool {
	return token.TokenType == other.TokenType &&
		token.Lexeme == other.Lexeme &&
		token.Literal == other.Literal &&
		token.Line == other.Line
}
