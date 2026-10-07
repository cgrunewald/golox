package interpreter

import "testing"

func resolveProgram(t *testing.T, program string) *Resolver {
	scanner := NewScanner(program)
	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		t.Fatalf("scanner error: %v", scanner.Errors())
	}

	parser := NewParser(tokens)
	stmts := parser.Parse()
	if parser.HasError() {
		t.Fatalf("parser error: %v", parser.Errors())
	}

	resolver := NewResolver(NewInterpreter(InterpreterConfig{}))
	resolver.ResolveStmts(stmts)
	return resolver
}

func TestResolverReportsTopLevelReturn(t *testing.T) {
	resolver := resolveProgram(t, `return "invalid";`)

	if len(resolver.Errors()) != 1 {
		t.Fatalf("expected 1 resolver error, got %v", resolver.Errors())
	}

	IfLoxError(resolver.Errors()[0], func(err *LoxError) {
		if err.runtimeErrorType != E_UNEXPECTED_RETURN {
			t.Errorf("expected error type %d, got %d", E_UNEXPECTED_RETURN, err.runtimeErrorType)
		}
	})
}
