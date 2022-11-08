package interpreter

type Expr interface {
  Accept(visitor ExprVisitor) interface{}
}

type ExprVisitor interface {
  VisitBinary(expr *Binary) interface{}
  VisitLogical(expr *Logical) interface{}
  VisitGrouping(expr *Grouping) interface{}
  VisitLiteral(expr *Literal) interface{}
  VisitUnary(expr *Unary) interface{}
  VisitTernaryCondition(expr *TernaryCondition) interface{}
  VisitAssign(expr *Assign) interface{}
  VisitVariable(expr *Variable) interface{}
  VisitCall(expr *Call) interface{}
  VisitGet(expr *Get) interface{}
  VisitSet(expr *Set) interface{}
  VisitLambda(expr *Lambda) interface{}
}

type Binary struct {
  Expr
  Left Expr
  Operator Token
  Right Expr
}

func (e *Binary) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitBinary(e)
}

type Logical struct {
  Expr
  Left Expr
  Operator Token
  Right Expr
}

func (e *Logical) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitLogical(e)
}

type Grouping struct {
  Expr
  Expression Expr
}

func (e *Grouping) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitGrouping(e)
}

type Literal struct {
  Expr
  Value interface{}
}

func (e *Literal) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitLiteral(e)
}

type Unary struct {
  Expr
  Operator Token
  Right Expr
}

func (e *Unary) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitUnary(e)
}

type TernaryCondition struct {
  Expr
  Condition Expr
  TrueBranch Expr
  FalseBranch Expr
}

func (e *TernaryCondition) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitTernaryCondition(e)
}

type Assign struct {
  Expr
  Name Token
  Value Expr
}

func (e *Assign) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitAssign(e)
}

type Variable struct {
  Expr
  Name Token
}

func (e *Variable) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitVariable(e)
}

type Call struct {
  Expr
  Callee Expr
  Paren Token
  Arguments []Expr
}

func (e *Call) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitCall(e)
}

type Get struct {
  Expr
  Object Expr
  Name Token
}

func (e *Get) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitGet(e)
}

type Set struct {
  Expr
  Object Expr
  Name Token
  Value Expr
}

func (e *Set) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitSet(e)
}

type Lambda struct {
  Expr
  Name Token
  Params []Token
  Body []Stmt
}

func (e *Lambda) Accept(visitor ExprVisitor) interface{} {
  return visitor.VisitLambda(e)
}


