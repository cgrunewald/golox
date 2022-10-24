package interpreter

type Environment struct {
	Enclosing *Environment
	Values    map[string]interface{}
}

func NewEnvironment() *Environment {
	return &Environment{nil, make(map[string]interface{})}
}

func NewEnclosedEnvironment(enclosing *Environment) *Environment {
	return &Environment{enclosing, make(map[string]interface{})}
}

func (e *Environment) Define(name Token, value interface{}) {
	e.Values[name.Lexeme] = value
}

func (e *Environment) Get(name Token) (interface{}, error) {
	if value, ok := e.Values[name.Lexeme]; ok {
		return value, nil
	}

	if e.Enclosing != nil {
		return e.Enclosing.Get(name)
	}

	return nil, NewRuntimeError(E_UNDEFINED_VARIABLE, name.Line, name.Lexeme, "Undefined variable")
}

func (e *Environment) Set(name Token, value interface{}) error {
	if _, ok := e.Values[name.Lexeme]; ok {
		e.Values[name.Lexeme] = value
		return nil
	}

	if e.Enclosing != nil {
		return e.Enclosing.Set(name, value)
	}

	return NewRuntimeError(E_UNDEFINED_VARIABLE, name.Line, name.Lexeme, "Undefined variable")
}
