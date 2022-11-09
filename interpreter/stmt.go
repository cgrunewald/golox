package interpreter

type Stmt interface {
  Accept(visitor StmtVisitor) interface{}
}

type StmtVisitor interface {
  VisitIfStmt(expr *IfStmt) interface{}
  VisitWhileStmt(expr *WhileStmt) interface{}
  VisitExprStmt(expr *ExprStmt) interface{}
  VisitPrintStmt(expr *PrintStmt) interface{}
  VisitVarStmt(expr *VarStmt) interface{}
  VisitFunctionStmt(expr *FunctionStmt) interface{}
  VisitClassStmt(expr *ClassStmt) interface{}
  VisitBlockStmt(expr *BlockStmt) interface{}
  VisitReturnStmt(expr *ReturnStmt) interface{}
}

type IfStmt struct {
  Expr
  Condition Expr
  ThenBranch Stmt
  ElseBranch Stmt
}

func (e *IfStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitIfStmt(e)
}

type WhileStmt struct {
  Expr
  Condition Expr
  Body Stmt
}

func (e *WhileStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitWhileStmt(e)
}

type ExprStmt struct {
  Expr
  Expression Expr
}

func (e *ExprStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitExprStmt(e)
}

type PrintStmt struct {
  Expr
  Expression Expr
}

func (e *PrintStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitPrintStmt(e)
}

type VarStmt struct {
  Expr
  Name Token
  Initializer Expr
}

func (e *VarStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitVarStmt(e)
}

type FunctionStmt struct {
  Expr
  Name Token
  Params []Token
  Body []Stmt
}

func (e *FunctionStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitFunctionStmt(e)
}

type ClassStmt struct {
  Expr
  Name Token
  SuperClass *Variable
  Methods []*FunctionStmt
}

func (e *ClassStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitClassStmt(e)
}

type BlockStmt struct {
  Expr
  Statements []Stmt
}

func (e *BlockStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitBlockStmt(e)
}

type ReturnStmt struct {
  Expr
  Keyword Token
  Expression Expr
}

func (e *ReturnStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitReturnStmt(e)
}


