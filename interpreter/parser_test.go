package interpreter

import "testing"

func runParse(t *testing.T, expression string, expected string) {
	scanner := NewScanner(expression)
	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		t.Errorf("scanner error: %v", scanner.Errors())
		return
	}

	parser := NewParser(tokens)
	expr := parser.ParseExpr()
	if parser.HasError() {
		t.Errorf("parser error: %v", parser.Errors())
		return
	}

	str := NewASTPrinter().Print(expr)

	if str != expected {
		t.Errorf("expected %s, got %s", expected, str)
		return
	}
}

func runParseStmt(t *testing.T, expression string, expected string) {
	scanner := NewScanner(expression)
	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		t.Errorf("scanner error: %v", scanner.Errors())
		return
	}

	parser := NewParser(tokens)
	stmt := parser.Parse()
	if parser.HasError() {
		t.Errorf("parser error: %v", parser.Errors())
		return
	}

	str := NewASTPrinter().PrintProgram(stmt)

	if str != expected {
		t.Errorf("expected %s, got %s", expected, str)
		return
	}
}

func runParseErrors(t *testing.T, expression string, expectedErrors int, expectedStmts int) {
	scanner := NewScanner(expression)
	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		t.Errorf("scanner error: %v", scanner.Errors())
		return
	}

	parser := NewParser(tokens)
	stmts := parser.Parse()
	if !parser.HasError() && expectedErrors > 0 {
		t.Errorf("expected error")
		return
	}

	if len(stmts) != expectedStmts {
		t.Errorf("expected %d statements, got %d", expectedStmts, len(stmts))
		return
	}

	if len(parser.Errors()) != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, len(parser.Errors()))
		return
	}
}

func TestParseExpressions(t *testing.T) {
	tests := []struct {
		expression string
		expected   string
	}{
		{"1 == 1 ? 4 + 4 * 3 : false", "(?: (== 1 1) (+ 4 (* 4 3)) false)"},
		{"3 + 4 + 5 * 6 * 7 + 1", "(+ (+ (+ 3 4) (* (* 5 6) 7)) 1)"},
		{"5 * (3 + 1)", "(* 5 (group (+ 3 1)))"},
		{"--4", "(- (- 4))"},
		{"true ? 1 : true ? 2 : true ? 3 : 4", "(?: true 1 (?: true 2 (?: true 3 4)))"},
		{"a == b", "(== (var a) (var b))"},
		{"a = 1", "(= (var a) 1)"},
		{"a = b = c = 1", "(= (var a) (= (var b) (= (var c) 1)))"},
		{"false or true", "(or false true)"},
		{"false and true", "(and false true)"},
		{"false or false or false and true", "(or (or false false) (and false true))"},
	}

	for _, test := range tests {
		runParse(t, test.expression, test.expected)
	}
}

func TestParseStatements(t *testing.T) {
	tests := []struct {
		expression string
		expected   string
	}{
		{"1;1 != 2;", "(scope 1 (!= 1 2))"},
		{"print 1;", "(scope (print 1))"},
		{"var a;", "(scope (def a))"},
		{"var a = \"test\";", "(scope (def a \"test\"))"},
		{"{ 1; {2 ;}}", "(scope (scope 1 (scope 2)))"},
	}

	for _, test := range tests {
		runParseStmt(t, test.expression, test.expected)
	}
}

func TestParseControlFlowStatements(t *testing.T) {
	tests := []struct {
		expression string
		expected   string
	}{
		{"if (true) 1; else 2;", "(scope (if true 1 2))"},
		{"if (3 + 3 > 1) {1; 2;} else {1; 2;}", "(scope (if (> (+ 3 3) 1) (scope 1 2) (scope 1 2)))"},
		{"while (true) {1; 2;}", "(scope (while true (scope 1 2)))"},
		{"while (true) 1;", "(scope (while true 1))"},
		{"for (;;) 1;", "(scope (while true 1))"},
		{"for (;;) {1;}", "(scope (while true (scope 1)))"},
		{"for (var i = 0; i < 10; i = i + 1) print i;", "(scope (scope (def i 0) (while (< (var i) 10) (scope (print (var i)) (= (var i) (+ (var i) 1))))))"},
		{"for (var i = 0; i < 10;) print i;", "(scope (scope (def i 0) (while (< (var i) 10) (print (var i)))))"},
	}

	for _, test := range tests {
		runParseStmt(t, test.expression, test.expected)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		expression     string
		expectedErrors int
		expectedStmts  int
	}{
		{"1;1 != 2;", 0, 2},
		{"=;1 != 2;", 1, 1},
		{"a=b; < != 2;print 3;", 1, 2},
	}

	for _, test := range tests {
		runParseErrors(t, test.expression, test.expectedErrors, test.expectedStmts)
	}
}
