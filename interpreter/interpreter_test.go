package interpreter

import (
	"errors"
	"testing"
)

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
		{"true and true", true, E_NO_ERROR},
		{"true and false", false, E_NO_ERROR},
		{"false and false", false, E_NO_ERROR},
		{"false or false", false, E_NO_ERROR},
		{"true and false", false, E_NO_ERROR},
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
				var b = a;
				var a = b * 3;
				print a;
			}
			print a;
			`,
			[]string{"3", "1"},
		},
		{
			`
			if (true) {
				print 1;
			}	else {
				print 2;
			}

			if (false) {
				print 2;
			} else {
				print 1;
			}

			if (false) print 1; if (false) print 2; else print 3;
			`,
			[]string{"1", "1", "3"},
		},
		{
			`
			for (var i = 0; i < 3; i = i + 1) print i;
			`,
			[]string{"0", "1", "2"},
		},
		{
			`
			var i = 3;
			while (i >= 0) {
				print i;
				i = i - 1;
			}
			`,
			[]string{"3", "2", "1", "0"},
		},
		{
			`
			print clock();
			print clock();
			`,
			[]string{"1", "2"},
		},
		{
			`
			fun foo() {
				print "bar";
			}
			foo();
			foo();

			fun bar(a, b) {
				print a + b;
			}
			bar("fizz", "buzz");
			`,
			[]string{"bar", "bar", "fizzbuzz"},
		},
		{
			`
			fun count(n) {
				if (n > 1) count(n - 1);
				print n;
			}
			count(4);
			`,
			[]string{"1", "2", "3", "4"},
		},
		{
			`
			fun mult2(n) {
				return 2 * n;
			}
			print mult2(2);

			fun nested() {
				var a = 1; 
				{
					var a = 2;
					{
						var a = 3;
						return a;
					}
				}
			}
			print nested();

			fun ifThen(b) {
				if (b) return 1; else return 2;
			}
			print ifThen(true);

			fun whileLoop(b) {
				while (b) {
					return 5;
				}
				return 6;
			}

			print whileLoop(true);
			print whileLoop(false);
			`,
			[]string{"4", "3", "1", "5", "6"},
		},
		{`
			fun createIncrementer() {
				var i = 0;
				fun incr() {
					i = i + 1;
					return i;
				}
				return incr;
			}

			var i1 = createIncrementer();
			var i2 = createIncrementer();

			print i1();
			print i1();
			print i2();
			print i2();
			`,
			[]string{"1", "2", "1", "2"},
		},
		{`
			fun testFor() {
				var first = nil;
				for (var i = 0; i < 2; i = i + 1) {
					fun ret() {
						return i;
					}
					if (first == nil) {
						first = ret;
					}
				}
				return first;
			}

			print testFor()();

			fun testFor2() {
				var first = nil;
				for (var i = 0; i < 2; i = i + 1) {
					var j = i;
					fun ret() {
						return j;
					}
					if (first == nil) {
						first = ret;
					}
				}
				return first;
			}
			print testFor2()();
			`,
			[]string{"2", "0"},
		},
		{
			`
				fun compose(f, g) {
					return fun (a) {
						return f(g(a));
					};
				}

				var f = compose(fun (a) { return a * 2; }, fun (b) { return b * 2; });
				print f(2);
			`,
			[]string{"8"},
		},
		{
			`
				var a = "global";
				{
					fun printA() {
						print a;
					}

					printA();
					var a = "local";

					printA();
				}
			`,
			[]string{"global", "global"},
		},
	}

	for _, test := range tests {
		doProgramTest(t, test.program, test.expectedOutput, []int32{})
	}
}

func TestBadPrograms(t *testing.T) {
	tests := []struct {
		program        string
		expectedOutput []string
		expectedErrors []int32
	}{
		{
			`
			fun bad() {
				var a = "a";
				var a = "b";
			}
			`,
			[]string{},
			[]int32{E_VAR_ALREADY_DEFINED},
		},
		{
			`
			return "invalid";
			`,
			[]string{},
			[]int32{E_UNEXPECTED_RETURN},
		},
	}
	for _, test := range tests {
		doProgramTest(t, test.program, test.expectedOutput, test.expectedErrors)
	}
}

func doProgramTest(t *testing.T, program string, expectedOutput []string, errorCount []int32) {
	output := make([]string, 0)
	clockIncr := 0
	config := InterpreterConfig{
		GlobalFuncOverrides: map[string]Callable{
			"clock": NewNativeCallable(0, func(i *Interpreter, arguments []interface{}) interface{} {
				clockIncr = clockIncr + 1
				return clockIncr
			}),
		},
		PrintFunc: func(value string) {
			output = append(output, value)
		},
	}

	errs := RunProgram(config, program)
	if len(errs) > 0 && len(errorCount) == 0 {
		t.Errorf("in program '%v':\nunexpected error(s):\n%v", program, errs)
		return
	}

	if len(errs) >= 0 && len(errorCount) > 0 {
		if len(errs) != len(errorCount) {
			t.Errorf("expected %d errors; got %d errors", len(errorCount), len(errs))
		} else {
			for i, err := range errorCount {
				var loxError *LoxError
				ok := errors.As(errs[i], &loxError)
				if ok {
					if loxError.runtimeErrorType != err {
						t.Errorf("Error[%d] expected type %d, got type %d", i, err, loxError.runtimeErrorType)
					}
				} else {
					t.Errorf("Expected a lox error, got %v", errs[i])
				}
			}
		}
	}

	if len(output) != len(expectedOutput) {
		t.Errorf("expected %d output lines, got %d", len(expectedOutput), len(output))
		return
	}

	for i, line := range output {
		if line != expectedOutput[i] {
			t.Errorf("expected output idx %d to be %q, got %q", i, expectedOutput[i], line)
		}
	}
}
