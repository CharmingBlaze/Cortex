package compiler

import (
	"cortex/internal/ast"
	"cortex/internal/config"
	"cortex/internal/errors"
	"fmt"
	"strings"
)

type SymbolType int

const (
	SymbolVariable SymbolType = iota
	SymbolFunction
	SymbolStruct
	SymbolEnum
	SymbolConst
	SymbolParameter
)

type Symbol struct {
	Name       string // Cortex name for resolution (e.g. "add")
	EmitName   string // C symbol to emit (e.g. "math__add"); if set, codegen uses this
	Type       string
	SymbolType SymbolType
	Node       ast.ASTNode
	Scope      *Scope
	Value      interface{} // For constants/enums: the constant value
}

type Scope struct {
	Parent   *Scope
	Symbols  map[string]*Symbol
	Children []*Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent:  parent,
		Symbols: make(map[string]*Symbol),
	}
}

func (s *Scope) Define(symbol *Symbol) error {
	// Case-insensitive check for existing symbol
	lowerName := strings.ToLower(symbol.Name)
	for key := range s.Symbols {
		if strings.ToLower(key) == lowerName {
			return fmt.Errorf("symbol '%s' already defined in scope", symbol.Name)
		}
	}
	s.Symbols[symbol.Name] = symbol
	symbol.Scope = s
	return nil
}

// Resolve looks up a symbol by name in this scope or parent scopes (case-insensitive).
func (s *Scope) Resolve(name string) *Symbol {
	// Case-insensitive lookup
	lowerName := strings.ToLower(name)
	for key, sym := range s.Symbols {
		if strings.ToLower(key) == lowerName {
			return sym
		}
	}
	if s.Parent != nil {
		return s.Parent.Resolve(name)
	}
	return nil
}

type SemanticAnalyzer struct {
	globalScope  *Scope
	currentScope *Scope
	errors       []error
	diagnostics  *errors.Collector // optional: when set, errors are also emitted as structured diagnostics
	features     config.FeatureSet
	hasInclude   bool
	inCoroutine  bool
	inAsync      bool
}

var blockchainBuiltins = map[string]struct{}{
	"sha256_hash":      {},
	"sha256_double":    {},
	"keccak256":        {},
	"merkle_root":      {},
	"hex_encode":       {},
	"hex_decode":       {},
	"base58_encode":    {},
	"base58_decode":    {},
	"blockheader_hash": {},
}

var qolBuiltins = map[string]struct{}{
	"make_vec2":    {},
	"make_vec3":    {},
	"dot":          {},
	"normalize":    {},
	"random_int":   {},
	"random_float": {},
	"get_time":     {},
	"sleep":        {},
	"wait":         {},
	"type_of":      {},
	"is_type":      {},
	"as_int":       {},
	"as_float":     {},
	"as_string":    {},
	"as_bool":      {},
	"as_dict":      {},
	"as_array":     {},
	"print":        {},
	"println":      {},
	"say":          {},
	"show":         {},
	"writeline":    {},
	"assert":       {},
	"clamp_float":  {},
	"sign_float":   {},
	"wrap_float":   {},
	"round_float":  {},
	"floor_float":  {},
	"ceil_float":   {},
	"array_create": {},
	"array_push":   {},
	"array_get":    {},
	"array_set":    {},
	"array_len":    {},
	"array_free":   {},
	"dict_create":  {},
	"dict_set":     {},
	"dict_get":     {},
	"dict_has":     {},
	"dict_len":     {},
	"dict_free":    {},
	"result_ok":    {},
	"result_err":   {},
	"result_is_ok": {},
	"result_value": {},
	"result_error": {},
	"array_pop":    {}, "array_insert": {}, "array_remove_at": {}, "array_capacity": {}, "array_reserve": {},
	"event_create": {}, "event_subscribe": {}, "event_unsubscribe": {}, "event_emit": {}, "event_free": {},
	"str_split": {}, "str_join": {}, "str_replace": {}, "str_trim": {},
	"starts_with": {}, "ends_with": {}, "to_lower": {}, "to_upper": {},
	"clamp_int": {}, "pow": {}, "random_choice": {},
	"file_exists": {}, "list_dir": {}, "path_join": {},
	"debug_log": {}, "debug_assert": {}, "dump": {},
	"assert_eq": {}, "assert_approx": {},
	"test_run_all": {},
	"json_parse":   {}, "json_stringify": {}, "parse_number": {}, "parse_int": {},
	"entity_create": {}, "add_component": {}, "get_component": {}, "has_component": {}, "entity_remove": {},
	"tcp_listen": {}, "tcp_accept": {}, "tcp_connect": {}, "tcp_send": {}, "tcp_recv": {}, "tcp_recv_string": {}, "tcp_close": {},
	"udp_socket": {}, "udp_send_to": {}, "udp_recv_from": {}, "udp_close": {},
	"http_get": {}, "http_post": {}, "http_get_with_header": {},
	"http_server_listen": {}, "http_server_read_request": {}, "http_server_send_response": {},
	"rpc_call": {}, "net_send_message": {}, "net_recv_message": {},
}

func NewSemanticAnalyzer(cfg config.Config) *SemanticAnalyzer {
	globalScope := NewScope(nil)
	analyzer := &SemanticAnalyzer{
		globalScope:  globalScope,
		currentScope: globalScope,
		features:     cfg.Features,
	}

	// Add built-in functions to global scope
	analyzer.RegisterBuiltins()

	return analyzer
}

func (a *SemanticAnalyzer) RegisterBuiltins() {
	// I/O
	a.globalScope.Define(&Symbol{
		Name:       "print",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "println",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "say",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "show",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "writeline",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "printf",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// JSON
	a.globalScope.Define(&Symbol{
		Name:       "json_parse",
		Type:       "dict",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "json_stringify",
		Type:       "string",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// String to number conversion
	a.globalScope.Define(&Symbol{
		Name:       "parse_number",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "parse_int",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "atoi",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Dict methods
	a.globalScope.Define(&Symbol{
		Name:       "get",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "set",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "has",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "keys",
		Type:       "array",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Array methods
	a.globalScope.Define(&Symbol{
		Name:       "push",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "pop",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "get",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "set",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "len",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Array builtins
	a.globalScope.Define(&Symbol{
		Name:       "array_create",
		Type:       "array",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "array_push",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "array_len",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "array_free",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "array_get",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "array_set",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Dict builtins
	a.globalScope.Define(&Symbol{
		Name:       "dict_create",
		Type:       "dict",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "dict_set",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "dict_get",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "dict_has",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "dict_len",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "dict_free",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Any type helpers
	a.globalScope.Define(&Symbol{
		Name:       "make_any_int",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "make_any_float",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "make_any_string",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "make_any_bool",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "as_int",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "as_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "as_string",
		Type:       "string",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "as_bool",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Type checking builtins
	a.globalScope.Define(&Symbol{
		Name:       "type_of",
		Type:       "string",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "is_type",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "as_array",
		Type:       "array",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "as_dict",
		Type:       "dict",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "make_any_dict",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// ECS builtins
	a.globalScope.Define(&Symbol{
		Name:       "entity_create",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "entity_remove",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "add_component",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "get_component",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "has_component",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// GUI builtins
	a.globalScope.Define(&Symbol{
		Name:       "gui_window_create",
		Type:       "gui_window",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "gui_window_show",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "gui_window_center",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "gui_dialog_info",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "gui_dialog_error",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "gui_dialog_confirm",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "gui_run",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "gui_label_create",
		Type:       "gui_widget",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "gui_button_create",
		Type:       "gui_widget",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "gui_container_add",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Result type builtins
	a.globalScope.Define(&Symbol{
		Name:       "result_ok",
		Type:       "result",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "result_err",
		Type:       "result",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "result_is_ok",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "result_value",
		Type:       "any",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "result_error",
		Type:       "string",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// I/O
	a.globalScope.Define(&Symbol{
		Name:       "input_line",
		Type:       "string",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "waitkey",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// File I/O
	a.globalScope.Define(&Symbol{
		Name:       "read_file",
		Type:       "string",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "write_file",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Math builtins (QoL)
	a.globalScope.Define(&Symbol{
		Name:       "make_vec2",
		Type:       "vec2",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "make_vec3",
		Type:       "vec3",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "dot",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "normalize",
		Type:       "vec2",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "vec2_length",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "vec2_distance",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "vec2_add",
		Type:       "vec2",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "vec2_sub",
		Type:       "vec2",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "vec2_scale",
		Type:       "vec2",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "clamp_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "lerp_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "min_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "max_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "random_int",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "random_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "get_time",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "sleep_func",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "round_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "floor_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "ceil_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "sign_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "wrap_float",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "assert",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Math functions
	a.globalScope.Define(&Symbol{
		Name:       "sqrt",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "srand",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "time",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "distance",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "abs",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "fabs",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "pow",
		Type:       "float",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Thread builtins
	a.globalScope.Define(&Symbol{
		Name:       "thread_spawn",
		Type:       "cortex_thread",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "thread_join",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "thread_is_running",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "thread_id",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "thread_sleep_ms",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})

	// Channel builtins
	a.globalScope.Define(&Symbol{
		Name:       "channel_create",
		Type:       "cortex_channel",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "channel_send",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "channel_recv",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "channel_try_send",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "channel_try_recv",
		Type:       "int",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "channel_close",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "channel_is_closed",
		Type:       "bool",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
	a.globalScope.Define(&Symbol{
		Name:       "channel_free",
		Type:       "void",
		SymbolType: SymbolFunction,
		Node:       nil,
	})
}

func (a *SemanticAnalyzer) Analyze(node ast.ASTNode) error {
	a.errors = []error{}
	a.currentScope = a.globalScope

	a.VisitNode(node)

	if len(a.errors) > 0 {
		return fmt.Errorf("semantic analysis found %d errors", len(a.errors))
	}
	return nil
}

func (a *SemanticAnalyzer) VisitNode(node ast.ASTNode) {
	switch n := node.(type) {
	case *ast.ProgramNode:
		a.VisitProgram(n)
	case *ast.IncludeNode:
		a.VisitInclude(n)
	case *ast.UseLibNode:
		a.VisitUseLib(n)
	case *ast.RawCNode:
		// Raw C is emitted verbatim; no semantic check
	case *ast.DefineNode:
		a.VisitDefine(n)
	case *ast.PragmaNode:
		a.VisitPragma(n)
	case *ast.LibraryNode:
		a.VisitLibrary(n)
	case *ast.ConfigNode:
		a.VisitConfig(n)
	case *ast.WrapperNode:
		a.VisitWrapper(n)
	case *ast.ExternDeclNode:
		a.VisitExternDecl(n)
	case *ast.PackageNode:
		// Package name is for tooling; no scope effect
	case *ast.ImportNode:
		// Import path is for multi-file merge; resolved at compile time
	case *ast.FunctionDeclNode:
		a.VisitFunctionDecl(n)
	case *ast.VariableDeclNode:
		a.VisitVariableDecl(n)
	case *ast.StructDeclNode:
		a.VisitStructDecl(n)
	case *ast.EnumDeclNode:
		a.VisitEnumDecl(n)
	case *ast.BlockNode:
		a.VisitBlock(n)
	case *ast.IfStmtNode:
		a.VisitIfStmt(n)
	case *ast.WhileStmtNode:
		a.VisitWhileStmt(n)
	case *ast.DoWhileStmtNode:
		a.VisitDoWhileStmt(n)
	case *ast.ForStmtNode:
		a.VisitForStmt(n)
	case *ast.ForInStmtNode:
		a.VisitForInStmt(n)
	case *ast.RepeatStmtNode:
		a.VisitRepeatStmt(n)
	case *ast.BreakStmtNode:
		a.VisitBreakStmt(n)
	case *ast.ContinueStmtNode:
		a.VisitContinueStmt(n)
	case *ast.SwitchStmtNode:
		a.VisitSwitchStmt(n)
	case *ast.TestStmtNode:
		a.VisitTestStmt(n)
	case *ast.DeferStmtNode:
		a.VisitDeferStmt(n)
	case *ast.MatchStmtNode:
		a.VisitMatchStmt(n)
	case *ast.ReturnStmtNode:
		a.VisitReturnStmt(n)
	case *ast.TupleExprNode:
		a.VisitTupleExpr(n)
	case *ast.ArrayLiteralNode:
		a.VisitArrayLiteral(n)
	case *ast.DictLiteralNode:
		a.VisitDictLiteral(n)
	case *ast.StructLiteralNode:
		a.VisitStructLiteral(n)
	case *ast.InterpolatedStringNode:
		a.VisitInterpolatedString(n)
	case *ast.BinaryExprNode:
		a.VisitBinaryExpr(n)
	case *ast.UnaryExprNode:
		a.VisitUnaryExpr(n)
	case *ast.CastExprNode:
		a.VisitCastExpr(n)
	case *ast.CallExprNode:
		a.VisitCallExpr(n)
	case *ast.LiteralNode:
		a.VisitLiteral(n)
	case *ast.IdentifierNode:
		a.VisitIdentifier(n)
	case *ast.AssignmentNode:
		a.VisitAssignment(n)
	case *ast.ArrayAccessNode:
		a.VisitArrayAccess(n)
	case *ast.IndexExprNode:
		a.VisitIndexExpr(n)
	case *ast.MemberAccessNode:
		a.VisitMemberAccess(n)
	case *ast.LambdaNode:
		a.VisitLambda(n)
	case *ast.YieldStmtNode:
		a.VisitYieldStmt(n)
	case *ast.AwaitExprNode:
		a.VisitAwaitExpr(n)
	case *ast.SpawnStmtNode:
		a.VisitSpawnStmt(n)
	default:
		a.AddError(fmt.Errorf("unknown node type: %T", node))
	}
}

func (a *SemanticAnalyzer) VisitProgram(node *ast.ProgramNode) {
	for _, decl := range node.Declarations {
		a.VisitNode(decl)
	}
}

func (a *SemanticAnalyzer) VisitInclude(node *ast.IncludeNode) {
	a.hasInclude = true
}

func (a *SemanticAnalyzer) VisitUseLib(node *ast.UseLibNode) {
	a.hasInclude = true
}

func (a *SemanticAnalyzer) VisitDefine(node *ast.DefineNode) {
	// Defines are handled at code generation time
}

func (a *SemanticAnalyzer) VisitPragma(node *ast.PragmaNode) {
	// Pragmas are handled at code generation time
}

func (a *SemanticAnalyzer) VisitLibrary(node *ast.LibraryNode) {
	// Libraries are handled at code generation time
}

func (a *SemanticAnalyzer) VisitConfig(node *ast.ConfigNode) {
	// Config is handled at code generation time
}

func (a *SemanticAnalyzer) VisitWrapper(node *ast.WrapperNode) {
	// Wrappers are handled at code generation time
}

func (a *SemanticAnalyzer) VisitExternDecl(node *ast.ExternDeclNode) {
	// Allow extern to redeclare builtins (e.g. malloc, free); otherwise add to global scope
	if existing := a.globalScope.Resolve(node.Name); existing != nil {
		// Redeclaration: keep the existing symbol (builtin); codegen will emit the user's extern declaration
		return
	}
	funcSymbol := &Symbol{
		Name:       node.Name,
		Type:       node.ReturnType,
		SymbolType: SymbolFunction,
		Node:       node,
	}
	if err := a.globalScope.Define(funcSymbol); err != nil {
		a.AddError(err)
	}
}

func (a *SemanticAnalyzer) VisitLambda(node *ast.LambdaNode) {
	// Resolve capture names from enclosing scope and record types for codegen
	for _, name := range node.Captures {
		sym := a.currentScope.Resolve(name)
		if sym == nil {
			a.AddError(fmt.Errorf("undefined capture '%s' in lambda at line %d", name, node.GetLine()))
			node.ResolvedCaptureTypes = append(node.ResolvedCaptureTypes, "any")
			continue
		}
		if sym.SymbolType != SymbolVariable && sym.SymbolType != SymbolParameter {
			a.AddError(fmt.Errorf("cannot capture non-variable '%s' in lambda at line %d", name, node.GetLine()))
			node.ResolvedCaptureTypes = append(node.ResolvedCaptureTypes, "any")
			continue
		}
		node.ResolvedCaptureTypes = append(node.ResolvedCaptureTypes, sym.Type)
	}
	// Create new scope for lambda parameters
	lambdaScope := NewScope(a.currentScope)
	a.currentScope = lambdaScope

	// Define parameters in scope
	for _, param := range node.Parameters {
		lambdaScope.Define(&Symbol{
			Name:       param.Name,
			Type:       param.Type,
			SymbolType: SymbolParameter,
			Node:       param,
		})
	}

	// Visit body
	a.VisitNode(node.Body)

	// Restore scope
	a.currentScope = lambdaScope.Parent
}

func (a *SemanticAnalyzer) VisitFunctionDecl(node *ast.FunctionDeclNode) {
	// Register function in current (global) scope first
	emitName := node.Name
	modulePrefix := ""
	if node.Module != "" {
		emitName = node.Module + "__" + node.Name
		modulePrefix = node.Module + "__"
	}
	funcSym := &Symbol{
		Name:       node.Name,
		EmitName:   emitName,
		Type:       node.ReturnType,
		SymbolType: SymbolFunction,
		Node:       node,
	}
	// Register with local name
	a.currentScope.Define(funcSym)
	// Also register with module-prefixed name for module-qualified calls
	if modulePrefix != "" {
		prefixedSym := &Symbol{
			Name:       modulePrefix + node.Name,
			EmitName:   emitName,
			Type:       node.ReturnType,
			SymbolType: SymbolFunction,
			Node:       node,
		}
		a.currentScope.Define(prefixedSym)
	}

	// Create new scope for function body
	funcScope := NewScope(a.currentScope)
	a.currentScope = funcScope

	// Define parameters in scope
	for _, param := range node.Parameters {
		funcScope.Define(&Symbol{
			Name:       param.Name,
			Type:       param.Type,
			SymbolType: SymbolParameter,
			Node:       param,
		})
		if param.DefaultValue != nil {
			// Type-check default value
			a.VisitNode(param.DefaultValue)
			defaultType := a.GetExpressionType(param.DefaultValue)
			if !a.IsAssignableTo(defaultType, param.Type) {
				a.AddError(fmt.Errorf("default value for parameter '%s' at line %d has incompatible type '%s', expected '%s'", param.Name, param.GetLine(), defaultType, param.Type))
			}
		}
	}

	// Set flag for coroutine context to validate yield statements
	if node.IsCoroutine {
		a.inCoroutine = true
	}
	// Set flag for async context to validate await expressions
	if node.IsAsync {
		a.inAsync = true
	}

	// Visit body
	a.VisitBlock(node.Body)

	// Reset coroutine flag after visiting body
	if node.IsCoroutine {
		a.inCoroutine = false
	}
	// Reset async flag after visiting body
	if node.IsAsync {
		a.inAsync = false
	}

	// Restore scope
	a.currentScope = funcScope.Parent
}

func (a *SemanticAnalyzer) VisitVariableDecl(node *ast.VariableDeclNode) {
	// Check if variable already defined in current scope only (allow shadowing from outer scopes)
	if _, exists := a.currentScope.Symbols[node.Name]; exists {
		a.AddError(fmt.Errorf("variable '%s' already defined in this scope", node.Name))
		return
	}

	emitName := node.Name
	if node.Module != "" {
		emitName = node.Module + "__" + node.Name
	}

	// Determine symbol type based on IsConst
	symType := SymbolVariable
	if node.IsConst {
		symType = SymbolConst
	}

	varSymbol := &Symbol{
		Name:       node.Name,
		EmitName:   emitName,
		Type:       node.Type,
		SymbolType: symType,
		Node:       node,
	}

	if err := a.currentScope.Define(varSymbol); err != nil {
		a.AddError(err)
	}

	// Visit initializer and check type when variable has static type
	if node.Initializer != nil {
		a.VisitNode(node.Initializer)
		// Type inference for var
		if node.Type == "var" {
			inferredType := a.GetExpressionType(node.Initializer)
			if inferredType != "" && inferredType != "unknown" {
				node.Type = inferredType
				varSymbol.Type = inferredType
			} else {
				// Default to any if we can't infer
				node.Type = "any"
				varSymbol.Type = "any"
			}
		}
		if node.Type != "var" && node.Type != "any" {
			if arr, ok := node.Initializer.(*ast.ArrayLiteralNode); ok {
				// Array initializer: each element must be assignable to element type (strip [] suffix)
				elemTargetType := node.Type
				if strings.HasSuffix(node.Type, "[]") {
					elemTargetType = strings.TrimSuffix(node.Type, "[]")
				}
				for _, elem := range arr.Elements {
					elemType := a.GetExpressionType(elem)
					if !a.IsAssignableTo(elemType, elemTargetType) {
						a.AddError(fmt.Errorf("type mismatch: array element %s not assignable to %s '%s' (line %d)", elemType, node.Type, node.Name, node.GetLine()))
						break
					}
				}
			} else {
				valueType := a.GetExpressionType(node.Initializer)
				if !a.IsAssignableTo(valueType, node.Type) {
					a.AddError(fmt.Errorf("type mismatch: initializer %s not assignable to %s '%s' (line %d)", valueType, node.Type, node.Name, node.GetLine()))
				}
			}
		}
	}
}

func (a *SemanticAnalyzer) VisitStructDecl(node *ast.StructDeclNode) {
	// Check if struct already defined
	if existing := a.currentScope.Resolve(node.Name); existing != nil {
		a.AddError(fmt.Errorf("struct '%s' already defined at line %d", node.Name, existing.Node.GetLine()))
		return
	}

	emitName := node.Name
	if node.Module != "" {
		emitName = node.Module + "__" + node.Name
	}
	structSymbol := &Symbol{
		Name:       node.Name,
		EmitName:   emitName,
		Type:       "struct",
		SymbolType: SymbolStruct,
		Node:       node,
	}

	if err := a.currentScope.Define(structSymbol); err != nil {
		a.AddError(err)
	}

	// Visit field declarations
	for _, field := range node.Fields {
		a.VisitNode(field)
	}
	// Visit method bodies with a scope that includes struct fields (so x, y etc. resolve)
	for _, m := range node.Methods {
		scope := NewScope(a.currentScope)
		for _, f := range node.Fields {
			scope.Define(&Symbol{Name: f.Name, Type: f.Type, SymbolType: SymbolVariable, Node: f})
		}
		for _, p := range m.Parameters {
			scope.Define(&Symbol{Name: p.Name, Type: p.Type, SymbolType: SymbolVariable, Node: p})
		}
		prev := a.currentScope
		a.currentScope = scope
		a.VisitNode(m.Body)
		a.currentScope = prev
	}
}

func (a *SemanticAnalyzer) VisitEnumDecl(node *ast.EnumDeclNode) {
	// Check if enum already defined
	if existing := a.currentScope.Resolve(node.Name); existing != nil {
		a.AddError(fmt.Errorf("enum '%s' already defined at line %d", node.Name, existing.Node.GetLine()))
		return
	}

	emitName := node.Name
	if node.Module != "" {
		emitName = node.Module + "__" + node.Name
	}
	enumSymbol := &Symbol{
		Name:       node.Name,
		EmitName:   emitName,
		Type:       "enum",
		SymbolType: SymbolEnum,
		Node:       node,
	}

	if err := a.currentScope.Define(enumSymbol); err != nil {
		a.AddError(err)
	}

	// Register each enum value as a constant with auto-incrementing value
	for i, value := range node.Values {
		valueSymbol := &Symbol{
			Name:       value,
			EmitName:   value,
			Type:       "int",
			SymbolType: SymbolConst,
			Node:       node,
			Value:      i, // Auto-increment: first value = 0, second = 1, etc.
		}
		if err := a.currentScope.Define(valueSymbol); err != nil {
			a.AddError(fmt.Errorf("enum value '%s' already defined at line %d", value, node.Line))
		}
	}
}

func (a *SemanticAnalyzer) VisitBlock(node *ast.BlockNode) {
	// Create new scope for block
	blockScope := NewScope(a.currentScope)
	a.currentScope = blockScope

	// Visit statements
	for _, stmt := range node.Statements {
		a.VisitNode(stmt)
	}

	// Restore scope
	a.currentScope = blockScope.Parent
}

func (a *SemanticAnalyzer) VisitIfStmt(node *ast.IfStmtNode) {
	// Visit condition
	a.VisitNode(node.Condition)

	// Visit then branch
	a.VisitNode(node.ThenBranch)

	// Visit else branch if present
	if node.ElseBranch != nil {
		a.VisitNode(node.ElseBranch)
	}
}

func (a *SemanticAnalyzer) VisitWhileStmt(node *ast.WhileStmtNode) {
	a.VisitNode(node.Condition)
	a.VisitNode(node.Body)
}

func (a *SemanticAnalyzer) VisitDoWhileStmt(node *ast.DoWhileStmtNode) {
	a.VisitNode(node.Body)
	a.VisitNode(node.Condition)
}

func (a *SemanticAnalyzer) VisitForStmt(node *ast.ForStmtNode) {
	// Create new scope for for loop
	forScope := NewScope(a.currentScope)
	a.currentScope = forScope

	// Visit initializer
	if node.Initializer != nil {
		a.VisitNode(node.Initializer)
	}

	// Visit condition
	if node.Condition != nil {
		a.VisitNode(node.Condition)
	}

	// Visit increment
	if node.Increment != nil {
		a.VisitNode(node.Increment)
	}

	// Visit body
	a.VisitNode(node.Body)

	// Restore scope
	a.currentScope = forScope.Parent
}

func (a *SemanticAnalyzer) VisitReturnStmt(node *ast.ReturnStmtNode) {
	if node.Value != nil {
		a.VisitNode(node.Value)
	}
}

func (a *SemanticAnalyzer) VisitDeferStmt(node *ast.DeferStmtNode) {
	a.VisitNode(node.Body)
}

func (a *SemanticAnalyzer) VisitMatchStmt(node *ast.MatchStmtNode) {
	a.VisitNode(node.Value)
	for _, c := range node.Cases {
		scope := NewScope(a.currentScope)
		a.currentScope = scope
		if c.VarName != "" && c.TypeName != "" {
			// Result pattern: Ok(v) -> v is any, Err(e) -> e is string; else use TypeName (e.g. type pattern)
			typ := c.TypeName
			if c.TypeName == "Ok" {
				typ = "any"
			} else if c.TypeName == "Err" {
				typ = "string"
			}
			scope.Define(&Symbol{Name: c.VarName, Type: typ, SymbolType: SymbolVariable, Node: c})
		}
		a.VisitNode(c.Body)
		a.currentScope = scope.Parent
	}
}

func (a *SemanticAnalyzer) VisitForInStmt(node *ast.ForInStmtNode) {
	a.VisitNode(node.Collection)
	scope := NewScope(a.currentScope)
	a.currentScope = scope
	scope.Define(&Symbol{Name: node.VarName, Type: "any", SymbolType: SymbolVariable, Node: node})
	a.VisitNode(node.Body)
	a.currentScope = scope.Parent
}

func (a *SemanticAnalyzer) VisitRepeatStmt(node *ast.RepeatStmtNode) {
	a.VisitNode(node.Count)
	a.VisitNode(node.Body)
}

func (a *SemanticAnalyzer) VisitBreakStmt(node *ast.BreakStmtNode) {}

func (a *SemanticAnalyzer) VisitContinueStmt(node *ast.ContinueStmtNode) {}

func (a *SemanticAnalyzer) VisitSwitchStmt(node *ast.SwitchStmtNode) {
	a.VisitNode(node.Value)
	for _, c := range node.Cases {
		if c.Constant != nil {
			a.VisitNode(c.Constant)
		}
		a.VisitNode(c.Body)
	}
}

func (a *SemanticAnalyzer) VisitTestStmt(node *ast.TestStmtNode) {
	a.VisitNode(node.Body)
}

func (a *SemanticAnalyzer) VisitTupleExpr(node *ast.TupleExprNode) {
	for _, e := range node.Elements {
		a.VisitNode(e)
	}
}

func (a *SemanticAnalyzer) VisitArrayLiteral(node *ast.ArrayLiteralNode) {
	if node.Dimensions == 2 {
		if node.RowCount == 0 || node.RowLength <= 0 {
			a.AddError(fmt.Errorf("2D array literal must have at least one row and one column (line %d)", node.GetLine()))
		}
		expectedLen := node.RowLength
		for idx, rowNode := range node.Elements {
			row, ok := rowNode.(*ast.ArrayLiteralNode)
			if !ok {
				a.AddError(fmt.Errorf("2D array literal row %d is not an array literal (line %d)", idx, rowNode.GetLine()))
				a.VisitNode(rowNode)
				continue
			}
			a.VisitArrayLiteral(row)
			if expectedLen >= 0 && len(row.Elements) != expectedLen {
				a.AddError(fmt.Errorf("2D array literal rows must have equal length (row %d has %d, expected %d)", idx, len(row.Elements), expectedLen))
			}
		}
		return
	}

	var elemType string
	for idx, e := range node.Elements {
		a.VisitNode(e)
		et := a.GetExpressionType(e)
		if elemType == "" || elemType == "any" {
			elemType = et
			continue
		}
		if et == "any" {
			continue
		}
		if et != elemType {
			if !(a.IsAssignableTo(et, elemType) || a.IsAssignableTo(elemType, et)) {
				a.AddError(fmt.Errorf("array literal element %d type %s does not match %s", idx, et, elemType))
			}
		}
	}
}

func (a *SemanticAnalyzer) VisitDictLiteral(node *ast.DictLiteralNode) {
	keySet := make(map[string]struct{})
	for _, ent := range node.Entries {
		if ent.Key == "" {
			a.AddError(fmt.Errorf("dict literal key cannot be empty at line %d", node.GetLine()))
		}
		if _, exists := keySet[ent.Key]; exists {
			a.AddError(fmt.Errorf("duplicate key '%s' in dict literal at line %d", ent.Key, node.GetLine()))
		} else {
			keySet[ent.Key] = struct{}{}
		}
		a.VisitNode(ent.Value)
	}
}

func (a *SemanticAnalyzer) VisitStructLiteral(node *ast.StructLiteralNode) {
	fieldSet := make(map[string]struct{})
	for _, field := range node.Fields {
		if field.Name == "" {
			a.AddError(fmt.Errorf("struct field name cannot be empty at line %d", node.GetLine()))
		}
		if _, exists := fieldSet[field.Name]; exists {
			a.AddError(fmt.Errorf("duplicate field '%s' in struct literal at line %d", field.Name, node.GetLine()))
		} else {
			fieldSet[field.Name] = struct{}{}
		}
		a.VisitNode(field.Value)
	}
}

func (a *SemanticAnalyzer) VisitInterpolatedString(node *ast.InterpolatedStringNode) {
	for _, p := range node.Parts {
		a.VisitNode(p)
		if id, ok := p.(*ast.IdentifierNode); ok {
			if sym := a.currentScope.Resolve(id.Name); sym != nil {
				id.ResolvedType = sym.Type
			}
		}
	}
}

func (a *SemanticAnalyzer) VisitBinaryExpr(node *ast.BinaryExprNode) {
	// Visit both operands first
	a.VisitNode(node.Left)
	a.VisitNode(node.Right)

	// Smart type checking and inference
	leftType := a.GetExpressionType(node.Left)
	rightType := a.GetExpressionType(node.Right)

	// Check for compatible types and perform smart type inference
	if !a.AreTypesCompatible(leftType, rightType, node.Operator) {
		a.AddError(fmt.Errorf("type mismatch: cannot %s %s and %s", node.Operator, leftType, rightType))
	}
}

func (a *SemanticAnalyzer) GetExpressionType(expr ast.ASTNode) string {
	switch e := expr.(type) {
	case *ast.LiteralNode:
		return e.Type
	case *ast.IdentifierNode:
		symbol := a.currentScope.Resolve(e.Name)
		if symbol != nil {
			return symbol.Type
		}
		return "unknown"
	case *ast.BinaryExprNode:
		return a.InferBinaryExpressionType(e)
	case *ast.UnaryExprNode:
		return a.GetExpressionType(e.Operand)
	case *ast.CallExprNode:
		return a.GetFunctionReturnType(e.Function)
	case *ast.ArrayLiteralNode:
		if len(e.Elements) == 0 {
			if e.Dimensions == 2 {
				return "any[][]"
			}
			return "any[]"
		}
		firstType := a.GetExpressionType(e.Elements[0])
		if firstType == "" {
			firstType = "any"
		}
		if e.Dimensions == 2 {
			if !strings.HasSuffix(firstType, "[]") {
				firstType += "[]"
			}
			return firstType + "[]"
		}
		return firstType + "[]"
	case *ast.DictLiteralNode:
		return "dict"
	case *ast.StructLiteralNode:
		if e.TypeName != "" {
			return e.TypeName
		}
		return "struct"
	case *ast.ArrayAccessNode:
		arrType := a.GetExpressionType(e.Array)
		if strings.HasSuffix(arrType, "[]") {
			return arrType[:len(arrType)-2]
		}
		return "any"
	case *ast.IndexExprNode:
		objType := a.GetExpressionType(e.Object)
		if strings.HasSuffix(objType, "[]") {
			return objType[:len(objType)-2]
		}
		return "any"
	case *ast.MemberAccessNode:
		// Try to resolve field type from struct definition
		objType := a.GetExpressionType(e.Object)
		if objType != "" && objType != "any" {
			// Look up struct definition
			if sym := a.globalScope.Resolve(objType); sym != nil && sym.SymbolType == SymbolStruct {
				if structNode, ok := sym.Node.(*ast.StructDeclNode); ok {
					for _, field := range structNode.Fields {
						if field.Name == e.Member {
							return field.Type
						}
					}
				}
			}
		}
		return "any"
	default:
		return "any"
	}
}

func (a *SemanticAnalyzer) InferBinaryExpressionType(expr *ast.BinaryExprNode) string {
	leftType := a.GetExpressionType(expr.Left)
	rightType := a.GetExpressionType(expr.Right)

	// Smart type inference for binary operations
	switch expr.Operator {
	case "+", "-", "*", "/":
		// Arithmetic operations: promote to higher precision type
		if a.IsNumericType(leftType) && a.IsNumericType(rightType) {
			return a.GetHigherPrecisionType(leftType, rightType)
		}
		// String concatenation
		if leftType == "string" || rightType == "string" {
			return "string"
		}
		return "any"
	case "==", "!=", "<", "<=", ">", ">=":
		// Comparison operations always return bool
		return "bool"
	case "&&", "||":
		// Logical operations always return bool
		return "bool"
	default:
		return "any"
	}
}

func (a *SemanticAnalyzer) IsNumericType(typeName string) bool {
	return typeName == "int" || typeName == "float" || typeName == "double" || typeName == "size_t" || typeName == "number"
}

func (a *SemanticAnalyzer) GetHigherPrecisionType(type1, type2 string) string {
	norm := func(t string) string {
		if t == "number" {
			return "int"
		}
		return t
	}
	type1, type2 = norm(type1), norm(type2)
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

func (a *SemanticAnalyzer) AreTypesCompatible(type1, type2, operator string) bool {
	switch operator {
	case "+", "-", "*", "/":
		if (a.IsNumericType(type1) && a.IsNumericType(type2)) ||
			(type1 == "string" || type2 == "string") ||
			(type1 == "any" || type2 == "any") {
			return true
		}
	case "==", "!=", "<", "<=", ">", ">=":
		// Comparisons: most types can be compared
		return true
	case "&&", "||":
		// Logical: both must be boolean or convertible to boolean
		return type1 == "bool" && type2 == "bool"
	default:
		// Default: allow any types for flexibility
		return true
	}
	return false
}

// isCompatibleType checks if actualType can be assigned to expectedType
func (a *SemanticAnalyzer) IsCompatibleType(expectedType, actualType string) bool {
	if expectedType == actualType {
		return true
	}
	if expectedType == "any" {
		return true
	}
	if actualType == "any" {
		return true
	}
	// Allow implicit conversion from int to float
	if expectedType == "float" && actualType == "int" {
		return true
	}
	return false
}

// isAssignableTo returns true if a value of type valueType can be assigned to a variable of type targetType.
// Used for static type checking: var/any accept anything; static types require compatibility.
func (a *SemanticAnalyzer) IsAssignableTo(valueType, targetType string) bool {
	// Dynamic typing: var and any accept any value
	if targetType == "var" || targetType == "any" {
		return true
	}
	// Unknown value type (e.g. undefined): allow for better error elsewhere or reject
	if valueType == "unknown" {
		return false
	}
	// Allow 'any' to be assigned to any type (dynamic typing)
	if valueType == "any" {
		return true
	}
	// Exact match
	if valueType == targetType {
		return true
	}
	// Numeric promotion: int -> float, double; float -> double
	if a.IsNumericType(targetType) && a.IsNumericType(valueType) {
		order := map[string]int{"int": 1, "size_t": 1, "float": 2, "double": 3}
		v, tok := order[valueType], order[targetType]
		if v == 0 {
			v = 1
		}
		if tok == 0 {
			tok = 1
		}
		return v <= tok
	}
	// Array: same element type or any[]
	if strings.HasSuffix(targetType, "[]") {
		return valueType == targetType || valueType == "any[]" || valueType == "any"
	}
	// Strict: no other cross-type assignment for static types
	return false
}

func (a *SemanticAnalyzer) GetFunctionReturnType(funcExpr ast.ASTNode) string {
	switch f := funcExpr.(type) {
	case *ast.IdentifierNode:
		symbol := a.currentScope.Resolve(f.Name)
		if symbol != nil && symbol.SymbolType == SymbolFunction {
			retType := symbol.Type
			if retType == "" || retType == "void" {
				// For void functions, return any to allow flexibility
				// but the actual return won't be used
				return "any"
			}
			return retType
		}
		// Check built-in functions
		switch f.Name {
		case "printf", "println", "print", "writeline":
			return "int"
		case "strlen", "sizeof":
			return "size_t"
		case "malloc", "calloc", "realloc":
			return "void*"
		case "fopen":
			return "void*"
		case "sqrt", "sin", "cos", "tan":
			return "double"
		case "type_of":
			return "string"
		case "is_type":
			return "bool"
		case "as_int":
			return "int"
		case "as_float":
			return "float"
		case "as_string":
			return "string"
		case "as_bool":
			return "bool"
		case "as_dict":
			return "dict"
		case "as_array":
			return "array"
		case "json_parse":
			return "dict"
		case "json_stringify":
			return "string"
		case "parse_number":
			return "float"
		case "parse_int":
			return "int"
		}
	case *ast.MemberAccessNode:
		// module.func() — resolve by member name (e.g. "add" from math.add)
		if id, ok := f.Object.(*ast.IdentifierNode); ok {
			// Try module__func lookup first
			moduleFuncName := id.Name + "__" + f.Member
			sym := a.currentScope.Resolve(moduleFuncName)
			if sym != nil && sym.SymbolType == SymbolFunction {
				return sym.Type
			}
			// Fallback: try just the member name
			sym = a.currentScope.Resolve(f.Member)
			if sym != nil && sym.SymbolType == SymbolFunction {
				return sym.Type
			}
		}
	}
	return "any"
}

func (a *SemanticAnalyzer) VisitCallExpr(node *ast.CallExprNode) {
	a.VisitNode(node.Function)

	// Visit positional arguments
	for _, arg := range node.Args {
		a.VisitNode(arg)
	}

	// Visit and validate named arguments
	for _, namedArg := range node.NamedArgs {
		a.VisitNode(namedArg.Value)

		// Resolve named argument against function signature
		if id, ok := node.Function.(*ast.IdentifierNode); ok {
			sym := a.currentScope.Resolve(id.Name)
			if sym != nil && sym.SymbolType == SymbolFunction {
				if fn, ok := sym.Node.(*ast.FunctionDeclNode); ok {
					// Check if parameter exists
					found := false
					for _, param := range fn.Parameters {
						if param.Name == namedArg.Name {
							found = true
							// Type check the argument against parameter type
							argType := a.GetExpressionType(namedArg.Value)
							if !a.IsAssignableTo(argType, param.Type) {
								a.AddError(fmt.Errorf("type mismatch: cannot pass %s to parameter '%s' of type %s (line %d)", argType, param.Name, param.Type, node.GetLine()))
							}
							break
						}
					}
					if !found {
						a.AddError(fmt.Errorf("unknown parameter '%s' for function '%s' (line %d)", namedArg.Name, id.Name, node.GetLine()))
					}
				}
			}
		}
	}

	// Check for duplicate named arguments
	seenNames := make(map[string]bool)
	for _, namedArg := range node.NamedArgs {
		if seenNames[namedArg.Name] {
			a.AddError(fmt.Errorf("duplicate named argument '%s' (line %d)", namedArg.Name, node.GetLine()))
		}
		seenNames[namedArg.Name] = true
	}
}

func (a *SemanticAnalyzer) VisitUnaryExpr(node *ast.UnaryExprNode) {
	// Visit operand
	a.VisitNode(node.Operand)
}

// VisitCastExpr visits a C-style cast expression
func (a *SemanticAnalyzer) VisitCastExpr(node *ast.CastExprNode) {
	// Visit operand
	a.VisitNode(node.Operand)
	// Cast expressions are type-safe in C, so we trust the target type
}

func (a *SemanticAnalyzer) VisitLiteral(node *ast.LiteralNode) {
	// Literals don't need semantic analysis
}

func (a *SemanticAnalyzer) VisitIdentifier(node *ast.IdentifierNode) {
	symbol := a.currentScope.Resolve(node.Name)
	if symbol == nil {
		if msg, gated := a.FeatureGateMessage(node.Name); gated {
			a.AddErrorAt(fmt.Errorf("%s at line %d", msg, node.GetLine()), node, errors.ErrSemanticUnknown, "Enable the required feature in config or use a different identifier.")
		} else if a.hasInclude {
			node.ResolvedType = "any"
		} else if guiBuiltins[node.Name] {
			// GUI builtins are always available
			node.ResolvedType = "any"
		} else if networkBuiltins[node.Name] {
			// Network builtins are always available
			node.ResolvedType = "any"
		} else if asyncBuiltins[node.Name] {
			// Async builtins are always available
			node.ResolvedType = "any"
		} else {
			a.AddErrorAt(fmt.Errorf("undefined identifier '%s' at line %d", node.Name, node.GetLine()), node, errors.ErrSemanticUnknown, "Check spelling or declare the identifier.")
		}
		return
	}
	node.ResolvedType = symbol.Type
	if symbol.EmitName != "" {
		node.EmitName = symbol.EmitName
	}
}

func (a *SemanticAnalyzer) FeatureGateMessage(name string) (string, bool) {
	if _, exists := blockchainBuiltins[name]; exists && !a.features.Blockchain {
		return fmt.Sprintf("feature 'blockchain' is disabled; '%s' is unavailable", name), true
	}
	if _, exists := qolBuiltins[name]; exists && !a.features.QoL {
		return fmt.Sprintf("feature 'qol' is disabled; '%s' is unavailable", name), true
	}
	return "", false
}

func (a *SemanticAnalyzer) VisitAssignment(node *ast.AssignmentNode) {
	a.VisitNode(node.Target)
	a.VisitNode(node.Value)
	// Check for const reassignment
	if ident, ok := node.Target.(*ast.IdentifierNode); ok {
		if sym := a.currentScope.Resolve(ident.Name); sym != nil {
			if sym.SymbolType == SymbolConst {
				a.AddError(fmt.Errorf("cannot assign to const '%s' (line %d)", ident.Name, node.GetLine()))
				return
			}
		}
	}
	// Static type checking: ensure value is assignable to target when target has static type
	var targetType string
	switch t := node.Target.(type) {
	case *ast.IdentifierNode:
		if sym := a.currentScope.Resolve(t.Name); sym != nil {
			targetType = sym.Type
		}
	case *ast.ArrayAccessNode:
		// Array element type
		arrType := a.GetExpressionType(t.Array)
		if strings.HasSuffix(arrType, "[]") {
			targetType = arrType[:len(arrType)-2]
		}
	}
	if targetType != "" {
		valueType := a.GetExpressionType(node.Value)
		if !a.IsAssignableTo(valueType, targetType) {
			a.AddError(fmt.Errorf("type mismatch: cannot assign %s to %s (line %d)", valueType, targetType, node.GetLine()))
		}
	}
}

func (a *SemanticAnalyzer) VisitArrayAccess(node *ast.ArrayAccessNode) {
	// Visit array and index
	a.VisitNode(node.Array)
	a.VisitNode(node.Index)
}

func (a *SemanticAnalyzer) VisitIndexExpr(node *ast.IndexExprNode) {
	// Visit object and index
	a.VisitNode(node.Object)
	a.VisitNode(node.Index)
}

func (a *SemanticAnalyzer) VisitMemberAccess(node *ast.MemberAccessNode) {
	// If object is a bare identifier (e.g. math in math.add), treat as module prefix — don't resolve so we don't error on "undefined"
	if id, ok := node.Object.(*ast.IdentifierNode); ok {
		if a.currentScope.Resolve(id.Name) == nil {
			return // module.func() style; skip resolving module name
		}
	}
	a.VisitNode(node.Object)
}

func (a *SemanticAnalyzer) VisitYieldStmt(node *ast.YieldStmtNode) {
	if !a.inCoroutine {
		a.AddError(fmt.Errorf("yield statement used outside of coroutine at line %d", node.GetLine()))
	}
	if node.Value != nil {
		a.VisitNode(node.Value)
	}
}

func (a *SemanticAnalyzer) VisitAwaitExpr(node *ast.AwaitExprNode) {
	if !a.inAsync {
		a.AddError(fmt.Errorf("await expression used outside of async function at line %d", node.GetLine()))
	}
	a.VisitNode(node.Expr)
}

func (a *SemanticAnalyzer) VisitSpawnStmt(node *ast.SpawnStmtNode) {
	// Visit the function and arguments
	a.VisitNode(node.Function)
	for _, arg := range node.Arguments {
		a.VisitNode(arg)
	}
	// If thread variable is specified, define it in current scope
	if node.ThreadVar != "" {
		a.currentScope.Define(&Symbol{
			Name:       node.ThreadVar,
			Type:       "cortex_thread",
			SymbolType: SymbolVariable,
		})
	}
}

func (a *SemanticAnalyzer) AddError(err error) {
	a.errors = append(a.errors, err)
}

// addErrorAt records an error and, when diagnostics collector is set, a structured diagnostic.
func (a *SemanticAnalyzer) AddErrorAt(err error, node ast.ASTNode, code errors.Code, suggestion string) {
	a.errors = append(a.errors, err)
	if a.diagnostics != nil && node != nil {
		line, col := node.GetLine(), node.GetColumn()
		if col <= 0 {
			col = 1
		}
		a.diagnostics.AddError(code, line, col, err.Error(), suggestion)
	}
}

func (a *SemanticAnalyzer) GetErrors() []error {
	return a.errors
}

// SetDiagnosticsCollector enables structured diagnostics (line, column, code, suggestion).
// When set, semantic errors are also added to the collector.
func (a *SemanticAnalyzer) SetDiagnosticsCollector(c *errors.Collector) {
	a.diagnostics = c
}
