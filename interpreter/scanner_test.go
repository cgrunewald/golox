package interpreter

import "testing"

func AssertScansEqual(t *testing.T, expected []Token, actual []Token) {
	if len(expected) != len(actual) {
		t.Fatalf("Length of tokens is not the same")
	}

	for i := 0; i < len(expected); i++ {
		if !expected[i].Equals(actual[i]) {
			t.Fatalf("Expected token %v to equal %v", actual[i], expected[i])
		}
	}
}

func TestScanSimple(t *testing.T) {
	scanner := NewScanner("();,.+-*/!<>=")
	expected := []Token{
		NewToken(TK_LEFT_PAREN, "(", nil, 1),
		NewToken(TK_RIGHT_PAREN, ")", nil, 1),
		NewToken(TK_SEMICOLON, ";", nil, 1),
		NewToken(TK_COMMA, ",", nil, 1),
		NewToken(TK_DOT, ".", nil, 1),
		NewToken(TK_PLUS, "+", nil, 1),
		NewToken(TK_MINUS, "-", nil, 1),
		NewToken(TK_STAR, "*", nil, 1),
		NewToken(TK_SLASH, "/", nil, 1),
		NewToken(TK_BANG, "!", nil, 1),
		NewToken(TK_LESS, "<", nil, 1),
		NewToken(TK_GREATER_EQUAL, ">=", nil, 1),
		NewToken(TK_EOF, "", nil, 1),
	}

	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		t.Fatalf("Encountered error: %v", scanner.Errors())
	}

	AssertScansEqual(t, expected, tokens)
}

func TestScanDigit(t *testing.T) {
	scanner := NewScanner("12.34 == 12.34")
	expected := []Token{
		NewToken(TK_NUMBER, "12.34", 12.34, 1),
		NewToken(TK_EQUAL_EQUAL, "==", nil, 1),
		NewToken(TK_NUMBER, "12.34", 12.34, 1),
		NewToken(TK_EOF, "", nil, 1),
	}

	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		t.Fatalf("Encountered error: %v", scanner.Errors())
	}

	AssertScansEqual(t, expected, tokens)
}

func TestScanIdentifier(t *testing.T) {
	scanner := NewScanner("fun a(b, c)\n{}")
	expected := []Token{
		NewToken(TK_FUN, "fun", nil, 1),
		NewToken(TK_IDENTIFIER, "a", "a", 1),
		NewToken(TK_LEFT_PAREN, "(", nil, 1),
		NewToken(TK_IDENTIFIER, "b", "b", 1),
		NewToken(TK_COMMA, ",", nil, 1),
		NewToken(TK_IDENTIFIER, "c", "c", 1),
		NewToken(TK_RIGHT_PAREN, ")", nil, 1),
		NewToken(TK_LEFT_BRACE, "{", nil, 2),
		NewToken(TK_RIGHT_BRACE, "}", nil, 2),
		NewToken(TK_EOF, "", nil, 2),
	}

	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		t.Fatalf("Encountered error: %v", scanner.Errors())
	}

	AssertScansEqual(t, expected, tokens)
}

func TestScanBadString(t *testing.T) {
	scanner := NewScanner("\"test")
	expected := []Token{
		NewToken(TK_EOF, "", nil, 1),
	}

	tokens := scanner.ScanTokens()
	if !scanner.HasError() && len(scanner.Errors()) != 1 {
		t.Fatalf("Expected 1 error")
	}

	AssertScansEqual(t, expected, tokens)
}

func TestScanBadChar(t *testing.T) {
	scanner := NewScanner("#")
	expected := []Token{
		NewToken(TK_EOF, "", nil, 1),
	}

	tokens := scanner.ScanTokens()
	if !scanner.HasError() && len(scanner.Errors()) != 1 {
		t.Fatalf("Expected 1 error")
	}

	AssertScansEqual(t, expected, tokens)
}

func TestScanComment(t *testing.T) {
	scanner := NewScanner("// this is a comment\n+")
	expected := []Token{
		NewToken(TK_PLUS, "+", nil, 2),
		NewToken(TK_EOF, "", nil, 2),
	}

	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		t.Fatalf("Encountered error: %v", scanner.Errors())
	}

	AssertScansEqual(t, expected, tokens)
}
