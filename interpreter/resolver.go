package interpreter

import (
	"github.com/cgrunewald/golox/interpreter/util"
)

type FunctionCallType int32

const (
	CALL_TYPE_NONE FunctionCallType = iota
	CALL_TYPE_FUNCTION
	CALL_TYPE_METHOD
	CALL_TYPE_INIT
)

type Resolver struct {
	scopes                  *util.Stack[map[string]bool]
	i                       *Interpreter
	errs                    []error
	currentFunctionCallType FunctionCallType
}

func NewResolver(i *Interpreter) *Resolver {
	return &Resolver{scopes: util.NewStack[map[string]bool](), i: i, errs: make([]error, 0), currentFunctionCallType: CALL_TYPE_NONE}
}

func (r *Resolver) define(name string) {
	if r.scopes.IsEmpty() {
		return
	}

	r.scopes.Peek()[name] = true
}

func (r *Resolver) declare(name Token) {
	if r.scopes.IsEmpty() {
		return
	}

	_, exists := r.scopes.Peek()[name.Lexeme]
	if exists {
		r.errs = append(r.errs, name.ToRuntimeError(E_VAR_ALREADY_DEFINED, "Already a variable with this name"))
	}

	r.scopes.Peek()[name.Lexeme] = false
}

func (r *Resolver) resolveLocal(expr Expr, name Token) {
	r.scopes.ForEach(func(i int, val map[string]bool) bool {
		if _, exists := val[name.Lexeme]; exists {
			r.i.resolve(expr, r.scopes.Length()-1-i)
			return false
		}
		return true
	})
}

func (r *Resolver) pushScope() {
	r.scopes.Push(make(map[string]bool))
}

func (r *Resolver) popScope() {
	r.scopes.Pop()
}

func (r *Resolver) ResolveExpr(expr Expr) {
	expr.Accept(r)
}

func (r *Resolver) ResolveStmts(stmts []Stmt) {
	for _, stmt := range stmts {
		stmt.Accept(r)
	}
}

func (r *Resolver) ResolveStmt(stmt Stmt) {
	stmt.Accept(r)
}

func (r *Resolver) VisitBinary(expr *Binary) interface{} {
	r.ResolveExpr(expr.Left)
	r.ResolveExpr(expr.Right)

	return nil
}

func (r *Resolver) VisitGrouping(expr *Grouping) interface{} {
	r.ResolveExpr(expr.Expression)

	return nil
}

func (r *Resolver) VisitLiteral(expr *Literal) interface{} {
	return nil
}

func (r *Resolver) VisitUnary(expr *Unary) interface{} {
	r.ResolveExpr(expr.Right)

	return nil
}

func (r *Resolver) VisitTernaryCondition(expr *TernaryCondition) interface{} {
	r.ResolveExpr(expr.Condition)
	r.ResolveExpr(expr.TrueBranch)
	r.ResolveExpr(expr.FalseBranch)

	return nil
}

func (r *Resolver) VisitVariable(expr *Variable) interface{} {
	if expr.Name.TokenType == TK_THIS && r.currentFunctionCallType != CALL_TYPE_METHOD && r.currentFunctionCallType != CALL_TYPE_INIT {
		r.errs = append(r.errs, expr.Name.ToRuntimeError(E_UNDEFINED_VARIABLE, "Cannot reference 'this' outside of a method"))
	}

	if !r.scopes.IsEmpty() {
		if val, ok := r.scopes.Peek()[expr.Name.Lexeme]; ok && !val {
			r.errs = append(r.errs, expr.Name.ToError("Can't read local variable in its own initializer"))
		}
	}

	r.resolveLocal(expr, expr.Name)
	return nil
}

func (r *Resolver) VisitLogical(expr *Logical) interface{} {
	r.ResolveExpr(expr.Left)
	r.ResolveExpr(expr.Right)

	return nil
}

func (r *Resolver) VisitVarStmt(stmt *VarStmt) interface{} {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		r.ResolveExpr(stmt.Initializer)
	}
	r.define(stmt.Name.Lexeme)
	return nil
}

func (r *Resolver) VisitFunctionStmt(stmt *FunctionStmt) interface{} {
	r.declare(stmt.Name)
	r.define(stmt.Name.Lexeme)

	r.resolveFunction(stmt.Params, stmt.Body, CALL_TYPE_FUNCTION)

	return nil
}

var ThisToken = Token{TokenType: TK_THIS, Lexeme: "this", Literal: nil, Line: 0}

func (r *Resolver) resolveMethod(params []Token, body []Stmt, callType FunctionCallType) {
	r.pushScope()

	r.declare(ThisToken)
	r.define("this")

	r.resolveFunction(params, body, callType)

	r.popScope()
}

func (r *Resolver) resolveFunction(params []Token, body []Stmt, callType FunctionCallType) {
	enclosingFunction := r.currentFunctionCallType
	r.currentFunctionCallType = callType

	r.pushScope()

	for _, param := range params {
		r.declare(param)
		r.define(param.Lexeme)
	}

	r.ResolveStmts(body)

	r.popScope()

	r.currentFunctionCallType = enclosingFunction
}

func (r *Resolver) VisitLambda(expr *Lambda) interface{} {
	r.resolveFunction(expr.Params, expr.Body, CALL_TYPE_FUNCTION)

	return nil
}

func (r *Resolver) VisitExprStmt(stmt *ExprStmt) interface{} {
	r.ResolveExpr(stmt.Expression)

	return nil
}

func (r *Resolver) VisitPrintStmt(stmt *PrintStmt) interface{} {
	r.ResolveExpr(stmt.Expression)

	return nil
}

func (r *Resolver) VisitBlockStmt(stmt *BlockStmt) interface{} {
	r.pushScope()
	r.ResolveStmts(stmt.Statements)
	r.popScope()

	return nil
}

func (r *Resolver) VisitAssign(expr *Assign) interface{} {
	r.ResolveExpr(expr.Value)
	r.resolveLocal(expr, expr.Name)

	return nil
}

func (r *Resolver) VisitIfStmt(stmt *IfStmt) interface{} {
	r.ResolveExpr(stmt.Condition)
	r.ResolveStmt(stmt.ThenBranch)
	if stmt.ElseBranch != nil {
		r.ResolveStmt(stmt.ThenBranch)
	}

	return nil
}

func (r *Resolver) VisitWhileStmt(stmt *WhileStmt) interface{} {
	r.ResolveExpr(stmt.Condition)
	r.ResolveStmt(stmt.Body)

	return nil
}

func (r *Resolver) VisitCall(expr *Call) interface{} {
	r.ResolveExpr(expr.Callee)

	for _, arg := range expr.Arguments {
		r.ResolveExpr(arg)
	}

	return nil
}

func (r *Resolver) VisitGet(expr *Get) interface{} {
	r.ResolveExpr(expr.Object)
	return nil
}

func (r *Resolver) VisitSet(expr *Set) interface{} {
	r.ResolveExpr(expr.Object)
	r.ResolveExpr(expr.Value)
	return nil
}

func (r *Resolver) VisitReturnStmt(stmt *ReturnStmt) interface{} {
	if r.currentFunctionCallType == CALL_TYPE_NONE {
		return stmt.Keyword.ToRuntimeError(E_UNEXPECTED_RETURN, "Unexpected return in global scope")
	}

	if stmt.Expression != nil {
		r.ResolveExpr(stmt.Expression)

		if r.currentFunctionCallType == CALL_TYPE_INIT {
			if v, ok := stmt.Expression.(*Variable); !ok || v.Name.TokenType != TK_THIS {
				r.errs = append(r.errs, stmt.Keyword.ToRuntimeError(E_UNEXPECTED_RETURN, "Unexpected return expression in `init`"))
			}
		}
	}

	return nil
}

func (r *Resolver) VisitClassStmt(stmt *ClassStmt) interface{} {
	r.declare(stmt.Name)
	r.define(stmt.Name.Lexeme)

	for _, m := range stmt.Methods {
		callType := CALL_TYPE_METHOD
		if m.Name.Lexeme == "init" {
			callType = CALL_TYPE_INIT
		}
		r.resolveMethod(m.Params, m.Body, callType)
	}
	return nil
}

func (r *Resolver) HasError() bool {
	return len(r.errs) > 0
}

func (r *Resolver) Errors() []error {
	return r.errs
}
