package compiler

import (
	"cortex/internal/ast"
	"cortex/internal/config"
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	tokens        []Token
	position      int
	features      config.FeatureSet
	currentModule string // set by "module \"name\";", used to prefix symbols (name__symbol)
	debug         bool   // enable debug output
}

func NewParser(cfg config.Config) *Parser {
	return &Parser{
		features: cfg.Features,
		debug:    cfg.Debug,
	}
}

func (p *Parser) Parse(tokens []Token) (ast.ASTNode, error) {
	p.tokens = tokens
	p.position = 0
	p.currentModule = "" // reset per file so each parsed file has its own module context

	program := &ast.ProgramNode{
		BaseNode: ast.BaseNode{Type: ast.NodeProgram, Line: 1, Column: 1},
	}

	for !p.IsAtEnd() {
		decl, err := p.ParseDeclaration()
		if err != nil {
			return nil, err
		}
		if decl != nil {
			program.Declarations = append(program.Declarations, decl)
		}
	}

	return program, nil
}

func (p *Parser) debugPrint(format string, args ...interface{}) {
	if p.debug {
		fmt.Printf("[PARSER] "+format+"\n", args...)
	}
}

func (p *Parser) debugToken(msg string) {
	if p.debug {
		tok := p.Peek()
		fmt.Printf("[PARSER] %s: pos=%d type=%d value=%q line=%d col=%d\n", msg, p.position, tok.Type, tok.Value, tok.Line, tok.Column)
	}
}

func (p *Parser) ParseDeclaration() (ast.ASTNode, error) {
	// Skip comments
	for p.Match(TokenComment) {
		continue
	}
	// End of block: do not try to parse "}" as a declaration
	if p.Check(TokenRBrace) {
		return nil, nil
	}

	p.debugToken("ParseDeclaration")

	// Feature gates: reject gated keywords when feature is disabled
	if p.Check(TokenAsync) || p.Check(TokenAwait) {
		if !p.features.Async {
			return nil, fmt.Errorf("line %d: async/await requires features.async (enable in config)", p.Peek().Line)
		}
	}
	// async void name(params) { body } — parse as normal function (no-op at codegen for now)
	if p.Match(TokenAsync) {
		typeToken := p.ConsumeType("Expected return type after 'async'")
		for p.Match(TokenMultiply) {
			typeToken.Value += "*"
		}
		nameToken := p.Consume(TokenIdentifier, "Expected function name after 'async'")
		return p.ParseFunctionDeclaration(typeToken, nameToken)
	}

	// Handle spawn statement: spawn function(args) or spawn var = function(args)
	if p.Match(TokenSpawn) {
		return p.ParseSpawnStatement()
	}

	if p.Match(TokenPackage) {
		name := p.Consume(TokenIdentifier, "Expected package name after 'package'")
		p.Consume(TokenSemicolon, "Expected ';' after package name")
		return &ast.PackageNode{
			BaseNode: ast.BaseNode{Type: ast.NodePackage, Line: name.Line, Column: name.Column},
			Name:     name.Value,
		}, nil
	}
	if p.Match(TokenImport) {
		pathTok := p.Consume(TokenString, "Expected import path string after 'import'")
		p.Consume(TokenSemicolon, "Expected ';' after import path")
		return &ast.ImportNode{
			BaseNode: ast.BaseNode{Type: ast.NodeImport, Line: pathTok.Line, Column: pathTok.Column},
			Path:     pathTok.Value,
		}, nil
	}
	if p.Match(TokenModule) {
		nameTok := p.Consume(TokenString, "Expected module name string after 'module'")
		p.Consume(TokenSemicolon, "Expected ';' after module name")
		p.currentModule = strings.Trim(nameTok.Value, "\"")
		return nil, nil // module directive does not add a declaration
	}
	// Handle preprocessor directives
	if p.Match(TokenInclude) {
		return p.ParseInclude()
	}
	if p.Match(TokenUse) {
		return p.ParseUseLib()
	}
	if p.Check(TokenRawC) {
		tok := p.Peek()
		p.Advance()
		return &ast.RawCNode{
			BaseNode: ast.BaseNode{Type: ast.NodeRawC, Line: tok.Line, Column: tok.Column},
			Content:  tok.Value,
		}, nil
	}
	if p.Match(TokenDefine) {
		return p.ParseDefine()
	}
	if p.Match(TokenPragma) {
		return p.ParsePragma()
	}
	if p.Match(TokenLibrary) {
		return p.ParseLibrary()
	}
	if p.Match(TokenConfig) {
		return p.ParseConfig()
	}
	if p.Match(TokenWrapper) {
		return p.ParseWrapper()
	}

	if p.Match(TokenAt) && p.Check(TokenIdentifier) && p.Peek().Value == "c" {
		p.Advance() // Consume 'c'
		return p.ParseRawCBlock()
	}

	// Handle extern declarations
	if p.Match(TokenExtern) {
		return p.ParseExternDeclaration()
	}

	// Handle coroutine functions: coroutine returnType name(params) { body }
	if p.Match(TokenCoroutine) {
		returnType := p.ConsumeType("Expected return type after 'coroutine'")
		for p.Match(TokenMultiply) {
			returnType.Value += "*"
		}
		nameToken := p.Consume(TokenIdentifier, "Expected function name after return type")
		// Parse as normal function but mark as coroutine
		fn, err := p.ParseFunctionDeclaration(returnType, nameToken)
		if err != nil {
			return nil, err
		}
		fn.(*ast.FunctionDeclNode).IsCoroutine = true
		return fn, nil
	}

	if p.Match(TokenStruct) {
		return p.ParseStructDeclaration()
	}
	if p.Match(TokenEnum) {
		return p.ParseEnumDeclaration()
	}

	// Tuple return type: ( type , type ) name ( ... ) — only when we see ( type , or ( type )
	if p.position+2 < len(p.tokens) && p.Check(TokenLParen) &&
		p.IsTypeToken(p.tokens[p.position+1].Type) &&
		(p.tokens[p.position+2].Type == TokenComma || p.tokens[p.position+2].Type == TokenRParen) {
		types, name, err := p.ParseTupleReturnTypeAndName()
		if err == nil && p.Check(TokenLParen) {
			return p.ParseFunctionDeclarationWithTuple(types, name)
		}
	}
	// Check for function or variable declaration (type followed by identifier, then ( for function)
	if p.position+2 < len(p.tokens) &&
		p.IsTypeToken(p.tokens[p.position].Type) &&
		p.tokens[p.position+1].Type == TokenIdentifier &&
		p.tokens[p.position+2].Type == TokenLParen {
		decl, err := p.ParseVariableOrFunctionDeclaration()
		if err != nil {
			return nil, err
		}
		return decl, nil
	}
	// Variable declaration: type identifier OR custom_type identifier (but not function call: identifier ( )
	// Only treat as declaration if: it's a type token, OR it's identifier followed by identifier NOT followed by (
	if p.position < len(p.tokens) {
		isType := p.IsTypeToken(p.tokens[p.position].Type)
		// Custom type declaration: MyType varName - next token after identifier should NOT be (
		isCustomTypeDecl := p.tokens[p.position].Type == TokenIdentifier &&
			p.position+1 < len(p.tokens) &&
			p.tokens[p.position+1].Type == TokenIdentifier &&
			(p.position+2 >= len(p.tokens) || p.tokens[p.position+2].Type != TokenLParen) // Not a function call like foo(bar)

		p.debugPrint("isType=%v isCustomTypeDecl=%v", isType, isCustomTypeDecl)

		if isType || isCustomTypeDecl {
			p.debugPrint("calling ParseVariableOrFunctionDeclaration")
			decl, err := p.ParseVariableOrFunctionDeclaration()
			if err != nil {
				return nil, err
			}
			if _, isVar := decl.(*ast.VariableDeclNode); isVar {
				p.Consume(TokenSemicolon, "Expected ';' after variable declaration")
			}
			return decl, nil
		}
	}

	p.debugPrint("falling through to ParseStatement")

	return p.ParseStatement()
}

func (p *Parser) ParseInclude() (ast.ASTNode, error) {
	line, col := p.Previous().Line, p.Previous().Column

	// Extract header from token value (e.g., "include <stdio.h>" -> "<stdio.h>")
	tokenValue := p.Previous().Value
	parts := strings.SplitN(tokenValue, " ", 2)
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		// Return nil to skip this include (it's likely an empty/malformed include)
		return nil, nil
	}
	header := strings.TrimSpace(parts[1])

	// Validate header format
	isSystem := false
	filename := ""
	if strings.HasPrefix(header, "<") && strings.HasSuffix(header, ">") {
		isSystem = true
		filename = header[1 : len(header)-1] // Remove < and >
	} else if strings.HasPrefix(header, "\"") && strings.HasSuffix(header, "\"") {
		isSystem = false
		filename = header[1 : len(header)-1] // Remove quotes
	} else {
		// Invalid header format, skip it
		return nil, nil
	}

	return &ast.IncludeNode{
		BaseNode: ast.BaseNode{Type: ast.NodeInclude, Line: line, Column: col},
		Header:   header,
		IsSystem: isSystem,
		Filename: filename,
	}, nil
}

func (p *Parser) ParseUseLib() (ast.ASTNode, error) {
	value := strings.TrimSpace(p.Previous().Value)
	// Value is like "use \"raylib\"" — find quoted lib name
	value = value[3:] // skip "use"
	value = strings.TrimSpace(value)
	libName := value
	if strings.HasPrefix(value, "\"") && strings.Contains(value, "\"") {
		end := strings.Index(value[1:], "\"")
		if end >= 0 {
			libName = value[1 : 1+end]
		}
	}
	if libName == "" {
		return nil, fmt.Errorf("expected library name in #use \"name\"")
	}
	return &ast.UseLibNode{
		BaseNode: ast.BaseNode{Type: ast.NodeUseLib, Line: p.Previous().Line, Column: p.Previous().Column},
		LibName:  libName,
	}, nil
}

func (p *Parser) ParseDefine() (ast.ASTNode, error) {
	value := p.Previous().Value
	parts := strings.Fields(value)
	if len(parts) < 2 {
		return nil, fmt.Errorf("Expected name after #define")
	}

	name := parts[1]
	var valueStr string
	if len(parts) > 2 {
		valueStr = strings.Join(parts[2:], " ")
	}

	return &ast.DefineNode{
		BaseNode: ast.BaseNode{Type: ast.NodeDefine, Line: p.Previous().Line, Column: p.Previous().Column},
		Name:     name,
		Value:    valueStr,
	}, nil
}

func (p *Parser) ParsePragma() (ast.ASTNode, error) {
	value := p.Previous().Value
	parts := strings.Fields(value)
	if len(parts) < 2 {
		return nil, fmt.Errorf("Expected directive after #pragma")
	}

	directive := parts[1]
	var content string
	if len(parts) > 2 {
		content = strings.Join(parts[2:], " ")
	}

	return &ast.PragmaNode{
		BaseNode:  ast.BaseNode{Type: ast.NodePragma, Line: p.Previous().Line, Column: p.Previous().Column},
		Directive: directive,
		Content:   content,
	}, nil
}

func (p *Parser) ParseLibrary() (ast.ASTNode, error) {
	value := p.Previous().Value
	parts := strings.Fields(value)
	if len(parts) < 2 {
		return nil, fmt.Errorf("Expected library name after library")
	}

	libName := parts[1]

	// For now, just store the library name
	return &ast.LibraryNode{
		BaseNode:  ast.BaseNode{Type: ast.NodeLibrary, Line: p.Previous().Line, Column: p.Previous().Column},
		Name:      libName,
		Functions: []ast.Node{},
	}, nil
}

func (p *Parser) ParseConfig() (ast.ASTNode, error) {
	return &ast.ConfigNode{
		BaseNode: ast.BaseNode{Type: ast.NodeConfig, Line: p.Previous().Line, Column: p.Previous().Column},
		Settings: make(map[string]interface{}),
	}, nil
}

func (p *Parser) ParseWrapper() (ast.ASTNode, error) {
	value := p.Previous().Value
	parts := strings.Fields(value)
	if len(parts) < 2 {
		return nil, fmt.Errorf("Expected wrapper name after wrapper")
	}

	wrapperName := parts[1]

	return &ast.WrapperNode{
		BaseNode:     ast.BaseNode{Type: ast.NodeWrapper, Line: p.Previous().Line, Column: p.Previous().Column},
		Name:         wrapperName,
		Declarations: []ast.ASTNode{},
	}, nil
}

func (p *Parser) ParseRawCBlock() (ast.ASTNode, error) {
	line, col := p.Previous().Line, p.Previous().Column
	p.Consume(TokenLBrace, "expected '{' after @c")
	content := ""
	braceCount := 1
	for !p.IsAtEnd() {
		if p.Match(TokenLBrace) {
			braceCount++
			content += p.Previous().Value
		} else if p.Match(TokenRBrace) {
			braceCount--
			if braceCount == 0 {
				break
			}
			content += p.Previous().Value
		} else {
			content += p.Peek().Value
			p.Advance()
		}
	}
	if braceCount != 0 {
		return nil, fmt.Errorf("unclosed brace in @c block at line %d", line)
	}
	return &ast.RawCNode{
		BaseNode: ast.BaseNode{Type: ast.NodeRawC, Line: line, Column: col},
		Content:  content,
	}, nil
}

func (p *Parser) ParseExternDeclaration() (ast.ASTNode, error) {
	returnType := p.ConsumeType("Expected return type after extern")
	for p.Match(TokenMultiply) {
		returnType.Value += "*"
	}

	name := p.Consume(TokenIdentifier, "Expected function name")
	p.Consume(TokenLParen, "Expected '(' after function name")

	var parameters []*ast.ParameterNode
	for !p.Check(TokenRParen) && !p.IsAtEnd() {
		paramType := p.ConsumeType("Expected parameter type")
		for p.Match(TokenMultiply) {
			paramType.Value += "*"
		}
		paramName := ""
		if p.Check(TokenIdentifier) {
			paramName = p.Advance().Value
		}
		parameters = append(parameters, &ast.ParameterNode{
			BaseNode: ast.BaseNode{Type: ast.NodeParameter, Line: paramType.Line, Column: paramType.Column},
			Name:     paramName,
			Type:     paramType.Value,
		})

		if !p.Check(TokenRParen) {
			p.Consume(TokenComma, "Expected ',' after parameter")
		}
	}

	p.Consume(TokenRParen, "Expected ')' after parameters")

	// Check for cleanup annotation: cleanup(funcName)
	var cleanupFunc string
	if p.Match(TokenCleanup) {
		p.Consume(TokenLParen, "Expected '(' after cleanup")
		cleanupTok := p.Consume(TokenIdentifier, "Expected cleanup function name")
		cleanupFunc = cleanupTok.Value
		p.Consume(TokenRParen, "Expected ')' after cleanup function name")
	}

	p.Consume(TokenSemicolon, "Expected ';' after extern declaration")

	return &ast.ExternDeclNode{
		BaseNode:    ast.BaseNode{Type: ast.NodeExternDecl, Line: name.Line, Column: name.Column},
		Name:        name.Value,
		ReturnType:  returnType.Value,
		Parameters:  parameters,
		CleanupFunc: cleanupFunc,
	}, nil
}

func (p *Parser) ParseStructDeclaration() (ast.ASTNode, error) {
	nameToken := p.Consume(TokenIdentifier, "Expected struct name")
	p.Consume(TokenLBrace, "Expected '{' after struct name")

	var fields []*ast.VariableDeclNode
	var methods []*ast.FunctionDeclNode
	for !p.Check(TokenRBrace) && !p.IsAtEnd() {
		if !p.IsTypeToken(p.Peek().Type) {
			return nil, fmt.Errorf("expected field or method at line %d", p.Peek().Line)
		}
		typeTok := p.ConsumeType("expected field/method type")
		nameTok := p.Consume(TokenIdentifier, "expected field or method name")
		if p.Check(TokenLParen) {
			// Method: name(params) { body }
			p.Advance()
			var params []*ast.ParameterNode
			for !p.Check(TokenRParen) && !p.IsAtEnd() {
				pt := p.ConsumeType("expected parameter type")
				pn := p.Consume(TokenIdentifier, "expected parameter name")
				params = append(params, &ast.ParameterNode{
					BaseNode: ast.BaseNode{Type: ast.NodeParameter, Line: pt.Line, Column: pt.Column},
					Type:     pt.Value,
					Name:     pn.Value,
				})
				if !p.Check(TokenRParen) {
					p.Consume(TokenComma, "expected ',' between parameters")
				}
			}
			p.Consume(TokenRParen, "expected ')' after parameters")
			body, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			methods = append(methods, &ast.FunctionDeclNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeFunctionDecl, Line: nameTok.Line, Column: nameTok.Column},
				Name:       nameTok.Value,
				Parameters: params,
				ReturnType: typeTok.Value,
				Body:       body.(*ast.BlockNode),
			})
		} else {
			// Field
			p.Consume(TokenSemicolon, "expected ';' after field declaration")
			fields = append(fields, &ast.VariableDeclNode{
				BaseNode:    ast.BaseNode{Type: ast.NodeVariableDecl, Line: nameTok.Line, Column: nameTok.Column},
				Name:        nameTok.Value,
				Type:        typeTok.Value,
				Initializer: nil,
			})
		}
	}

	p.Consume(TokenRBrace, "Expected '}' after struct")

	return &ast.StructDeclNode{
		BaseNode: ast.BaseNode{Type: ast.NodeStructDecl, Line: nameToken.Line, Column: nameToken.Column},
		Name:     nameToken.Value,
		Module:   p.currentModule,
		Fields:   fields,
		Methods:  methods,
	}, nil
}

func (p *Parser) ParseEnumDeclaration() (ast.ASTNode, error) {
	nameToken := p.Consume(TokenIdentifier, "Expected enum name")

	p.Consume(TokenLBrace, "Expected '{' after enum name")

	var values []string
	stringValues := make(map[string]string)

	for !p.Check(TokenRBrace) && !p.IsAtEnd() {
		valueToken := p.Consume(TokenIdentifier, "Expected enum value")
		values = append(values, valueToken.Value)

		// Check for explicit value: = "string" or = number
		if p.Match(TokenAssign) {
			if p.Check(TokenString) {
				// String enum: Red = "red"
				strTok := p.Advance()
				stringValues[valueToken.Value] = strTok.Value
			}
			// For numeric values, we could parse them but for now just skip
			// as auto-increment is the default
		}

		if !p.Check(TokenRBrace) {
			p.Consume(TokenComma, "Expected ',' after enum value")
		}
	}

	p.Consume(TokenRBrace, "Expected '}' after enum values")

	return &ast.EnumDeclNode{
		BaseNode:     ast.BaseNode{Type: ast.NodeEnumDecl, Line: nameToken.Line, Column: nameToken.Column},
		Name:         nameToken.Value,
		Module:       p.currentModule,
		Values:       values,
		StringValues: stringValues,
	}, nil
}

func (p *Parser) ParseTupleReturnTypeAndName() (types []string, name string, err error) {
	p.Advance() // consume (
	var list []string
	t := p.ConsumeType("Expected type in tuple return")
	for p.Match(TokenMultiply) {
		t.Value += "*"
	}
	list = append(list, t.Value)
	for p.Match(TokenComma) {
		t := p.ConsumeType("Expected type in tuple return")
		for p.Match(TokenMultiply) {
			t.Value += "*"
		}
		list = append(list, t.Value)
	}
	p.Consume(TokenRParen, "Expected ')' after tuple return type")
	nameTok := p.Consume(TokenIdentifier, "Expected function name")
	return list, nameTok.Value, nil
}

func (p *Parser) ParseFunctionDeclarationWithTuple(returnTypes []string, name string) (ast.ASTNode, error) {
	p.Consume(TokenLParen, "Expected '(' after function name")
	var parameters []*ast.ParameterNode
	for !p.Check(TokenRParen) && !p.IsAtEnd() {
		paramType := p.ConsumeType("Expected parameter type")
		for p.Match(TokenMultiply) {
			paramType.Value += "*"
		}
		paramName := p.Consume(TokenIdentifier, "Expected parameter name")
		parameters = append(parameters, &ast.ParameterNode{
			BaseNode: ast.BaseNode{Type: ast.NodeParameter, Line: paramType.Line, Column: paramType.Column},
			Name:     paramName.Value,
			Type:     paramType.Value,
		})
		if !p.Check(TokenRParen) {
			p.Consume(TokenComma, "Expected ',' after parameter")
		}
	}
	p.Consume(TokenRParen, "Expected ')' after parameters")
	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.FunctionDeclNode{
		BaseNode:    ast.BaseNode{Type: ast.NodeFunctionDecl, Line: body.GetLine(), Column: body.GetColumn()},
		Name:        name,
		Module:      p.currentModule,
		Parameters:  parameters,
		ReturnTypes: returnTypes,
		Body:        body.(*ast.BlockNode),
	}, nil
}

func (p *Parser) ParseVariableOrFunctionDeclaration() (ast.ASTNode, error) {
	typeToken := p.ConsumeTypeOrVar()
	for p.Match(TokenMultiply) {
		typeToken.Value += "*"
	}

	// Handle const specially - it's a modifier, not the actual type
	// After const, we need: actualType name OR name (for type inference)
	if typeToken.Type == TokenConst {
		// Get the actual type or use var for type inference
		var actualType Token
		if p.IsTypeToken(p.Peek().Type) {
			actualType = p.ConsumeTypeOrVar()
			for p.Match(TokenMultiply) {
				actualType.Value += "*"
			}
		} else {
			// const without type - infer from initializer
			actualType = Token{Type: TokenVar, Value: "var"}
		}
		// Get the name
		nameToken := p.Consume(TokenIdentifier, "Expected variable name after const")

		// Check if this is a function by looking for opening parenthesis
		if p.Check(TokenLParen) {
			return p.ParseFunctionDeclaration(actualType, nameToken)
		} else {
			// Mark as const and pass actual type
			return p.ParseVariableDeclarationWithConst(actualType, nameToken)
		}
	}

	// Allow type keywords to be used as names (e.g., "result" as variable name)
	var nameToken Token
	if p.IsTypeToken(p.Peek().Type) {
		nameToken = p.Advance()
	} else {
		nameToken = p.Consume(TokenIdentifier, "Expected variable or function name")
	}

	// Check if this is a function by looking for opening parenthesis
	if p.Check(TokenLParen) {
		return p.ParseFunctionDeclaration(typeToken, nameToken)
	} else {
		return p.ParseVariableDeclarationWithTypes(typeToken, nameToken)
	}
}

func (p *Parser) ParseFunctionDeclaration(typeToken, nameToken Token) (ast.ASTNode, error) {
	line, col := p.Peek().Line, p.Peek().Column
	isAsync := p.Match(TokenAsync)
	isCoroutine := p.Match(TokenCoroutine)
	p.Consume(TokenLParen, "Expected '(' after function name")

	var parameters []*ast.ParameterNode
	for !p.Check(TokenRParen) && !p.IsAtEnd() {
		paramType := p.ConsumeType("Expected parameter type")
		for p.Match(TokenMultiply) {
			paramType.Value += "*"
		}
		paramName := p.Consume(TokenIdentifier, "Expected parameter name")
		var defaultVal ast.ASTNode
		if p.Match(TokenAssign) { // Use TokenAssign (=) for default values
			def, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			defaultVal = def
		}
		parameters = append(parameters, &ast.ParameterNode{
			BaseNode:     ast.BaseNode{Type: ast.NodeParameter, Line: paramType.Line, Column: paramType.Column},
			Name:         paramName.Value,
			Type:         paramType.Value,
			DefaultValue: defaultVal,
		})

		if !p.Check(TokenRParen) {
			p.Consume(TokenComma, "Expected ',' after parameter")
		}
	}

	p.Consume(TokenRParen, "Expected ')' after parameters")

	// Use the type token as return type (C-style: int func() {})
	returnType := typeToken.Value
	// Also support -> for explicit return type override
	if p.Match(TokenArrow) {
		returnTypeTok := p.ConsumeType("Expected return type after ->")
		returnType = returnTypeTok.Value
	}
	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.FunctionDeclNode{
		BaseNode:    ast.BaseNode{Type: ast.NodeFunctionDecl, Line: line, Column: col},
		Name:        nameToken.Value,
		Module:      p.currentModule,
		Parameters:  parameters,
		ReturnType:  returnType,
		Body:        body.(*ast.BlockNode),
		IsAsync:     isAsync,
		IsCoroutine: isCoroutine,
	}, nil
}

func (p *Parser) ParseVariableDeclarationWithTypes(typeToken, nameToken Token) (ast.ASTNode, error) {
	var initializer ast.ASTNode
	var arraySize int = -1 // -1 means not an array, 0 means empty size, >0 means fixed size
	isConst := typeToken.Type == TokenConst

	// The actual type - for const, this was already resolved by the caller
	actualType := typeToken.Value

	// Check for array size declaration: Type name[size]
	if p.Match(TokenLBracket) {
		if p.Check(TokenNumber) {
			sizeTok := p.Advance()
			arraySize = 0
			if s, err := strconv.Atoi(sizeTok.Value); err == nil {
				arraySize = s
			}
		} else if p.Check(TokenRBracket) {
			arraySize = 0 // empty size like Type name[]
		}
		p.Consume(TokenRBracket, "Expected ']' after array size")
	}

	if p.Match(TokenAssign) {
		init, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		initializer = init
	}

	node := &ast.VariableDeclNode{
		BaseNode:    ast.BaseNode{Type: ast.NodeVariableDecl, Line: nameToken.Line, Column: nameToken.Column},
		Name:        nameToken.Value,
		Module:      p.currentModule,
		Type:        actualType,
		Initializer: initializer,
		IsConst:     isConst,
	}
	if arraySize >= 0 {
		node.Type = actualType + "[]"
		node.ArraySize = arraySize
	}
	return node, nil
}

func (p *Parser) ParseVariableDeclaration() (ast.ASTNode, error) {
	typeToken := p.ConsumeType("Expected type")
	nameToken := p.Consume(TokenIdentifier, "Expected variable name")
	return p.ParseVariableDeclarationWithTypes(typeToken, nameToken)
}

// ParseVariableDeclarationWithConst parses a const declaration with the actual type already resolved.
func (p *Parser) ParseVariableDeclarationWithConst(actualTypeToken, nameToken Token) (ast.ASTNode, error) {
	var initializer ast.ASTNode
	var arraySize int = -1 // -1 means not an array, 0 means empty size, >0 means fixed size

	// Check for array size declaration: Type name[size]
	if p.Match(TokenLBracket) {
		if p.Check(TokenNumber) {
			sizeTok := p.Advance()
			arraySize = 0
			if s, err := strconv.Atoi(sizeTok.Value); err == nil {
				arraySize = s
			}
		} else if p.Check(TokenRBracket) {
			arraySize = 0 // empty size like Type name[]
		}
		p.Consume(TokenRBracket, "Expected ']' after array size")
	}

	if p.Match(TokenAssign) {
		init, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		initializer = init
	}

	node := &ast.VariableDeclNode{
		BaseNode:    ast.BaseNode{Type: ast.NodeVariableDecl, Line: nameToken.Line, Column: nameToken.Column},
		Name:        nameToken.Value,
		Module:      p.currentModule,
		Type:        actualTypeToken.Value,
		Initializer: initializer,
		IsConst:     true,
	}
	if arraySize >= 0 {
		node.Type = actualTypeToken.Value + "[]"
		node.ArraySize = arraySize
	}
	return node, nil
}

// parseVariableDeclarationRest parses name and optional initializer when type was already consumed (e.g. in for-loop).
func (p *Parser) ParseVariableDeclarationRest(typeToken Token) (ast.ASTNode, error) {
	nameToken := p.Consume(TokenIdentifier, "Expected variable name")
	return p.ParseVariableDeclarationWithTypes(typeToken, nameToken)
}

func (p *Parser) ParseStatement() (ast.ASTNode, error) {
	if p.Match(TokenInclude) {
		return p.ParseInclude()
	}
	if p.Match(TokenAt) && p.Check(TokenIdentifier) && p.Peek().Value == "c" {
		p.Advance() // Consume 'c'
		return p.ParseRawCBlock()
	}
	if p.Match(TokenIf) {
		return p.ParseIfStatement()
	}
	if p.Match(TokenDo) {
		return p.ParseDoWhileStatement()
	}
	if p.Match(TokenWhile) {
		return p.ParseWhileStatement()
	}
	// for ( x in collection ) vs for ( init; cond; incr ) — use lookahead
	if p.Check(TokenFor) && p.position+3 < len(p.tokens) &&
		p.tokens[p.position+1].Type == TokenLParen &&
		p.tokens[p.position+2].Type == TokenIdentifier &&
		p.tokens[p.position+3].Type == TokenIn {
		p.Advance() // for
		p.Advance() // (
		idName := p.Peek().Value
		p.Advance() // id
		p.Advance() // in
		collection, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		p.Consume(TokenRParen, "Expected ')' after for-in")
		body, err := p.ParseBlock()
		if err != nil {
			return nil, err
		}
		return &ast.ForInStmtNode{
			BaseNode:   ast.BaseNode{Type: ast.NodeForInStmt, Line: p.Previous().Line, Column: p.Previous().Column},
			VarName:    idName,
			Collection: collection,
			Body:       body.(*ast.BlockNode),
		}, nil
	}
	if p.Match(TokenFor) {
		return p.ParseForStatement()
	}
	if p.Match(TokenMatch) {
		return p.ParseMatchStatement()
	}
	if p.Match(TokenDefer) {
		return p.ParseDeferStatement()
	}
	if p.Match(TokenRepeat) {
		return p.ParseRepeatStatement()
	}
	if p.Match(TokenSwitch) {
		return p.ParseSwitchStatement()
	}
	if p.Match(TokenTest) {
		return p.ParseTestStatement()
	}
	if p.Match(TokenBreak) {
		return p.ParseBreakStatement()
	}
	if p.Match(TokenContinue) {
		return p.ParseContinueStatement()
	}
	if p.Match(TokenReturn) {
		return p.ParseReturnStatement()
	}
	if p.Match(TokenYield) {
		return p.ParseYieldStatement()
	}
	if p.Match(TokenLBrace) {
		return p.ParseBlock()
	}

	// Check for variable declaration in statement position
	if p.position < len(p.tokens) && p.IsTypeToken(p.tokens[p.position].Type) {
		decl, err := p.ParseVariableDeclaration()
		if err != nil {
			return nil, err
		}
		// Variable declarations in statements need semicolons
		if !p.Match(TokenSemicolon) {
			return nil, fmt.Errorf("Expected ';' after variable declaration")
		}
		return decl, nil
	}

	return p.ParseExpressionStatement()
}

func (p *Parser) ParseIfStatement() (ast.ASTNode, error) {
	p.Consume(TokenLParen, "Expected '(' after 'if'")
	condition, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	p.Consume(TokenRParen, "Expected ')' after if condition")

	thenBranch, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}
	// ParseBlock already consumed the closing '}' if present

	var elseBranch *ast.BlockNode
	if p.Match(TokenElse) {
		if p.Check(TokenIf) {
			p.Advance() // consume "if" so parseIfStatement sees '('
			elseNode, err := p.ParseIfStatement()
			if err != nil {
				return nil, err
			}
			// Wrap else-if in a block
			elseBranch = &ast.BlockNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: elseNode.GetLine(), Column: elseNode.GetColumn()},
				Statements: []ast.ASTNode{elseNode},
			}
		} else {
			elseBlock, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			elseBranch = elseBlock.(*ast.BlockNode)
		}
	}

	return &ast.IfStmtNode{
		BaseNode:   ast.BaseNode{Type: ast.NodeIfStmt, Line: condition.GetLine(), Column: condition.GetColumn()},
		Condition:  condition,
		ThenBranch: thenBranch.(*ast.BlockNode),
		ElseBranch: elseBranch,
	}, nil
}

func (p *Parser) ParseWhileStatement() (ast.ASTNode, error) {
	p.Consume(TokenLParen, "Expected '(' after 'while'")
	condition, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	p.Consume(TokenRParen, "Expected ')' after while condition")

	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.WhileStmtNode{
		BaseNode:  ast.BaseNode{Type: ast.NodeWhileStmt, Line: condition.GetLine(), Column: condition.GetColumn()},
		Condition: condition,
		Body:      body.(*ast.BlockNode),
	}, nil
}

func (p *Parser) ParseDoWhileStatement() (ast.ASTNode, error) {
	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}
	p.Consume(TokenWhile, "Expected 'while' after do block")
	p.Consume(TokenLParen, "Expected '(' after 'while'")
	condition, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	p.Consume(TokenRParen, "Expected ')' after do-while condition")
	if !p.Match(TokenSemicolon) {
		return nil, fmt.Errorf("Expected ';' after do-while condition")
	}
	return &ast.DoWhileStmtNode{
		BaseNode:  ast.BaseNode{Type: ast.NodeDoWhileStmt, Line: body.GetLine(), Column: body.GetColumn()},
		Body:      body.(*ast.BlockNode),
		Condition: condition,
	}, nil
}

func (p *Parser) ParseRepeatStatement() (ast.ASTNode, error) {
	p.Consume(TokenLParen, "Expected '(' after 'repeat'")
	count, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	p.Consume(TokenRParen, "Expected ')' after repeat count")
	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.RepeatStmtNode{
		BaseNode: ast.BaseNode{Type: ast.NodeRepeatStmt, Line: count.GetLine(), Column: count.GetColumn()},
		Count:    count,
		Body:     body.(*ast.BlockNode),
	}, nil
}

func (p *Parser) ParseBreakStatement() (ast.ASTNode, error) {
	if !p.Match(TokenSemicolon) {
		p.Consume(TokenSemicolon, "Expected ';' after 'break'")
	}
	return &ast.BreakStmtNode{
		BaseNode: ast.BaseNode{Type: ast.NodeBreakStmt, Line: p.Previous().Line, Column: p.Previous().Column},
	}, nil
}

func (p *Parser) ParseContinueStatement() (ast.ASTNode, error) {
	if !p.Match(TokenSemicolon) {
		p.Consume(TokenSemicolon, "Expected ';' after 'continue'")
	}
	return &ast.ContinueStmtNode{
		BaseNode: ast.BaseNode{Type: ast.NodeContinueStmt, Line: p.Previous().Line, Column: p.Previous().Column},
	}, nil
}

func (p *Parser) ParseSwitchStatement() (ast.ASTNode, error) {
	p.Consume(TokenLParen, "Expected '(' after 'switch'")
	value, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	p.Consume(TokenRParen, "Expected ')' after switch value")
	p.Consume(TokenLBrace, "Expected '{' after switch")
	var cases []*ast.SwitchCaseNode
	for !p.Check(TokenRBrace) && !p.IsAtEnd() {
		if p.Match(TokenCase) {
			constant, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			p.Consume(TokenColon, "Expected ':' after case value")
			body, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			cases = append(cases, &ast.SwitchCaseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeSwitchCase, Line: constant.GetLine(), Column: constant.GetColumn()},
				Constant: constant,
				Body:     body.(*ast.BlockNode),
			})
		} else if p.Match(TokenDefault) {
			p.Consume(TokenColon, "Expected ':' after default")
			body, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			cases = append(cases, &ast.SwitchCaseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeSwitchCase, Line: p.Previous().Line, Column: p.Previous().Column},
				Constant: nil,
				Body:     body.(*ast.BlockNode),
			})
		} else {
			return nil, fmt.Errorf("Expected 'case' or 'default' in switch at line %d", p.Peek().Line)
		}
	}
	p.Consume(TokenRBrace, "Expected '}' after switch")
	return &ast.SwitchStmtNode{
		BaseNode: ast.BaseNode{Type: ast.NodeSwitchStmt, Line: value.GetLine(), Column: value.GetColumn()},
		Value:    value,
		Cases:    cases,
	}, nil
}

func (p *Parser) ParseTestStatement() (ast.ASTNode, error) {
	line, col := p.Peek().Line, p.Peek().Column
	nameTok := p.Consume(TokenString, "Expected test name string after 'test'")
	name := nameTok.Value
	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.TestStmtNode{
		BaseNode: ast.BaseNode{Type: ast.NodeTestStmt, Line: line, Column: col},
		Name:     name,
		Body:     body.(*ast.BlockNode),
	}, nil
}

func (p *Parser) ParseForStatement() (ast.ASTNode, error) {
	p.Consume(TokenLParen, "Expected '(' after 'for'")

	var initializer ast.ASTNode
	var err error

	if !p.Check(TokenSemicolon) {
		if p.MatchType() {
			typeToken := p.Previous()
			initializer, err = p.ParseVariableDeclarationRest(typeToken)
		} else {
			initializer, err = p.ParseExpression()
		}
		if err != nil {
			return nil, err
		}
	}

	p.Consume(TokenSemicolon, "Expected ';' after for initializer")

	var condition ast.ASTNode
	if !p.Check(TokenSemicolon) {
		condition, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	p.Consume(TokenSemicolon, "Expected ';' after for condition")

	var increment ast.ASTNode
	if !p.Check(TokenRParen) {
		increment, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	p.Consume(TokenRParen, "Expected ')' after for clauses")

	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.ForStmtNode{
		BaseNode:    ast.BaseNode{Type: ast.NodeForStmt, Line: body.GetLine(), Column: body.GetColumn()},
		Initializer: initializer,
		Condition:   condition,
		Increment:   increment,
		Body:        body.(*ast.BlockNode),
	}, nil
}

func (p *Parser) ParseDeferStatement() (ast.ASTNode, error) {
	// Check if next is a block { ... } or a single statement
	if p.Check(TokenLBrace) {
		// Original syntax: defer { ... }
		body, err := p.ParseBlock()
		if err != nil {
			return nil, err
		}
		return &ast.DeferStmtNode{
			BaseNode: ast.BaseNode{Type: ast.NodeDeferStmt, Line: p.Previous().Line, Column: p.Previous().Column},
			Body:     body.(*ast.BlockNode),
		}, nil
	}

	// Go-style syntax: defer expression;
	// Parse a single expression and wrap it in a block
	startTok := p.Previous()
	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	// Consume the semicolon
	p.Consume(TokenSemicolon, "Expected ';' after defer expression")

	// Wrap the expression in a block - the expression will be emitted as a statement
	body := &ast.BlockNode{
		BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: startTok.Line, Column: startTok.Column},
		Statements: []ast.ASTNode{expr},
	}

	return &ast.DeferStmtNode{
		BaseNode: ast.BaseNode{Type: ast.NodeDeferStmt, Line: startTok.Line, Column: startTok.Column},
		Body:     body,
	}, nil
}

func (p *Parser) ParseMatchStatement() (ast.ASTNode, error) {
	p.Consume(TokenLParen, "Expected '(' after match")
	value, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}
	p.Consume(TokenRParen, "Expected ')' after match value")
	p.Consume(TokenLBrace, "Expected '{' after match")

	var cases []*ast.CaseClauseNode
	for !p.Check(TokenRBrace) && !p.IsAtEnd() {
		if p.Match(TokenDefault) {
			p.Consume(TokenColon, "Expected ':' after default")
			body, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			cases = append(cases, &ast.CaseClauseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeCaseClause, Line: p.Previous().Line, Column: p.Previous().Column},
				TypeName: "",
				VarName:  "",
				Body:     body.(*ast.BlockNode),
			})
			continue
		}
		if !p.Match(TokenCase) {
			return nil, fmt.Errorf("expected case or default in match")
		}
		// case Ok(var): or case Err(var): or case Ok var: (Result pattern)
		if p.Check(TokenIdentifier) && (p.Peek().Value == "Ok" || p.Peek().Value == "Err") {
			typeName := p.Advance().Value
			varName := ""
			if p.Match(TokenLParen) {
				varName = p.Consume(TokenIdentifier, "Expected variable name in case Ok(var) or case Err(var)").Value
				p.Consume(TokenRParen, "Expected ')' after variable in case Ok(var)")
			} else if p.Check(TokenIdentifier) {
				varName = p.Advance().Value
			}
			p.Consume(TokenColon, "Expected ':' after case Ok/Err")
			body, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			cases = append(cases, &ast.CaseClauseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeCaseClause, Line: p.Previous().Line, Column: p.Previous().Column},
				TypeName: typeName,
				VarName:  varName,
				Body:     body.(*ast.BlockNode),
			})
			continue
		}
		// case type var: or case literal:
		if p.IsTypeToken(p.Peek().Type) {
			typeName := p.Advance().Value
			varName := ""
			if p.Check(TokenIdentifier) {
				varName = p.Advance().Value
			}
			p.Consume(TokenColon, "Expected ':' after case")
			body, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			cases = append(cases, &ast.CaseClauseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeCaseClause, Line: p.Previous().Line, Column: p.Previous().Column},
				TypeName: typeName,
				VarName:  varName,
				Body:     body.(*ast.BlockNode),
			})
		} else {
			lit, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			p.Consume(TokenColon, "Expected ':' after case literal")
			body, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			cases = append(cases, &ast.CaseClauseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeCaseClause, Line: p.Previous().Line, Column: p.Previous().Column},
				Literal:  lit,
				Body:     body.(*ast.BlockNode),
			})
		}
	}
	p.Consume(TokenRBrace, "Expected '}' after match")
	return &ast.MatchStmtNode{
		BaseNode: ast.BaseNode{Type: ast.NodeMatchStmt, Line: value.GetLine(), Column: value.GetColumn()},
		Value:    value,
		Cases:    cases,
	}, nil
}

func (p *Parser) ParseReturnStatement() (ast.ASTNode, error) {
	var value ast.ASTNode
	if !p.Check(TokenSemicolon) {
		var err error
		value, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	p.Consume(TokenSemicolon, "Expected ';' after return value")

	return &ast.ReturnStmtNode{
		BaseNode: ast.BaseNode{Type: ast.NodeReturnStmt, Line: p.Previous().Line, Column: p.Previous().Column},
		Value:    value,
	}, nil
}

func (p *Parser) ParseYieldStatement() (ast.ASTNode, error) {
	var value ast.ASTNode
	if !p.Check(TokenSemicolon) {
		var err error
		value, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	p.Consume(TokenSemicolon, "Expected ';' after yield value")

	return &ast.YieldStmtNode{
		BaseNode: ast.BaseNode{Type: ast.NodeYieldStmt, Line: p.Previous().Line, Column: p.Previous().Column},
		Value:    value,
	}, nil
}

func (p *Parser) ParseSpawnStatement() (ast.ASTNode, error) {
	line, col := p.Peek().Line, p.Peek().Column

	// Check for optional variable assignment: spawn var = func(args)
	var threadVar string
	if p.Check(TokenIdentifier) && p.position+1 < len(p.tokens) && p.tokens[p.position+1].Type == TokenAssign {
		threadVar = p.Advance().Value
		p.Consume(TokenAssign, "Expected '=' after spawn variable")
	}

	// Parse the function call (or just function name for no-arg functions)
	var funcExpr ast.ASTNode
	var args []ast.ASTNode

	// Get function name
	if p.Check(TokenIdentifier) {
		funcName := p.Advance()
		funcExpr = &ast.IdentifierNode{
			BaseNode: ast.BaseNode{Type: ast.NodeIdentifier, Line: funcName.Line, Column: funcName.Column},
			Name:     funcName.Value,
		}

		// Check for arguments - if no parens, call with no args
		if p.Check(TokenLParen) {
			p.Advance() // consume '('
			for !p.Check(TokenRParen) && !p.IsAtEnd() {
				arg, err := p.ParseExpression()
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
				if !p.Check(TokenRParen) {
					p.Consume(TokenComma, "Expected ',' after argument")
				}
			}
			p.Consume(TokenRParen, "Expected ')' after arguments")
		}
		// No parens = no arguments (simpler syntax)
	} else {
		// Parse as full expression
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		call, ok := expr.(*ast.CallExprNode)
		if !ok {
			return nil, fmt.Errorf("line %d: expected function call after 'spawn'", line)
		}
		funcExpr = call.Function
		args = call.Args
	}

	p.Consume(TokenSemicolon, "Expected ';' after spawn statement")

	return &ast.SpawnStmtNode{
		BaseNode:  ast.BaseNode{Type: ast.NodeSpawnStmt, Line: line, Column: col},
		Function:  funcExpr,
		Arguments: args,
		ThreadVar: threadVar,
	}, nil
}

func (p *Parser) ParseBlock() (ast.ASTNode, error) {
	if p.Check(TokenLBrace) {
		p.Advance()
	}
	tok := p.Peek()
	if p.IsAtEnd() {
		tok = p.Previous()
	}
	block := &ast.BlockNode{
		BaseNode: ast.BaseNode{Type: ast.NodeBlock, Line: tok.Line, Column: tok.Column},
	}

	for !p.Check(TokenRBrace) && !p.Check(TokenElse) && !p.IsAtEnd() {
		stmt, err := p.ParseDeclaration()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	if !p.IsAtEnd() && p.Check(TokenRBrace) {
		p.Consume(TokenRBrace, "Expected '}' after block")
	}

	return block, nil
}

func (p *Parser) ParseLambda() (ast.ASTNode, error) {
	line, col := p.Peek().Line, p.Peek().Column
	p.Consume(TokenLBracket, "expected '['")
	var captures []string
	if !p.Check(TokenRBracket) {
		for {
			tok := p.Consume(TokenIdentifier, "expected capture name or ']' in lambda")
			captures = append(captures, tok.Value)
			if p.Check(TokenRBracket) {
				break
			}
			p.Consume(TokenComma, "expected ',' or ']' after capture")
		}
	}
	p.Consume(TokenRBracket, "expected ']' after lambda")
	p.Consume(TokenLParen, "expected '(' after []")
	var parameters []*ast.ParameterNode
	for !p.Check(TokenRParen) && !p.IsAtEnd() {
		paramType := p.ConsumeType("expected parameter type")
		paramName := p.Consume(TokenIdentifier, "expected parameter name")
		parameters = append(parameters, &ast.ParameterNode{
			BaseNode: ast.BaseNode{Type: ast.NodeParameter, Line: paramType.Line, Column: paramType.Column},
			Type:     paramType.Value,
			Name:     paramName.Value,
		})
		if !p.Check(TokenRParen) {
			p.Consume(TokenComma, "expected ',' between parameters")
		}
	}
	p.Consume(TokenRParen, "expected ')' after parameters")
	returnType := ""
	if p.Match(TokenArrow) {
		returnType = p.ConsumeType("expected return type after ->").Value
	}
	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.LambdaNode{
		BaseNode:   ast.BaseNode{Type: ast.NodeLambda, Line: line, Column: col},
		Captures:   captures,
		Parameters: parameters,
		ReturnType: returnType,
		Body:       body.(*ast.BlockNode),
	}, nil
}

func (p *Parser) ParseExpressionStatement() (ast.ASTNode, error) {
	p.debugToken("ParseExpressionStatement start")
	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	p.Consume(TokenSemicolon, "Expected ';' after expression")
	return expr, nil
}

func (p *Parser) ParseExpression() (ast.ASTNode, error) {
	return p.ParseAssignment()
}

func (p *Parser) ParseAssignment() (ast.ASTNode, error) {
	expr, err := p.ParseLogicalOr()
	if err != nil {
		return nil, err
	}

	if p.Match(TokenAssign) {
		value, err := p.ParseAssignment()
		if err != nil {
			return nil, err
		}

		return &ast.AssignmentNode{
			BaseNode: ast.BaseNode{Type: ast.NodeAssignment, Line: expr.GetLine(), Column: expr.GetColumn()},
			Target:   expr,
			Value:    value,
		}, nil
	}

	return expr, nil
}

func (p *Parser) ParseLogicalOr() (ast.ASTNode, error) {
	expr, err := p.ParseLogicalAnd()
	if err != nil {
		return nil, err
	}

	for p.Match(TokenOr) {
		operator := p.Previous()
		right, err := p.ParseLogicalAnd()
		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeBinaryExpr, Line: operator.Line, Column: operator.Column},
			Left:     expr,
			Operator: operator.Value,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) ParseLogicalAnd() (ast.ASTNode, error) {
	expr, err := p.ParseEquality()
	if err != nil {
		return nil, err
	}

	for p.Match(TokenAnd) {
		operator := p.Previous()
		right, err := p.ParseEquality()
		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeBinaryExpr, Line: operator.Line, Column: operator.Column},
			Left:     expr,
			Operator: operator.Value,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) ParseEquality() (ast.ASTNode, error) {
	expr, err := p.ParseComparison()
	if err != nil {
		return nil, err
	}

	for p.Match(TokenEqual, TokenNotEqual) {
		operator := p.Previous()
		right, err := p.ParseComparison()
		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeBinaryExpr, Line: operator.Line, Column: operator.Column},
			Left:     expr,
			Operator: operator.Value,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) ParseComparison() (ast.ASTNode, error) {
	expr, err := p.ParseTerm()
	if err != nil {
		return nil, err
	}

	for p.Match(TokenLess, TokenLessEqual, TokenGreater, TokenGreaterEqual) {
		operator := p.Previous()
		right, err := p.ParseTerm()
		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeBinaryExpr, Line: operator.Line, Column: operator.Column},
			Left:     expr,
			Operator: operator.Value,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) ParseTerm() (ast.ASTNode, error) {
	expr, err := p.ParseFactor()
	if err != nil {
		return nil, err
	}

	for p.Match(TokenPlus, TokenMinus) {
		operator := p.Previous()
		right, err := p.ParseFactor()
		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeBinaryExpr, Line: operator.Line, Column: operator.Column},
			Left:     expr,
			Operator: operator.Value,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) ParseFactor() (ast.ASTNode, error) {
	expr, err := p.ParseUnary()
	if err != nil {
		return nil, err
	}

	for p.Match(TokenMultiply, TokenDivide, TokenModulo) {
		operator := p.Previous()
		right, err := p.ParseUnary()
		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeBinaryExpr, Line: operator.Line, Column: operator.Column},
			Left:     expr,
			Operator: operator.Value,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) ParseUnary() (ast.ASTNode, error) {
	// C-style cast: (type)expr - check for LParen followed by type name
	if p.Check(TokenLParen) {
		// Look ahead to see if this is a cast (type name followed by ))
		castType := p.tryParseCastType()
		if castType != "" {
			// We have a cast! Parse the operand
			operand, err := p.ParseUnary()
			if err != nil {
				return nil, err
			}
			return &ast.CastExprNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeCastExpr, Line: p.Peek().Line, Column: p.Peek().Column},
				TargetType: castType,
				Operand:    operand,
			}, nil
		}
	}

	// await expr — when async enabled, parse and return expr (no-op at codegen for now)
	if p.Check(TokenAwait) {
		if !p.features.Async {
			return nil, fmt.Errorf("line %d: await requires features.async (enable in config)", p.Peek().Line)
		}
		line, col := p.Previous().Line, p.Previous().Column
		p.Advance()
		expr, err := p.ParseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.AwaitExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeAwaitExpr, Line: line, Column: col},
			Expr:     expr,
		}, nil
	}
	if p.Match(TokenNot, TokenMinus, TokenIncrement, TokenDecrement) {
		operator := p.Previous()
		right, err := p.ParseUnary()
		if err != nil {
			return nil, err
		}

		return &ast.UnaryExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeUnaryExpr, Line: operator.Line, Column: operator.Column},
			Operator: operator.Value,
			Operand:  right,
		}, nil
	}

	return p.ParseCall()
}

// tryParseCastType attempts to parse a C-style cast type like (char*) or (int).
// Returns the type string if successful, or empty string if not a cast.
func (p *Parser) tryParseCastType() string {
	// Save parser state for rollback
	savedPos := p.position

	// Consume LParen
	if !p.Match(TokenLParen) {
		return ""
	}

	// Check for type identifier or type keyword (int, float, char, etc.)
	tok := p.Peek()
	var typeName string
	if p.isTypeToken(tok.Type) {
		typeName = tok.Value
		p.Advance()
	} else if tok.Type == TokenIdentifier {
		typeName = tok.Value
		p.Advance()
	} else {
		p.position = savedPos
		return ""
	}

	// Check for pointer suffix (e.g., char*, void**)
	for p.Match(TokenMultiply) {
		typeName += "*"
	}

	// Must have closing paren
	if !p.Match(TokenRParen) {
		p.position = savedPos
		return ""
	}

	// Verify this looks like a type (common C types or custom types)
	if !p.isValidCastType(typeName) {
		p.position = savedPos
		return ""
	}

	return typeName
}

// isTypeToken checks if the token is a type keyword
func (p *Parser) isTypeToken(tt TokenType) bool {
	switch tt {
	case TokenInt, TokenFloat, TokenDouble, TokenCharType, TokenBool, TokenVoid:
		return true
	default:
		return false
	}
}

// isValidCastType checks if the type name is a valid cast target
func (p *Parser) isValidCastType(typeName string) bool {
	// Common C types
	builtinTypes := map[string]bool{
		"int": true, "float": true, "double": true, "char": true,
		"void": true, "long": true, "short": true, "bool": true,
		"size_t": true, "int8_t": true, "int16_t": true, "int32_t": true,
		"int64_t": true, "uint8_t": true, "uint16_t": true, "uint32_t": true,
		"uint64_t": true, "FILE": true,
	}
	// Check base type (without pointer suffix)
	baseType := typeName
	for strings.HasSuffix(baseType, "*") {
		baseType = baseType[:len(baseType)-1]
	}
	return builtinTypes[baseType] || p.isCustomType(baseType)
}

// isCustomType checks if the name is a known custom type
func (p *Parser) isCustomType(name string) bool {
	// Check semantic analyzer for custom types (structs, typedefs, etc.)
	// For now, allow any identifier that starts with uppercase or common patterns
	return len(name) > 0 && (name[0] >= 'A' && name[0] <= 'Z') ||
		strings.HasSuffix(name, "_t")
}

func (p *Parser) ParseCall() (ast.ASTNode, error) {
	p.debugToken("ParseCall start")
	expr, err := p.ParsePrimary()
	if err != nil {
		return nil, err
	}
	p.debugToken("ParseCall after ParsePrimary")

	for {
		if p.Match(TokenLParen) {
			p.debugToken("ParseCall matched (, starting args")
			line, col := p.Peek().Line, p.Peek().Column
			var args []ast.ASTNode
			var namedArgs []*ast.NamedArgumentNode
			for !p.Check(TokenRParen) && !p.IsAtEnd() {
				p.debugToken("ParseCall arg loop iteration")
				// Check for named argument: identifier : value (don't consume identifier yet!)
				if p.Check(TokenIdentifier) && p.PeekNext() == TokenColon {
					p.Advance() // consume identifier
					name := p.Previous().Value
					p.Consume(TokenColon, "expected ':' after named argument")
					value, err := p.ParseExpression()
					if err != nil {
						return nil, err
					}
					namedArgs = append(namedArgs, &ast.NamedArgumentNode{
						BaseNode: ast.BaseNode{Type: ast.NodeNamedArgument, Line: line, Column: col},
						Name:     name,
						Value:    value,
					})
				} else {
					arg, err := p.ParseExpression()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
				}
				if !p.Check(TokenRParen) {
					p.Consume(TokenComma, "expected ',' between arguments")
				}
			}
			p.Consume(TokenRParen, "expected ')' after arguments")
			expr = &ast.CallExprNode{
				BaseNode:  ast.BaseNode{Type: ast.NodeCallExpr, Line: line, Column: col},
				Function:  expr,
				Args:      args,
				NamedArgs: namedArgs,
			}
		} else if p.Match(TokenDot) {
			member := p.Consume(TokenIdentifier, "Expected property name after '.'")
			expr = &ast.MemberAccessNode{
				BaseNode: ast.BaseNode{Type: ast.NodeMemberAccess, Line: member.Line, Column: member.Column},
				Object:   expr,
				Member:   member.Value,
			}
		} else if p.Match(TokenLBracket) {
			index, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			p.Consume(TokenRBracket, "Expected ']' after index")

			expr = &ast.IndexExprNode{
				BaseNode: ast.BaseNode{Type: ast.NodeArrayAccess, Line: expr.GetLine(), Column: expr.GetColumn()},
				Object:   expr,
				Index:    index,
			}
		} else if p.Match(TokenIncrement) {
			expr = &ast.UnaryExprNode{
				BaseNode:  ast.BaseNode{Type: ast.NodeUnaryExpr, Line: expr.GetLine(), Column: expr.GetColumn()},
				Operator:  "++",
				Operand:   expr,
				IsPostfix: true,
			}
		} else if p.Match(TokenDecrement) {
			expr = &ast.UnaryExprNode{
				BaseNode:  ast.BaseNode{Type: ast.NodeUnaryExpr, Line: expr.GetLine(), Column: expr.GetColumn()},
				Operator:  "--",
				Operand:   expr,
				IsPostfix: true,
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) ParsePrimary() (ast.ASTNode, error) {
	p.debugToken("ParsePrimary")
	if p.Match(TokenTrue) {
		return &ast.LiteralNode{
			BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: p.Previous().Line, Column: p.Previous().Column},
			Value:    true,
			Type:     "bool",
		}, nil
	}

	if p.Match(TokenFalse) {
		return &ast.LiteralNode{
			BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: p.Previous().Line, Column: p.Previous().Column},
			Value:    false,
			Type:     "bool",
		}, nil
	}

	if p.Match(TokenNull) {
		return &ast.LiteralNode{
			BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: p.Previous().Line, Column: p.Previous().Column},
			Value:    nil,
			Type:     "null",
		}, nil
	}

	// Allow type keywords to be used as identifiers in expressions (e.g., var result = ...)
	if p.IsTypeToken(p.Peek().Type) && !p.Check(TokenVoid) && !p.Check(TokenVar) {
		tok := p.Advance()
		return &ast.IdentifierNode{
			BaseNode: ast.BaseNode{Type: ast.NodeIdentifier, Line: tok.Line, Column: tok.Column},
			Name:     tok.Value,
		}, nil
	}

	if p.Match(TokenNumber) {
		return &ast.LiteralNode{
			BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: p.Previous().Line, Column: p.Previous().Column},
			Value:    p.Previous().Value,
			Type:     "number",
		}, nil
	}

	if p.Match(TokenString) {
		val := p.Previous().Value
		line, col := p.Previous().Line, p.Previous().Column
		if strings.Contains(val, "${") {
			var parts []ast.ASTNode
			remaining := val
			for strings.Contains(remaining, "${") {
				i := strings.Index(remaining, "${")
				if i > 0 {
					parts = append(parts, &ast.LiteralNode{
						BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: line, Column: col},
						Value:    remaining[:i],
						Type:     "string",
					})
				}
				remaining = remaining[i+2:]
				j := strings.Index(remaining, "}")
				if j < 0 {
					return nil, fmt.Errorf("unclosed ${ in string at line %d", line)
				}
				id := strings.TrimSpace(remaining[:j])
				if id == "" {
					return nil, fmt.Errorf("empty ${} in string at line %d", line)
				}
				parts = append(parts, &ast.IdentifierNode{
					BaseNode: ast.BaseNode{Type: ast.NodeIdentifier, Line: line, Column: col},
					Name:     id,
				})
				remaining = remaining[j+1:]
			}
			if remaining != "" {
				parts = append(parts, &ast.LiteralNode{
					BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: line, Column: col},
					Value:    remaining,
					Type:     "string",
				})
			}
			return &ast.InterpolatedStringNode{
				BaseNode: ast.BaseNode{Type: ast.NodeInterpolatedString, Line: line, Column: col},
				Parts:    parts,
			}, nil
		}
		return &ast.LiteralNode{
			BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: line, Column: col},
			Value:    val,
			Type:     "string",
		}, nil
	}

	if p.Match(TokenChar) {
		return &ast.LiteralNode{
			BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: p.Previous().Line, Column: p.Previous().Column},
			Value:    p.Previous().Value,
			Type:     "char",
		}, nil
	}

	if p.Match(TokenIdentifier) {
		return &ast.IdentifierNode{
			BaseNode: ast.BaseNode{Type: ast.NodeIdentifier, Line: p.Previous().Line, Column: p.Previous().Column},
			Name:     p.Previous().Value,
		}, nil
	}

	if p.Match(TokenLParen) {
		// Check for arrow function: () => expr or (params) => expr
		savedPos := p.position - 1 // position of the '('
		isArrowFunc := false

		// Look ahead to detect arrow function pattern
		if p.Check(TokenRParen) {
			// () => ... - empty params
			p.Advance() // consume ')'
			if p.Check(TokenFatArrow) {
				isArrowFunc = true
			} else {
				// Not arrow func, restore to before '(' for expression parsing
				p.position = savedPos
			}
		} else if p.Check(TokenIdentifier) {
			// Check for (x, y) => pattern
			paramNames := []string{}
			for p.Check(TokenIdentifier) {
				paramNames = append(paramNames, p.Peek().Value)
				p.Advance()
				if p.Check(TokenComma) {
					p.Advance()
				}
			}
			if p.Check(TokenRParen) {
				p.Advance() // consume ')'
				if p.Check(TokenFatArrow) {
					isArrowFunc = true
					// Create parameters from collected names
					_ = paramNames // Will use these for typed params
				} else {
					// Not arrow func, restore position
					p.position = savedPos
				}
			} else {
				// Not arrow func, restore position
				p.position = savedPos
			}
		} else {
			// Not arrow func, restore position for expression parsing
			p.position = savedPos
		}

		if isArrowFunc {
			p.Advance() // consume '=>'

			// Parse body - either single expression or block
			var body *ast.BlockNode
			if p.Check(TokenLBrace) {
				bodyNode, err := p.ParseBlock()
				if err != nil {
					return nil, err
				}
				body = bodyNode.(*ast.BlockNode)
			} else {
				// Single expression body
				exprBody, err := p.ParseExpression()
				if err != nil {
					return nil, err
				}
				body = &ast.BlockNode{
					BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: exprBody.GetLine(), Column: exprBody.GetColumn()},
					Statements: []ast.ASTNode{exprBody},
				}
			}

			return &ast.LambdaNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeLambda, Line: p.Peek().Line, Column: p.Peek().Column},
				Captures:   []string{},
				Parameters: []*ast.ParameterNode{},
				ReturnType: "",
				Body:       body,
			}, nil
		}

		// Regular parenthesized expression - re-consume the '('
		p.Advance() // consume '(' again
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		if p.Match(TokenComma) {
			// Tuple: ( expr , expr , ... )
			elements := []ast.ASTNode{expr}
			for {
				e, err := p.ParseExpression()
				if err != nil {
					return nil, err
				}
				elements = append(elements, e)
				if !p.Match(TokenComma) {
					break
				}
			}
			p.Consume(TokenRParen, "Expected ')' after tuple")
			return &ast.TupleExprNode{
				BaseNode: ast.BaseNode{Type: ast.NodeTupleExpr, Line: expr.GetLine(), Column: expr.GetColumn()},
				Elements: elements,
			}, nil
		}
		p.Consume(TokenRParen, "Expected ')' after expression")
		return expr, nil
	}

	// Dict literal: { "key": value, ... } — keys must be string literals
	if p.Check(TokenLBrace) && p.PeekAhead(1).Type == TokenString && p.PeekAhead(2).Type == TokenColon {
		return p.ParseDictLiteral()
	}

	// Struct literal: { field: value, ... } — keys are identifiers
	if p.Check(TokenLBrace) && p.PeekAhead(1).Type == TokenIdentifier && p.PeekAhead(2).Type == TokenColon {
		return p.ParseStructLiteral()
	}

	// Lambda: [] (params) or [capture...] (params) — must come before array literal
	if p.Check(TokenLBracket) && p.CanBeLambda() {
		return p.ParseLambda()
	}
	if p.Match(TokenLBracket) {
		var elements []ast.ASTNode
		for !p.Check(TokenRBracket) && !p.IsAtEnd() {
			e, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, e)
			if !p.Check(TokenRBracket) {
				p.Consume(TokenComma, "Expected ',' or ']' in array literal")
			}
		}
		p.Consume(TokenRBracket, "Expected ']' after array literal")
		node := &ast.ArrayLiteralNode{
			BaseNode: ast.BaseNode{Type: ast.NodeArrayLiteral, Line: p.Previous().Line, Column: p.Previous().Column},
			Elements: elements,
		}
		p.AnnotateArrayLiteral(node)
		return node, nil
	}

	return nil, fmt.Errorf("Unexpected token '%s' at line %d, column %d",
		p.Peek().Value, p.Peek().Line, p.Peek().Column)
}

func (p *Parser) ParseDictLiteral() (ast.ASTNode, error) {
	line, col := p.Peek().Line, p.Peek().Column
	p.Consume(TokenLBrace, "Expected '{' for dict literal")
	var entries []ast.DictEntry
	for !p.Check(TokenRBrace) && !p.IsAtEnd() {
		if !p.Check(TokenString) {
			return nil, fmt.Errorf("dict literal key must be a string literal at line %d", p.Peek().Line)
		}
		keyTok := p.Advance()
		keyStr := keyTok.Value // lexer already returns string content without quotes
		p.Consume(TokenColon, "Expected ':' after dict key")
		val, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		entries = append(entries, ast.DictEntry{Key: keyStr, Value: val})
		if !p.Check(TokenRBrace) {
			p.Consume(TokenComma, "Expected ',' or '}' in dict literal")
		}
	}
	p.Consume(TokenRBrace, "Expected '}' after dict literal")
	return &ast.DictLiteralNode{
		BaseNode: ast.BaseNode{Type: ast.NodeDictLiteral, Line: line, Column: col},
		Entries:  entries,
	}, nil
}

func (p *Parser) ParseStructLiteral() (ast.ASTNode, error) {
	line, col := p.Peek().Line, p.Peek().Column
	p.Consume(TokenLBrace, "Expected '{' for struct literal")
	var fields []ast.StructFieldInit
	for !p.Check(TokenRBrace) && !p.IsAtEnd() {
		if !p.Check(TokenIdentifier) {
			return nil, fmt.Errorf("struct field name must be an identifier at line %d", p.Peek().Line)
		}
		nameTok := p.Advance()

		var val ast.ASTNode
		var err error

		// Check for shorthand: field name without colon means field: field (variable with same name)
		if p.Check(TokenComma) || p.Check(TokenRBrace) {
			// Shorthand: use identifier as both name and value
			val = &ast.IdentifierNode{
				BaseNode: ast.BaseNode{Type: ast.NodeIdentifier, Line: nameTok.Line, Column: nameTok.Column},
				Name:     nameTok.Value,
			}
		} else {
			// Regular: field: value
			p.Consume(TokenColon, "Expected ':' after struct field name")
			val, err = p.ParseExpression()
			if err != nil {
				return nil, err
			}
		}
		fields = append(fields, ast.StructFieldInit{Name: nameTok.Value, Value: val})
		if !p.Check(TokenRBrace) {
			p.Consume(TokenComma, "Expected ',' or '}' in struct literal")
		}
	}
	p.Consume(TokenRBrace, "Expected '}' after struct literal")
	return &ast.StructLiteralNode{
		BaseNode: ast.BaseNode{Type: ast.NodeStructLiteral, Line: line, Column: col},
		Fields:   fields,
	}, nil
}

func (p *Parser) AnnotateArrayLiteral(node *ast.ArrayLiteralNode) {
	if node == nil {
		return
	}
	node.Dimensions = 1
	node.RowCount = len(node.Elements)
	node.RowLength = len(node.Elements)
	if len(node.Elements) == 0 {
		return
	}
	allRows := true
	for _, elem := range node.Elements {
		if child, ok := elem.(*ast.ArrayLiteralNode); ok {
			p.AnnotateArrayLiteral(child)
		} else {
			allRows = false
		}
	}
	if !allRows {
		return
	}
	rowLen := -1
	for _, elem := range node.Elements {
		child, ok := elem.(*ast.ArrayLiteralNode)
		if !ok || child.Dimensions != 1 {
			return
		}
		if rowLen == -1 {
			rowLen = len(child.Elements)
		}
	}
	node.Dimensions = 2
	node.RowCount = len(node.Elements)
	if rowLen < 0 {
		rowLen = 0
	}
	node.RowLength = rowLen
}

// Helper methods
func (p *Parser) Match(types ...TokenType) bool {
	for _, tokenType := range types {
		if p.Check(tokenType) {
			p.Advance()
			return true
		}
	}
	return false
}

func (p *Parser) MatchType() bool {
	typeTokens := []TokenType{TokenVoid, TokenInt, TokenFloat, TokenDouble, TokenCharType, TokenBool, TokenStringType, TokenVec2, TokenVec3, TokenVar, TokenAny, TokenConst, TokenArray, TokenDict, TokenResult, TokenEvent, TokenGuiWindow, TokenGuiWidget, TokenGuiContainer, TokenGuiEvent}
	return p.Match(typeTokens...)
}

func (p *Parser) ConsumeTypeOrVar() Token {
	typeTokens := []TokenType{TokenVoid, TokenInt, TokenFloat, TokenDouble, TokenCharType, TokenBool, TokenStringType, TokenVec2, TokenVec3, TokenVar, TokenAny, TokenConst, TokenArray, TokenDict, TokenResult, TokenEvent, TokenGuiWindow, TokenGuiWidget, TokenGuiContainer, TokenGuiEvent}
	for _, tokenType := range typeTokens {
		if p.Check(tokenType) {
			return p.Advance()
		}
	}
	if p.Check(TokenIdentifier) {
		return p.Advance()
	}
	panic(fmt.Sprintf("Expected type or 'var'. Got %s instead", p.Peek().Value))
}

func (p *Parser) ConsumeType(errorMessage string) Token {
	typeTokens := []TokenType{TokenVoid, TokenInt, TokenFloat, TokenDouble, TokenCharType, TokenBool, TokenStringType, TokenVec2, TokenVec3, TokenVar, TokenAny, TokenConst, TokenArray, TokenDict, TokenResult, TokenEvent, TokenGuiWindow, TokenGuiWidget, TokenGuiContainer, TokenGuiEvent}
	for _, tokenType := range typeTokens {
		if p.Check(tokenType) {
			tok := p.Advance()
			// Check for optional type suffix: int? means optional int
			if p.Check(TokenQuestion) {
				p.Advance() // consume '?'
				tok.Value = tok.Value + "?"
			}
			// Check for union type: int | float | string
			for p.Check(TokenPipe) {
				p.Advance() // consume '|'
				tok.Value = tok.Value + " | "
				// Get next type in union
				nextTok := p.ConsumeType(errorMessage)
				tok.Value = tok.Value + nextTok.Value
			}
			return tok
		}
	}

	tok := p.Consume(TokenIdentifier, errorMessage)
	// Check for optional type suffix on identifier types (struct names)
	if p.Check(TokenQuestion) {
		p.Advance() // consume '?'
		tok.Value = tok.Value + "?"
	}
	// Check for union type
	for p.Check(TokenPipe) {
		p.Advance() // consume '|'
		tok.Value = tok.Value + " | "
		nextTok := p.ConsumeType(errorMessage)
		tok.Value = tok.Value + nextTok.Value
	}
	return tok
}

func (p *Parser) IsTypeToken(tokenType TokenType) bool {
	typeTokens := []TokenType{TokenVoid, TokenInt, TokenFloat, TokenDouble, TokenCharType, TokenBool, TokenStringType, TokenVec2, TokenVec3, TokenVar, TokenAny, TokenConst, TokenArray, TokenDict, TokenResult, TokenEvent, TokenGuiWindow, TokenGuiWidget, TokenGuiContainer, TokenGuiEvent}
	for _, tt := range typeTokens {
		if tt == tokenType {
			return true
		}
	}
	return false
}

func (p *Parser) PeekNext() TokenType {
	if p.position+1 >= len(p.tokens) {
		return TokenEOF
	}
	return p.tokens[p.position+1].Type
}

func (p *Parser) Check(tokenType TokenType) bool {
	if p.IsAtEnd() {
		return false
	}
	return p.Peek().Type == tokenType
}

func (p *Parser) Advance() Token {
	if !p.IsAtEnd() {
		p.position++
	}
	return p.Previous()
}

func (p *Parser) IsAtEnd() bool {
	return p.Peek().Type == TokenEOF
}

func (p *Parser) Peek() Token {
	return p.tokens[p.position]
}

func (p *Parser) PeekAhead(n int) Token {
	pos := p.position + n
	if pos < 0 || pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[pos]
}

// canBeLambda returns true if the current token is [ and the following tokens form a lambda: [] ( or [id, ...] (
func (p *Parser) CanBeLambda() bool {
	if p.Peek().Type != TokenLBracket {
		return false
	}
	// No-capture: ] (
	if p.PeekAhead(1).Type == TokenRBracket && p.PeekAhead(2).Type == TokenLParen {
		return true
	}
	// Capture list: id ( "," id )* ] (
	i := 1
	for {
		t := p.PeekAhead(i)
		if t.Type == TokenRBracket {
			return p.PeekAhead(i+1).Type == TokenLParen
		}
		if t.Type == TokenIdentifier {
			i++
			if p.PeekAhead(i).Type == TokenComma {
				i++ // skip comma
				continue
			}
			if p.PeekAhead(i).Type == TokenRBracket {
				return p.PeekAhead(i+1).Type == TokenLParen
			}
			return false
		}
		return false
	}
}

func (p *Parser) Previous() Token {
	return p.tokens[p.position-1]
}

func (p *Parser) Consume(tokenType TokenType, message string) Token {
	if p.Check(tokenType) {
		return p.Advance()
	}

	panic(fmt.Sprintf("%s. Got %s instead", message, p.Peek().Value))
}
