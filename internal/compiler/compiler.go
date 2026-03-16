package compiler

import (
	"cortex/internal/ast"
	"cortex/internal/clibs"
	"cortex/internal/config"
	"cortex/internal/errors"
	"cortex/internal/optimizer"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// findRuntimeDir returns the directory containing runtime/core.c (for includes and core.c).
// Checks: CORTEX_ROOT env, then "runtime" under cwd, then "runtime" next to the executable.
func FindRuntimeDir() string {
	if root := os.Getenv("CORTEX_ROOT"); root != "" {
		if p := filepath.Join(root, "runtime"); dirHasCoreC(p) {
			return p
		}
	}
	if cwd, _ := os.Getwd(); cwd != "" {
		if p := filepath.Join(cwd, "runtime"); dirHasCoreC(p) {
			return p
		}
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		if p := filepath.Join(dir, "runtime"); dirHasCoreC(p) {
			return p
		}
		if p := filepath.Join(dir, "..", "runtime"); dirHasCoreC(p) {
			return p
		}
	}
	return ""
}

func dirHasCoreC(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "core.c"))
	return err == nil && !info.IsDir()
}

var networkBuiltins = map[string]bool{
	"tcp_listen": true, "tcp_accept": true, "tcp_connect": true, "tcp_send": true, "tcp_recv": true,
	"tcp_recv_string": true, "tcp_close": true, "udp_socket": true, "udp_send_to": true,
	"udp_recv_from": true, "udp_close": true, "http_get": true, "http_post": true,
	"http_get_with_header": true, "http_server_listen": true, "http_server_read_request": true,
	"http_server_send_response": true, "rpc_call": true, "net_send_message": true, "net_recv_message": true,
}

var guiBuiltins = map[string]bool{
	"gui_window_create": true, "gui_window_show": true, "gui_window_hide": true, "gui_window_close": true,
	"gui_window_set_title": true, "gui_window_center": true, "gui_window_set_fixed_size": true,
	"gui_window_set_fullscreen": true, "gui_window_set_content": true, "gui_label_create": true,
	"gui_label_set_text": true, "gui_button_create": true, "gui_entry_create": true,
	"gui_entry_get_text": true, "gui_entry_set_text": true, "gui_textarea_create": true,
	"gui_textarea_get_text": true, "gui_textarea_set_text": true, "gui_checkbox_create": true,
	"gui_checkbox_get_state": true, "gui_checkbox_set_state": true, "gui_slider_create": true,
	"gui_slider_get_value": true, "gui_slider_set_value": true, "gui_progress_create": true,
	"gui_progress_set_value": true, "gui_image_create": true, "gui_image_set_fill": true,
	"gui_rectangle_create": true, "gui_circle_create": true, "gui_line_create": true,
	"gui_line_set_color": true, "gui_vbox_create": true, "gui_hbox_create": true,
	"gui_grid_create": true, "gui_container_add": true, "gui_dialog_info": true,
	"gui_dialog_error": true, "gui_dialog_confirm": true, "gui_dialog_file_open": true,
	"gui_dialog_file_save": true, "gui_refresh": true, "gui_resize": true, "gui_move": true,
	"gui_enable": true, "gui_disable": true, "gui_is_enabled": true, "gui_run": true,
	"gui_quit": true,
}

// asyncBuiltins are functions that require runtime/async.h
var asyncBuiltins = map[string]bool{
	"co_create": true, "co_resume": true, "co_yield": true, "co_free": true, "co_current": true, "co_finished": true,
	"async_create": true, "async_await": true, "async_is_complete": true, "async_run_all": true,
}

// threadBuiltins are functions that require runtime/thread.h
var threadBuiltins = map[string]bool{
	"thread_spawn": true, "thread_join": true, "thread_is_running": true, "thread_id": true, "thread_sleep_ms": true,
	"channel_create": true, "channel_of": true, "channel_send": true, "channel_recv": true,
	"channel_try_send": true, "channel_try_recv": true, "channel_close": true, "channel_is_closed": true, "channel_free": true,
}

func usesNetworkBuiltins(node ast.ASTNode) bool {
	switch n := node.(type) {
	case *ast.ProgramNode:
		for _, d := range n.Declarations {
			if usesNetworkBuiltins(d) {
				return true
			}
		}
		return false
	case *ast.FunctionDeclNode:
		if n.Body != nil {
			return usesNetworkBuiltins(n.Body)
		}
		return false
	case *ast.BlockNode:
		for _, s := range n.Statements {
			if usesNetworkBuiltins(s) {
				return true
			}
		}
		return false
	case *ast.CallExprNode:
		if id, ok := n.Function.(*ast.IdentifierNode); ok {
			return networkBuiltins[id.Name]
		}
		return false
	case *ast.IfStmtNode:
		return usesNetworkBuiltins(n.Condition) || (n.ThenBranch != nil && usesNetworkBuiltins(n.ThenBranch)) || (n.ElseBranch != nil && usesNetworkBuiltins(n.ElseBranch))
	case *ast.WhileStmtNode:
		return usesNetworkBuiltins(n.Condition) || (n.Body != nil && usesNetworkBuiltins(n.Body))
	case *ast.DoWhileStmtNode:
		return usesNetworkBuiltins(n.Condition) || (n.Body != nil && usesNetworkBuiltins(n.Body))
	case *ast.ForStmtNode:
		u := n.Body != nil && usesNetworkBuiltins(n.Body)
		if n.Initializer != nil {
			u = u || usesNetworkBuiltins(n.Initializer)
		}
		if n.Condition != nil {
			u = u || usesNetworkBuiltins(n.Condition)
		}
		if n.Increment != nil {
			u = u || usesNetworkBuiltins(n.Increment)
		}
		return u
	case *ast.ReturnStmtNode:
		if n.Value != nil {
			return usesNetworkBuiltins(n.Value)
		}
		return false
	case *ast.AssignmentNode:
		return usesNetworkBuiltins(n.Target) || usesNetworkBuiltins(n.Value)
	case *ast.BinaryExprNode:
		return usesNetworkBuiltins(n.Left) || usesNetworkBuiltins(n.Right)
	case *ast.UnaryExprNode:
		return usesNetworkBuiltins(n.Operand)
	case *ast.VariableDeclNode:
		if n.Initializer != nil {
			return usesNetworkBuiltins(n.Initializer)
		}
		return false
	case *ast.StructDeclNode:
		for _, m := range n.Methods {
			if usesNetworkBuiltins(m) {
				return true
			}
		}
		return false
	case *ast.ForInStmtNode:
		return (n.Collection != nil && usesNetworkBuiltins(n.Collection)) || (n.Body != nil && usesNetworkBuiltins(n.Body))
	case *ast.RepeatStmtNode:
		return n.Body != nil && usesNetworkBuiltins(n.Body)
	case *ast.MatchStmtNode:
		u := n.Value != nil && usesNetworkBuiltins(n.Value)
		for _, c := range n.Cases {
			if usesNetworkBuiltins(c) {
				u = true
				break
			}
		}
		return u
	case *ast.CaseClauseNode:
		if n.Body != nil {
			return usesNetworkBuiltins(n.Body)
		}
		return false
	case *ast.SwitchStmtNode:
		u := n.Value != nil && usesNetworkBuiltins(n.Value)
		for _, c := range n.Cases {
			if usesNetworkBuiltins(c) {
				u = true
				break
			}
		}
		return u
	case *ast.MemberAccessNode:
		return usesNetworkBuiltins(n.Object)
	case *ast.ArrayAccessNode:
		return usesNetworkBuiltins(n.Array) || usesNetworkBuiltins(n.Index)
	case *ast.DeferStmtNode:
		return n.Body != nil && usesNetworkBuiltins(n.Body)
	default:
		return false
	}
}

func usesGuiBuiltins(node ast.ASTNode) bool {
	switch n := node.(type) {
	case *ast.ProgramNode:
		for _, d := range n.Declarations {
			if usesGuiBuiltins(d) {
				return true
			}
		}
		return false
	case *ast.FunctionDeclNode:
		if n.Body != nil {
			return usesGuiBuiltins(n.Body)
		}
		return false
	case *ast.BlockNode:
		for _, s := range n.Statements {
			if usesGuiBuiltins(s) {
				return true
			}
		}
		return false
	case *ast.CallExprNode:
		if id, ok := n.Function.(*ast.IdentifierNode); ok {
			return guiBuiltins[id.Name]
		}
		return false
	case *ast.IfStmtNode:
		return usesGuiBuiltins(n.Condition) || (n.ThenBranch != nil && usesGuiBuiltins(n.ThenBranch)) || (n.ElseBranch != nil && usesGuiBuiltins(n.ElseBranch))
	case *ast.WhileStmtNode:
		return usesGuiBuiltins(n.Condition) || (n.Body != nil && usesGuiBuiltins(n.Body))
	case *ast.DoWhileStmtNode:
		return usesGuiBuiltins(n.Condition) || (n.Body != nil && usesGuiBuiltins(n.Body))
	case *ast.ForStmtNode:
		u := n.Body != nil && usesGuiBuiltins(n.Body)
		if n.Initializer != nil {
			u = u || usesGuiBuiltins(n.Initializer)
		}
		if n.Condition != nil {
			u = u || usesGuiBuiltins(n.Condition)
		}
		if n.Increment != nil {
			u = u || usesGuiBuiltins(n.Increment)
		}
		return u
	case *ast.ReturnStmtNode:
		if n.Value != nil {
			return usesGuiBuiltins(n.Value)
		}
		return false
	case *ast.AssignmentNode:
		return usesGuiBuiltins(n.Target) || usesGuiBuiltins(n.Value)
	case *ast.BinaryExprNode:
		return usesGuiBuiltins(n.Left) || usesGuiBuiltins(n.Right)
	case *ast.UnaryExprNode:
		return usesGuiBuiltins(n.Operand)
	case *ast.VariableDeclNode:
		if n.Initializer != nil {
			return usesGuiBuiltins(n.Initializer)
		}
		return false
	case *ast.StructDeclNode:
		for _, m := range n.Methods {
			if usesGuiBuiltins(m) {
				return true
			}
		}
		return false
	case *ast.ForInStmtNode:
		return (n.Collection != nil && usesGuiBuiltins(n.Collection)) || (n.Body != nil && usesGuiBuiltins(n.Body))
	case *ast.RepeatStmtNode:
		return n.Body != nil && usesGuiBuiltins(n.Body)
	case *ast.MatchStmtNode:
		u := n.Value != nil && usesGuiBuiltins(n.Value)
		for _, c := range n.Cases {
			if usesGuiBuiltins(c) {
				u = true
				break
			}
		}
		return u
	case *ast.CaseClauseNode:
		if n.Body != nil {
			return usesGuiBuiltins(n.Body)
		}
		return false
	case *ast.SwitchStmtNode:
		u := n.Value != nil && usesGuiBuiltins(n.Value)
		for _, c := range n.Cases {
			if usesGuiBuiltins(c) {
				u = true
				break
			}
		}
		return u
	case *ast.MemberAccessNode:
		return usesGuiBuiltins(n.Object)
	case *ast.ArrayAccessNode:
		return usesGuiBuiltins(n.Array) || usesGuiBuiltins(n.Index)
	case *ast.DeferStmtNode:
		return n.Body != nil && usesGuiBuiltins(n.Body)
	default:
		return false
	}
}

func usesAsyncBuiltins(node ast.ASTNode) bool {
	switch n := node.(type) {
	case *ast.ProgramNode:
		for _, d := range n.Declarations {
			if usesAsyncBuiltins(d) {
				return true
			}
		}
		return false
	case *ast.FunctionDeclNode:
		// Coroutine functions need async runtime
		if n.IsCoroutine {
			return true
		}
		if n.Body != nil {
			return usesAsyncBuiltins(n.Body)
		}
		return false
	case *ast.BlockNode:
		for _, s := range n.Statements {
			if usesAsyncBuiltins(s) {
				return true
			}
		}
		return false
	case *ast.CallExprNode:
		if id, ok := n.Function.(*ast.IdentifierNode); ok {
			return asyncBuiltins[id.Name]
		}
		return false
	case *ast.YieldStmtNode:
		// Yield statements need async runtime
		return true
	case *ast.IfStmtNode:
		return usesAsyncBuiltins(n.Condition) || (n.ThenBranch != nil && usesAsyncBuiltins(n.ThenBranch)) || (n.ElseBranch != nil && usesAsyncBuiltins(n.ElseBranch))
	case *ast.WhileStmtNode:
		return usesAsyncBuiltins(n.Condition) || (n.Body != nil && usesAsyncBuiltins(n.Body))
	case *ast.ForStmtNode:
		u := n.Body != nil && usesAsyncBuiltins(n.Body)
		if n.Initializer != nil {
			u = u || usesAsyncBuiltins(n.Initializer)
		}
		if n.Condition != nil {
			u = u || usesAsyncBuiltins(n.Condition)
		}
		if n.Increment != nil {
			u = u || usesAsyncBuiltins(n.Increment)
		}
		return u
	case *ast.ReturnStmtNode:
		if n.Value != nil {
			return usesAsyncBuiltins(n.Value)
		}
		return false
	case *ast.AssignmentNode:
		return usesAsyncBuiltins(n.Target) || usesAsyncBuiltins(n.Value)
	case *ast.VariableDeclNode:
		if n.Initializer != nil {
			return usesAsyncBuiltins(n.Initializer)
		}
		return false
	default:
		return false
	}
}

func usesThreadBuiltins(node ast.ASTNode) bool {
	switch n := node.(type) {
	case *ast.ProgramNode:
		for _, d := range n.Declarations {
			if usesThreadBuiltins(d) {
				return true
			}
		}
		return false
	case *ast.FunctionDeclNode:
		if n.Body != nil {
			return usesThreadBuiltins(n.Body)
		}
		return false
	case *ast.BlockNode:
		for _, s := range n.Statements {
			if usesThreadBuiltins(s) {
				return true
			}
		}
		return false
	case *ast.SpawnStmtNode:
		// Spawn statements need thread runtime
		return true
	case *ast.IfStmtNode:
		return usesThreadBuiltins(n.Condition) || (n.ThenBranch != nil && usesThreadBuiltins(n.ThenBranch)) || (n.ElseBranch != nil && usesThreadBuiltins(n.ElseBranch))
	case *ast.ForStmtNode:
		u := n.Body != nil && usesThreadBuiltins(n.Body)
		if n.Initializer != nil {
			u = u || usesThreadBuiltins(n.Initializer)
		}
		if n.Condition != nil {
			u = u || usesThreadBuiltins(n.Condition)
		}
		if n.Increment != nil {
			u = u || usesThreadBuiltins(n.Increment)
		}
		return u
	case *ast.ReturnStmtNode:
		if n.Value != nil {
			return usesThreadBuiltins(n.Value)
		}
		return false
	case *ast.AssignmentNode:
		return usesThreadBuiltins(n.Target) || usesThreadBuiltins(n.Value)
	case *ast.VariableDeclNode:
		if n.Initializer != nil {
			return usesThreadBuiltins(n.Initializer)
		}
		return false
	case *ast.CallExprNode:
		// Check if calling a thread/channel builtin
		if id, ok := n.Function.(*ast.IdentifierNode); ok {
			if threadBuiltins[id.Name] {
				return true
			}
		}
		// Check arguments
		for _, arg := range n.Args {
			if usesThreadBuiltins(arg) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// usesManagedBuiltins checks if the AST uses cleanup annotations (requires managed.h)
func usesManagedBuiltins(node ast.ASTNode) bool {
	switch n := node.(type) {
	case *ast.ProgramNode:
		for _, d := range n.Declarations {
			if usesManagedBuiltins(d) {
				return true
			}
		}
		return false
	case *ast.ExternDeclNode:
		// Extern with cleanup annotation needs managed runtime
		return n.CleanupFunc != ""
	default:
		return false
	}
}

type Compiler struct {
	lexer       *Lexer
	parser      *Parser
	analyzer    *SemanticAnalyzer
	generator   *CodeGenerator
	config      config.Config
	diagnostics *errors.Collector
	includes    []string
	libraries   []string
}

// SetDiagnosticsCollector sets the optional collector for structured diagnostics (line, column, code, suggestion).
// When set, lex/semantic (and parse via recover) errors are added here.
func (c *Compiler) SetDiagnosticsCollector(col *errors.Collector) {
	c.diagnostics = col
}

func NewCompiler(cfg config.Config) *Compiler {
	astCfg := ast.Config{
		Backend:      cfg.Backend,
		IncludePaths: cfg.IncludePaths,
		LibraryPaths: cfg.LibraryPaths,
		Libraries:    cfg.Libraries,
	}
	astCfg.Features.Async = cfg.Features.Async
	astCfg.Features.Actors = cfg.Features.Actors
	astCfg.Features.Blockchain = cfg.Features.Blockchain
	astCfg.Features.QoL = cfg.Features.QoL
	return &Compiler{
		lexer:     NewLexer(),
		parser:    NewParser(cfg),
		analyzer:  NewSemanticAnalyzer(cfg),
		generator: NewCodeGenerator(astCfg),
		config:    cfg,
	}
}

// Compile compiles a single source file to an executable.
func (c *Compiler) Compile(inputFile, outputFile string) error {
	return c.CompileMulti([]string{inputFile}, outputFile)
}

// GenerateC runs parse, analyze, and codegen and returns the generated C source (no file write, no gcc).
// Useful for tests that need to assert on generated code.
func (c *Compiler) GenerateC(inputFiles []string) (string, error) {
	if len(inputFiles) == 0 {
		return "", fmt.Errorf("no input files")
	}
	expanded, err := c.resolveImportPaths(inputFiles)
	if err != nil {
		return "", err
	}
	inputFiles = expanded
	if c.diagnostics != nil {
		c.analyzer.SetDiagnosticsCollector(c.diagnostics)
	}
	merged := &ast.ProgramNode{
		BaseNode:     ast.BaseNode{Type: ast.NodeProgram, Line: 1, Column: 1},
		Declarations: nil,
	}
	for _, inputFile := range inputFiles {
		source, err := os.ReadFile(inputFile)
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", inputFile, err)
		}
		s := string(source)
		if len(s) >= 3 && s[0] == '\xef' && s[1] == '\xbb' && s[2] == '\xbf' {
			s = s[3:]
		}
		tokens, err := c.lexer.Tokenize(s)
		if err != nil {
			if c.diagnostics != nil {
				c.diagnostics.AddError(errors.ErrLexUnexpectedChar, 0, 0, err.Error(), "Check for invalid characters or unterminated string.")
			}
			return "", fmt.Errorf("%s: lexical error: %w", inputFile, err)
		}
		var astNode ast.ASTNode
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%v", r)
					if c.diagnostics != nil {
						c.diagnostics.AddError(errors.ErrParseUnexpected, 0, 0, fmt.Sprint(r), "Check syntax (e.g. missing semicolon, bracket).")
					}
				}
			}()
			astNode, err = c.parser.Parse(tokens)
		}()
		if err != nil {
			return "", fmt.Errorf("%s: syntax error: %w", inputFile, err)
		}
		if astNode == nil {
			return "", fmt.Errorf("%s: parse failed", inputFile)
		}
		prog, ok := astNode.(*ast.ProgramNode)
		if !ok {
			return "", fmt.Errorf("%s: expected program node", inputFile)
		}
		merged.Declarations = append(merged.Declarations, prog.Declarations...)
	}
	if err := c.analyzer.Analyze(merged); err != nil {
		msg := fmt.Sprintf("Compilation error: %v", err)
		for _, semanticErr := range c.analyzer.GetErrors() {
			msg += fmt.Sprintf("\n  - %v", semanticErr)
		}
		return "", fmt.Errorf("%s", msg)
	}
	merged = optimizer.Run(merged, optimizer.Options{ConstantFolding: true}).(*ast.ProgramNode)
	c.generator.SetUsesNetwork(usesNetworkBuiltins(merged))
	c.generator.SetAutoExternFunctions(c.analyzer.GetAutoExternFunctions())
	code, err := c.generator.Generate(merged)
	if err != nil {
		return "", fmt.Errorf("code generation error: %w", err)
	}
	return code, nil
}

// resolveImportPaths expands inputFiles with any files referenced by import "path"; in the source.
// Resolves paths relative to the directory of the file containing the import.
// Lookup order: same_dir/path.cx, same_dir/path/mod.cx.
func (c *Compiler) resolveImportPaths(inputFiles []string) ([]string, error) {
	seen := make(map[string]bool)
	var ordered []string
	queue := make([]string, 0, len(inputFiles))
	for _, f := range inputFiles {
		abs, err := filepath.Abs(f)
		if err != nil {
			abs = f
		}
		if !seen[abs] {
			seen[abs] = true
			ordered = append(ordered, abs)
			queue = append(queue, abs)
		}
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		source, err := os.ReadFile(current)
		if err != nil {
			continue
		}
		s := string(source)
		if len(s) >= 3 && s[0] == '\xef' && s[1] == '\xbb' && s[2] == '\xbf' {
			s = s[3:]
		}
		tokens, err := c.lexer.Tokenize(s)
		if err != nil {
			continue
		}
		astNode, err := c.parser.Parse(tokens)
		if err != nil {
			continue
		}
		prog, ok := astNode.(*ast.ProgramNode)
		if !ok {
			continue
		}
		dir := filepath.Dir(current)
		for _, decl := range prog.Declarations {
			if imp, ok := decl.(*ast.ImportNode); ok {
				path := strings.Trim(imp.Path, `"`)
				resolved := c.resolveOneImport(path, dir)
				if resolved != "" {
					abs, _ := filepath.Abs(resolved)
					if abs == "" {
						abs = resolved
					}
					if !seen[abs] {
						seen[abs] = true
						ordered = append(ordered, abs)
						queue = append(queue, abs)
					}
				}
			}
		}
	}
	return ordered, nil
}

func (c *Compiler) resolveOneImport(importPath, fromDir string) string {
	try := []string{
		filepath.Join(fromDir, importPath+".cx"),
		filepath.Join(fromDir, importPath, "mod.cx"),
	}
	for _, p := range try {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

// CompileMulti compiles one or more source files into a single executable.
// All files are parsed and merged into one program (shared global scope).
// If sources contain import "path";, those files are resolved and merged automatically.
func (c *Compiler) CompileMulti(inputFiles []string, outputFile string) error {
	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files")
	}
	expanded, err := c.resolveImportPaths(inputFiles)
	if err != nil {
		return err
	}
	inputFiles = expanded
	merged := &ast.ProgramNode{
		BaseNode:     ast.BaseNode{Type: ast.NodeProgram, Line: 1, Column: 1},
		Declarations: nil,
	}
	for _, inputFile := range inputFiles {
		source, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", inputFile, err)
		}
		s := string(source)
		if len(s) >= 3 && s[0] == '\xef' && s[1] == '\xbb' && s[2] == '\xbf' {
			s = s[3:]
		}
		tokens, err := c.lexer.Tokenize(s)
		if err != nil {
			return fmt.Errorf("%s: lexical error: %w", inputFile, err)
		}
		astNode, err := c.parser.Parse(tokens)
		if err != nil {
			return fmt.Errorf("%s: syntax error: %w", inputFile, err)
		}
		prog, ok := astNode.(*ast.ProgramNode)
		if !ok {
			return fmt.Errorf("%s: expected program node", inputFile)
		}
		merged.Declarations = append(merged.Declarations, prog.Declarations...)
	}
	program := merged

	if err := c.analyzer.Analyze(program); err != nil {
		msg := fmt.Sprintf("Compilation error: %v", err)
		for _, semanticErr := range c.analyzer.GetErrors() {
			msg += fmt.Sprintf("\n  - %v", semanticErr)
		}
		return fmt.Errorf("%s", msg)
	}
	program = optimizer.Run(program, optimizer.Options{ConstantFolding: true}).(*ast.ProgramNode)
	usesNet := usesNetworkBuiltins(program)
	c.generator.SetUsesNetwork(usesNet)
	c.generator.SetUsesGui(usesGuiBuiltins(program))
	c.generator.SetUsesAsync(usesAsyncBuiltins(program))
	c.generator.SetUsesThread(usesThreadBuiltins(program))
	c.generator.SetUsesManaged(usesManagedBuiltins(program))
	code, err := c.generator.Generate(program)
	if err != nil {
		return fmt.Errorf("code generation error: %w", err)
	}

	cFile := strings.TrimSuffix(outputFile, filepath.Ext(outputFile)) + ".c"
	if err := os.WriteFile(cFile, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write C file: %w", err)
	}

	c.collectIncludes(program)
	c.inferLibraries()

	linkLibs := c.libraries
	usesGui := usesGuiBuiltins(program)
	usesAsync := usesAsyncBuiltins(program)
	usesThread := usesThreadBuiltins(program)
	return c.compileCCode(cFile, outputFile, linkLibs, usesNet, usesGui, usesAsync, usesThread)
}

func (c *Compiler) collectIncludes(program *ast.ProgramNode) {
	for _, decl := range program.Declarations {
		if include, ok := decl.(*ast.IncludeNode); ok {
			c.includes = append(c.includes, include.Filename)
		}
	}
}

func (c *Compiler) inferLibraries() {
	for _, header := range c.includes {
		lib := clibs.InferLibraryFromHeader(header)
		if lib != "" {
			c.libraries = append(c.libraries, lib)
		}
	}
	c.libraries = clibs.DedupeLibraries(c.libraries)
}

func linksRaylib(libs []string, pragmas []string) bool {
	for _, l := range libs {
		if l == "raylib" {
			return true
		}
	}
	for _, p := range pragmas {
		if p == "raylib" {
			return true
		}
	}
	return false
}

// standardCHeaders are include names that are part of the C runtime; we do not infer -l for them.
var standardCHeaders = map[string]bool{
	"stdio.h": true, "stdlib.h": true, "string.h": true, "time.h": true,
	"stdbool.h": true, "math.h": true, "stddef.h": true, "limits.h": true,
	"ctype.h": true, "errno.h": true, "assert.h": true, "signal.h": true,
	"gui_runtime.h": true, "core.h": true, "game.h": true, "network.h": true,
	"async.h": true, "thread.h": true, "managed.h": true,
}

// collectLinkPragmas returns library names from #pragma link, #use, and #include <name.h> (C-style: include implies -l name).
func collectLinkPragmas(node ast.ASTNode) []string {
	program, ok := node.(*ast.ProgramNode)
	if !ok {
		return nil
	}
	seen := make(map[string]bool)
	var libs []string
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			libs = append(libs, name)
		}
	}
	for _, decl := range program.Declarations {
		if p, ok := decl.(*ast.PragmaNode); ok && p.Directive == "link" && p.Content != "" {
			name := strings.TrimSpace(p.Content)
			name = strings.Trim(name, "\"")
			add(name)
		}
		if u, ok := decl.(*ast.UseLibNode); ok && u.LibName != "" {
			add(u.LibName)
		}
		if inc, ok := decl.(*ast.IncludeNode); ok && inc.Filename != "" {
			base := filepath.Base(inc.Filename)
			if strings.HasSuffix(base, ".h") {
				if !standardCHeaders[base] {
					lib := strings.TrimSuffix(base, ".h")
					add(lib)
				}
			}
		}
	}
	return libs
}

func (c *Compiler) compileCCode(cFile, outputFile string, linkPragmas []string, usesNetwork bool, usesGui bool, usesAsync bool, usesThread bool) error {
	runtimeDir := FindRuntimeDir()
	if runtimeDir == "" {
		return fmt.Errorf("could not find runtime directory (run from project root or set CORTEX_ROOT)")
	}
	runtimeSource := filepath.Join(runtimeDir, "core.c")
	gameSource := filepath.Join(runtimeDir, "game.c")
	includeDir := filepath.Dir(runtimeDir)

	args := []string{}
	if isWindows() {
		args = append(args, "-mconsole") // Force console mode so main() works
	}
	args = append(args, "-I", includeDir)
	for _, p := range c.config.IncludePaths {
		args = append(args, "-I", p)
	}
	for _, p := range c.config.LibraryPaths {
		args = append(args, "-L", p)
	}
	args = append(args, featureDefineArgs(c.config.Features)...)
	args = append(args, "-o", outputFile, cFile, runtimeSource)
	if _, err := os.Stat(gameSource); err == nil {
		args = append(args, gameSource)
	}
	if usesNetwork {
		networkSource := filepath.Join(runtimeDir, "network.c")
		if _, err := os.Stat(networkSource); err == nil {
			args = append(args, networkSource)
			if isWindows() {
				args = append(args, "-lws2_32")
			}
		}
	}
	if usesGui {
		// Add runtime directory to include path for gui_runtime.h
		args = append(args, "-I", runtimeDir)
		// Use GTK4 on all platforms
		if isWindows() {
			// MSYS2 GTK4 paths on Windows
			msysInclude := "C:/msys64/mingw64/include"
			msysLib := "C:/msys64/mingw64/lib"
			args = append(args, "-I", msysInclude+"/gtk-4.0")
			args = append(args, "-I", msysInclude+"/pango-1.0")
			args = append(args, "-I", msysInclude+"/cairo")
			args = append(args, "-I", msysInclude+"/gdk-pixbuf-2.0")
			args = append(args, "-I", msysInclude+"/graphene-1.0")
			args = append(args, "-I", msysLib+"/graphene-1.0/include")
			args = append(args, "-I", msysInclude+"/glib-2.0")
			args = append(args, "-I", msysLib+"/glib-2.0/include")
			args = append(args, "-I", msysInclude+"/harfbuzz")
			args = append(args, "-I", msysInclude+"/freetype2")
			args = append(args, "-I", msysInclude+"/libpng16")
			args = append(args, "-I", msysInclude+"/pixman-1")
			args = append(args, "-L", msysLib)
		}
		// Add GTK4 GUI implementation for all platforms
		guiGtkDir := filepath.Join(runtimeDir, "..", "internal", "gui_gtk4")
		gtkSources := []string{"gui_core.c", "gui_widgets.c", "gui_containers.c", "gui_dialogs.c"}
		for _, src := range gtkSources {
			srcPath := filepath.Join(guiGtkDir, src)
			if _, err := os.Stat(srcPath); err == nil {
				args = append(args, srcPath)
			}
		}
		// Link GTK4 libraries
		args = append(args, "-lgtk-4", "-lpangocairo-1.0", "-lpangowin32-1.0", "-lpango-1.0",
			"-lgdk_pixbuf-2.0", "-lcairo-gobject", "-lcairo", "-lharfbuzz", "-lvulkan-1",
			"-lgraphene-1.0", "-lgio-2.0", "-lgobject-2.0", "-lglib-2.0", "-lintl")
	}
	if usesAsync {
		asyncSource := filepath.Join(runtimeDir, "async.c")
		if _, err := os.Stat(asyncSource); err == nil {
			args = append(args, asyncSource)
		}
	}
	if usesThread {
		threadSource := filepath.Join(runtimeDir, "thread.c")
		if _, err := os.Stat(threadSource); err == nil {
			args = append(args, threadSource)
		}
	}
	helperSource := filepath.Join(runtimeDir, "raylib_helper.c")
	if linksRaylib(c.config.Libraries, linkPragmas) {
		if _, err := os.Stat(helperSource); err == nil {
			args = append(args, helperSource)
		}
	}
	for _, lib := range c.config.Libraries {
		args = append(args, "-l"+lib)
	}
	for _, lib := range linkPragmas {
		args = append(args, "-l"+lib)
	}
	for _, lib := range c.libraries {
		args = append(args, "-l"+lib)
	}

	backend := c.config.Backend
	if backend == "" {
		backend = "auto"
	}

	// Priority: zig cc (bundled), zig (PATH), tcc, gcc
	useZig := backend == "zig" || (backend == "auto")
	useTcc := backend == "tcc" || (backend == "auto")

	// Try Zig CC first
	if useZig {
		if zigPath := findBundledZig(); zigPath != "" {
			args = append(args, "-lm")
			cmd := exec.Command(zigPath, append([]string{"cc"}, args...)...)
			output, err := cmd.CombinedOutput()
			os.Remove(cFile)
			if err != nil {
				return fmt.Errorf("Zig CC compilation failed: %v\nOutput: %s", err, string(output))
			}
			return nil
		}
		if zigExe := findZig(); zigExe != "" {
			args = append(args, "-lm")
			cmd := exec.Command(zigExe, append([]string{"cc"}, args...)...)
			output, err := cmd.CombinedOutput()
			os.Remove(cFile)
			if err != nil {
				return fmt.Errorf("Zig CC compilation failed: %v\nOutput: %s", err, string(output))
			}
			return nil
		}
	}

	// Try TCC
	if useTcc {
		if tccExe := findTcc(); tccExe != "" {
			args = append(args, "-lm")
			cmd := exec.Command(tccExe, args...)
			output, err := cmd.CombinedOutput()
			os.Remove(cFile)
			if err != nil {
				return fmt.Errorf("TCC compilation failed: %v\nOutput: %s", err, string(output))
			}
			return nil
		}
		if backend == "tcc" {
			return fmt.Errorf("backend=tcc but tcc not found (put tcc.exe in tools/ next to cortex or in PATH)")
		}
	}

	// Fallback to GCC
	args = append(args, "-lm")
	cmd := exec.Command("gcc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("C compilation failed: %v\nOutput: %s", err, string(output))
	}

	os.Remove(cFile)
	return nil
}

// findTcc returns the path to tcc (tcc.exe on Windows) if found: next to executable in tools/, same dir, then PATH.
func findTcc() string {
	exeName := "tcc"
	if isWindows() {
		exeName = "tcc.exe"
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for _, d := range []string{filepath.Join(dir, "tools"), dir} {
			p := filepath.Join(d, exeName)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	if p, err := exec.LookPath(exeName); err == nil {
		return p
	}
	return ""
}

// findBundledZig returns the path to bundled zig if found (relative to cortex binary)
func findBundledZig() string {
	exeName := "zig"
	if isWindows() {
		exeName = "zig.exe"
	}
	if exe, err := os.Executable(); err == nil {
		// Release structure: bin/cortex, zig/zig
		binDir := filepath.Dir(exe)
		releaseDir := filepath.Dir(binDir)
		zigPath := filepath.Join(releaseDir, "zig", exeName)
		if _, err := os.Stat(zigPath); err == nil {
			return zigPath
		}
		// Also check tools/zig
		toolsZig := filepath.Join(binDir, "tools", exeName)
		if _, err := os.Stat(toolsZig); err == nil {
			return toolsZig
		}
	}
	return ""
}

// findZig returns the path to zig in PATH
func findZig() string {
	exeName := "zig"
	if isWindows() {
		exeName = "zig.exe"
	}
	if p, err := exec.LookPath(exeName); err == nil {
		return p
	}
	return ""
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func featureDefineArgs(features config.FeatureSet) []string {
	args := []string{}
	appendFlag := func(name string, enabled bool) {
		value := "0"
		if enabled {
			value = "1"
		}
		args = append(args, fmt.Sprintf("-D%s=%s", name, value))
	}
	appendFlag("CORTEX_FEATURE_ASYNC", features.Async)
	appendFlag("CORTEX_FEATURE_ACTORS", features.Actors)
	appendFlag("CORTEX_FEATURE_BLOCKCHAIN", features.Blockchain)
	appendFlag("CORTEX_FEATURE_QOL", features.QoL)
	return args
}
