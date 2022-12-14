package interpreter

func RunProgram(config InterpreterConfig, program string) []error {
	scanner := NewScanner(program)
	tokens := scanner.ScanTokens()
	if scanner.HasError() {
		return scanner.Errors()
	}

	parser := NewParser(tokens)
	stmts := parser.Parse()
	if parser.HasError() {
		return parser.Errors()
	}

	i := NewInterpreter(config)

	resolver := NewResolver(i)
	resolver.ResolveStmts(stmts)
	if resolver.HasError() {
		return resolver.Errors()
	}

	_, err := i.Interpret(stmts)
	if err != nil {
		return []error{err}
	}

	return nil
}
