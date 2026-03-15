package compiler

import (
	"cortex/internal/ast"
	"fmt"
	"strings"
)

// EscapeStringForC escapes a string for use inside a C string or char literal.
func EscapeStringForC(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\'':
			b.WriteString(`\'`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 32 || r > 126 {
				b.WriteString(fmt.Sprintf("\\x%02x", r))
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

type CodeGenerator struct {
	indentation            int
	output                 strings.Builder
	outputTarget           *strings.Builder // normally &output; set to &lambdaDefs when emitting lambda body
	lambdaDefs             strings.Builder
	lambdaCounter          int
	testDefs               strings.Builder
	testRegistrations      []struct{ name, funcName string }
	testCounter            int
	cfg                    ast.Config
	structMethods          map[string]map[string]*ast.FunctionDeclNode // struct name -> method name -> method node
	functionParams         map[string][]*ast.ParameterNode             // function name -> parameters for defaults
	currentMethodStruct    string                                      // when generating method body
	currentMethodReceiver  string
	currentMethodFields    map[string]bool
	omitTrailingSemicolon  bool
	deferStack             [][]*ast.BlockNode
	currentFunctionReturns []string          // tuple return types when inside a function
	arrayDimensions        map[string]int    // var name -> 1 or 2 for 1D/2D array (used for bounds check)
	typeEmitNames          map[string]string // Cortex type -> C name (e.g. Vec2 -> math__Vec2 when module set)
	closureCounter         int
	closureCaptureMap      map[string]string       // capture name -> env field when emitting lambda body
	lambdaClosureCache     map[*ast.LambdaNode]int // lambda -> closure id (so we emit struct/fn once)
	usesNetwork            bool                    // if true, emit #include "runtime/network.h"
	usesGui                bool                    // if true, emit #include "runtime/gui_runtime.h"
	usesAsync              bool                    // if true, emit #include "runtime/async.h"
	usesThread             bool                    // if true, emit #include "runtime/thread.h"
	usesManaged            bool                    // if true, emit #include "runtime/managed.h"
	includedHeaders        map[string]bool         // track headers to prevent duplicates
	cleanupFunctions       map[string]string       // function name -> cleanup function name
}

func BoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func NewCodeGenerator(cfg ast.Config) *CodeGenerator {
	return &CodeGenerator{
		cfg:                cfg,
		closureCaptureMap:  make(map[string]string),
		lambdaClosureCache: make(map[*ast.LambdaNode]int),
		includedHeaders:    make(map[string]bool),
	}
}

// SetUsesNetwork sets whether to emit #include "runtime/network.h" (set by compiler when AST uses network APIs).
func (g *CodeGenerator) SetUsesNetwork(v bool) { g.usesNetwork = v }

// SetUsesGui sets whether to emit #include "runtime/gui_runtime.h" (set by compiler when AST uses GUI APIs).
func (g *CodeGenerator) SetUsesGui(v bool) { g.usesGui = v }

// SetUsesAsync sets whether to emit #include "runtime/async.h" (set by compiler when AST uses async APIs).
func (g *CodeGenerator) SetUsesAsync(v bool) { g.usesAsync = v }

// SetUsesThread sets whether to emit #include "runtime/thread.h" (set by compiler when AST uses thread APIs).
func (g *CodeGenerator) SetUsesThread(v bool) { g.usesThread = v }

// SetUsesManaged sets whether to emit #include "runtime/managed.h" (set by compiler when AST uses cleanup annotations).
func (g *CodeGenerator) SetUsesManaged(v bool) { g.usesManaged = v }

func (g *CodeGenerator) Generate(node ast.ASTNode) (string, error) {
	g.output.Reset()
	g.lambdaDefs.Reset()
	g.lambdaCounter = 0
	g.testDefs.Reset()
	g.testRegistrations = nil
	g.testCounter = 0
	g.structMethods = make(map[string]map[string]*ast.FunctionDeclNode)
	g.functionParams = make(map[string][]*ast.ParameterNode)
	g.arrayDimensions = make(map[string]int)
	g.typeEmitNames = make(map[string]string) // Cortex type name -> C name (e.g. Vec2 -> math__Vec2)
	g.closureCounter = 0
	g.closureCaptureMap = nil
	g.lambdaClosureCache = make(map[*ast.LambdaNode]int)
	g.currentMethodStruct = ""
	g.currentMethodReceiver = ""
	g.currentMethodFields = nil
	g.indentation = 0
	g.outputTarget = &g.output

	g.CollectTests(node)
	g.VisitNode(node)

	var registerTests strings.Builder
	if len(g.testRegistrations) > 0 {
		registerTests.WriteString("static void cortex_register_all_tests(void) {\n")
		for _, r := range g.testRegistrations {
			registerTests.WriteString("  test_register(\"" + r.name + "\", " + r.funcName + ");\n")
		}
		registerTests.WriteString("}\n")
	}
	return g.lambdaDefs.String() + g.testDefs.String() + registerTests.String() + g.output.String(), nil
}

func (g *CodeGenerator) CollectTests(node ast.ASTNode) {
	switch n := node.(type) {
	case *ast.ProgramNode:
		for _, d := range n.Declarations {
			g.CollectTests(d)
		}
	case *ast.BlockNode:
		for _, s := range n.Statements {
			g.CollectTests(s)
		}
	case *ast.TestStmtNode:
		g.testCounter++
		fnName := fmt.Sprintf("cortex_test_%d", g.testCounter)
		oldTarget := g.outputTarget
		g.outputTarget = &g.testDefs
		g.Write("static void " + fnName + "(void) {\n")
		g.Indent()
		g.VisitBlock(n.Body)
		g.Dedent()
		g.Write("}\n")
		g.outputTarget = oldTarget
		g.testRegistrations = append(g.testRegistrations, struct{ name, funcName string }{n.Name, fnName})
	case *ast.IfStmtNode:
		g.CollectTests(n.ThenBranch)
		if n.ElseBranch != nil {
			g.CollectTests(n.ElseBranch)
		}
	case *ast.FunctionDeclNode:
		g.CollectTests(n.Body)
	case *ast.WhileStmtNode:
		g.CollectTests(n.Body)
	case *ast.DoWhileStmtNode:
		g.CollectTests(n.Body)
	case *ast.ForStmtNode:
		g.CollectTests(n.Body)
	case *ast.RepeatStmtNode:
		g.CollectTests(n.Body)
	case *ast.SwitchStmtNode:
		for _, c := range n.Cases {
			g.CollectTests(c.Body)
		}
	case *ast.MatchStmtNode:
		for _, c := range n.Cases {
			g.CollectTests(c.Body)
		}
	}
}

func (g *CodeGenerator) VisitNode(node ast.ASTNode) {
	switch n := node.(type) {
	case *ast.ProgramNode:
		g.VisitProgram(n)
	case *ast.IncludeNode:
		g.VisitInclude(n)
	case *ast.UseLibNode:
		g.VisitUseLib(n)
	case *ast.RawCNode:
		g.VisitRawC(n)
	case *ast.DefineNode:
		g.VisitDefine(n)
	case *ast.PragmaNode:
		g.VisitPragma(n)
	case *ast.LibraryNode:
		g.VisitLibrary(n)
	case *ast.ConfigNode:
		g.VisitConfig(n)
	case *ast.WrapperNode:
		g.VisitWrapper(n)
	case *ast.ExternDeclNode:
		g.VisitExternDecl(n)
	case *ast.PackageNode:
		g.Write("// package " + n.Name + "\n")
	case *ast.ImportNode:
		g.Write("// import \"" + n.Path + "\"\n")
	case *ast.FunctionDeclNode:
		g.VisitFunctionDecl(n)
	case *ast.VariableDeclNode:
		g.VisitVariableDecl(n)
	case *ast.StructDeclNode:
		g.VisitStructDecl(n)
	case *ast.EnumDeclNode:
		g.VisitEnumDecl(n)
	case *ast.BlockNode:
		g.VisitBlock(n)
	case *ast.IfStmtNode:
		g.VisitIfStmt(n)
	case *ast.WhileStmtNode:
		g.VisitWhileStmt(n)
	case *ast.DoWhileStmtNode:
		g.VisitDoWhileStmt(n)
	case *ast.ForStmtNode:
		g.VisitForStmt(n)
	case *ast.ForInStmtNode:
		g.VisitForInStmt(n)
	case *ast.RepeatStmtNode:
		g.VisitRepeatStmt(n)
	case *ast.BreakStmtNode:
		g.VisitBreakStmt(n)
	case *ast.ContinueStmtNode:
		g.VisitContinueStmt(n)
	case *ast.SwitchStmtNode:
		g.VisitSwitchStmt(n)
	case *ast.TestStmtNode:
		// already generated in collectTests; skip in main output
	case *ast.DeferStmtNode:
		g.VisitDeferStmt(n)
	case *ast.MatchStmtNode:
		g.VisitMatchStmt(n)
	case *ast.ReturnStmtNode:
		g.VisitReturnStmt(n)
	case *ast.TupleExprNode:
		g.VisitTupleExpr(n)
	case *ast.ArrayLiteralNode:
		g.VisitArrayLiteral(n)
	case *ast.DictLiteralNode:
		g.VisitDictLiteral(n)
	case *ast.StructLiteralNode:
		g.VisitStructLiteral(n)
	case *ast.InterpolatedStringNode:
		g.VisitInterpolatedString(n)
	case *ast.BinaryExprNode:
		g.VisitBinaryExpr(n)
	case *ast.UnaryExprNode:
		g.VisitUnaryExpr(n)
	case *ast.CastExprNode:
		g.VisitCastExpr(n)
	case *ast.CallExprNode:
		g.VisitCallExpr(n)
	case *ast.LiteralNode:
		g.VisitLiteral(n)
	case *ast.IdentifierNode:
		g.VisitIdentifier(n)
	case *ast.AssignmentNode:
		g.VisitAssignment(n)
	case *ast.ArrayAccessNode:
		g.VisitArrayAccess(n)
	case *ast.IndexExprNode:
		g.VisitIndexExpr(n)
	case *ast.MemberAccessNode:
		g.VisitMemberAccess(n)
	case *ast.LambdaNode:
		g.VisitLambda(n)
	case *ast.YieldStmtNode:
		g.VisitYieldStmt(n)
	case *ast.AwaitExprNode:
		g.VisitAwaitExpr(n)
	case *ast.SpawnStmtNode:
		g.VisitSpawnStmt(n)
	default:
		g.Write(fmt.Sprintf("// Unknown node type: %T\n", node))
	}
}

func (g *CodeGenerator) VisitProgram(node *ast.ProgramNode) {
	g.Write("// Generated Cortex Program\n")
	g.Write(fmt.Sprintf("#define CORTEX_FEATURE_ASYNC %d\n", BoolToInt(g.cfg.Features.Async)))
	g.Write(fmt.Sprintf("#define CORTEX_FEATURE_ACTORS %d\n", BoolToInt(g.cfg.Features.Actors)))
	g.Write(fmt.Sprintf("#define CORTEX_FEATURE_BLOCKCHAIN %d\n", BoolToInt(g.cfg.Features.Blockchain)))
	g.Write(fmt.Sprintf("#define CORTEX_FEATURE_QOL %d\n\n", BoolToInt(g.cfg.Features.QoL)))

	// Track standard includes to prevent duplicates
	g.includedHeaders["stdio.h"] = true
	g.includedHeaders["stdlib.h"] = true
	g.includedHeaders["math.h"] = true
	g.includedHeaders["stdbool.h"] = true
	g.includedHeaders["time.h"] = true
	g.includedHeaders["string.h"] = true
	g.includedHeaders["core.h"] = true
	g.includedHeaders["game.h"] = true

	g.Write("#include <stdio.h>\n")
	g.Write("#include <stdlib.h>\n")
	g.Write("#include <math.h>\n")
	g.Write("#include <stdbool.h>\n")
	g.Write("#include <time.h>\n")
	g.Write("#include <string.h>\n")
	g.Write("#include \"runtime/core.h\"\n")
	if g.usesNetwork {
		g.Write("#include \"runtime/network.h\"\n")
	}
	if g.usesGui {
		g.Write("#include \"runtime/gui_runtime.h\"\n")
	}
	if g.usesAsync {
		g.Write("#include \"runtime/async.h\"\n")
	}
	if g.usesThread {
		g.Write("#include \"runtime/thread.h\"\n")
	}
	if g.usesManaged {
		g.Write("#include \"runtime/managed.h\"\n")
	}
	g.Write("#include \"runtime/game.h\"\n\n")

	// Generate declarations
	for _, decl := range node.Declarations {
		g.VisitNode(decl)
		g.Write("\n")
	}
}

func (g *CodeGenerator) VisitInclude(node *ast.IncludeNode) {
	// Skip empty includes
	header := strings.TrimSpace(node.Header)
	if header == "" {
		return
	}
	if !(strings.HasPrefix(header, "<") || strings.HasPrefix(header, "\"")) {
		return
	}
	if node.Filename == "" {
		return
	}

	// Check for duplicates
	if g.includedHeaders[node.Filename] {
		return
	}
	g.includedHeaders[node.Filename] = true

	// Write the include
	if node.IsSystem {
		g.Write(fmt.Sprintf("#include <%s>\n", node.Filename))
	} else {
		g.Write(fmt.Sprintf("#include \"%s\"\n", node.Filename))
	}
}

func (g *CodeGenerator) VisitUseLib(node *ast.UseLibNode) {
	// #use "name" -> #include <name.h> and link -l name (link collected in collectLinkPragmas)
	header := node.LibName + ".h"
	g.Write(fmt.Sprintf("#include <%s>\n", header))
}

func (g *CodeGenerator) VisitRawC(node *ast.RawCNode) {
	g.Write(node.Content)
}

func (g *CodeGenerator) VisitDefine(node *ast.DefineNode) {
	if node.Value != "" {
		g.Write(fmt.Sprintf("#define %s %s\n", node.Name, node.Value))
	} else {
		g.Write(fmt.Sprintf("#define %s\n", node.Name))
	}
}

func (g *CodeGenerator) VisitPragma(node *ast.PragmaNode) {
	if node.Directive == "link" {
		g.Write(fmt.Sprintf("// pragma link: %s\n", node.Content))
	} else {
		g.Write(fmt.Sprintf("#pragma %s %s\n", node.Directive, node.Content))
	}
}

func (g *CodeGenerator) VisitLibrary(node *ast.LibraryNode) {
	g.Write(fmt.Sprintf("// Library: %s\n", node.Name))
	for _, fn := range node.Functions {
		if externFn, ok := fn.(*ast.ExternDeclNode); ok {
			g.VisitExternDecl(externFn)
		}
	}
}

func (g *CodeGenerator) VisitConfig(node *ast.ConfigNode) {
	g.Write("// Configuration\n")
	for key, value := range node.Settings {
		g.Write(fmt.Sprintf("// %s: %v\n", key, value))
	}
}

func (g *CodeGenerator) VisitWrapper(node *ast.WrapperNode) {
	g.Write(fmt.Sprintf("// Wrapper: %s\n", node.Name))
	for _, decl := range node.Declarations {
		g.VisitNode(decl)
	}
}

func (g *CodeGenerator) VisitExternDecl(node *ast.ExternDeclNode) {
	returnType := g.ConvertType(node.ReturnType)
	g.Write(fmt.Sprintf("extern %s %s(", returnType, node.Name))

	for i, param := range node.Parameters {
		if i > 0 {
			g.Write(", ")
		}
		paramType := g.ConvertType(param.Type)
		if param.Name != "" {
			g.Write(fmt.Sprintf("%s %s", paramType, param.Name))
		} else {
			g.Write(paramType)
		}
	}

	g.Write(");\n")

	// Store cleanup function mapping for auto-defer
	if node.CleanupFunc != "" {
		if g.cleanupFunctions == nil {
			g.cleanupFunctions = make(map[string]string)
		}
		g.cleanupFunctions[node.Name] = node.CleanupFunc
		g.usesManaged = true
	}
}

func (g *CodeGenerator) VisitFunctionDecl(node *ast.FunctionDeclNode) {
	cRet := g.ConvertType(node.ReturnType)
	var params []string
	for _, p := range node.Parameters {
		params = append(params, g.ConvertType(p.Type)+" "+p.Name)
	}
	paramStr := strings.Join(params, ", ")

	// Store function parameters for default value handling in calls
	if g.functionParams == nil {
		g.functionParams = make(map[string][]*ast.ParameterNode)
	}
	g.functionParams[node.Name] = node.Parameters

	// Use module-prefixed name for functions from modules
	funcName := node.Name
	if node.Module != "" {
		funcName = node.Module + "__" + node.Name
	}

	// Coroutine functions need special handling - use stackful coroutines from async.c
	if node.IsCoroutine {
		// Generate a struct to hold coroutine arguments
		frameFields := make([]string, len(node.Parameters))
		for i, p := range node.Parameters {
			frameFields[i] = fmt.Sprintf("%s %s", g.ConvertType(p.Type), p.Name)
		}
		frameDef := strings.Join(frameFields, "; ")
		if len(frameFields) > 0 {
			frameDef = frameDef + ";"
		}
		g.Write(fmt.Sprintf("typedef struct { %s } %s_frame;\n\n", frameDef, funcName))

		// Generate the coroutine entry function (called by co_create)
		g.Write(fmt.Sprintf("static void %s_entry(void* _arg) {\n", funcName))
		g.Indent()
		g.Write(fmt.Sprintf("%s_frame* _f = (%s_frame*)_arg;\n", funcName, funcName))
		// Copy parameters to local variables (on coroutine's stack, preserved across yields)
		for _, p := range node.Parameters {
			cType := g.ConvertType(p.Type)
			g.Write(fmt.Sprintf("%s %s = _f->%s;\n", cType, p.Name, p.Name))
		}
		// Generate body with proper semicolons
		for _, stmt := range node.Body.Statements {
			g.VisitNode(stmt)
			// Add semicolon if the statement doesn't already have one
			if !g.omitTrailingSemicolon {
				g.Write(";")
			}
			g.Write("\n")
		}
		g.Dedent()
		g.Write("}\n\n")

		// Generate the wrapper function that users call
		g.Write(fmt.Sprintf("%s %s(%s) {\n", cRet, funcName, paramStr))
		g.Indent()
		g.Write(fmt.Sprintf("%s_frame* _frame = malloc(sizeof(%s_frame));\n", funcName, funcName))
		for _, p := range node.Parameters {
			g.Write(fmt.Sprintf("_frame->%s = %s;\n", p.Name, p.Name))
		}
		g.Write(fmt.Sprintf("co_t _co = co_create(%s_entry, _frame, 0);\n", funcName))
		g.Write("while (!co_finished(_co)) { co_resume(_co); }\n") // Run until complete
		g.Write("co_free(_co);\n")
		g.Write("free(_frame);\n")
		if cRet != "void" {
			g.Write("return 0;\n")
		}
		g.Dedent()
		g.Write("}\n")
		return
	}

	// For async functions, we might need to adjust the function signature or return type in the future
	if node.IsAsync {
		g.Write("// Async function\n")
	}

	g.Write(cRet + " " + funcName + "(" + paramStr + ") ")
	g.VisitBlock(node.Body)
	g.Write("\n")
}

func (g *CodeGenerator) VisitVariableDecl(node *ast.VariableDeclNode) {
	cName := node.Name
	if node.Module != "" {
		cName = node.Module + "__" + node.Name
	}
	if dl, ok := node.Initializer.(*ast.DictLiteralNode); ok {
		g.Write("cortex_dict* " + cName + " = dict_create()")
		if !g.omitTrailingSemicolon {
			g.Write(";")
		}
		g.Write("\n")
		for _, ent := range dl.Entries {
			g.Indent()
			g.Write("dict_set(" + cName + ", " + fmt.Sprintf("%q", ent.Key) + ", ")
			g.EmitExprAsAny(ent.Value)
			g.Write(")")
			if !g.omitTrailingSemicolon {
				g.Write(";")
			}
			g.Write("\n")
			g.Dedent()
		}
		return
	}
	if arr, ok := node.Initializer.(*ast.ArrayLiteralNode); ok {
		// Use declared element type if explicitly specified (e.g., int numbers[5])
		elemType := ""
		if strings.HasSuffix(node.Type, "[]") {
			elemType = strings.TrimSuffix(node.Type, "[]")
		} else if node.ArraySize > 0 {
			// Fixed-size array: int numbers[5] -> element type is int
			elemType = node.Type
		}
		if elemType == "" || elemType == "any" {
			elemType = g.ArrayLiteralElementType(arr)
		}
		ct := g.ConvertType(elemType)
		if arr.Dimensions == 2 {
			rows := arr.RowCount
			cols := arr.RowLength
			if rows <= 0 {
				rows = 1
			}
			if cols <= 0 {
				cols = 1
			}
			g.Write(fmt.Sprintf("%s %s[%d][%d] = ", ct, cName, rows, cols))
			g.VisitArrayLiteral2D(arr)
			g.Write(";\n")
			g.Indent()
			g.Write(fmt.Sprintf("int %s_rows = %d; int %s_cols = %d;", cName, arr.RowCount, cName, arr.RowLength))
			g.Dedent()
			if g.arrayDimensions == nil {
				g.arrayDimensions = make(map[string]int)
			}
			g.arrayDimensions[cName] = 2
			return
		}
		g.Write(fmt.Sprintf("%s %s[] = ", ct, cName))
		g.VisitArrayLiteral(arr)
		g.Write(";\n")
		g.Indent()
		g.Write(fmt.Sprintf("int %s_len = %d;", cName, len(arr.Elements)))
		g.Dedent()
		return
	}
	varType := g.ConvertType(node.Type)
	// Check if this is an array type (Type ends with [])
	if strings.HasSuffix(node.Type, "[]") {
		baseType := strings.TrimSuffix(node.Type, "[]")
		baseCType := g.ConvertType(baseType)
		if node.ArraySize > 0 {
			// Fixed-size array: Type name[size] -> int name[2]
			g.Write(fmt.Sprintf("%s %s[%d]", baseCType, cName, node.ArraySize))
		} else {
			// Dynamic array: Type name[] -> use cortex_array
			g.Write(fmt.Sprintf("cortex_array* %s", cName))
		}
	} else {
		// Check if this is a cleanup-annotated call - use managed handle
		needsManaged := false
		var cleanupFunc string
		if node.Initializer != nil {
			if call, ok := node.Initializer.(*ast.CallExprNode); ok {
				if id, ok := call.Function.(*ast.IdentifierNode); ok {
					if cf, exists := g.cleanupFunctions[id.Name]; exists {
						needsManaged = true
						cleanupFunc = cf
					}
				}
			}
		}

		if needsManaged {
			// Use managed handle with cleanup attribute
			g.Write(fmt.Sprintf("__attribute__((cleanup(cortex_managed_cleanup))) cortex_managed __managed_%s = { ", cName))
			g.VisitNode(node.Initializer)
			g.Write(fmt.Sprintf(", (void(*)(void*))%s }; ", cleanupFunc))
			g.Write(fmt.Sprintf("%s %s = (%s)__managed_%s.ptr", varType, cName, varType, cName))
		} else {
			g.Write(fmt.Sprintf("%s %s", varType, cName))
		}
	}
	if node.Initializer != nil && !strings.Contains(node.Type, "[]") {
		// Check if already handled by managed
		needsManaged := false
		if call, ok := node.Initializer.(*ast.CallExprNode); ok {
			if id, ok := call.Function.(*ast.IdentifierNode); ok {
				if _, exists := g.cleanupFunctions[id.Name]; exists {
					needsManaged = true
				}
			}
		}
		if !needsManaged {
			g.Write(" = ")
			g.VisitNode(node.Initializer)
		}
	}
	if !g.omitTrailingSemicolon {
		g.Write(";")
	}
}

func (g *CodeGenerator) VisitStructDecl(node *ast.StructDeclNode) {
	cName := node.Name
	if node.Module != "" {
		cName = node.Module + "__" + node.Name
	}
	g.Write(fmt.Sprintf("typedef struct {\n"))
	g.Indent()
	for _, field := range node.Fields {
		fieldType := g.ConvertType(field.Type)
		g.Write(fmt.Sprintf("%s %s;\n", fieldType, field.Name))
	}
	g.Dedent()
	g.Write(fmt.Sprintf("} %s;\n", cName))
	if g.typeEmitNames != nil {
		g.typeEmitNames[node.Name] = cName
	}

	if g.structMethods[node.Name] == nil {
		g.structMethods[node.Name] = make(map[string]*ast.FunctionDeclNode)
	}
	for _, m := range node.Methods {
		g.structMethods[node.Name][m.Name] = m
		// Emit static void StructName_method(StructName* self, params) { body }
		retType := g.ConvertType(m.ReturnType)
		g.Write(fmt.Sprintf("static %s %s_%s(%s* self", retType, cName, m.Name, cName))
		for _, param := range m.Parameters {
			g.Write(fmt.Sprintf(", %s %s", g.ConvertType(param.Type), param.Name))
		}
		g.Write(") ")
		fieldSet := make(map[string]bool)
		for _, f := range node.Fields {
			fieldSet[f.Name] = true
		}
		g.currentMethodStruct = node.Name
		g.currentMethodReceiver = "self"
		g.currentMethodFields = fieldSet
		g.VisitBlock(m.Body)
		g.currentMethodStruct = ""
		g.currentMethodReceiver = ""
		g.currentMethodFields = nil
		g.Write("\n")
	}
}

func (g *CodeGenerator) VisitEnumDecl(node *ast.EnumDeclNode) {
	cName := node.Name
	if node.Module != "" {
		cName = node.Module + "__" + node.Name
	}
	g.Write(fmt.Sprintf("typedef enum {\n"))
	g.Indent()

	for i, value := range node.Values {
		if i > 0 {
			g.Write(",\n")
		}
		g.Write(value)
	}

	g.Dedent()
	g.Write(fmt.Sprintf("\n} %s;\n", cName))
	if g.typeEmitNames != nil {
		g.typeEmitNames[node.Name] = cName
	}
}

func (g *CodeGenerator) NeedsStatementSemicolon(node ast.ASTNode) bool {
	switch node.(type) {
	case *ast.CallExprNode, *ast.BinaryExprNode, *ast.UnaryExprNode, *ast.IdentifierNode:
		return true
	case *ast.BreakStmtNode, *ast.ContinueStmtNode:
		return true
	default:
		return false
	}
}

func (g *CodeGenerator) VisitBlock(node *ast.BlockNode) {
	g.VisitBlockWithPrefix(node, "")
}

func (g *CodeGenerator) VisitBlockWithPrefix(node *ast.BlockNode, prefix string) {
	g.Write("{\n")
	g.Indent()
	if prefix != "" {
		g.Write(prefix)
	}
	g.deferStack = append(g.deferStack, nil)

	for _, stmt := range node.Statements {
		if d, ok := stmt.(*ast.DeferStmtNode); ok {
			g.deferStack[len(g.deferStack)-1] = append(g.deferStack[len(g.deferStack)-1], d.Body)
		} else {
			g.VisitNode(stmt)
			if g.NeedsStatementSemicolon(stmt) {
				g.Write(";")
			}
			g.Write("\n")
		}
	}

	// Run defers in LIFO order
	cur := g.deferStack[len(g.deferStack)-1]
	for i := len(cur) - 1; i >= 0; i-- {
		g.VisitBlock(cur[i])
		g.Write("\n")
	}
	g.deferStack = g.deferStack[:len(g.deferStack)-1]
	g.Dedent()
	g.Write("}")
}

func (g *CodeGenerator) VisitIfStmt(node *ast.IfStmtNode) {
	g.Write("if (")
	g.VisitNode(node.Condition)
	g.Write(") ")

	g.VisitNode(node.ThenBranch)

	if node.ElseBranch != nil {
		g.Write(" else ")
		g.VisitNode(node.ElseBranch)
	}
	g.Write("\n")
}

func (g *CodeGenerator) VisitWhileStmt(node *ast.WhileStmtNode) {
	g.Write("while (")
	g.VisitNode(node.Condition)
	g.Write(") ")
	g.VisitNode(node.Body)
}

func (g *CodeGenerator) VisitDoWhileStmt(node *ast.DoWhileStmtNode) {
	g.Write("do ")
	g.VisitNode(node.Body)
	g.Write(" while (")
	g.VisitNode(node.Condition)
	g.Write(");\n")
}

func (g *CodeGenerator) VisitForStmt(node *ast.ForStmtNode) {
	g.Write("for (")
	if node.Initializer != nil {
		g.omitTrailingSemicolon = true
		g.VisitNode(node.Initializer)
		g.omitTrailingSemicolon = false
	} else {
		g.Write(";")
	}
	g.Write("; ")
	if node.Condition != nil {
		g.VisitNode(node.Condition)
	}
	g.Write("; ")
	if node.Increment != nil {
		g.VisitNode(node.Increment)
	}
	g.Write(") ")
	g.VisitNode(node.Body)
}

func (g *CodeGenerator) VisitRepeatStmt(node *ast.RepeatStmtNode) {
	g.Write("for (int _repeat_i = 0; _repeat_i < (")
	if node.Count != nil {
		g.VisitNode(node.Count)
	}
	g.Write("); _repeat_i++) ")
	g.VisitNode(node.Body)
}

func (g *CodeGenerator) VisitBreakStmt(node *ast.BreakStmtNode) {
	g.Write("break")
}

func (g *CodeGenerator) VisitContinueStmt(node *ast.ContinueStmtNode) {
	g.Write("continue")
}

func (g *CodeGenerator) VisitSwitchStmt(node *ast.SwitchStmtNode) {
	g.Write("switch (")
	g.VisitNode(node.Value)
	g.Write(") {\n")
	g.Indent()
	for _, c := range node.Cases {
		if c.Constant != nil {
			g.Write("case ")
			g.VisitNode(c.Constant)
			g.Write(":\n")
		} else {
			g.Write("default:\n")
		}
		g.Indent()
		g.VisitBlock(c.Body)
		g.Dedent()
	}
	g.Dedent()
	g.Write("}\n")
}

func (g *CodeGenerator) VisitLambda(node *ast.LambdaNode) {
	if len(node.Captures) > 0 {
		// Captured lambda: generate closure; when used as call arg, visitCallExpr emits (fn, &env)
		fnName, _ := g.GenerateClosure(node)
		g.Write(fnName)
		return
	}
	g.lambdaCounter++
	name := fmt.Sprintf("cortex_lambda_%d", g.lambdaCounter-1)
	retType := node.ReturnType
	if retType == "" {
		retType = "void"
	}
	cRet := g.ConvertType(retType)
	var params []string
	for _, p := range node.Parameters {
		params = append(params, g.ConvertType(p.Type)+" "+p.Name)
	}
	paramStr := strings.Join(params, ", ")

	oldTarget := g.outputTarget
	g.outputTarget = &g.lambdaDefs
	g.Write("static " + cRet + " " + name + "(" + paramStr + ") {\n")
	g.Indent()
	g.VisitBlock(node.Body)
	g.Dedent()
	g.Write("}\n")
	g.outputTarget = oldTarget

	g.Write(name)
}

// generateClosure emits closure struct and function to lambdaDefs (once per lambda), returns (fnName, structTypeName).
func (g *CodeGenerator) GenerateClosure(node *ast.LambdaNode) (fnName, structTypeName string) {
	if id, ok := g.lambdaClosureCache[node]; ok {
		return fmt.Sprintf("cortex_closure_%d_fn", id), fmt.Sprintf("cortex_closure_%d_t", id)
	}
	id := g.closureCounter
	g.closureCounter++
	g.lambdaClosureCache[node] = id
	fnName = fmt.Sprintf("cortex_closure_%d_fn", id)
	structTypeName = fmt.Sprintf("cortex_closure_%d_t", id)

	// Build struct fields from capture types
	types := node.ResolvedCaptureTypes
	if len(types) < len(node.Captures) {
		types = make([]string, len(node.Captures))
		for i := range node.Captures {
			if i < len(node.ResolvedCaptureTypes) {
				types[i] = node.ResolvedCaptureTypes[i]
			} else {
				types[i] = "any"
			}
		}
	}
	var fields []string
	for i, t := range types {
		if i >= len(node.Captures) {
			break
		}
		fields = append(fields, g.ConvertType(t)+" c"+fmt.Sprint(i))
	}
	oldTarget := g.outputTarget
	g.outputTarget = &g.lambdaDefs
	g.Write("typedef struct { " + strings.Join(fields, "; ") + " } " + structTypeName + ";\n")

	retType := node.ReturnType
	if retType == "" {
		retType = "void"
	}
	cRet := g.ConvertType(retType)
	var params []string
	params = append(params, "void* _env")
	for _, p := range node.Parameters {
		params = append(params, g.ConvertType(p.Type)+" "+p.Name)
	}
	paramStr := strings.Join(params, ", ")

	g.closureCaptureMap = make(map[string]string)
	for i, capName := range node.Captures {
		g.closureCaptureMap[capName] = "c" + fmt.Sprint(i)
	}
	g.Write("static " + cRet + " " + fnName + "(" + paramStr + ") { " + structTypeName + "* env = (" + structTypeName + "*)_env; ")
	g.VisitBlock(node.Body)
	g.Write(" }\n")
	g.closureCaptureMap = nil
	g.outputTarget = oldTarget
	return fnName, structTypeName
}

func (g *CodeGenerator) EmitDefers() {
	for level := len(g.deferStack) - 1; level >= 0; level-- {
		for i := len(g.deferStack[level]) - 1; i >= 0; i-- {
			g.VisitBlock(g.deferStack[level][i])
			g.Write("\n")
		}
	}
}

func (g *CodeGenerator) VisitReturnStmt(node *ast.ReturnStmtNode) {
	g.EmitDefers()
	g.Write("return")
	if node.Value != nil {
		g.Write(" ")
		if tuple, ok := node.Value.(*ast.TupleExprNode); ok && len(g.currentFunctionReturns) > 1 {
			g.Write("(struct { ")
			for i, t := range g.currentFunctionReturns {
				if i > 0 {
					g.Write(" ")
				}
				g.Write(g.ConvertType(t) + " f" + fmt.Sprint(i) + ";")
			}
			g.Write(" }){ ")
			for i, e := range tuple.Elements {
				if i > 0 {
					g.Write(", ")
				}
				g.VisitNode(e)
			}
			g.Write(" }")
		} else {
			g.VisitNode(node.Value)
		}
	}
	g.Write(";")
}

func (g *CodeGenerator) VisitDeferStmt(node *ast.DeferStmtNode) {
	// Defers are collected in visitBlock and emitted at block end / before return
}

func (g *CodeGenerator) VisitMatchStmt(node *ast.MatchStmtNode) {
	for i, c := range node.Cases {
		if i > 0 {
			g.Write(" else ")
		}
		if c.TypeName == "" && c.Literal == nil {
			g.Write("{\n")
			g.Indent()
			g.VisitBlock(c.Body)
			g.Dedent()
			g.Write("\n}")
			continue
		}
		if c.TypeName == "Ok" {
			g.Write("if (result_is_ok(")
			g.VisitNode(node.Value)
			g.Write(")) ")
			if c.VarName != "" {
				g.Write("{\n")
				g.Indent()
				g.Write("AnyValue " + c.VarName + " = result_value(")
				g.VisitNode(node.Value)
				g.Write(");\n")
				g.VisitBlock(c.Body)
				g.Dedent()
				g.Write("\n}")
			} else {
				g.VisitNode(c.Body)
			}
			continue
		}
		if c.TypeName == "Err" {
			g.Write("if (!result_is_ok(")
			g.VisitNode(node.Value)
			g.Write(")) ")
			if c.VarName != "" {
				g.Write("{\n")
				g.Indent()
				g.Write("char* " + c.VarName + " = result_error(")
				g.VisitNode(node.Value)
				g.Write(");\n")
				g.VisitBlock(c.Body)
				g.Dedent()
				g.Write("\n}")
			} else {
				g.VisitNode(c.Body)
			}
			continue
		}
		if c.Literal != nil {
			g.Write("if (")
			g.VisitNode(node.Value)
			g.Write(" == ")
			g.VisitNode(c.Literal)
			g.Write(") ")
		} else {
			g.Write("if (is_type(")
			g.VisitNode(node.Value)
			g.Write(", \"" + c.TypeName + "\")) ")
			if c.VarName != "" {
				g.Write("{\n")
				g.Indent()
				g.Write(fmt.Sprintf("%s %s = as_%s(", c.TypeName, c.VarName, c.TypeName))
				g.VisitNode(node.Value)
				g.Write(");\n")
				g.VisitBlock(c.Body)
				g.Dedent()
				g.Write("\n}")
			} else {
				g.VisitNode(c.Body)
			}
			continue
		}
		g.VisitNode(c.Body)
	}
}

func (g *CodeGenerator) VisitForInStmt(node *ast.ForInStmtNode) {
	// for (x in arr) -> for (int _i = 0; _i < arr_len; _i++) { type x = arr[_i]; body }
	colName := ""
	if id, ok := node.Collection.(*ast.IdentifierNode); ok {
		colName = id.Name
	}
	if colName == "" {
		colName = "arr" // fallback
	}
	g.Write("for (int _i = 0; _i < " + colName + "_len; _i++) {\n")
	g.Indent()
	g.Write(fmt.Sprintf("int %s = ", node.VarName))
	g.VisitNode(node.Collection)
	g.Write("[_i];\n")
	g.VisitBlock(node.Body)
	g.Dedent()
	g.Write("}\n")
}

func (g *CodeGenerator) VisitTupleExpr(node *ast.TupleExprNode) {
	// Emit as compound literal or initializer list for return
	if len(node.Elements) == 0 {
		return
	}
	g.Write("(")
	for i, e := range node.Elements {
		if i > 0 {
			g.Write(", ")
		}
		g.VisitNode(e)
	}
	g.Write(")")
}

func (g *CodeGenerator) VisitArrayLiteral(node *ast.ArrayLiteralNode) {
	if node.Dimensions == 2 {
		g.VisitArrayLiteral2D(node)
		return
	}
	g.Write("{")
	for i, e := range node.Elements {
		if i > 0 {
			g.Write(", ")
		}
		g.VisitNode(e)
	}
	g.Write("}")
}

// visitArrayLiteral2D emits {{row0}, {row1}, ...} for 2D C array initializer.
func (g *CodeGenerator) VisitArrayLiteral2D(node *ast.ArrayLiteralNode) {
	g.Write("{")
	for i, row := range node.Elements {
		if i > 0 {
			g.Write(", ")
		}
		if inner, ok := row.(*ast.ArrayLiteralNode); ok {
			g.Write("{")
			for j, e := range inner.Elements {
				if j > 0 {
					g.Write(", ")
				}
				g.VisitNode(e)
			}
			g.Write("}")
		} else {
			g.VisitNode(row)
		}
	}
	g.Write("}")
}

func (g *CodeGenerator) ArrayLiteralElementType(node *ast.ArrayLiteralNode) string {
	t := g.GetExpressionType(node)
	for strings.HasSuffix(t, "[]") {
		t = t[:len(t)-2]
	}
	if t == "" || t == "any" {
		t = "int"
	}
	return t
}

func (g *CodeGenerator) VisitDictLiteral(node *ast.DictLiteralNode) {
	// Emit as GNU C statement expression so dict literal can appear in any expression context
	g.Write("({ cortex_dict* __d = dict_create();\n")
	for _, ent := range node.Entries {
		g.Write("dict_set(__d, " + fmt.Sprintf("%q", ent.Key) + ", ")
		g.EmitExprAsAny(ent.Value)
		g.Write(");\n")
	}
	g.Write("__d; })")
}

func (g *CodeGenerator) VisitStructLiteral(node *ast.StructLiteralNode) {
	// Emit as GNU C statement expression: ({ StructType __s = { .field = value, ... }; __s; })
	// The type name should be set by the semantic analyzer or inferred from context
	typeName := node.TypeName
	if typeName == "" {
		typeName = "struct_" // fallback, should be set properly
	}
	g.Write("({ " + typeName + " __s = { ")
	for i, field := range node.Fields {
		if i > 0 {
			g.Write(", ")
		}
		g.Write("." + field.Name + " = ")
		g.VisitNode(field.Value)
	}
	g.Write(" }; __s; })")
}

func (g *CodeGenerator) VisitInterpolatedString(node *ast.InterpolatedStringNode) {
	if len(node.Parts) == 0 {
		g.Write(`""`)
		return
	}
	if len(node.Parts) == 1 {
		if lit, ok := node.Parts[0].(*ast.LiteralNode); ok && lit.Type == "string" {
			g.Write(fmt.Sprintf(`"%s"`, EscapeStringForC(lit.Value.(string))))
		} else {
			g.Write(g.ToStringFuncForType(g.GetExpressionType(node.Parts[0])) + "(")
			g.VisitNode(node.Parts[0])
			g.Write(")")
		}
		return
	}
	for i := 0; i < len(node.Parts)-1; i++ {
		g.Write("cortex_strcat(")
	}
	if lit, ok := node.Parts[0].(*ast.LiteralNode); ok && lit.Type == "string" {
		g.Write(fmt.Sprintf(`"%s"`, EscapeStringForC(lit.Value.(string))))
	} else {
		g.Write(g.ToStringFuncForType(g.GetExpressionType(node.Parts[0])) + "(")
		g.VisitNode(node.Parts[0])
		g.Write(")")
	}
	for i := 1; i < len(node.Parts); i++ {
		g.Write(", ")
		if lit, ok := node.Parts[i].(*ast.LiteralNode); ok && lit.Type == "string" {
			g.Write(fmt.Sprintf(`"%s"`, EscapeStringForC(lit.Value.(string))))
		} else {
			g.Write(g.ToStringFuncForType(g.GetExpressionType(node.Parts[i])) + "(")
			g.VisitNode(node.Parts[i])
			g.Write(")")
		}
		g.Write(")")
	}
}

func (g *CodeGenerator) VisitBinaryExpr(node *ast.BinaryExprNode) {
	if node.FoldedLiteral != nil {
		if lit, ok := node.FoldedLiteral.(*ast.LiteralNode); ok {
			g.VisitLiteral(lit)
			return
		}
	}
	leftType := g.GetExpressionType(node.Left)
	rightType := g.GetExpressionType(node.Right)

	// Smart type conversion and code generation
	switch node.Operator {
	case "+":
		// Handle string concatenation vs numeric addition
		if leftType == "string" || rightType == "string" {
			g.GenerateStringConcatenation(node.Left, node.Right, leftType, rightType)
		} else {
			g.GenerateNumericOperation(node.Left, node.Right, node.Operator, leftType, rightType)
		}
	case "-", "*", "/":
		g.GenerateNumericOperation(node.Left, node.Right, node.Operator, leftType, rightType)
	case "==", "!=", "<", "<=", ">", ">=":
		g.GenerateComparison(node.Left, node.Right, node.Operator, leftType, rightType)
	case "&&", "||":
		g.GenerateLogicalOperation(node.Left, node.Right, node.Operator, leftType, rightType)
	default:
		// Default case
		g.Write("(")
		g.VisitNode(node.Left)
		g.Write(fmt.Sprintf(" %s ", node.Operator))
		g.VisitNode(node.Right)
		g.Write(")")
	}
}

func (g *CodeGenerator) ToStringFuncForType(t string) string {
	switch t {
	case "int", "size_t":
		return "toString_int"
	case "float":
		return "toString_float"
	case "double":
		return "toString_double"
	case "bool":
		return "toString_bool"
	default:
		return "toString_int" // fallback for any/numeric
	}
}

func (g *CodeGenerator) GenerateStringConcatenation(left, right ast.ASTNode, leftType, rightType string) {
	// C has no + for strings; use runtime helper (caller must free for nested concat)
	g.Write("(cortex_strcat(")
	if leftType != "string" {
		g.Write(g.ToStringFuncForType(leftType) + "(")
		g.VisitNode(left)
		g.Write(")")
	} else {
		g.VisitNode(left)
	}
	g.Write(", ")
	if rightType != "string" {
		g.Write(g.ToStringFuncForType(rightType) + "(")
		g.VisitNode(right)
		g.Write(")")
	} else {
		g.VisitNode(right)
	}
	g.Write("))")
}

func (g *CodeGenerator) GenerateNumericOperation(left, right ast.ASTNode, op string, leftType, rightType string) {
	targetType := g.GetHigherPrecisionType(leftType, rightType)
	cTargetType := g.ConvertType(targetType) // "number" -> "double" for valid C cast

	g.Write("(")

	if leftType != targetType {
		g.Write(fmt.Sprintf("(%s)", cTargetType))
		g.VisitNode(left)
	} else {
		g.VisitNode(left)
	}

	g.Write(fmt.Sprintf(" %s ", op))

	if rightType != targetType {
		g.Write(fmt.Sprintf("(%s)", cTargetType))
		g.VisitNode(right)
	} else {
		g.VisitNode(right)
	}

	g.Write(")")
}

func (g *CodeGenerator) GenerateComparison(left, right ast.ASTNode, op string, leftType, rightType string) {
	g.Write("(")
	g.VisitNode(left)
	g.Write(fmt.Sprintf(" %s ", op))
	g.VisitNode(right)
	g.Write(")")
}

func (g *CodeGenerator) GenerateLogicalOperation(left, right ast.ASTNode, op string, leftType, rightType string) {
	g.Write("(")
	g.VisitNode(left)
	g.Write(fmt.Sprintf(" %s ", op))
	g.VisitNode(right)
	g.Write(")")
}

// emitExprAsAny emits C code that produces an AnyValue from the given expression (for array_push, dict_set, result_ok).
func (g *CodeGenerator) EmitExprAsAny(expr ast.ASTNode) {
	t := g.GetExpressionType(expr)
	switch t {
	case "int", "number":
		g.Write("make_any_int(")
		g.VisitNode(expr)
		g.Write(")")
	case "float", "double":
		g.Write("make_any_float(")
		g.VisitNode(expr)
		g.Write(")")
	case "string":
		g.Write("make_any_string(")
		g.VisitNode(expr)
		g.Write(")")
	case "bool":
		g.Write("make_any_bool(")
		g.VisitNode(expr)
		g.Write(")")
	case "vec2":
		g.Write("make_any_vec2(")
		g.VisitNode(expr)
		g.Write(")")
	case "vec3":
		g.Write("make_any_vec3(")
		g.VisitNode(expr)
		g.Write(")")
	case "dict":
		g.Write("make_any_dict(")
		g.VisitNode(expr)
		g.Write(")")
	case "array":
		g.Write("make_any_array(")
		g.VisitNode(expr)
		g.Write(")")
	case "any":
		g.VisitNode(expr)
	default:
		g.Write("make_any_null()")
	}
}

func (g *CodeGenerator) GetExpressionType(expr ast.ASTNode) string {
	switch e := expr.(type) {
	case *ast.LiteralNode:
		return e.Type
	case *ast.IdentifierNode:
		if e.ResolvedType != "" {
			return e.ResolvedType
		}
		return "any"
	case *ast.TupleExprNode:
		return "tuple"
	case *ast.ArrayLiteralNode:
		if len(e.Elements) > 0 {
			return g.GetExpressionType(e.Elements[0]) + "[]"
		}
		return "int[]"
	case *ast.DictLiteralNode:
		return "dict"
	case *ast.InterpolatedStringNode:
		return "string"
	case *ast.BinaryExprNode:
		if e.FoldedLiteral != nil {
			if lit, ok := e.FoldedLiteral.(*ast.LiteralNode); ok {
				return lit.Type
			}
		}
		return g.InferBinaryExpressionType(e)
	case *ast.CallExprNode:
		if id, ok := e.Function.(*ast.IdentifierNode); ok && id.ResolvedType != "" {
			return id.ResolvedType
		}
		return "any"
	case *ast.MemberAccessNode:
		objType := g.GetExpressionType(e.Object)
		if objType != "" && objType != "any" {
			return "any" // Field type would require struct info; use any for safety
		}
		return "any"
	case *ast.LambdaNode:
		return "any" // no-capture lambda; use var to hold function pointer
	default:
		return "any"
	}
}

func (g *CodeGenerator) InferBinaryExpressionType(expr *ast.BinaryExprNode) string {
	leftType := g.GetExpressionType(expr.Left)
	rightType := g.GetExpressionType(expr.Right)

	switch expr.Operator {
	case "+", "-", "*", "/":
		if g.IsNumericType(leftType) && g.IsNumericType(rightType) {
			return g.GetHigherPrecisionType(leftType, rightType)
		}
		if leftType == "string" || rightType == "string" {
			return "string"
		}
		return "any"
	case "==", "!=", "<", "<=", ">", ">=":
		return "bool"
	case "&&", "||":
		return "bool"
	default:
		return "any"
	}
}

func (g *CodeGenerator) IsNumericType(typeName string) bool {
	return typeName == "int" || typeName == "float" || typeName == "double" || typeName == "size_t"
}

func (g *CodeGenerator) GetHigherPrecisionType(type1, type2 string) string {
	typeOrder := map[string]int{
		"int":    1,
		"float":  2,
		"double": 3,
		"size_t": 2,
	}

	if typeOrder[type1] > typeOrder[type2] {
		return type1
	}
	return type2
}

func (g *CodeGenerator) VisitUnaryExpr(node *ast.UnaryExprNode) {
	if node.IsPostfix && (node.Operator == "++" || node.Operator == "--") {
		g.VisitNode(node.Operand)
		g.Write(node.Operator)
	} else {
		g.Write(fmt.Sprintf("(%s", node.Operator))
		g.VisitNode(node.Operand)
		g.Write(")")
	}
}

// VisitCastExpr generates C-style cast: (type)expr
func (g *CodeGenerator) VisitCastExpr(node *ast.CastExprNode) {
	g.Write("(")
	g.Write(node.TargetType)
	g.Write(")")
	g.VisitNode(node.Operand)
}

func (g *CodeGenerator) VisitCallExpr(node *ast.CallExprNode) {
	// Check if this is a method call: obj.method(args)
	if member, ok := node.Function.(*ast.MemberAccessNode); ok {
		objType := g.GetExpressionType(member.Object)
		// Check if objType is a struct with this method
		if methods, exists := g.structMethods[objType]; exists {
			if _, hasMethod := methods[member.Member]; hasMethod {
				// Generate: StructName_method(&obj, args...)
				emitName := objType
				if g.typeEmitNames != nil && g.typeEmitNames[objType] != "" {
					emitName = g.typeEmitNames[objType]
				}
				g.Write(fmt.Sprintf("%s_%s(", emitName, member.Member))
				// Pass object by reference
				g.Write("&")
				g.VisitNode(member.Object)
				// Pass other arguments
				for _, arg := range node.Args {
					g.Write(", ")
					g.VisitNode(arg)
				}
				g.Write(")")
				return
			}
		}
		// Check if this is a module-qualified call: module.func(args)
		if id, ok := member.Object.(*ast.IdentifierNode); ok {
			// Generate: module__func(args)
			g.Write(fmt.Sprintf("%s__%s(", id.Name, member.Member))
			for i, arg := range node.Args {
				if i > 0 {
					g.Write(", ")
				}
				g.VisitNode(arg)
			}
			g.Write(")")
			return
		}
	}

	if id, ok := node.Function.(*ast.IdentifierNode); ok {
		name := id.Name
		if id.EmitName != "" {
			name = id.EmitName
		}
		// Case-insensitive matching for built-in functions
		switch strings.ToLower(name) {
		case "print", "say":
			g.Write("print_string")
		case "println", "show", "writeline":
			g.Write("println_string")
		case "printf":
			g.Write("printf")
		case "as_string":
			// Special handling: use toString_* for primitives, as_string for AnyValue
			if len(node.Args) > 0 {
				argType := g.GetExpressionType(node.Args[0])
				switch argType {
				case "int", "int32", "int64":
					g.Write("toString_int(")
					g.VisitNode(node.Args[0])
					g.Write(")")
				case "float", "float32", "double", "float64":
					g.Write("toString_float(")
					g.VisitNode(node.Args[0])
					g.Write(")")
				case "bool":
					g.Write("toString_bool(")
					g.VisitNode(node.Args[0])
					g.Write(")")
				default:
					// For AnyValue or unknown types, wrap in AnyValue first
					g.Write("as_string(")
					g.EmitExprAsAny(node.Args[0])
					g.Write(")")
				}
			} else {
				g.Write("\"\"")
			}
			return
		case "as_int":
			// Convert to int from AnyValue
			if len(node.Args) > 0 {
				argType := g.GetExpressionType(node.Args[0])
				if argType == "int" {
					g.VisitNode(node.Args[0])
				} else {
					g.Write("as_int(")
					g.EmitExprAsAny(node.Args[0])
					g.Write(")")
				}
			} else {
				g.Write("0")
			}
			return
		case "as_float":
			// Convert to float from AnyValue
			if len(node.Args) > 0 {
				argType := g.GetExpressionType(node.Args[0])
				if argType == "float" || argType == "double" {
					g.VisitNode(node.Args[0])
				} else if argType == "int" {
					g.Write("(float)(")
					g.VisitNode(node.Args[0])
					g.Write(")")
				} else {
					g.Write("as_float(")
					g.EmitExprAsAny(node.Args[0])
					g.Write(")")
				}
			} else {
				g.Write("0.0f")
			}
			return
		case "as_bool":
			// Convert to bool from AnyValue
			if len(node.Args) > 0 {
				argType := g.GetExpressionType(node.Args[0])
				if argType == "bool" {
					g.VisitNode(node.Args[0])
				} else {
					g.Write("as_bool(")
					g.EmitExprAsAny(node.Args[0])
					g.Write(")")
				}
			} else {
				g.Write("false")
			}
			return
		case "type_of":
			g.Write("type_of(")
			if len(node.Args) > 0 {
				g.EmitExprAsAny(node.Args[0])
			}
			g.Write(")")
			return
		case "is_type":
			g.Write("is_type(")
			if len(node.Args) > 0 {
				g.EmitExprAsAny(node.Args[0])
				if len(node.Args) > 1 {
					g.Write(", ")
					g.VisitNode(node.Args[1])
				}
			}
			g.Write(")")
			return
		case "as_array":
			g.Write("as_array(")
			if len(node.Args) > 0 {
				g.EmitExprAsAny(node.Args[0])
			}
			g.Write(")")
			return
		case "as_dict":
			g.Write("as_dict(")
			if len(node.Args) > 0 {
				g.EmitExprAsAny(node.Args[0])
			}
			g.Write(")")
			return
		case "dict_get":
			g.Write("dict_get(")
			for i, arg := range node.Args {
				if i > 0 {
					g.Write(", ")
				}
				g.VisitNode(arg)
			}
			g.Write(")")
			return
		case "make_any_dict":
			g.Write("make_any_dict(")
			for i, arg := range node.Args {
				if i > 0 {
					g.Write(", ")
				}
				g.VisitNode(arg)
			}
			g.Write(")")
			return
		case "entity_create":
			g.Write("entity_create()")
			return
		case "entity_remove":
			g.Write("entity_remove(")
			for i, arg := range node.Args {
				if i > 0 {
					g.Write(", ")
				}
				g.VisitNode(arg)
			}
			g.Write(")")
			return
		case "add_component":
			g.Write("add_component(")
			for i, arg := range node.Args {
				if i > 0 {
					g.Write(", ")
				}
				if i == 2 {
					// Third argument needs to be AnyValue
					g.EmitExprAsAny(arg)
				} else {
					g.VisitNode(arg)
				}
			}
			g.Write(")")
			return
		case "get_component":
			g.Write("get_component(")
			for i, arg := range node.Args {
				if i > 0 {
					g.Write(", ")
				}
				g.VisitNode(arg)
			}
			g.Write(")")
			return
		case "has_component":
			g.Write("has_component(")
			for i, arg := range node.Args {
				if i > 0 {
					g.Write(", ")
				}
				g.VisitNode(arg)
			}
			g.Write(")")
			return
		case "json_stringify":
			g.Write("json_stringify_any(")
			if len(node.Args) > 0 {
				g.EmitExprAsAny(node.Args[0])
			}
			g.Write(")")
			return
		case "json_parse":
			g.Write("json_parse(")
			for i, arg := range node.Args {
				if i > 0 {
					g.Write(", ")
				}
				g.VisitNode(arg)
			}
			g.Write(")")
			return
		case "gui_window_create", "gui_window_show", "gui_window_center",
			"gui_dialog_info", "gui_dialog_error", "gui_dialog_confirm", "gui_run",
			"gui_label_create", "gui_button_create", "gui_container_add":
			g.Write(name + "(")
			for i, arg := range node.Args {
				if i > 0 {
					g.Write(", ")
				}
				g.VisitNode(arg)
			}
			g.Write(")")
			return
		case "assert":
			// Special handling: assert(cond) -> ((cond) ? (void)0 : cortex_assert_fail(line, "assertion failed"))
			g.Write("((")
			if len(node.Args) > 0 {
				g.VisitNode(node.Args[0])
			}
			g.Write(") ? (void)0 : cortex_assert_fail(")
			g.Write(fmt.Sprintf("%d", node.GetLine()))
			g.Write(", \"assertion failed\"))")
			return
		default:
			g.Write(name)
		}
	} else if lambda, ok := node.Function.(*ast.LambdaNode); ok {
		if len(lambda.Captures) > 0 {
			fnName, structTypeName := g.GenerateClosure(lambda)
			// Emit environment struct initialization with captured values
			envName := fmt.Sprintf("env_%d", g.closureCounter-1)
			g.Write(fmt.Sprintf("%s %s = { ", structTypeName, envName))
			for i, capName := range lambda.Captures {
				if i > 0 {
					g.Write(", ")
				}
				g.Write(capName)
			}
			g.Write(" }; ")
			// Emit function pointer and environment pointer as arguments
			g.Write(fmt.Sprintf("%s, &%s", fnName, envName))
			return
		} else {
			g.VisitLambda(lambda)
		}
	} else {
		g.VisitNode(node.Function)
	}

	g.Write("(")
	// Get function signature for default parameter handling
	var fnParams []*ast.ParameterNode
	if id, ok := node.Function.(*ast.IdentifierNode); ok {
		fnParams = g.functionParams[id.Name]
	}

	// Handle positional arguments with potential pointer conversion
	for i, arg := range node.Args {
		if i > 0 {
			g.Write(", ")
		}
		// Check if argument needs pointer conversion
		argType := g.GetExpressionType(arg)
		if strings.HasPrefix(argType, "array_") || strings.HasPrefix(argType, "struct_") {
			g.Write("&")
			g.VisitNode(arg)
		} else {
			g.VisitNode(arg)
		}
	}

	// Fill in default values for missing parameters
	if fnParams != nil {
		for i := len(node.Args); i < len(fnParams); i++ {
			if i > 0 {
				g.Write(", ")
			}
			if fnParams[i].DefaultValue != nil {
				g.VisitNode(fnParams[i].DefaultValue)
			} else {
				// No default value - this is an error but emit 0 to avoid C errors
				g.Write("0")
			}
		}
	}

	// Handle named arguments (assuming order is resolved at semantic stage or by function definition)
	if len(node.NamedArgs) > 0 {
		if len(node.Args) > 0 {
			g.Write(", ")
		}
		for i, namedArg := range node.NamedArgs {
			if i > 0 {
				g.Write(", ")
			}
			g.VisitNode(namedArg.Value)
		}
	}
	g.Write(")")
}

func (g *CodeGenerator) VisitLiteral(node *ast.LiteralNode) {
	switch node.Type {
	case "bool":
		if node.Value.(bool) {
			g.Write("true")
		} else {
			g.Write("false")
		}
	case "null":
		g.Write("NULL")
	case "string":
		g.Write(fmt.Sprintf(`"%s"`, EscapeStringForC(node.Value.(string))))
	case "char":
		g.Write(fmt.Sprintf(`'%s'`, EscapeStringForC(node.Value.(string))))
	default:
		// For numbers, just output the value directly for C compatibility
		g.Write(fmt.Sprintf("%v", node.Value))
	}
}

func (g *CodeGenerator) VisitIdentifier(node *ast.IdentifierNode) {
	if g.closureCaptureMap != nil {
		if field, ok := g.closureCaptureMap[node.Name]; ok {
			g.Write("env->" + field)
			return
		}
	}
	if g.currentMethodFields != nil && g.currentMethodFields[node.Name] {
		g.Write(g.currentMethodReceiver + "->" + node.Name)
		return
	}
	switch node.Name {
	case "print", "say":
		g.Write("print_string")
	case "println", "show", "writeline":
		g.Write("println_string")
	default:
		name := node.Name
		if node.EmitName != "" {
			name = node.EmitName
		}
		g.Write(name)
	}
}

func (g *CodeGenerator) VisitAssignment(node *ast.AssignmentNode) {
	g.VisitNode(node.Target)
	g.Write(" = ")
	g.VisitNode(node.Value)
	g.Write(";")
}

func (g *CodeGenerator) GetArrayAccessBase(node *ast.ArrayAccessNode) *ast.IdentifierNode {
	switch a := node.Array.(type) {
	case *ast.IdentifierNode:
		return a
	case *ast.ArrayAccessNode:
		return g.GetArrayAccessBase(a)
	default:
		return nil
	}
}

func (g *CodeGenerator) EmitName(id *ast.IdentifierNode) string {
	if id != nil && id.EmitName != "" {
		return id.EmitName
	}
	if id != nil {
		return id.Name
	}
	return ""
}

func (g *CodeGenerator) VisitArrayAccess(node *ast.ArrayAccessNode) {
	line := node.GetLine()
	base := g.GetArrayAccessBase(node)
	baseName := g.EmitName(base)
	dims := 1
	if base != nil && g.arrayDimensions != nil && baseName != "" {
		if d, ok := g.arrayDimensions[baseName]; ok {
			dims = d
		}
	}
	if id, ok := node.Array.(*ast.IdentifierNode); ok {
		g.VisitNode(node.Array)
		bn := g.EmitName(id)
		if bn == "" {
			bn = id.Name
		}
		if dims == 2 {
			g.Write("[cortex_bounds_check(" + bn + "_rows, ")
		} else {
			g.Write("[cortex_bounds_check(" + bn + "_len, ")
		}
		g.VisitNode(node.Index)
		g.Write(fmt.Sprintf(", %d)]", line))
		return
	}
	// Chained: arr[i][j]; node.Array is inner ArrayAccessNode
	g.VisitNode(node.Array)
	if dims == 2 && base != nil {
		bn := g.EmitName(base)
		if bn == "" {
			bn = base.Name
		}
		g.Write("[cortex_bounds_check(" + bn + "_cols, ")
		g.VisitNode(node.Index)
		g.Write(fmt.Sprintf(", %d)]", line))
	} else {
		g.Write("[")
		g.VisitNode(node.Index)
		g.Write("]")
	}
}

func (g *CodeGenerator) VisitIndexExpr(node *ast.IndexExprNode) {
	// Simple array indexing: obj[index]
	g.VisitNode(node.Object)
	g.Write("[")
	g.VisitNode(node.Index)
	g.Write("]")
}

func (g *CodeGenerator) VisitMemberAccess(node *ast.MemberAccessNode) {
	g.VisitNode(node.Object)
	g.Write(fmt.Sprintf(".%s", node.Member))
}

func (g *CodeGenerator) ConvertType(cortexType string) string {
	switch cortexType {
	case "void":
		return "void"
	case "int":
		return "int"
	case "float":
		return "float"
	case "double":
		return "double"
	case "number":
		return "double" // number pseudo-type maps to double in C
	case "char":
		return "char"
	case "bool":
		return "int" // C bool is often int
	case "string":
		return "char*" // Automatically convert string to char*
	case "any":
		return "AnyValue" // Handle any as AnyValue struct for runtime type support
	case "array":
		return "cortex_array*" // Dynamic array type
	case "dict":
		return "cortex_dict*" // Dictionary type
	case "gui_window":
		return "gui_window" // GUI window handle
	case "gui_widget":
		return "gui_widget" // GUI widget handle
	case "result", "result_ok", "result_err":
		return "cortex_result" // Result type
	default:
		if strings.HasPrefix(cortexType, "array_") {
			baseType := strings.TrimPrefix(cortexType, "array_")
			return g.ConvertType(baseType) + "*" // Convert array to pointer
		} else if strings.HasPrefix(cortexType, "struct_") {
			return cortexType + "*" // Convert struct to pointer
		} else if strings.HasPrefix(cortexType, "fn_") {
			// Handle function pointers
			return "void*" // Placeholder for function pointer conversion
		}
		return cortexType // Fallback to original type
	}
}

func (g *CodeGenerator) VisitYieldStmt(node *ast.YieldStmtNode) {
	// Yield pauses the coroutine and returns control to the caller
	g.Write("co_yield();")
}

func (g *CodeGenerator) VisitAwaitExpr(node *ast.AwaitExprNode) {
	// await expr -> async_await(expr)
	g.Write("async_await(")
	g.VisitNode(node.Expr)
	g.Write(")")
}

func (g *CodeGenerator) VisitSpawnStmt(node *ast.SpawnStmtNode) {
	// spawn fn(args) -> thread_spawn(fn, args)
	// spawn var = fn(args) -> cortex_thread var = thread_spawn(fn, args)
	if node.ThreadVar != "" {
		g.Write("cortex_thread " + node.ThreadVar + " = ")
	}
	g.Write("thread_spawn((void(*)(void*))")
	g.VisitNode(node.Function)
	g.Write(", ")
	if len(node.Arguments) == 0 {
		g.Write("NULL")
	} else if len(node.Arguments) == 1 {
		g.VisitNode(node.Arguments[0])
	} else {
		// Multiple args need a struct - for now use NULL
		g.Write("NULL")
	}
	g.Write(");")
}

func (g *CodeGenerator) Write(text string) {
	if g.outputTarget != nil {
		g.outputTarget.WriteString(text)
	} else {
		g.output.WriteString(text)
	}
}

func (g *CodeGenerator) Indent() {
	g.indentation++
	g.Write(strings.Repeat("    ", g.indentation))
}

func (g *CodeGenerator) Dedent() {
	g.indentation--
	if g.indentation < 0 {
		g.indentation = 0
	}
}
