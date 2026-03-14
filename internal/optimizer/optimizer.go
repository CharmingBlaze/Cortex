// Package optimizer provides optional AST passes for code generation (e.g. constant folding).
// Use Run to apply optimizations before codegen; keeps the compiler modular.
package optimizer

import (
	"cortex/internal/ast"
	"math"
	"strconv"
)

// Options configures which optimizations to run.
type Options struct {
	ConstantFolding bool
}

// Run applies enabled optimizations. Currently runs a constant-folding pass that
// replaces constant binary expressions with literals where the codegen supports it.
// Returns the same root (optimizations are in-place or return replacement).
func Run(root ast.Node, opts Options) ast.Node {
	if opts.ConstantFolding {
		FoldConstants(root)
	}
	return root
}

// foldConstants replaces constant numeric binary expressions with literal nodes.
// Only folds when both operands are numeric literals; leaves AST shape valid.
func FoldConstants(node ast.Node) {
	switch n := node.(type) {
	case *ast.ProgramNode:
		for _, d := range n.Declarations {
			FoldConstants(d)
		}
	case *ast.FunctionDeclNode:
		if n.Body != nil {
			FoldConstants(n.Body)
		}
	case *ast.VariableDeclNode:
		if n.Initializer != nil {
			FoldConstants(n.Initializer)
		}
	case *ast.BinaryExprNode:
		TryFoldBinary(n)
	case *ast.BlockNode:
		for _, s := range n.Statements {
			FoldConstants(s)
		}
	case *ast.IfStmtNode:
		FoldConstants(n.Condition)
		FoldConstants(n.ThenBranch)
		if n.ElseBranch != nil {
			FoldConstants(n.ElseBranch)
		}
	case *ast.WhileStmtNode:
		FoldConstants(n.Condition)
		FoldConstants(n.Body)
	case *ast.ForStmtNode:
		if n.Initializer != nil {
			FoldConstants(n.Initializer)
		}
		if n.Condition != nil {
			FoldConstants(n.Condition)
		}
		if n.Increment != nil {
			FoldConstants(n.Increment)
		}
		FoldConstants(n.Body)
	case *ast.CallExprNode:
		FoldConstants(n.Function)
		for _, a := range n.Args {
			FoldConstants(a)
		}
	case *ast.UnaryExprNode:
		FoldConstants(n.Operand)
	case *ast.AssignmentNode:
		FoldConstants(n.Target)
		FoldConstants(n.Value)
	case *ast.ReturnStmtNode:
		if n.Value != nil {
			FoldConstants(n.Value)
		}
	}
}

func TryFoldBinary(n *ast.BinaryExprNode) {
	left, okL := n.Left.(*ast.LiteralNode)
	right, okR := n.Right.(*ast.LiteralNode)
	if !okL || !okR {
		return
	}
	// Support int, float, and string (lexer often stores numbers as string) literals
	getNum := func(l *ast.LiteralNode) (intVal int, floatVal float64, isFloat bool) {
		switch v := l.Value.(type) {
		case int:
			return v, float64(v), false
		case int64:
			return int(v), float64(v), false
		case float64:
			return int(v), v, true
		case float32:
			return int(v), float64(v), true
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return int(i), float64(i), false
			}
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return int(f), f, true
			}
			return 0, 0, false
		default:
			return 0, 0, false
		}
	}
	lI, lF, lFloat := getNum(left)
	rI, rF, rFloat := getNum(right)
	useFloat := lFloat || rFloat
	op := n.Operator

	var result interface{}
	switch op {
	case "+":
		if useFloat {
			result = lF + rF
		} else {
			result = lI + rI
		}
	case "-":
		if useFloat {
			result = lF - rF
		} else {
			result = lI - rI
		}
	case "*":
		if useFloat {
			result = lF * rF
		} else {
			result = lI * rI
		}
	case "/":
		if useFloat {
			if rF == 0 {
				return
			}
			result = lF / rF
		} else {
			if rI == 0 {
				return
			}
			result = lI / rI
		}
	case "%":
		if useFloat {
			return
		}
		if rI == 0 {
			return
		}
		result = lI % rI
	case "<":
		if useFloat {
			result = lF < rF
		} else {
			result = lI < rI
		}
	case ">":
		if useFloat {
			result = lF > rF
		} else {
			result = lI > rI
		}
	case "<=":
		if useFloat {
			result = lF <= rF
		} else {
			result = lI <= rI
		}
	case ">=":
		if useFloat {
			result = lF >= rF
		} else {
			result = lI >= rI
		}
	case "==":
		if useFloat {
			result = math.Abs(lF-rF) < 1e-9
		} else {
			result = lI == rI
		}
	case "!=":
		if useFloat {
			result = math.Abs(lF-rF) >= 1e-9
		} else {
			result = lI != rI
		}
	case "&&":
		result = ToBool(left) && ToBool(right)
	case "||":
		result = ToBool(left) || ToBool(right)
	default:
		return
	}
	n.FoldedLiteral = &ast.LiteralNode{
		BaseNode: n.BaseNode,
		Value:    result,
		Type:     TypeOf(result),
	}
}

func ToBool(l *ast.LiteralNode) bool {
	if l == nil {
		return false
	}
	switch v := l.Value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case float64:
		return v != 0
	case float32:
		return v != 0
	}
	return false
}

func TypeOf(v interface{}) string {
	switch v.(type) {
	case int:
		return "int"
	case float64, float32:
		return "float"
	case bool:
		return "bool"
	default:
		return "int"
	}
}
