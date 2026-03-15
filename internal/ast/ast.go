package ast

import "fmt"

// NodeType represents the type of an AST node.
type NodeType int

const (
	NodeBase = iota
	NodeIdentifier
	NodeNumberLiteral
	NodeStringLiteral
	NodeBoolLiteral
	NodeNullLiteral
	NodeBinaryExpr
	NodeUnaryExpr
	NodeCallExpr
	NodeMemberExpr
	NodeIndexExpr
	NodeVarDecl
	NodeFunctionDecl
	NodeBlock
	NodeIfStmt
	NodeWhileStmt
	NodeForStmt
	NodeReturnStmt
	NodeBreakStmt
	NodeContinueStmt
	NodeParameter
	NodeDictLiteral
	NodeArrayLiteral
	NodeLambda
	NodeNamedArgument
	NodeStructLiteral
	NodeYieldStmt
	NodeAwaitExpr
	NodeCastExpr
	NodeInclude
	NodeRawC
)

// Node is the interface implemented by all AST nodes.
type Node interface {
	GetType() NodeType
	GetLine() int
	GetColumn() int
	String() string
}

// ASTNode is an alias for Node for compatibility.
type ASTNode = Node

// BaseNode carries common metadata for nodes.
type BaseNode struct {
	Type   NodeType
	Line   int
	Column int
}

func (n BaseNode) GetType() NodeType { return n.Type }
func (n BaseNode) GetLine() int      { return n.Line }
func (n BaseNode) GetColumn() int    { return n.Column }

// IncludeNode represents a #include directive.
type IncludeNode struct {
	BaseNode
	Header   string
	IsSystem bool
	Filename string
}

func (n *IncludeNode) String() string {
	return fmt.Sprintf("Include{Header: %s, IsSystem: %v, Filename: %s}", n.Header, n.IsSystem, n.Filename)
}

// IdentifierNode represents an identifier.
type IdentifierNode struct {
	BaseNode
	Name         string
	ResolvedType string
	EmitName     string
}

func (n *IdentifierNode) String() string {
	return fmt.Sprintf("Identifier{Name: %s, ResolvedType: %s, EmitName: %s}", n.Name, n.ResolvedType, n.EmitName)
}

// NumberLiteralNode represents a number literal.
type NumberLiteralNode struct {
	BaseNode
	Value string
}

func (n *NumberLiteralNode) String() string {
	return fmt.Sprintf("NumberLiteral{Value: %s}", n.Value)
}

// StringLiteralNode represents a string literal.
type StringLiteralNode struct {
	BaseNode
	Value string
}

func (n *StringLiteralNode) String() string {
	return fmt.Sprintf("StringLiteral{Value: %s}", n.Value)
}

// BoolLiteralNode represents a boolean literal.
type BoolLiteralNode struct {
	BaseNode
	Value bool
}

func (n *BoolLiteralNode) String() string {
	return fmt.Sprintf("BoolLiteral{Value: %v}", n.Value)
}

// NullLiteralNode represents a null literal.
type NullLiteralNode struct {
	BaseNode
}

func (n *NullLiteralNode) String() string {
	return fmt.Sprintf("NullLiteral")
}

// BinaryExprNode represents a binary expression.
type BinaryExprNode struct {
	BaseNode
	Left          Node
	Operator      string
	Right         Node
	FoldedLiteral Node
}

func (n *BinaryExprNode) String() string {
	return fmt.Sprintf("BinaryExpr{Left: %v, Operator: %s, Right: %v, FoldedLiteral: %v}", n.Left, n.Operator, n.Right, n.FoldedLiteral)
}

// UnaryExprNode represents a unary expression.
type UnaryExprNode struct {
	BaseNode
	Operator  string
	Operand   Node
	IsPostfix bool
}

func (n *UnaryExprNode) String() string {
	return fmt.Sprintf("UnaryExpr{Operator: %s, Operand: %v, IsPostfix: %v}", n.Operator, n.Operand, n.IsPostfix)
}

// CastExprNode represents a C-style cast expression (type)expr.
type CastExprNode struct {
	BaseNode
	TargetType string // The target type to cast to
	Operand    Node   // The expression being cast
}

func (n *CastExprNode) String() string {
	return fmt.Sprintf("CastExpr{TargetType: %s, Operand: %v}", n.TargetType, n.Operand)
}

// CallExprNode represents a function call.
type CallExprNode struct {
	BaseNode
	Function  Node
	Args      []Node
	NamedArgs []*NamedArgumentNode
}

func (n *CallExprNode) String() string {
	return fmt.Sprintf("CallExpr{Function: %v, Args: %v, NamedArgs: %v}", n.Function, n.Args, n.NamedArgs)
}

// MemberExprNode represents a member access.
type MemberExprNode struct {
	BaseNode
	Object   Node
	Property string
}

func (n *MemberExprNode) String() string {
	return fmt.Sprintf("MemberExpr{Object: %v, Property: %s}", n.Object, n.Property)
}

// IndexExprNode represents an indexing operation.
type IndexExprNode struct {
	BaseNode
	Object Node
	Index  Node
}

func (n *IndexExprNode) String() string {
	return fmt.Sprintf("IndexExpr{Object: %v, Index: %v}", n.Object, n.Index)
}

// VarDeclNode represents a variable declaration.
type VarDeclNode struct {
	BaseNode
	Name        string
	Type        string
	Initializer Node
	IsConst     bool
	Module      string
	ArraySize   int // -1 for non-array, 0 for empty [], >0 for fixed size [N]
}

func (n *VarDeclNode) String() string {
	return fmt.Sprintf("VarDecl{Name: %s, Type: %s, Initializer: %v, IsConst: %v}", n.Name, n.Type, n.Initializer, n.IsConst)
}

// FunctionDeclNode represents a function declaration.
type FunctionDeclNode struct {
	BaseNode
	Name        string
	Parameters  []*ParameterNode
	ReturnType  string
	ReturnTypes []string
	Body        *BlockNode
	IsAsync     bool
	IsCoroutine bool
	Module      string
}

func (n *FunctionDeclNode) String() string {
	return fmt.Sprintf("FunctionDecl{Name: %s, Parameters: %v, ReturnType: %s, ReturnTypes: %v, Body: %v, IsAsync: %v, IsCoroutine: %v, Module: %s}", n.Name, n.Parameters, n.ReturnType, n.ReturnTypes, n.Body, n.IsAsync, n.IsCoroutine, n.Module)
}

// ParameterNode represents a function parameter.
type ParameterNode struct {
	BaseNode
	Type         string
	Name         string
	DefaultValue Node
}

func (n *ParameterNode) String() string {
	return fmt.Sprintf("Parameter{Type: %s, Name: %s, DefaultValue: %v}", n.Type, n.Name, n.DefaultValue)
}

// BlockNode represents a block of statements.
type BlockNode struct {
	BaseNode
	Statements []Node
}

func (n *BlockNode) String() string {
	return fmt.Sprintf("Block{Statements: %v}", n.Statements)
}

// IfStmtNode represents an if statement.
type IfStmtNode struct {
	BaseNode
	Condition  Node
	ThenBranch *BlockNode
	ElseBranch *BlockNode
}

func (n *IfStmtNode) String() string {
	return fmt.Sprintf("IfStmt{Condition: %v, Then: %v, Else: %v}", n.Condition, n.ThenBranch, n.ElseBranch)
}

// WhileStmtNode represents a while loop.
type WhileStmtNode struct {
	BaseNode
	Condition Node
	Body      *BlockNode
}

func (n *WhileStmtNode) String() string {
	return fmt.Sprintf("WhileStmt{Condition: %v, Body: %v}", n.Condition, n.Body)
}

// ForStmtNode represents a for loop.
type ForStmtNode struct {
	BaseNode
	Initializer Node
	Condition   Node
	Increment   Node
	Body        *BlockNode
}

func (n *ForStmtNode) String() string {
	return fmt.Sprintf("ForStmt{Initializer: %v, Condition: %v, Increment: %v, Body: %v}", n.Initializer, n.Condition, n.Increment, n.Body)
}

// ReturnStmtNode represents a return statement.
type ReturnStmtNode struct {
	BaseNode
	Value Node
}

func (n *ReturnStmtNode) String() string {
	return fmt.Sprintf("ReturnStmt{Value: %v}", n.Value)
}

// BreakStmtNode represents a break statement.
type BreakStmtNode struct {
	BaseNode
}

func (n *BreakStmtNode) String() string {
	return fmt.Sprintf("BreakStmt")
}

// ContinueStmtNode represents a continue statement.
type ContinueStmtNode struct {
	BaseNode
}

func (n *ContinueStmtNode) String() string {
	return fmt.Sprintf("ContinueStmt")
}

// DictLiteralNode represents a dictionary literal.
type DictLiteralNode struct {
	BaseNode
	Entries []DictEntry
}

func (n *DictLiteralNode) String() string {
	return fmt.Sprintf("DictLiteral{Entries: %v}", n.Entries)
}

// ArrayLiteralNode represents an array literal.
type ArrayLiteralNode struct {
	BaseNode
	Elements   []Node
	Dimensions int
	RowCount   int
	RowLength  int
}

func (n *ArrayLiteralNode) String() string {
	return fmt.Sprintf("ArrayLiteral{Elements: %v, Dimensions: %d, RowCount: %d, RowLength: %d}", n.Elements, n.Dimensions, n.RowCount, n.RowLength)
}

// StructLiteralNode represents a struct initializer like { field: value, ... }
type StructLiteralNode struct {
	BaseNode
	TypeName string            // The struct type name
	Fields   []StructFieldInit // Field initializers
}

// StructFieldInit represents a field initialization in a struct literal
type StructFieldInit struct {
	Name  string
	Value Node
}

func (n *StructLiteralNode) String() string {
	return fmt.Sprintf("StructLiteral{Type: %s, Fields: %d}", n.TypeName, len(n.Fields))
}

// LambdaNode represents a lambda expression.
type LambdaNode struct {
	BaseNode
	Captures             []string
	Parameters           []*ParameterNode
	ReturnType           string
	Body                 *BlockNode
	ResolvedCaptureTypes []string
}

func (n *LambdaNode) String() string {
	return fmt.Sprintf("Lambda{Captures: %v, Parameters: %v, ReturnType: %s, Body: %v, ResolvedCaptureTypes: %v}", n.Captures, n.Parameters, n.ReturnType, n.Body, n.ResolvedCaptureTypes)
}

// NamedArgumentNode represents a named argument.
type NamedArgumentNode struct {
	BaseNode
	Name  string
	Value Node
}

func (n *NamedArgumentNode) String() string {
	return fmt.Sprintf("NamedArgument{Name: %s, Value: %v}", n.Name, n.Value)
}

// YieldStmtNode represents a yield statement.
type YieldStmtNode struct {
	BaseNode
	Value Node
}

func (n *YieldStmtNode) String() string {
	return fmt.Sprintf("YieldStmt{Value: %v}", n.Value)
}

// AwaitExprNode represents an await expression.
type AwaitExprNode struct {
	BaseNode
	Expr Node
}

func (n *AwaitExprNode) String() string {
	return fmt.Sprintf("AwaitExpr{Expr: %v}", n.Expr)
}

// RawCNode represents a raw C block.
type RawCNode struct {
	BaseNode
	Content string
}

func (n *RawCNode) String() string {
	return fmt.Sprintf("RawC{Content: %s}", n.Content)
}

// Missing node type aliases for compatibility

// UseLibNode represents a library use directive.
type UseLibNode struct {
	BaseNode
	LibName string
}

func (n *UseLibNode) String() string {
	return fmt.Sprintf("UseLib{LibName: %s}", n.LibName)
}

// DefineNode represents a define directive.
type DefineNode struct {
	BaseNode
	Name  string
	Value string
}

func (n *DefineNode) String() string {
	return fmt.Sprintf("Define{Name: %s, Value: %s}", n.Name, n.Value)
}

// PragmaNode represents a pragma directive.
type PragmaNode struct {
	BaseNode
	Content   string
	Directive string
}

func (n *PragmaNode) String() string {
	return fmt.Sprintf("Pragma{Content: %s}", n.Content)
}

// LibraryNode represents a library declaration.
type LibraryNode struct {
	BaseNode
	Name      string
	Functions []Node
}

func (n *LibraryNode) String() string {
	return fmt.Sprintf("Library{Name: %s}", n.Name)
}

// ConfigNode represents a config declaration.
type ConfigNode struct {
	BaseNode
	Settings map[string]interface{}
}

func (n *ConfigNode) String() string {
	return fmt.Sprintf("Config{Settings: %v}", n.Settings)
}

// WrapperNode represents a wrapper declaration.
type WrapperNode struct {
	BaseNode
	Name         string
	Declarations []Node
}

func (n *WrapperNode) String() string {
	return fmt.Sprintf("Wrapper{Name: %s}", n.Name)
}

// ExternDeclNode represents an extern declaration.
type ExternDeclNode struct {
	BaseNode
	Name        string
	ReturnType  string
	Parameters  []*ParameterNode
	CleanupFunc string // Optional cleanup function for automatic memory management
}

func (n *ExternDeclNode) String() string {
	if n.CleanupFunc != "" {
		return fmt.Sprintf("ExternDecl{Name: %s, ReturnType: %s, Cleanup: %s}", n.Name, n.ReturnType, n.CleanupFunc)
	}
	return fmt.Sprintf("ExternDecl{Name: %s, ReturnType: %s}", n.Name, n.ReturnType)
}

// PackageNode represents a package declaration.
type PackageNode struct {
	BaseNode
	Name string
}

func (n *PackageNode) String() string {
	return fmt.Sprintf("Package{Name: %s}", n.Name)
}

// ImportNode represents an import declaration.
type ImportNode struct {
	BaseNode
	Path string
}

func (n *ImportNode) String() string {
	return fmt.Sprintf("Import{Path: %s}", n.Path)
}

// ProgramNode represents the root of the AST.
type ProgramNode struct {
	BaseNode
	Declarations []Node
}

func (n *ProgramNode) String() string {
	return fmt.Sprintf("Program{Declarations: %d}", len(n.Declarations))
}

// VariableDeclNode represents a variable declaration.
type VariableDeclNode = VarDeclNode

// StructDeclNode represents a struct declaration.
type StructDeclNode struct {
	BaseNode
	Name    string
	Module  string
	Fields  []*VarDeclNode
	Methods []*FunctionDeclNode
}

func (n *StructDeclNode) String() string {
	return fmt.Sprintf("Struct{Name: %s, Fields: %d}", n.Name, len(n.Fields))
}

// EnumDeclNode represents an enum declaration.
type EnumDeclNode struct {
	BaseNode
	Name   string
	Module string
	Values []string
}

func (n *EnumDeclNode) String() string {
	return fmt.Sprintf("Enum{Name: %s, Values: %d}", n.Name, len(n.Values))
}

// DoWhileStmtNode represents a do-while statement.
type DoWhileStmtNode struct {
	BaseNode
	Body      *BlockNode
	Condition Node
}

func (n *DoWhileStmtNode) String() string {
	return fmt.Sprintf("DoWhile{Condition: %v}", n.Condition)
}

// DeferStmtNode represents a defer statement.
type DeferStmtNode struct {
	BaseNode
	Call *CallExprNode
	Body *BlockNode
}

func (n *DeferStmtNode) String() string {
	return fmt.Sprintf("Defer{Call: %v}", n.Call)
}

// MatchStmtNode represents a match statement.
type MatchStmtNode struct {
	BaseNode
	Value Node
	Cases []*CaseClauseNode
}

func (n *MatchStmtNode) String() string {
	return fmt.Sprintf("Match{Value: %v, Cases: %d}", n.Value, len(n.Cases))
}

// CaseClauseNode represents a case clause in a match.
type CaseClauseNode struct {
	BaseNode
	Pattern  Node
	Body     *BlockNode
	TypeName string
	VarName  string
	Literal  Node
}

func (n *CaseClauseNode) String() string {
	return fmt.Sprintf("Case{Pattern: %v}", n.Pattern)
}

// ForInStmtNode represents a for-in statement.
type ForInStmtNode struct {
	BaseNode
	VarName    string
	Collection Node
	Body       *BlockNode
}

func (n *ForInStmtNode) String() string {
	return fmt.Sprintf("ForIn{VarName: %s, Collection: %v}", n.VarName, n.Collection)
}

// RepeatStmtNode represents a repeat-until statement.
type RepeatStmtNode struct {
	BaseNode
	Body      *BlockNode
	Condition Node
	Count     Node
}

func (n *RepeatStmtNode) String() string {
	return fmt.Sprintf("Repeat{Condition: %v}", n.Condition)
}

// SwitchStmtNode represents a switch statement.
type SwitchStmtNode struct {
	BaseNode
	Value Node
	Cases []*SwitchCaseNode
}

func (n *SwitchStmtNode) String() string {
	return fmt.Sprintf("Switch{Value: %v, Cases: %d}", n.Value, len(n.Cases))
}

// SwitchCaseNode represents a case in a switch.
type SwitchCaseNode struct {
	BaseNode
	Constant Node
	Body     *BlockNode
}

func (n *SwitchCaseNode) String() string {
	return fmt.Sprintf("SwitchCase{Constant: %v}", n.Constant)
}

// TestStmtNode represents a test statement.
type TestStmtNode struct {
	BaseNode
	Name string
	Body *BlockNode
}

func (n *TestStmtNode) String() string {
	return fmt.Sprintf("Test{Name: %s}", n.Name)
}

// TupleTypeNode represents a tuple type.
type TupleTypeNode struct {
	BaseNode
	Types []string
}

func (n *TupleTypeNode) String() string {
	return fmt.Sprintf("TupleType{Types: %v}", n.Types)
}

// TupleExprNode represents a tuple expression.
type TupleExprNode struct {
	BaseNode
	Elements []Node
}

func (n *TupleExprNode) String() string {
	return fmt.Sprintf("TupleExpr{Elements: %d}", len(n.Elements))
}

// DictEntry represents a dictionary entry.
type DictEntry struct {
	Key   string
	Value Node
}

// InterpolatedStringNode represents an interpolated string.
type InterpolatedStringNode struct {
	BaseNode
	Parts []Node
}

func (n *InterpolatedStringNode) String() string {
	return fmt.Sprintf("InterpolatedString{Parts: %d}", len(n.Parts))
}

// NamedArg represents a named argument.
type NamedArg = NamedArgumentNode

// LiteralNode represents a literal value.
type LiteralNode struct {
	BaseNode
	Type  string
	Value interface{}
}

func (n *LiteralNode) String() string {
	return fmt.Sprintf("Literal{Value: %v}", n.Value)
}

// AssignmentNode represents an assignment.
type AssignmentNode struct {
	BaseNode
	Target Node
	Value  Node
}

func (n *AssignmentNode) String() string {
	return fmt.Sprintf("Assignment{Target: %v, Value: %v}", n.Target, n.Value)
}

// ArrayAccessNode represents an array access.
type ArrayAccessNode struct {
	BaseNode
	Array Node
	Index Node
}

func (n *ArrayAccessNode) String() string {
	return fmt.Sprintf("ArrayAccess{Array: %v, Index: %v}", n.Array, n.Index)
}

// MemberAccessNode represents a member access.
type MemberAccessNode struct {
	BaseNode
	Object Node
	Member string
}

func (n *MemberAccessNode) String() string {
	return fmt.Sprintf("MemberAccess{Object: %v, Member: %s}", n.Object, n.Member)
}

// SpawnStmtNode represents a spawn statement that runs a function in a new thread.
type SpawnStmtNode struct {
	BaseNode
	Function  Node   // Function to spawn (identifier or expression)
	Arguments []Node // Arguments to pass
	ThreadVar string // Optional variable to store thread handle
}

func (n *SpawnStmtNode) String() string {
	return fmt.Sprintf("Spawn{Function: %v, Args: %v, ThreadVar: %s}", n.Function, n.Arguments, n.ThreadVar)
}

// ChannelExprNode represents channel operations (create, send, recv).
type ChannelExprNode struct {
	BaseNode
	Operation string // "create", "send", "recv", "try_send", "try_recv", "close"
	Channel   Node   // Channel handle (for send/recv/close)
	Value     Node   // Value to send (for send operations)
	ElemType  string // Element type (for create)
	Capacity  int    // Buffer capacity (for create)
}

func (n *ChannelExprNode) String() string {
	return fmt.Sprintf("Channel{Op: %s, Channel: %v, Value: %v, ElemType: %s, Cap: %d}",
		n.Operation, n.Channel, n.Value, n.ElemType, n.Capacity)
}

// Node constants for compatibility
const (
	NodeProgram            = NodeInclude
	NodeUseLib             = NodeInclude
	NodeDefine             = NodeInclude
	NodePragma             = NodeInclude
	NodeLibrary            = NodeInclude
	NodeConfig             = NodeInclude
	NodeWrapper            = NodeInclude
	NodeExternDecl         = NodeInclude
	NodePackage            = NodeInclude
	NodeImport             = NodeInclude
	NodeVariableDecl       = NodeVarDecl
	NodeStructDecl         = NodeInclude
	NodeEnumDecl           = NodeInclude
	NodeDoWhileStmt        = NodeWhileStmt
	NodeDeferStmt          = NodeInclude
	NodeMatchStmt          = NodeInclude
	NodeCaseClause         = NodeInclude
	NodeForInStmt          = NodeForStmt
	NodeRepeatStmt         = NodeWhileStmt
	NodeSwitchStmt         = NodeIfStmt
	NodeSwitchCase         = NodeInclude
	NodeTestStmt           = NodeInclude
	NodeTupleType          = NodeInclude
	NodeTupleExpr          = NodeInclude
	NodeInterpolatedString = NodeStringLiteral
	NodeLiteral            = NodeNumberLiteral
	NodeAssignment         = NodeInclude
	NodeArrayAccess        = NodeIndexExpr
	NodeMemberAccess       = NodeMemberExpr
	NodeSpawnStmt          = NodeInclude
	NodeChannelExpr        = NodeInclude
)

// Config holds configuration for code generation.
type Config struct {
	Features struct {
		Async      bool
		Actors     bool
		Blockchain bool
		QoL        bool
	}
	Backend      string
	IncludePaths []string
	LibraryPaths []string
	Libraries    []string
}
