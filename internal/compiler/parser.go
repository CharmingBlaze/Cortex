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
		// Support: async fn name(), async returnType name(), async varName = expr
		var typeToken Token
		var nameToken Token

		// Check for 'fn' BEFORE ConsumeType since fn is in IsTypeToken
		if p.Match(TokenFn) {
			// async fn name(params) -> returnType
			nameToken = p.Consume(TokenIdentifier, "Expected function name after 'async fn'")
			typeToken = Token{Type: TokenFn, Value: "fn"}
		} else if p.Check(TokenIdentifier) {
			// Could be: async name = expr (variable) or async Type name() (function)
			// Peek ahead to check for = vs (
			savedPos := p.position
			nameToken = p.Advance()
			if p.Match(TokenAssign) {
				// async varName = expr - variable declaration with async modifier
				value, err := p.ParseExpression()
				if err != nil {
					return nil, err
				}
				p.Consume(TokenSemicolon, "Expected ';' after async variable declaration")
				return &ast.VariableDeclNode{
					BaseNode:    ast.BaseNode{Type: ast.NodeVariableDecl, Line: nameToken.Line, Column: nameToken.Column},
					Name:        nameToken.Value,
					Type:        "async",
					Initializer: value,
				}, nil
			}
			// Not a variable, restore and continue as function
			p.position = savedPos
			typeToken = p.ConsumeType("Expected return type after 'async'")
			for p.Match(TokenMultiply) {
				typeToken.Value += "*"
			}
			nameToken = p.Consume(TokenIdentifier, "Expected function name after 'async'")
		} else {
			// async returnType name(params)
			typeToken = p.ConsumeType("Expected return type after 'async'")
			for p.Match(TokenMultiply) {
				typeToken.Value += "*"
			}
			nameToken = p.Consume(TokenIdentifier, "Expected function name after 'async'")
		}
		fn, err := p.ParseFunctionDeclaration(typeToken, nameToken)
		if err != nil {
			return nil, err
		}
		fn.(*ast.FunctionDeclNode).IsAsync = true
		// If using 'fn' syntax, default to void and check for -> returnType
		if typeToken.Type == TokenFn {
			fn.(*ast.FunctionDeclNode).ReturnType = "void"
			if p.Match(TokenArrow) {
				returnTypeTok := p.ConsumeType("Expected return type after ->")
				fn.(*ast.FunctionDeclNode).ReturnType = returnTypeTok.Value
			}
		}
		return fn, nil
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

	// Handle visibility modifiers (public/private)
	isPublic := p.Match(TokenPublic)
	isPrivate := p.Match(TokenPrivate)
	_ = isPublic // Track for semantic analysis
	_ = isPrivate

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
	if p.Match(TokenTypeKeyword) {
		return p.ParseTypeDeclaration()
	}
	if p.Match(TokenEnum) {
		return p.ParseEnumDeclaration()
	}

	// Handle fn keyword for function declarations: fn name(params) -> returnType
	if p.Match(TokenFn) {
		// Accept identifier or keyword as function name (e.g., "test" is a keyword but valid fn name)
		var nameToken Token
		if p.Check(TokenIdentifier) || p.IsTypeToken(p.Peek().Type) || p.Peek().Type == TokenTest {
			nameToken = p.Advance()
		} else {
			nameToken = p.Consume(TokenIdentifier, "Expected function name after 'fn'")
		}
		fn, err := p.ParseFunctionDeclaration(Token{Type: TokenFn, Value: "fn"}, nameToken)
		if err != nil {
			return nil, err
		}
		// ParseFunctionDeclaration already handles -> returnType, so we don't need to override here
		return fn, nil
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
	// Accept identifier or type token as struct name (e.g., Vec2 is a type token but valid struct name)
	var nameToken Token
	if p.Check(TokenIdentifier) || p.IsTypeToken(p.Peek().Type) {
		nameToken = p.Advance()
	} else {
		nameToken = p.Consume(TokenIdentifier, "Expected struct name")
	}
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
		// Support both identifier and string literal enum values
		var valueName string
		if p.Check(TokenString) {
			// String literal as enum value: "active", "inactive"
			strTok := p.Advance()
			valueName = strTok.Value
			stringValues[valueName] = strTok.Value
		} else {
			valueToken := p.Consume(TokenIdentifier, "Expected enum value")
			valueName = valueToken.Value
		}
		values = append(values, valueName)

		// Check for explicit value: = "string" or = number
		if p.Match(TokenAssign) {
			if p.Check(TokenString) {
				// String enum: Red = "red"
				strTok := p.Advance()
				stringValues[valueName] = strTok.Value
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

// ParseTypeDeclaration parses TYPE...ENDTYPE user-defined type blocks
// Syntax: TYPE TypeName
//
//	  FieldName1
//	  FieldName2 AS Type
//	ENDTYPE
func (p *Parser) ParseTypeDeclaration() (ast.ASTNode, error) {
	// Type name can be an identifier OR a type keyword (e.g., Vec2, Vec3)
	var nameToken Token
	if p.IsTypeToken(p.Peek().Type) {
		nameToken = p.Advance()
	} else {
		nameToken = p.Consume(TokenIdentifier, "Expected type name after 'type'")
	}

	var fields []*ast.VariableDeclNode

	// Parse fields until ENDTYPE
	for !p.Check(TokenEndType) && !p.IsAtEnd() {
		// Field name comes first - can be identifier or type keyword
		var fieldNameTok Token
		if p.IsTypeToken(p.Peek().Type) {
			fieldNameTok = p.Advance()
		} else {
			fieldNameTok = p.Consume(TokenIdentifier, "Expected field name in type definition")
		}

		// Check for AS Type syntax, otherwise default to int
		fieldType := "int" // default type
		if p.Match(TokenAs) {
			// Parse the type - can be a keyword type or identifier (custom type)
			if p.IsTypeToken(p.Peek().Type) {
				typeTok := p.ConsumeType("Expected type after 'as'")
				fieldType = typeTok.Value
			} else if p.Check(TokenIdentifier) {
				// Custom type reference (nested types)
				typeTok := p.Advance()
				fieldType = typeTok.Value
			} else {
				return nil, fmt.Errorf("expected type after 'as' at line %d", p.Peek().Line)
			}
		}

		fields = append(fields, &ast.VariableDeclNode{
			BaseNode:    ast.BaseNode{Type: ast.NodeVariableDecl, Line: fieldNameTok.Line, Column: fieldNameTok.Column},
			Name:        fieldNameTok.Value,
			Type:        fieldType,
			Initializer: nil,
		})

		// Optional semicolon or newline between fields
		p.Match(TokenSemicolon)
	}

	p.Consume(TokenEndType, "Expected 'endtype' to close type definition")

	return &ast.StructDeclNode{
		BaseNode: ast.BaseNode{Type: ast.NodeStructDecl, Line: nameToken.Line, Column: nameToken.Column},
		Name:     nameToken.Value,
		Module:   p.currentModule,
		Fields:   fields,
		Methods:  nil,
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

	// Handle array type syntax: Type[] name or Type[][] name
	for p.Match(TokenLBracket) {
		p.Consume(TokenRBracket, "Expected ']' after '[' for array type")
		typeToken.Value += "[]"
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
			// Handle array type syntax for const: const Type[] name or const Type[][] name
			for p.Match(TokenLBracket) {
				p.Consume(TokenRBracket, "Expected ']' after '[' for array type")
				actualType.Value += "[]"
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

	// Check for visibility modifiers (public/private)
	isPublic := p.Match(TokenPublic)
	isPrivate := p.Match(TokenPrivate)
	_ = isPublic // Track visibility for semantic analysis
	_ = isPrivate

	p.Consume(TokenLParen, "Expected '(' after function name")

	var parameters []*ast.ParameterNode
	for !p.Check(TokenRParen) && !p.IsAtEnd() {
		paramType := p.ConsumeType("Expected parameter type")
		for p.Match(TokenMultiply) {
			paramType.Value += "*"
		}
		paramName := p.Consume(TokenIdentifier, "Expected parameter name")
		var defaultVal ast.ASTNode
		// Support both = and : for default values
		if p.Match(TokenAssign) || p.Match(TokenColon) {
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
	// Handle 'fn' keyword - default to void
	if typeToken.Type == TokenFn {
		returnType = "void"
	}
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
	isVar := typeToken.Type == TokenVar || typeToken.Type == TokenLet || typeToken.Type == TokenFn

	// The actual type - for const, this was already resolved by the caller
	actualType := typeToken.Value

	// Handle var/let/fn as type inference
	if isVar {
		actualType = "var"
	}

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
	// Handle async variable declaration inside functions: async name = expr;
	if p.Match(TokenAsync) {
		// Check if this is a variable declaration: async name = expr
		if p.Check(TokenIdentifier) {
			nameToken := p.Advance()
			if p.Match(TokenAssign) {
				value, err := p.ParseExpression()
				if err != nil {
					return nil, err
				}
				p.Consume(TokenSemicolon, "Expected ';' after async variable declaration")
				return &ast.VariableDeclNode{
					BaseNode:    ast.BaseNode{Type: ast.NodeVariableDecl, Line: nameToken.Line, Column: nameToken.Column},
					Name:        nameToken.Value,
					Type:        "async",
					Initializer: value,
				}, nil
			}
			// Not a variable assignment, put tokens back
			p.position -= 2 // back up over name and async
		} else {
			// Not a variable, put async token back
			p.position--
		}
	}
	if p.Match(TokenIf) {
		return p.ParseIfStatement()
	}
	if p.Match(TokenUnless) {
		return p.ParseUnlessStatement()
	}
	if p.Match(TokenDo) {
		return p.ParseDoWhileStatement()
	}
	if p.Match(TokenWhile) {
		return p.ParseWhileStatement()
	}
	if p.Match(TokenLoop) {
		return p.ParseLoopStatement()
	}
	// Handle for statement - check for for-in vs standard for
	if p.Match(TokenFor) {
		p.Consume(TokenLParen, "Expected '(' after 'for'")

		// Check for for-in: for (var in collection)
		// Peek at the token after the first identifier
		if p.Check(TokenIdentifier) || p.Check(TokenVar) {
			// Save position in case this is a standard for loop
			savedPos := p.position

			// Try to parse as for-in
			var varName string
			if p.Match(TokenVar) {
				varName = p.Consume(TokenIdentifier, "Expected variable name after 'var'").Value
			} else {
				varName = p.Advance().Value
			}

			if p.Match(TokenIn) {
				// This is for-in
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
					VarName:    varName,
					Collection: collection,
					Body:       body.(*ast.BlockNode),
				}, nil
			}

			// Not for-in, restore position and parse as standard for
			p.position = savedPos
		}

		// Standard for loop: for (init; cond; incr)
		var initializer ast.ASTNode
		if !p.Check(TokenSemicolon) {
			// Could be a variable declaration or expression
			if p.IsTypeToken(p.Peek().Type) || p.Check(TokenVar) {
				decl, err := p.ParseVariableDeclaration()
				if err != nil {
					return nil, err
				}
				initializer = decl
			} else {
				init, err := p.ParseExpression()
				if err != nil {
					return nil, err
				}
				initializer = init
			}
		}
		p.Consume(TokenSemicolon, "Expected ';' after for initializer")

		var condition ast.ASTNode
		if !p.Check(TokenSemicolon) {
			cond, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			condition = cond
		}
		p.Consume(TokenSemicolon, "Expected ';' after for condition")

		var increment ast.ASTNode
		if !p.Check(TokenRParen) {
			incr, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			increment = incr
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
	if p.Match(TokenSelect) {
		return p.ParseSelectStatement()
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
	if p.Match(TokenTry) {
		return p.ParseTryStatement()
	}
	if p.Match(TokenThrow) {
		return p.ParseThrowStatement()
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
		return decl, nil
	}

	// Check for var AS Type syntax (user-defined type style)
	// Pattern: identifier AS Type
	if p.position+2 < len(p.tokens) && p.tokens[p.position].Type == TokenIdentifier && p.tokens[p.position+1].Type == TokenAs {
		// Parse: varName AS Type
		nameTok := p.Advance()
		p.Advance() // consume AS
		typeTok := p.ConsumeType("Expected type after 'as'")
		var initializer ast.ASTNode
		var arraySize int = -1

		// Check for array size: var AS Type[size]
		if p.Match(TokenLBracket) {
			if p.Check(TokenNumber) {
				sizeTok := p.Advance()
				arraySize = 0
				if s, err := strconv.Atoi(sizeTok.Value); err == nil {
					arraySize = s
				}
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
		p.Consume(TokenSemicolon, "Expected ';' after variable declaration")

		node := &ast.VariableDeclNode{
			BaseNode:    ast.BaseNode{Type: ast.NodeVariableDecl, Line: nameTok.Line, Column: nameTok.Column},
			Name:        nameTok.Value,
			Module:      p.currentModule,
			Type:        typeTok.Value,
			Initializer: initializer,
		}
		if arraySize >= 0 {
			node.Type = typeTok.Value + "[]"
			node.ArraySize = arraySize
		}
		return node, nil
	}

	// Variable declarations in statements need semicolons
	if p.position < len(p.tokens) && p.IsTypeToken(p.tokens[p.position].Type) {
		decl, err := p.ParseVariableDeclaration()
		if err != nil {
			return nil, err
		}
		if !p.Match(TokenSemicolon) {
			return nil, fmt.Errorf("Expected ';' after variable declaration")
		}
		return decl, nil
	}

	return p.ParseExpressionStatement()
}

func (p *Parser) ParseIfStatement() (ast.ASTNode, error) {
	// Check for if let pattern matching
	if p.Match(TokenLet) {
		return p.ParseIfLetStatement()
	}

	// Optional parentheses around condition
	var condition ast.ASTNode
	var err error

	if p.Match(TokenLParen) {
		// Traditional: if (condition) { ... }
		condition, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
		p.Consume(TokenRParen, "Expected ')' after if condition")
	} else {
		// Modern: if condition { ... } or if condition statement;
		condition, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	// Body can be a block { } or a single statement
	var thenBranch *ast.BlockNode
	if p.Check(TokenLBrace) {
		body, err := p.ParseBlock()
		if err != nil {
			return nil, err
		}
		thenBranch = body.(*ast.BlockNode)
	} else {
		// Single statement body
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		thenBranch = &ast.BlockNode{
			BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: stmt.GetLine(), Column: stmt.GetColumn()},
			Statements: []ast.ASTNode{stmt},
		}
	}

	var elseBranch *ast.BlockNode
	if p.Match(TokenElse) {
		if p.Check(TokenIf) {
			p.Advance() // consume "if" so parseIfStatement sees optional parens
			elseNode, err := p.ParseIfStatement()
			if err != nil {
				return nil, err
			}
			// Wrap else-if in a block
			elseBranch = &ast.BlockNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: elseNode.GetLine(), Column: elseNode.GetColumn()},
				Statements: []ast.ASTNode{elseNode},
			}
		} else if p.Check(TokenElif) {
			// elif is sugar for else if
			p.Advance() // consume "elif"
			elseNode, err := p.ParseIfStatement()
			if err != nil {
				return nil, err
			}
			elseBranch = &ast.BlockNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: elseNode.GetLine(), Column: elseNode.GetColumn()},
				Statements: []ast.ASTNode{elseNode},
			}
		} else if p.Check(TokenLBrace) {
			elseBlock, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			elseBranch = elseBlock.(*ast.BlockNode)
		} else {
			// Single statement else
			elseStmt, err := p.ParseStatement()
			if err != nil {
				return nil, err
			}
			elseBranch = &ast.BlockNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: elseStmt.GetLine(), Column: elseStmt.GetColumn()},
				Statements: []ast.ASTNode{elseStmt},
			}
		}
	} else if p.Match(TokenElif) {
		// elif without else (elif as direct continuation)
		elseNode, err := p.ParseIfStatement()
		if err != nil {
			return nil, err
		}
		elseBranch = &ast.BlockNode{
			BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: elseNode.GetLine(), Column: elseNode.GetColumn()},
			Statements: []ast.ASTNode{elseNode},
		}
	}

	return &ast.IfStmtNode{
		BaseNode:   ast.BaseNode{Type: ast.NodeIfStmt, Line: condition.GetLine(), Column: condition.GetColumn()},
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

func (p *Parser) ParseUnlessStatement() (ast.ASTNode, error) {
	// unless condition { } is sugar for if !(condition) { }
	// Optional parentheses around condition
	var condition ast.ASTNode
	var err error

	if p.Match(TokenLParen) {
		// Traditional: unless (condition) { ... }
		condition, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
		p.Consume(TokenRParen, "Expected ')' after unless condition")
	} else {
		// Modern: unless condition { ... }
		condition, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}

	// Wrap condition in logical NOT
	negatedCondition := &ast.UnaryExprNode{
		BaseNode: ast.BaseNode{Type: ast.NodeUnaryExpr, Line: condition.GetLine(), Column: condition.GetColumn()},
		Operator: "!",
		Operand:  condition,
	}

	// Body can be a block { } or a single statement
	var thenBranch *ast.BlockNode
	if p.Check(TokenLBrace) {
		body, err := p.ParseBlock()
		if err != nil {
			return nil, err
		}
		thenBranch = body.(*ast.BlockNode)
	} else {
		// Single statement body
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		thenBranch = &ast.BlockNode{
			BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: stmt.GetLine(), Column: stmt.GetColumn()},
			Statements: []ast.ASTNode{stmt},
		}
	}

	// unless can have else (but not elif, since that would be confusing)
	var elseBranch *ast.BlockNode
	if p.Match(TokenElse) {
		if p.Check(TokenLBrace) {
			elseBlock, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			elseBranch = elseBlock.(*ast.BlockNode)
		} else {
			// Single statement else
			elseStmt, err := p.ParseStatement()
			if err != nil {
				return nil, err
			}
			elseBranch = &ast.BlockNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: elseStmt.GetLine(), Column: elseStmt.GetColumn()},
				Statements: []ast.ASTNode{elseStmt},
			}
		}
	}

	return &ast.IfStmtNode{
		BaseNode:   ast.BaseNode{Type: ast.NodeIfStmt, Line: condition.GetLine(), Column: condition.GetColumn()},
		Condition:  negatedCondition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

func (p *Parser) ParseIfLetStatement() (ast.ASTNode, error) {
	// if let value = optional { ... }
	// Parse pattern: identifier = expression
	line, col := p.Peek().Line, p.Peek().Column

	// Get the variable name
	varNameTok := p.Consume(TokenIdentifier, "Expected variable name after 'let'")
	varName := varNameTok.Value

	p.Consume(TokenAssign, "Expected '=' after variable name in if let")

	// Parse the expression being matched
	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	// Create a condition that checks if the expression is not null/none
	// For now, we'll create a simple binary expression: expr != null
	nullLiteral := &ast.LiteralNode{
		BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: line, Column: col},
		Value:    nil,
		Type:     "null",
	}

	condition := &ast.BinaryExprNode{
		BaseNode: ast.BaseNode{Type: ast.NodeBinaryExpr, Line: line, Column: col},
		Left:     expr,
		Operator: "!=",
		Right:    nullLiteral,
	}

	// Body can be a block { } or a single statement
	var thenBranch *ast.BlockNode
	if p.Check(TokenLBrace) {
		body, err := p.ParseBlock()
		if err != nil {
			return nil, err
		}
		thenBranch = body.(*ast.BlockNode)
	} else {
		// Single statement body
		stmt, err := p.ParseStatement()
		if err != nil {
			return nil, err
		}
		thenBranch = &ast.BlockNode{
			BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: stmt.GetLine(), Column: stmt.GetColumn()},
			Statements: []ast.ASTNode{stmt},
		}
	}

	// Prepend a variable declaration to the then branch
	// var varName = expr (casted/unwrapped)
	varDecl := &ast.VariableDeclNode{
		BaseNode:    ast.BaseNode{Type: ast.NodeVarDecl, Line: line, Column: col},
		Name:        varName,
		Initializer: expr,
	}

	// Insert at beginning of then branch
	thenBranch.Statements = append([]ast.ASTNode{varDecl}, thenBranch.Statements...)

	// else branch is optional
	var elseBranch *ast.BlockNode
	if p.Match(TokenElse) {
		if p.Check(TokenLBrace) {
			elseBlock, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			elseBranch = elseBlock.(*ast.BlockNode)
		} else {
			elseStmt, err := p.ParseStatement()
			if err != nil {
				return nil, err
			}
			elseBranch = &ast.BlockNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: elseStmt.GetLine(), Column: elseStmt.GetColumn()},
				Statements: []ast.ASTNode{elseStmt},
			}
		}
	}

	return &ast.IfStmtNode{
		BaseNode:   ast.BaseNode{Type: ast.NodeIfStmt, Line: line, Column: col},
		Condition:  condition,
		ThenBranch: thenBranch,
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

func (p *Parser) ParseLoopStatement() (ast.ASTNode, error) {
	// loop { } is sugar for while (true) { }
	body, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}

	// Create a literal "true" as the condition
	trueLiteral := &ast.LiteralNode{
		BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: p.Previous().Line, Column: p.Previous().Column},
		Value:    true,
		Type:     "bool",
	}

	return &ast.WhileStmtNode{
		BaseNode:  ast.BaseNode{Type: ast.NodeWhileStmt, Line: p.Previous().Line, Column: p.Previous().Column},
		Condition: trueLiteral,
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

// ParseSelectStatement parses BASIC-style SELECT CASE statements
// Syntax: SELECT CASE expression
//
//	  CASE value1, value2
//	    statements
//	  CASE start TO end
//	    statements
//	  CASE ELSE
//	    statements
//	END SELECT
func (p *Parser) ParseSelectStatement() (ast.ASTNode, error) {
	p.Consume(TokenCase, "Expected 'CASE' after 'SELECT'")
	value, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	var cases []*ast.SwitchCaseNode

	for !p.Check(TokenEnd) && !p.IsAtEnd() {
		if !p.Match(TokenCase) {
			// Check for CASE ELSE
			if p.Check(TokenElse) {
				p.Advance()
				body, err := p.ParseBlock()
				if err != nil {
					return nil, err
				}
				cases = append(cases, &ast.SwitchCaseNode{
					BaseNode: ast.BaseNode{Type: ast.NodeSwitchCase, Line: p.Previous().Line, Column: p.Previous().Column},
					Constant: nil, // default case
					Body:     body.(*ast.BlockNode),
				})
				break
			}
			return nil, fmt.Errorf("expected 'CASE' in select statement at line %d", p.Peek().Line)
		}

		// Check for CASE ELSE (default case)
		if p.Match(TokenElse) {
			// Parse body until next CASE or END SELECT
			var statements []ast.ASTNode
			for !p.Check(TokenCase) && !p.Check(TokenEnd) && !p.IsAtEnd() {
				stmt, err := p.ParseStatement()
				if err != nil {
					return nil, err
				}
				statements = append(statements, stmt)
			}
			body := &ast.BlockNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: p.Previous().Line, Column: p.Previous().Column},
				Statements: statements,
			}
			cases = append(cases, &ast.SwitchCaseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeSwitchCase, Line: p.Previous().Line, Column: p.Previous().Column},
				Constant: nil, // default case
				Body:     body,
			})
			continue
		}

		// Parse case values (can be multiple: CASE 1, 2, 3)
		// Or ranges: CASE 1 TO 10
		var constants []ast.ASTNode

		for {
			// Parse a simple value (not full expression to avoid consuming return/etc)
			var startConst ast.ASTNode
			if p.Check(TokenNumber) {
				tok := p.Advance()
				startConst = &ast.NumberLiteralNode{Value: tok.Value}
			} else if p.Check(TokenString) {
				tok := p.Advance()
				startConst = &ast.StringLiteralNode{Value: tok.Value}
			} else if p.Check(TokenIdentifier) {
				tok := p.Advance()
				startConst = &ast.IdentifierNode{Name: tok.Value}
			} else if p.Check(TokenTrue) || p.Check(TokenFalse) {
				tok := p.Advance()
				startConst = &ast.BoolLiteralNode{Value: tok.Type == TokenTrue}
			} else {
				return nil, fmt.Errorf("expected case value at line %d", p.Peek().Line)
			}
			constants = append(constants, startConst)

			// Check for TO (range)
			if p.Match(TokenTo) {
				// Parse end value
				var endConst ast.ASTNode
				if p.Check(TokenNumber) {
					tok := p.Advance()
					endConst = &ast.NumberLiteralNode{Value: tok.Value}
				} else if p.Check(TokenString) {
					tok := p.Advance()
					endConst = &ast.StringLiteralNode{Value: tok.Value}
				} else if p.Check(TokenIdentifier) {
					tok := p.Advance()
					endConst = &ast.IdentifierNode{Name: tok.Value}
				} else {
					return nil, fmt.Errorf("expected end value for range at line %d", p.Peek().Line)
				}
				// Create a range expression node
				constants = append(constants, &ast.BinaryExprNode{
					BaseNode: ast.BaseNode{Type: ast.NodeBinaryExpr, Line: startConst.GetLine(), Column: startConst.GetColumn()},
					Left:     startConst,
					Operator: "..",
					Right:    endConst,
				})
			}

			// Check for comma (multiple values)
			if !p.Match(TokenComma) {
				break
			}
		}

		// Parse body until next CASE or END SELECT
		var statements []ast.ASTNode
		for !p.Check(TokenCase) && !p.Check(TokenEnd) && !p.Check(TokenElse) && !p.IsAtEnd() {
			stmt, err := p.ParseStatement()
			if err != nil {
				return nil, err
			}
			statements = append(statements, stmt)
		}

		body := &ast.BlockNode{
			BaseNode:   ast.BaseNode{Type: ast.NodeBlock, Line: constants[0].GetLine(), Column: constants[0].GetColumn()},
			Statements: statements,
		}

		// Add each constant as a separate case (or range as single case)
		for _, c := range constants {
			cases = append(cases, &ast.SwitchCaseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeSwitchCase, Line: c.GetLine(), Column: c.GetColumn()},
				Constant: c,
				Body:     body,
			})
		}
	}

	// Consume END SELECT
	p.Consume(TokenEnd, "Expected 'END' to close select statement")
	p.Consume(TokenSelect, "Expected 'SELECT' after 'END'")

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
	// Optional parentheses - allow match(x) or match x
	var value ast.ASTNode
	var err error
	if p.Match(TokenLParen) {
		value, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
		p.Consume(TokenRParen, "Expected ')' after match")
	} else {
		// No parens - parse expression directly
		value, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}
	p.Consume(TokenLBrace, "Expected '{' after match")

	var cases []*ast.CaseClauseNode
	for !p.Check(TokenRBrace) && !p.IsAtEnd() {
		// Handle _ wildcard case
		if p.Match(TokenUnderscore) {
			p.Consume(TokenFatArrow, "Expected '=>' after _")
			body, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			cases = append(cases, &ast.CaseClauseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeCaseClause, Line: p.Previous().Line, Column: p.Previous().Column},
				TypeName: "_",
				VarName:  "",
				Body:     body.(*ast.BlockNode),
			})
			continue
		}
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
		// Support modern syntax: type var => body OR literal => body (without 'case' keyword)
		if p.Check(TokenIdentifier) || p.IsTypeToken(p.Peek().Type) || p.Check(TokenNumber) || p.Check(TokenString) {
			// Check if this is a modern-style case (without 'case' keyword)
			// Peek ahead for =>
			startPos := p.position
			var typeName string
			var varName string
			var lit ast.ASTNode

			if p.IsTypeToken(p.Peek().Type) {
				typeName = p.Advance().Value
				if p.Check(TokenIdentifier) {
					varName = p.Advance().Value
				}
			} else if p.Check(TokenNumber) || p.Check(TokenString) {
				lit, _ = p.ParseExpression()
			} else if p.Check(TokenIdentifier) {
				// Could be enum value or type name
				typeName = p.Advance().Value
				if p.Check(TokenIdentifier) {
					varName = p.Advance().Value
				}
			}

			// Check for => (modern syntax) or : (traditional syntax)
			if p.Match(TokenFatArrow) {
				// Modern syntax: type var => body
				body, err := p.ParseBlock()
				if err != nil {
					return nil, err
				}
				cases = append(cases, &ast.CaseClauseNode{
					BaseNode: ast.BaseNode{Type: ast.NodeCaseClause, Line: p.Previous().Line, Column: p.Previous().Column},
					TypeName: typeName,
					VarName:  varName,
					Literal:  lit,
					Body:     body.(*ast.BlockNode),
				})
				continue
			} else if p.Check(TokenColon) {
				// Traditional syntax with colon
				p.Advance()
				body, err := p.ParseBlock()
				if err != nil {
					return nil, err
				}
				cases = append(cases, &ast.CaseClauseNode{
					BaseNode: ast.BaseNode{Type: ast.NodeCaseClause, Line: p.Previous().Line, Column: p.Previous().Column},
					TypeName: typeName,
					VarName:  varName,
					Literal:  lit,
					Body:     body.(*ast.BlockNode),
				})
				continue
			}
			// Not a match case, restore position and fall through to case keyword check
			p.position = startPos
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

// ParseMatchExpression parses match as an expression that returns a value
func (p *Parser) ParseMatchExpression() (ast.ASTNode, error) {
	// Optional parentheses - allow match(x) or match x
	var value ast.ASTNode
	var err error
	if p.Match(TokenLParen) {
		value, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
		p.Consume(TokenRParen, "Expected ')' after match")
	} else {
		// No parens - parse expression directly
		value, err = p.ParseExpression()
		if err != nil {
			return nil, err
		}
	}
	p.Consume(TokenLBrace, "Expected '{' after match")

	var cases []*ast.CaseClauseNode
	for !p.Check(TokenRBrace) && !p.IsAtEnd() {
		// Handle _ wildcard case
		if p.Match(TokenUnderscore) {
			p.Consume(TokenFatArrow, "Expected '=>' after _")
			// For expression, parse single expression instead of block
			var bodyExpr ast.ASTNode
			if p.Check(TokenLBrace) {
				blk, err := p.ParseBlock()
				if err != nil {
					return nil, err
				}
				bodyExpr = blk
			} else {
				bodyExpr, err = p.ParseExpression()
				if err != nil {
					return nil, err
				}
			}
			cases = append(cases, &ast.CaseClauseNode{
				BaseNode: ast.BaseNode{Type: ast.NodeCaseClause, Line: p.Previous().Line, Column: p.Previous().Column},
				TypeName: "_",
				ExprBody: bodyExpr,
			})
			continue
		}

		// Parse the case value (literal or identifier)
		var lit ast.ASTNode
		var typeName string
		var varName string

		if p.Check(TokenNumber) || p.Check(TokenString) {
			lit, _ = p.ParsePrimary()
		} else if p.Check(TokenIdentifier) {
			// Could be enum value or type name
			typeName = p.Advance().Value
			if p.Check(TokenIdentifier) {
				varName = p.Advance().Value
			}
			// Check if this was actually a literal (enum value) - look ahead for =>
			if p.Check(TokenFatArrow) {
				// It was a literal enum value, use typeName as the literal
				lit = &ast.IdentifierNode{
					BaseNode: ast.BaseNode{Type: ast.NodeIdentifier, Line: p.Previous().Line, Column: p.Previous().Column},
					Name:     typeName,
				}
				typeName = ""
			}
		} else if p.IsTypeToken(p.Peek().Type) {
			typeName = p.Advance().Value
			if p.Check(TokenIdentifier) {
				varName = p.Advance().Value
			}
		}

		// Expect => followed by expression
		if !p.Match(TokenFatArrow) {
			return nil, fmt.Errorf("expected '=>' after case value, got %s", p.Peek().Value)
		}

		// Parse expression for this case
		var bodyExpr ast.ASTNode
		if p.Check(TokenLBrace) {
			blk, err := p.ParseBlock()
			if err != nil {
				return nil, err
			}
			bodyExpr = blk
		} else {
			bodyExpr, err = p.ParseExpression()
			if err != nil {
				return nil, err
			}
		}
		cases = append(cases, &ast.CaseClauseNode{
			BaseNode: ast.BaseNode{Type: ast.NodeCaseClause, Line: p.Previous().Line, Column: p.Previous().Column},
			TypeName: typeName,
			VarName:  varName,
			Literal:  lit,
			ExprBody: bodyExpr,
		})
		// Optional comma separator
		p.Match(TokenComma)
	}
	p.Consume(TokenRBrace, "Expected '}' after match")
	// Return as a ternary-like expression node - use MatchStmtNode for now
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

func (p *Parser) ParseTryStatement() (ast.ASTNode, error) {
	line, col := p.Peek().Line, p.Peek().Column

	// Parse try block
	tryBlock, err := p.ParseBlock()
	if err != nil {
		return nil, err
	}

	var catchBlocks []*ast.CatchClauseNode
	var finally ast.ASTNode

	// Parse catch clauses
	for p.Match(TokenCatch) {
		catchLine, catchCol := p.Previous().Line, p.Previous().Column

		var exceptionType string
		var exceptionVar string

		// Check for catch with type and variable: catch (Type var)
		if p.Match(TokenLParen) {
			// Parse exception type (optional)
			if p.Check(TokenIdentifier) {
				exceptionType = p.Advance().Value
			}
			// Parse exception variable name
			if p.Check(TokenIdentifier) {
				exceptionVar = p.Advance().Value
			}
			p.Consume(TokenRParen, "Expected ')' after catch clause")
		}

		// Parse catch block
		catchBody, err := p.ParseBlock()
		if err != nil {
			return nil, err
		}

		catchBlocks = append(catchBlocks, &ast.CatchClauseNode{
			BaseNode:      ast.BaseNode{Type: ast.NodeCatchClause, Line: catchLine, Column: catchCol},
			ExceptionType: exceptionType,
			ExceptionVar:  exceptionVar,
			Body:          catchBody,
		})
	}

	// Parse finally block (optional)
	if p.Match(TokenIdentifier) && p.Previous().Value == "finally" {
		var err error
		finally, err = p.ParseBlock()
		if err != nil {
			return nil, err
		}
	}

	return &ast.TryStmtNode{
		BaseNode:    ast.BaseNode{Type: ast.NodeTryStmt, Line: line, Column: col},
		TryBlock:    tryBlock,
		CatchBlocks: catchBlocks,
		Finally:     finally,
	}, nil
}

func (p *Parser) ParseThrowStatement() (ast.ASTNode, error) {
	line, col := p.Peek().Line, p.Peek().Column

	// Parse the expression to throw
	expr, err := p.ParseExpression()
	if err != nil {
		return nil, err
	}

	p.Consume(TokenSemicolon, "Expected ';' after throw expression")

	return &ast.ThrowStmtNode{
		BaseNode:   ast.BaseNode{Type: ast.NodeThrowStmt, Line: line, Column: col},
		Expression: expr,
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

	// Handle compound assignment operators: +=, -=, *=, /=
	if p.Match(TokenPlusAssign) {
		value, err := p.ParseAssignment()
		if err != nil {
			return nil, err
		}
		return &ast.CompoundAssignmentNode{
			BaseNode: ast.BaseNode{Type: ast.NodeCompoundAssignment, Line: expr.GetLine(), Column: expr.GetColumn()},
			Target:   expr,
			Value:    value,
			Operator: "+",
		}, nil
	}
	if p.Match(TokenMinusAssign) {
		value, err := p.ParseAssignment()
		if err != nil {
			return nil, err
		}
		return &ast.CompoundAssignmentNode{
			BaseNode: ast.BaseNode{Type: ast.NodeCompoundAssignment, Line: expr.GetLine(), Column: expr.GetColumn()},
			Target:   expr,
			Value:    value,
			Operator: "-",
		}, nil
	}
	if p.Match(TokenMultiplyAssign) {
		value, err := p.ParseAssignment()
		if err != nil {
			return nil, err
		}
		return &ast.CompoundAssignmentNode{
			BaseNode: ast.BaseNode{Type: ast.NodeCompoundAssignment, Line: expr.GetLine(), Column: expr.GetColumn()},
			Target:   expr,
			Value:    value,
			Operator: "*",
		}, nil
	}
	if p.Match(TokenDivideAssign) {
		value, err := p.ParseAssignment()
		if err != nil {
			return nil, err
		}
		return &ast.CompoundAssignmentNode{
			BaseNode: ast.BaseNode{Type: ast.NodeCompoundAssignment, Line: expr.GetLine(), Column: expr.GetColumn()},
			Target:   expr,
			Value:    value,
			Operator: "/",
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

	// Handle null coalescing: a ?? b
	for p.Match(TokenNullCoalesce) {
		operator := p.Previous()
		right, err := p.ParseLogicalAnd()
		if err != nil {
			return nil, err
		}

		expr = &ast.BinaryExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeBinaryExpr, Line: operator.Line, Column: operator.Column},
			Left:     expr,
			Operator: "??",
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

	// Check for range operators: 0..10 or 0..<n
	if p.Match(TokenRange) {
		right, err := p.ParseCall()
		if err != nil {
			return nil, err
		}
		expr = &ast.RangeNode{
			BaseNode:  ast.BaseNode{Type: ast.NodeRange, Line: expr.GetLine(), Column: expr.GetColumn()},
			Start:     expr,
			End:       right,
			Exclusive: false,
		}
	} else if p.Match(TokenRangeExclusive) {
		right, err := p.ParseCall()
		if err != nil {
			return nil, err
		}
		expr = &ast.RangeNode{
			BaseNode:  ast.BaseNode{Type: ast.NodeRange, Line: expr.GetLine(), Column: expr.GetColumn()},
			Start:     expr,
			End:       right,
			Exclusive: true,
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
	if p.Match(TokenNot, TokenMinus) {
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
	// Handle prefix ++ and -- separately
	if p.Match(TokenIncrement) {
		target, err := p.ParseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.IncrementNode{
			BaseNode:    ast.BaseNode{Type: ast.NodeIncrement, Line: p.Previous().Line, Column: p.Previous().Column},
			Target:      target,
			IsIncrement: true,
			IsPrefix:    true,
		}, nil
	}
	if p.Match(TokenDecrement) {
		target, err := p.ParseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.IncrementNode{
			BaseNode:    ast.BaseNode{Type: ast.NodeIncrement, Line: p.Previous().Line, Column: p.Previous().Column},
			Target:      target,
			IsIncrement: false,
			IsPrefix:    true,
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
		} else if p.Match(TokenOptionalChain) {
			// Optional chaining: obj?.member
			member := p.Consume(TokenIdentifier, "Expected property name after '?.'")
			expr = &ast.MemberAccessNode{
				BaseNode: ast.BaseNode{Type: ast.NodeMemberAccess, Line: member.Line, Column: member.Column},
				Object:   expr,
				Member:   member.Value,
				Optional: true,
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
			expr = &ast.IncrementNode{
				BaseNode:    ast.BaseNode{Type: ast.NodeIncrement, Line: expr.GetLine(), Column: expr.GetColumn()},
				Target:      expr,
				IsIncrement: true,
				IsPrefix:    false,
			}
		} else if p.Match(TokenDecrement) {
			expr = &ast.IncrementNode{
				BaseNode:    ast.BaseNode{Type: ast.NodeIncrement, Line: expr.GetLine(), Column: expr.GetColumn()},
				Target:      expr,
				IsIncrement: false,
				IsPrefix:    false,
			}
		} else if p.Match(TokenQuestion) {
			// Postfix ? - optional check (returns bool)
			expr = &ast.UnaryExprNode{
				BaseNode:  ast.BaseNode{Type: ast.NodeUnaryExpr, Line: expr.GetLine(), Column: expr.GetColumn()},
				Operator:  "?",
				Operand:   expr,
				IsPostfix: true,
			}
		} else if p.Match(TokenNot) {
			// Postfix ! - force unwrap
			expr = &ast.UnaryExprNode{
				BaseNode:  ast.BaseNode{Type: ast.NodeUnaryExpr, Line: expr.GetLine(), Column: expr.GetColumn()},
				Operator:  "!",
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

	// Handle match expression: match value { cases }
	if p.Match(TokenMatch) {
		return p.ParseMatchExpression()
	}

	// Handle channel<T>(size) constructor
	if p.Match(TokenChannel) {
		line, col := p.Previous().Line, p.Previous().Column
		// Parse type parameter: <T>
		elemType := "any" // default type
		if p.Match(TokenLess) {
			// Accept either identifier or type keyword (int, float, etc.)
			if p.IsTypeToken(p.Peek().Type) {
				tok := p.Advance()
				elemType = tok.Value
			} else if p.Check(TokenIdentifier) {
				tok := p.Advance()
				elemType = tok.Value
			} else {
				return nil, fmt.Errorf("expected type in channel<T> at line %d", line)
			}
			p.Consume(TokenGreater, "Expected '>' after channel type")
		}
		// Check if this is a constructor call: channel<T>(size) or just type: channel<T>
		if !p.Check(TokenLParen) {
			// Just a type reference, return as identifier
			return &ast.IdentifierNode{
				BaseNode: ast.BaseNode{Type: ast.NodeIdentifier, Line: line, Column: col},
				Name:     "channel<" + elemType + ">",
			}, nil
		}
		// Parse constructor arguments: (size)
		p.Consume(TokenLParen, "Expected '(' after channel")
		var sizeArg ast.ASTNode
		if !p.Check(TokenRParen) {
			arg, err := p.ParseExpression()
			if err != nil {
				return nil, err
			}
			sizeArg = arg
		}
		p.Consume(TokenRParen, "Expected ')' after channel arguments")
		// Use channel_of(T, N) macro which expands to channel_create(sizeof(T), N)
		return &ast.CallExprNode{
			BaseNode: ast.BaseNode{Type: ast.NodeCallExpr, Line: line, Column: col},
			Function: &ast.IdentifierNode{
				BaseNode: ast.BaseNode{Type: ast.NodeIdentifier, Line: line, Column: col},
				Name:     "channel_of",
			},
			Args: []ast.ASTNode{
				&ast.LiteralNode{
					BaseNode: ast.BaseNode{Type: ast.NodeLiteral, Line: line, Column: col},
					Value:    elemType,
					Type:     "type",
				},
				sizeArg,
			},
		}, nil
	}

	// Allow type keywords to be used as identifiers in expressions (e.g., var result = ...)
	if p.IsTypeToken(p.Peek().Type) && !p.Check(TokenVoid) && !p.Check(TokenVar) && !p.Check(TokenLet) && !p.Check(TokenFn) {
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
		// Check for arrow function: () => expr or (params) => expr or (type param) => expr
		savedPos := p.position - 1 // position of the '('
		isArrowFunc := false
		paramNames := []string{} // Declared at outer scope for use later
		paramTypes := []string{} // For typed parameters

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
		} else if p.Check(TokenIdentifier) || p.IsTypeToken(p.Peek().Type) {
			// Check for (x, y) => pattern or (int x, int y) => pattern
			for p.Check(TokenIdentifier) || p.IsTypeToken(p.Peek().Type) {
				// Check if this is a typed parameter (type name) or untyped (just name)
				if p.IsTypeToken(p.Peek().Type) {
					// Typed parameter: int x
					paramType := p.Advance().Value
					if p.Check(TokenIdentifier) {
						paramNames = append(paramNames, p.Peek().Value)
						paramTypes = append(paramTypes, paramType)
						p.Advance()
					} else {
						// Not a valid pattern, restore
						p.position = savedPos
						break
					}
				} else if p.Check(TokenIdentifier) {
					// Untyped parameter: x
					paramNames = append(paramNames, p.Peek().Value)
					paramTypes = append(paramTypes, "var")
					p.Advance()
				}
				if p.Check(TokenComma) {
					p.Advance()
				}
			}
			if p.Check(TokenRParen) {
				p.Advance() // consume ')'
				if p.Check(TokenFatArrow) {
					isArrowFunc = true
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

			// Build parameters from the collected names and types
			params := make([]*ast.ParameterNode, len(paramNames))
			for i, name := range paramNames {
				paramType := "var"
				if i < len(paramTypes) {
					paramType = paramTypes[i]
				}
				params[i] = &ast.ParameterNode{
					BaseNode: ast.BaseNode{Type: ast.NodeParameter, Line: body.GetLine(), Column: body.GetColumn()},
					Name:     name,
					Type:     paramType,
				}
			}

			return &ast.LambdaNode{
				BaseNode:   ast.BaseNode{Type: ast.NodeLambda, Line: body.GetLine(), Column: body.GetColumn()},
				Captures:   []string{},
				Parameters: params,
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

	// Array literal with curly braces: {1, 2, 3} — must come before struct/dict detection
	if p.Check(TokenLBrace) {
		// Look ahead to determine if this is an array literal
		// Array literal: {number, ...} or {expr, ...} where first element is not string:"key":
		next := p.PeekAhead(1)
		if next.Type == TokenNumber || next.Type == TokenMinus || next.Type == TokenTrue || next.Type == TokenFalse || next.Type == TokenLBrace || next.Type == TokenString {
			// Check if it's actually a dict (string followed by colon)
			if next.Type == TokenString && p.PeekAhead(2).Type == TokenColon {
				// This is a dict literal, fall through
			} else {
				// This is likely an array literal
				return p.ParseArrayLiteralBrace()
			}
		}
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

// ParseArrayLiteralBrace parses array literals with curly braces: {1, 2, 3}
func (p *Parser) ParseArrayLiteralBrace() (ast.ASTNode, error) {
	p.Consume(TokenLBrace, "Expected '{' for array literal")
	var elements []ast.ASTNode
	for !p.Check(TokenRBrace) && !p.IsAtEnd() {
		e, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, e)
		if !p.Check(TokenRBrace) {
			p.Consume(TokenComma, "Expected ',' or '}' in array literal")
		}
	}
	p.Consume(TokenRBrace, "Expected '}' after array literal")
	node := &ast.ArrayLiteralNode{
		BaseNode: ast.BaseNode{Type: ast.NodeArrayLiteral, Line: p.Previous().Line, Column: p.Previous().Column},
		Elements: elements,
	}
	p.AnnotateArrayLiteral(node)
	return node, nil
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
	typeTokens := []TokenType{TokenVoid, TokenInt, TokenFloat, TokenDouble, TokenCharType, TokenBool, TokenStringType, TokenVec2, TokenVec3, TokenVar, TokenLet, TokenFn, TokenAny, TokenConst, TokenArray, TokenDict, TokenResult, TokenEvent, TokenGuiWindow, TokenGuiWidget, TokenGuiContainer, TokenGuiEvent, TokenChannel}
	return p.Match(typeTokens...)
}

func (p *Parser) ConsumeTypeOrVar() Token {
	typeTokens := []TokenType{TokenVoid, TokenInt, TokenFloat, TokenDouble, TokenCharType, TokenBool, TokenStringType, TokenVec2, TokenVec3, TokenVar, TokenLet, TokenFn, TokenAny, TokenConst, TokenArray, TokenDict, TokenResult, TokenEvent, TokenGuiWindow, TokenGuiWidget, TokenGuiContainer, TokenGuiEvent, TokenChannel}
	for _, tokenType := range typeTokens {
		if p.Check(tokenType) {
			tok := p.Advance()
			// Check for generic type parameters: vector<T> or optional<T>
			if p.Check(TokenLess) {
				p.Advance() // consume '<'
				// Parse type parameter
				typeParam := p.ConsumeType("Expected type parameter after '<'")
				p.Consume(TokenGreater, "Expected '>' after type parameter")
				tok.Value = tok.Value + "<" + typeParam.Value + ">"
			}
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
				nextTok := p.ConsumeType("Expected type after '|' in union")
				tok.Value = tok.Value + nextTok.Value
			}
			return tok
		}
	}
	if p.Check(TokenIdentifier) {
		tok := p.Advance()
		// Check for generic type parameters on identifier types: MyStruct<T>
		if p.Check(TokenLess) {
			p.Advance() // consume '<'
			typeParam := p.ConsumeType("Expected type parameter after '<'")
			p.Consume(TokenGreater, "Expected '>' after type parameter")
			tok.Value = tok.Value + "<" + typeParam.Value + ">"
		}
		// Check for union type with identifier types
		for p.Check(TokenPipe) {
			p.Advance() // consume '|'
			tok.Value = tok.Value + " | "
			nextTok := p.ConsumeType("Expected type after '|' in union")
			tok.Value = tok.Value + nextTok.Value
		}
		return tok
	}
	panic(fmt.Sprintf("Expected type or 'var'. Got %s instead", p.Peek().Value))
}

func (p *Parser) ConsumeType(errorMessage string) Token {
	typeTokens := []TokenType{TokenVoid, TokenInt, TokenFloat, TokenDouble, TokenCharType, TokenBool, TokenStringType, TokenVec2, TokenVec3, TokenVar, TokenLet, TokenFn, TokenAny, TokenConst, TokenArray, TokenDict, TokenResult, TokenEvent, TokenGuiWindow, TokenGuiWidget, TokenGuiContainer, TokenGuiEvent, TokenChannel}
	for _, tokenType := range typeTokens {
		if p.Check(tokenType) {
			tok := p.Advance()
			// Check for generic type parameters: vector<T> or optional<T>
			if p.Check(TokenLess) {
				p.Advance() // consume '<'
				typeParam := p.ConsumeType("Expected type parameter after '<'")
				p.Consume(TokenGreater, "Expected '>' after type parameter")
				tok.Value = tok.Value + "<" + typeParam.Value + ">"
			}
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
	// Check for generic type parameters on identifier types: MyStruct<T>
	if p.Check(TokenLess) {
		p.Advance() // consume '<'
		typeParam := p.ConsumeType("Expected type parameter after '<'")
		p.Consume(TokenGreater, "Expected '>' after type parameter")
		tok.Value = tok.Value + "<" + typeParam.Value + ">"
	}
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
	typeTokens := []TokenType{TokenVoid, TokenInt, TokenFloat, TokenDouble, TokenCharType, TokenBool, TokenStringType, TokenVec2, TokenVec3, TokenVar, TokenLet, TokenFn, TokenAny, TokenConst, TokenArray, TokenDict, TokenResult, TokenEvent, TokenGuiWindow, TokenGuiWidget, TokenGuiContainer, TokenGuiEvent, TokenChannel}
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
