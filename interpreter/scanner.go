package interpreter

import (
	"fmt"
	"strconv"
)

type Scanner struct {
	source []rune
	tokens []Token
	errors []error

	start   int
	current int
	line    int
}

func NewScanner(source string) *Scanner {
	return &Scanner{
		source: []rune(source),
		tokens: make([]Token, 0),
		errors: make([]error, 0),
		start:  0, current: 0, line: 1}
}

func (s *Scanner) Errors() []error {
	return s.errors
}

func (s *Scanner) HasError() bool {
	return len(s.errors) > 0
}

func (scanner *Scanner) ScanTokens() []Token {

	for !scanner.isAtEnd() {
		scanner.start = scanner.current
		scanner.scanToken()
	}

	scanner.tokens = append(scanner.tokens, Token{TK_EOF, "", nil, scanner.line})
	return scanner.tokens
}

func (scanner *Scanner) isAtEnd() bool {
	return scanner.current >= len(scanner.source)
}

func (scanner *Scanner) addToken(tokType TokenType, literal interface{}) {
	scanner.tokens = append(scanner.tokens, NewToken(tokType, string(scanner.source[scanner.start:scanner.current]), literal, scanner.line))
}

func (scanner *Scanner) scanToken() {
	c := scanner.advance()
	switch c {
	case "(":
		scanner.addToken(TK_LEFT_PAREN, nil)
		break
	case ")":
		scanner.addToken(TK_RIGHT_PAREN, nil)
		break
	case "{":
		scanner.addToken(TK_LEFT_BRACE, nil)
		break
	case "}":
		scanner.addToken(TK_RIGHT_BRACE, nil)
		break
	case ",":
		scanner.addToken(TK_COMMA, nil)
		break
	case ".":
		scanner.addToken(TK_DOT, nil)
		break
	case "+":
		scanner.addToken(TK_PLUS, nil)
		break
	case "-":
		scanner.addToken(TK_MINUS, nil)
		break
	case ";":
		scanner.addToken(TK_SEMICOLON, nil)
		break
	case ":":
		scanner.addToken(TK_COLON, nil)
		break
	case "?":
		scanner.addToken(TK_QUESTION, nil)
		break
	case "/":
		if scanner.match("/") {
			for scanner.peek() != "\n" && !scanner.isAtEnd() {
				scanner.advance()
			}
		} else {
			scanner.addToken(TK_SLASH, nil)
		}
		break
	case "*":
		scanner.addToken(TK_STAR, nil)
		break
	case "!":
		if scanner.match("=") {
			scanner.addToken(TK_BANG_EQUAL, nil)
		} else {
			scanner.addToken(TK_BANG, nil)
		}
		break
	case "=":
		if scanner.match("=") {
			scanner.addToken(TK_EQUAL_EQUAL, nil)
		} else {
			scanner.addToken(TK_EQUAL, nil)
		}
		break
	case ">":
		if scanner.match("=") {
			scanner.addToken(TK_GREATER_EQUAL, nil)
		} else {
			scanner.addToken(TK_GREATER, nil)
		}
	case "<":
		if scanner.match("=") {
			scanner.addToken(TK_LESS_EQUAL, nil)
		} else {
			scanner.addToken(TK_LESS, nil)
		}
		break
	case "\"":
		scanner.string()
		break
	case " ", "\r", "\t":
		break
	case "\n":
		scanner.line++
		break
	default:
		if isDigit(c) {
			scanner.number()
		} else if isAlpha(c) {
			scanner.identifier()
		} else {
			scanner.errors = append(scanner.errors, NewError(scanner.line, fmt.Sprintf("Unexpected character '%s'", c)))
		}
	}
}

func (scanner *Scanner) identifier() {
	for isAlphaNumeric(scanner.peek()) {
		scanner.advance()
	}

	text := string(scanner.source[scanner.start:scanner.current])

	tokType, ok := TokenTypeKeywords[text]
	if !ok {
		scanner.addToken(TK_IDENTIFIER, text)
	} else {
		scanner.addToken(tokType, nil)
	}

}

func (scanner *Scanner) number() {
	for isDigit(scanner.peek()) {
		scanner.advance()
	}

	if scanner.peek() == "." && isDigit(scanner.peekNext()) {
		scanner.advance()

		for isDigit(scanner.peek()) {
			scanner.advance()
		}
	}

	value, err := strconv.ParseFloat(string(scanner.source[scanner.start:scanner.current]), 64)
	if err != nil {
		scanner.errors = append(
			scanner.errors,
			NewError(
				scanner.line,
				fmt.Sprintf("Invalid number '%s'.", string(scanner.source[scanner.start:scanner.current]))))
	} else {
		scanner.addToken(TK_NUMBER, value)
	}
}

func (scanner *Scanner) peekNext() string {
	if scanner.current+1 >= len(scanner.source) {
		return ""
	}

	return string(scanner.source[scanner.current+1])
}

func (scanner *Scanner) string() {
	for scanner.peek() != "\"" && !scanner.isAtEnd() {
		if scanner.peek() == "\n" {
			scanner.line++
		}

		scanner.advance()
	}

	if scanner.isAtEnd() {
		scanner.errors = append(scanner.errors, NewError(scanner.line, fmt.Sprintf("Unterminated string '%s'.", string(scanner.source[scanner.start:scanner.current]))))
		return
	}

	scanner.advance() // consume last "

	value := string(scanner.source[scanner.start+1 : scanner.current-1])
	scanner.addToken(TK_STRING, value)
}

func isDigit(c string) bool {
	return c >= "0" && c <= "9"
}

func isAlpha(c string) bool {
	return (c >= "a" && c <= "z") || (c >= "A" && c <= "Z") || c == "_"
}

func isAlphaNumeric(c string) bool {
	return isAlpha(c) || isDigit(c)
}

func (scanner *Scanner) peek() string {
	if scanner.isAtEnd() {
		return ""
	}

	return string(scanner.source[scanner.current])
}

func (scanner *Scanner) match(expected string) bool {
	if scanner.isAtEnd() {
		return false
	}

	if string(scanner.source[scanner.current]) != expected {
		return false
	}

	scanner.current++
	return true
}

func (scanner *Scanner) advance() string {
	current := scanner.current
	scanner.current++

	return string(scanner.source[current])
}
