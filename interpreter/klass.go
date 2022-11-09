package interpreter

import "fmt"

type Klass struct {
	name    Token
	methods map[string]*FunctionStmt
	env     *Environment
	super   *Klass
}

func NewKlass(name Token, methods []*FunctionStmt, env *Environment, super *Klass) *Klass {
	methodMap := make(map[string]*FunctionStmt)
	for _, method := range methods {
		methodMap[method.Name.Lexeme] = method
	}

	if super != nil {
		env = NewEnclosedEnvironment(env)
		env.Define("super", super)
	}

	return &Klass{name: name, methods: methodMap, env: env, super: super}
}

func (k *Klass) Arity() int {
	if m, ok := k.methods["init"]; ok {
		return len(m.Params)
	}

	return 0
}

func (k *Klass) Call(i *Interpreter, arguments []interface{}) interface{} {
	instance := NewInstance(k)

	klass, init := k.FindMethod("init")
	if init != nil {
		method := instance.bind("init", init, klass)
		method.Call(i, arguments)
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

type Gettable interface {
	Get(property string) (interface{}, bool)
}

func NewInstance(klass *Klass) *KlassInstance {
	return &KlassInstance{klass, make(map[string]interface{})}
}

func (i *KlassInstance) String() string {
	return fmt.Sprintf("%v instance", i.klass)
}

func (k *Klass) GetSuperMethod(method Token, instance *KlassInstance) (interface{}, error) {
	klass, methodDef := k.FindMethod(method.Lexeme)
	if methodDef == nil {
		return nil, NewRuntimeError(E_UNDEFINED_OBJECT_PROPERTY, method.Line, method.Lexeme, "Method does not exist on super")
	}

	// Bind and cache the binding
	boundMethod := instance.bind(method.Lexeme, methodDef, klass)
	return boundMethod, nil

}

func (k *Klass) FindMethod(property string) (*Klass, *FunctionStmt) {
	klass := k
	for klass != nil {
		method, ok := klass.methods[property]
		if ok {
			return klass, method
		}

		klass = klass.super
	}

	return nil, nil
}

func (i *KlassInstance) Get(property string) (interface{}, bool) {
	val, ok := i.properties[property]
	if ok {
		return val, ok
	}

	klass, method := i.klass.FindMethod(property)
	if method == nil {
		return nil, false
	}

	// Bind and cache the binding
	boundMethod := i.bind(property, method, klass)
	return boundMethod, true
}

func (i *KlassInstance) Set(property string, value interface{}) {
	i.properties[property] = value
}

func (i *KlassInstance) bind(property string, f *FunctionStmt, klass *Klass) Callable {
	methodEnv := NewEnclosedEnvironment(klass.env)
	methodEnv.Define("this", i)

	if property == "init" {
		return NewInitFunctionCallable(f, methodEnv)
	}
	return NewFunctionCallable(f, methodEnv)
}
