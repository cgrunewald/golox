package interpreter

import "fmt"

type Klass struct {
	name    Token
	methods map[string]*FunctionStmt
	env     *Environment
}

func NewKlass(name Token, methods []*FunctionStmt, env *Environment) *Klass {
	methodMap := make(map[string]*FunctionStmt)
	for _, method := range methods {
		methodMap[method.Name.Lexeme] = method
	}

	return &Klass{name: name, methods: methodMap, env: env}
}

func (k *Klass) Arity() int {
	if m, ok := k.methods["init"]; ok {
		return len(m.Params)
	}

	return 0
}

func (k *Klass) Call(i *Interpreter, arguments []interface{}) interface{} {
	instance := NewInstance(k)
	if _, ok := k.methods["init"]; ok {
		initMethod, _ := instance.Get("init")
		initMethod.(Callable).Call(i, arguments)
	}

	return instance
}

func (k *Klass) String() string {
	return k.name.Lexeme
}

type KlassInstance struct {
	klass      *Klass
	properties map[string]interface{}
}

func NewInstance(klass *Klass) *KlassInstance {
	return &KlassInstance{klass, make(map[string]interface{})}
}

func (i *KlassInstance) String() string {
	return fmt.Sprintf("%v instance", i.klass)
}

func (i *KlassInstance) Get(property string) (interface{}, bool) {
	val, ok := i.properties[property]
	if ok {
		return val, ok
	}

	method, ok := i.klass.methods[property]
	if !ok {
		return nil, ok
	}

	// Bind and cache the binding
	boundMethod := i.bind(property, method)
	i.Set(property, boundMethod)
	return boundMethod, true
}

func (i *KlassInstance) Set(property string, value interface{}) {
	i.properties[property] = value
}

func (i *KlassInstance) bind(property string, f *FunctionStmt) Callable {
	methodEnv := NewEnclosedEnvironment(i.klass.env)
	methodEnv.Define("this", i)

	if property == "init" {
		return NewInitFunctionCallable(f, methodEnv)
	}
	return NewFunctionCallable(f, methodEnv)
}
