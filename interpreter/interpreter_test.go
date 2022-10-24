package interpreter

import "testing"

func doTest(t *testing.T, expression string, expected interface{}, expectedErr int32) {
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

	interpreter := NewInterpreter(InterpreterConfig{})
	result, err := interpreter.InterpretExpr(expr)
	if err != nil {
		if expectedErr != E_NO_ERROR {
			IfLoxError(err, func(err *LoxError) {
				if err.runtimeErrorType != expectedErr {
					t.Errorf("expected error type %d, got %d", expectedErr, err.runtimeErrorType)
				}
			})
		} else {
			t.Errorf("interpreter error: %v", err)
		}
		return
	}

	if err == nil && expectedErr != E_NO_ERROR {
		t.Errorf("expected error %d, got %v", expectedErr, result)
	}

	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
		return
	}
}

func TestInterpretExpressions(t *testing.T) {
	tests := []struct {
		expression       string
		expected         interface{}
		runtimeErrorType int32
	}{
		{"1 == 1 ? 4 + 4 * 3 : false", 16.0, E_NO_ERROR},
		{"\"ab\" + \"cd\"", "abcd", E_NO_ERROR},
		{"5 + \"cd\"", "5cd", E_NO_ERROR},
		{"-4", -4.0, E_NO_ERROR},
		{"!!true", true, E_NO_ERROR},
		{"4 <= 3", false, E_NO_ERROR},
		{"4 > 3", true, E_NO_ERROR},
		{"\"a\" < \"b\"", true, E_NO_ERROR},
		{"5 / 0", nil, E_DIVIDE_BY_ZERO},
	}

	for _, test := range tests {
		doTest(t, test.expression, test.expected, test.runtimeErrorType)
	}
}

func TestRuntimeErrors(t *testing.T) {

}

func TestPrograms(t *testing.T) {
	tests := []struct {
		program        string
		expectedOutput []string
	}{
		{
			`
			var a = 1;
			a = 3;
			print a;
			`,
			[]string{"3"},
		},
		{
			`
			var a = 1;
			{
				var a = a * 3;
				print a;
			}
			print a;
			`,
			[]string{"3", "1"},
		},
	}

	for _, test := range tests {
		doProgramTest(t, test.program, test.expectedOutput)
	}
}

func doProgramTest(t *testing.T, program string, expectedOutput []string) {
	output := make([]string, 0)
	config := InterpreterConfig{
		PrintFunc: func(value string) {
			output = append(output, value)
		},
	}

	errs := RunProgram(config, program)
	if len(errs) > 0 {
		t.Errorf("unexpected error: %v", errs)
		return
	}

	if len(output) != len(expectedOutput) {
		t.Errorf("expected %d output lines, got %d", len(expectedOutput), len(output))
		return
	}

	for i, line := range output {
		if line != expectedOutput[i] {
			t.Errorf("expected output line '%d' to be '%q', got '%q'", i, expectedOutput[i], line)
		}
	}
}
