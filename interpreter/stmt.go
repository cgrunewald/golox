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
  VisitBlockStmt(expr *BlockStmt) interface{}
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

type BlockStmt struct {
  Expr
  Statements []Stmt
}

func (e *BlockStmt) Accept(visitor StmtVisitor) interface{} {
  return visitor.VisitBlockStmt(e)
}


