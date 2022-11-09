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

func (e *Environment) Define(name string, value interface{}) {
	e.Values[name] = value
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

func (e *Environment) GetAt(name Token, distance int) (interface{}, error) {
	return e.ancestor(distance).Get(name)
}

func (e *Environment) ancestor(distance int) *Environment {
	env := e
	for i := 0; i < distance; i = i + 1 {
		env = env.Enclosing
	}
	return env
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

func (e *Environment) SetAt(name Token, value interface{}, distance int) error {
	return e.ancestor(distance).Set(name, value)
}
