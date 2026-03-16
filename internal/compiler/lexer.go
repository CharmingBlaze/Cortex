package compiler

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenNumber
	TokenString
	TokenChar

	// Keywords
	TokenVoid
	TokenInt
	TokenFloat
	TokenDouble
	TokenCharType
	TokenBool
	TokenStringType
	TokenVec2
	TokenVec3
	TokenStruct
	TokenEnum
	TokenAsync
	TokenAwait
	TokenActor
	TokenChannel
	TokenSpawn
	TokenIf
	TokenElif
	TokenElse
	TokenUnless
	TokenFor
	TokenWhile
	TokenLoop
	TokenDo
	TokenReturn
	TokenDefer
	TokenMatch
	TokenCase
	TokenDefault
	TokenRepeat
	TokenBreak
	TokenContinue
	TokenSwitch
	TokenSelect
	TokenIn
	TokenTrue
	TokenFalse
	TokenNull
	TokenVar
	TokenLet
	TokenFn
	TokenAny
	TokenConst
	TokenArray
	TokenDict
	TokenResult
	TokenEvent
	TokenGuiWindow
	TokenGuiWidget
	TokenGuiContainer
	TokenGuiEvent
	TokenTest
	TokenCoroutine
	TokenYield
	TokenTypeKeyword
	TokenEndType
	TokenAs
	TokenEnd
	TokenPublic
	TokenPrivate

	// Preprocessor
	TokenInclude
	TokenUse
	TokenDefine
	TokenUndef
	TokenIfDef
	TokenIfNDef
	TokenEndIf
	TokenPragma
	TokenLibrary
	TokenConfig
	TokenWrapper
	TokenRawC
	TokenExtern
	TokenPackage
	TokenImport
	TokenModule
	TokenCleanup

	// Operators
	TokenAssign
	TokenPlusAssign     // +=
	TokenMinusAssign    // -=
	TokenMultiplyAssign // *=
	TokenDivideAssign   // /=
	TokenPlus
	TokenMinus
	TokenMultiply
	TokenDivide
	TokenModulo
	TokenEqual
	TokenNotEqual
	TokenLess
	TokenLessEqual
	TokenGreater
	TokenGreaterEqual
	TokenAnd
	TokenOr
	TokenPipe
	TokenNot
	TokenIncrement
	TokenDecrement
	TokenAt
	TokenUnderscore

	// Delimiters
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenSemicolon
	TokenComma
	TokenDot
	TokenColon
	TokenQuestion
	TokenArrow
	TokenFatArrow
	TokenNullCoalesce   // ?? - null coalescing
	TokenOptionalChain  // ?. - optional chaining
	TokenRange          // .. - inclusive range
	TokenRangeExclusive // ..< - exclusive range
	TokenTry
	TokenCatch
	TokenThrow

	// Special
	TokenComment
	TokenWhitespace
)

type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %v, Value: %q, Line: %d, Column: %d}", t.Type, t.Value, t.Line, t.Column)
}

type Lexer struct {
	input    string
	position int
	line     int
	column   int
}

func NewLexer() *Lexer {
	return &Lexer{
		line:   1,
		column: 1,
	}
}

func (l *Lexer) Tokenize(input string) ([]Token, error) {
	l.input = input
	l.position = 0
	l.line = 1
	l.column = 1

	var tokens []Token

	for l.position < len(l.input) {
		char := l.CurrentChar()

		if unicode.IsSpace(rune(char)) {
			if char == '\n' {
				l.line++
				l.column = 1
			} else {
				l.column++
			}
			l.position++
			continue
		}

		// Handle preprocessor directives
		if char == '#' {
			directive := l.ReadPreprocessorDirective()
			tokens = append(tokens, directive)
			// Skip to end of line after preprocessor directive
			for l.position < len(l.input) && l.input[l.position] != '\n' {
				l.position++
			}
			continue
		}

		if char == '/' && l.PeekChar() == '/' {
			comment := l.ReadSingleLineComment()
			tokens = append(tokens, Token{Type: TokenComment, Value: comment, Line: l.line, Column: l.column})
			continue
		}

		if char == '/' && l.PeekChar() == '*' {
			comment := l.ReadMultiLineComment()
			tokens = append(tokens, Token{Type: TokenComment, Value: comment, Line: l.line, Column: l.column})
			continue
		}

		if unicode.IsDigit(rune(char)) {
			number := l.ReadNumber()
			tokens = append(tokens, Token{Type: TokenNumber, Value: number, Line: l.line, Column: l.column})
			continue
		}

		if char == '"' {
			str := l.ReadString()
			tokens = append(tokens, Token{Type: TokenString, Value: str, Line: l.line, Column: l.column})
			continue
		}

		if char == '\'' {
			charLit := l.ReadChar()
			tokens = append(tokens, Token{Type: TokenChar, Value: charLit, Line: l.line, Column: l.column})
			continue
		}

		// Handle standalone underscore _ (wildcard in match) before identifier parsing
		if char == '_' {
			// Check if this is a standalone _ or part of an identifier
			nextChar := l.PeekChar()
			if !l.IsIdentifierChar(nextChar) {
				// Standalone _ - treat as wildcard token
				tokens = append(tokens, Token{Type: TokenUnderscore, Value: "_", Line: l.line, Column: l.column})
				l.position++
				l.column++
				continue
			}
			// Otherwise fall through to identifier parsing
		}

		if l.IsIdentifierChar(char) {
			identifier := l.ReadIdentifier()
			tokenType := l.GetKeywordType(identifier)
			tokens = append(tokens, Token{Type: tokenType, Value: identifier, Line: l.line, Column: l.column})
			continue
		}

		if char == '@' {
			tokens = append(tokens, Token{Type: TokenAt, Value: "@", Line: l.line, Column: l.column})
			l.position++
			l.column++
			continue
		}

		tokenType := l.GetOperatorType(char)
		if tokenType != TokenEOF {
			value := string(char)

			// Handle multi-character operators
			if char == '=' && l.PeekChar() == '=' {
				l.position++
				value = "=="
				tokenType = TokenEqual
			} else if char == '!' && l.PeekChar() == '=' {
				l.position++
				value = "!="
				tokenType = TokenNotEqual
			} else if char == '<' && l.PeekChar() == '=' {
				l.position++
				value = "<="
				tokenType = TokenLessEqual
			} else if char == '>' && l.PeekChar() == '=' {
				l.position++
				value = ">="
				tokenType = TokenGreaterEqual
			} else if char == '+' && l.PeekChar() == '+' {
				l.position++
				value = "++"
				tokenType = TokenIncrement
			} else if char == '-' && l.PeekChar() == '-' {
				l.position++
				value = "--"
				tokenType = TokenDecrement
			} else if char == '+' && l.PeekChar() == '=' {
				l.position++
				value = "+="
				tokenType = TokenPlusAssign
			} else if char == '-' && l.PeekChar() == '=' {
				l.position++
				value = "-="
				tokenType = TokenMinusAssign
			} else if char == '*' && l.PeekChar() == '=' {
				l.position++
				value = "*="
				tokenType = TokenMultiplyAssign
			} else if char == '/' && l.PeekChar() == '=' {
				l.position++
				value = "/="
				tokenType = TokenDivideAssign
			} else if char == '-' && l.PeekChar() == '>' {
				l.position++
				value = "->"
				tokenType = TokenArrow
			} else if char == '=' && l.PeekChar() == '>' {
				l.position++
				value = "=>"
				tokenType = TokenFatArrow
			} else if char == '&' && l.PeekChar() == '&' {
				l.position++
				value = "&&"
				tokenType = TokenAnd
			} else if char == '|' && l.PeekChar() == '|' {
				l.position++
				value = "||"
				tokenType = TokenOr
			} else if char == '?' && l.PeekChar() == '?' {
				l.position++
				value = "??"
				tokenType = TokenNullCoalesce
			} else if char == '?' && l.PeekChar() == '.' {
				l.position++
				value = "?."
				tokenType = TokenOptionalChain
			} else if char == '.' && l.PeekChar() == '.' {
				l.position++
				// Check for ..< (exclusive range) vs .. (inclusive range)
				if l.PeekChar() == '<' {
					l.position++
					value = "..<"
					tokenType = TokenRangeExclusive
				} else {
					value = ".."
					tokenType = TokenRange
				}
			} else if char == '|' {
				// Single | for union types
				tokenType = TokenPipe
			}

			tokens = append(tokens, Token{Type: tokenType, Value: value, Line: l.line, Column: l.column})
			l.column++
			l.position++
			continue
		}

		return nil, fmt.Errorf("unexpected character '%c' at line %d, column %d", char, l.line, l.column)
	}

	tokens = append(tokens, Token{Type: TokenEOF, Value: "", Line: l.line, Column: l.column})
	return tokens, nil
}

func (l *Lexer) CurrentChar() byte {
	if l.position >= len(l.input) {
		return 0
	}
	return l.input[l.position]
}

func (l *Lexer) PeekChar() byte {
	if l.position+1 >= len(l.input) {
		return 0
	}
	return l.input[l.position+1]
}

func (l *Lexer) ReadNumber() string {
	start := l.position
	for l.position < len(l.input) && (unicode.IsDigit(rune(l.input[l.position])) || l.input[l.position] == '.') {
		l.position++
	}
	return l.input[start:l.position]
}

func (l *Lexer) ReadEscape() (rune, bool) {
	if l.position >= len(l.input) {
		return 0, false
	}
	c := l.input[l.position]
	l.position++
	switch c {
	case 'n':
		return '\n', true
	case 't':
		return '\t', true
	case 'r':
		return '\r', true
	case '0':
		return 0, true
	case '\\':
		return '\\', true
	case '"':
		return '"', true
	case '\'':
		return '\'', true
	default:
		return rune(c), true
	}
}

func (l *Lexer) ReadString() string {
	l.position++ // skip opening quote
	var value []rune
	for l.position < len(l.input) && l.input[l.position] != '"' {
		if l.input[l.position] == '\\' {
			l.position++
			if r, ok := l.ReadEscape(); ok {
				value = append(value, r)
			}
		} else {
			value = append(value, rune(l.input[l.position]))
			l.position++
		}
	}
	l.position++ // skip closing quote
	return string(value)
}

func (l *Lexer) ReadChar() string {
	l.position++ // skip opening quote
	var value []rune
	for l.position < len(l.input) && l.input[l.position] != '\'' {
		if l.input[l.position] == '\\' {
			l.position++
			if r, ok := l.ReadEscape(); ok {
				value = append(value, r)
			}
		} else {
			value = append(value, rune(l.input[l.position]))
			l.position++
		}
	}
	l.position++ // skip closing quote
	return string(value)
}

func (l *Lexer) ReadIdentifier() string {
	start := l.position
	for l.position < len(l.input) && l.IsIdentifierChar(l.input[l.position]) {
		l.position++
	}
	return l.input[start:l.position]
}

func (l *Lexer) ReadPreprocessorDirective() Token {
	l.position++ // skip #
	start := l.position
	startColumn := l.column

	// Read directive name
	for l.position < len(l.input) && l.IsIdentifierChar(l.input[l.position]) {
		l.position++
		l.column++
	}

	directive := l.input[start:l.position]

	// Skip whitespace
	for l.position < len(l.input) && unicode.IsSpace(rune(l.input[l.position])) {
		if l.input[l.position] == '\n' {
			break
		}
		l.position++
		l.column++
	}

	// Read directive content until end of line
	contentStart := l.position
	for l.position < len(l.input) && l.input[l.position] != '\n' {
		l.position++
	}
	content := strings.TrimSpace(l.input[contentStart:l.position])

	// Determine token type based on directive
	var tokenType TokenType
	switch directive {
	case "include":
		tokenType = TokenInclude
	case "use":
		tokenType = TokenUse
	case "define":
		tokenType = TokenDefine
	case "undef":
		tokenType = TokenUndef
	case "ifdef":
		tokenType = TokenIfDef
	case "ifndef":
		tokenType = TokenIfNDef
	case "endif":
		tokenType = TokenEndIf
	case "pragma":
		tokenType = TokenPragma
	case "library":
		tokenType = TokenLibrary
	case "cortex/internal/config":
		tokenType = TokenConfig
	case "wrapper":
		tokenType = TokenWrapper
	default:
		tokenType = TokenIdentifier
	}

	// Only return token if content is not empty for directives that need content
	if (directive == "include" || directive == "use" || directive == "define") && content == "" {
		// Return a token with empty content - parser will handle it
		return Token{
			Type:   tokenType,
			Value:  directive,
			Line:   l.line,
			Column: startColumn,
		}
	}

	return Token{
		Type:   tokenType,
		Value:  directive + " " + content,
		Line:   l.line,
		Column: startColumn,
	}
}

func (l *Lexer) ReadSingleLineComment() string {
	l.position += 2 // skip //
	start := l.position
	for l.position < len(l.input) && l.input[l.position] != '\n' {
		l.position++
	}
	return l.input[start:l.position]
}

func (l *Lexer) ReadMultiLineComment() string {
	l.position += 2 // skip /*
	start := l.position
	for l.position < len(l.input) && !(l.input[l.position] == '*' && l.PeekChar() == '/') {
		if l.input[l.position] == '\n' {
			l.line++
			l.column = 1
		}
		l.position++
	}
	l.position += 2 // skip */
	return l.input[start:l.position]
}

func (l *Lexer) IsIdentifierChar(char byte) bool {
	return unicode.IsLetter(rune(char)) || unicode.IsDigit(rune(char)) || char == '_'
}

func (l *Lexer) GetKeywordType(identifier string) TokenType {
	// Case-insensitive keyword matching
	lowerIdent := strings.ToLower(identifier)
	keywords := map[string]TokenType{
		"void":      TokenVoid,
		"int":       TokenInt,
		"float":     TokenFloat,
		"double":    TokenDouble,
		"char":      TokenCharType,
		"bool":      TokenBool,
		"string":    TokenStringType,
		"vec2":      TokenVec2,
		"vec3":      TokenVec3,
		"struct":    TokenStruct,
		"enum":      TokenEnum,
		"async":     TokenAsync,
		"await":     TokenAwait,
		"actor":     TokenActor,
		"channel":   TokenChannel,
		"spawn":     TokenSpawn,
		"var":       TokenVar,
		"let":       TokenLet,
		"fn":        TokenFn,
		"any":       TokenAny,
		"const":     TokenConst,
		"extern":    TokenExtern,
		"package":   TokenPackage,
		"import":    TokenImport,
		"module":    TokenModule,
		"cleanup":   TokenCleanup,
		"if":        TokenIf,
		"elif":      TokenElif,
		"else":      TokenElse,
		"unless":    TokenUnless,
		"for":       TokenFor,
		"while":     TokenWhile,
		"loop":      TokenLoop,
		"do":        TokenDo,
		"return":    TokenReturn,
		"defer":     TokenDefer,
		"match":     TokenMatch,
		"case":      TokenCase,
		"default":   TokenDefault,
		"repeat":    TokenRepeat,
		"break":     TokenBreak,
		"continue":  TokenContinue,
		"switch":    TokenSwitch,
		"select":    TokenSelect,
		"in":        TokenIn,
		"true":      TokenTrue,
		"false":     TokenFalse,
		"null":      TokenNull,
		"array":     TokenArray,
		"dict":      TokenDict,
		"result":    TokenResult,
		"event":     TokenEvent,
		"test":      TokenTest,
		"coroutine": TokenCoroutine,
		"yield":     TokenYield,
		"try":       TokenTry,
		"catch":     TokenCatch,
		"throw":     TokenThrow,
		"type":      TokenTypeKeyword,
		"endtype":   TokenEndType,
		"end":       TokenEnd,
		"as":        TokenAs,
		"public":    TokenPublic,
		"private":   TokenPrivate,
	}

	if tokenType, exists := keywords[lowerIdent]; exists {
		return tokenType
	}
	return TokenIdentifier
}

func (l *Lexer) GetOperatorType(char byte) TokenType {
	operators := map[byte]TokenType{
		'=': TokenAssign,
		'+': TokenPlus,
		'-': TokenMinus,
		'*': TokenMultiply,
		'/': TokenDivide,
		'%': TokenModulo,
		'<': TokenLess,
		'>': TokenGreater,
		'&': TokenAnd,
		'|': TokenOr,
		'!': TokenNot,
		'(': TokenLParen,
		')': TokenRParen,
		'{': TokenLBrace,
		'}': TokenRBrace,
		'[': TokenLBracket,
		']': TokenRBracket,
		';': TokenSemicolon,
		',': TokenComma,
		'.': TokenDot,
		':': TokenColon,
		'?': TokenQuestion,
	}

	if tokenType, exists := operators[char]; exists {
		return tokenType
	}
	return TokenEOF
}
