package binder

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CType represents a C type with modifiers
type CType struct {
	Name       string
	IsPointer  bool
	IsConst    bool
	IsArray    bool
	ArraySize  string
	IsUnsigned bool
	IsStruct   bool
}

// CFunction represents a parsed C function
type CFunction struct {
	Name       string
	ReturnType CType
	Params     []CParam
	IsVariadic bool
}

// CParam represents a function parameter
type CParam struct {
	Name string
	Type CType
}

// CStruct represents a parsed C struct
type CStruct struct {
	Name   string
	Fields []CField
}

// CField represents a struct field
type CField struct {
	Name string
	Type CType
}

// CEnum represents a parsed C enum
type CEnum struct {
	Name   string
	Values []CEnumValue
}

// CEnumValue represents an enum value
type CEnumValue struct {
	Name  string
	Value string
}

// Binder converts C headers to Cortex bindings
type Binder struct {
	functions []CFunction
	structs   []CStruct
	enums     []CEnum
	typedefs  map[string]CType
	defines   map[string]string
	libName   string
}

// NewBinder creates a new Binder
func NewBinder(libName string) *Binder {
	return &Binder{
		libName:  libName,
		typedefs: make(map[string]CType),
		defines:  make(map[string]string),
	}
}

// ParseHeader parses a C header file
func (b *Binder) ParseHeader(headerPath string) error {
	data, err := os.ReadFile(headerPath)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	content := string(data)

	// Remove comments
	content = removeComments(content)

	// Parse different constructs
	b.parseTypedefs(content)
	b.parseEnums(content)
	b.parseStructs(content)
	b.parseFunctions(content)
	b.parseDefines(content)

	return nil
}

// removeComments strips C-style comments
func removeComments(content string) string {
	// Remove single-line comments
	re := regexp.MustCompile(`//.*`)
	content = re.ReplaceAllString(content, "")

	// Remove multi-line comments
	re = regexp.MustCompile(`(?s)/\*.*?\*/`)
	content = re.ReplaceAllString(content, "")

	return content
}

// parseTypedefs extracts typedef declarations
func (b *Binder) parseTypedefs(content string) {
	// Match: typedef <type> <name>;
	re := regexp.MustCompile(`typedef\s+([\w\s\*]+?)\s+(\w+)\s*;`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			typeStr := strings.TrimSpace(m[1])
			name := strings.TrimSpace(m[2])
			b.typedefs[name] = parseCType(typeStr)
		}
	}

	// Match: typedef struct { ... } <name>;
	re = regexp.MustCompile(`(?s)typedef\s+struct\s*\{[^}]*\}\s*(\w+)\s*;`)
	matches = re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			name := strings.TrimSpace(m[1])
			b.typedefs[name] = CType{Name: name, IsStruct: true}
		}
	}

	// Match: typedef struct <tag> <name>;
	re = regexp.MustCompile(`typedef\s+struct\s+(\w+)\s+(\w+)\s*;`)
	matches = re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			tag := strings.TrimSpace(m[1])
			name := strings.TrimSpace(m[2])
			b.typedefs[name] = CType{Name: tag, IsStruct: true}
		}
	}
}

// parseEnums extracts enum declarations
func (b *Binder) parseEnums(content string) {
	// Match: typedef enum { ... } <name>;
	re := regexp.MustCompile(`(?s)typedef\s+enum\s*(\w*)\s*\{([^}]+)\}\s*(\w+)\s*;`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 4 {
			enumName := strings.TrimSpace(m[3])
			valuesStr := m[2]

			enum := CEnum{Name: enumName}
			// Parse enum values
			valueRe := regexp.MustCompile(`(\w+)\s*(?:=\s*([^,]+))?`)
			valueMatches := valueRe.FindAllStringSubmatch(valuesStr, -1)
			for _, vm := range valueMatches {
				if len(vm) >= 2 && vm[1] != "" {
					val := ""
					if len(vm) >= 3 {
						val = strings.TrimSpace(vm[2])
					}
					enum.Values = append(enum.Values, CEnumValue{
						Name:  vm[1],
						Value: val,
					})
				}
			}
			b.enums = append(b.enums, enum)
		}
	}

	// Match: enum <name> { ... };
	re = regexp.MustCompile(`(?s)enum\s+(\w+)\s*\{([^}]+)\}\s*;`)
	matches = re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			enumName := strings.TrimSpace(m[1])
			valuesStr := m[2]

			enum := CEnum{Name: enumName}
			valueRe := regexp.MustCompile(`(\w+)\s*(?:=\s*([^,]+))?`)
			valueMatches := valueRe.FindAllStringSubmatch(valuesStr, -1)
			for _, vm := range valueMatches {
				if len(vm) >= 2 && vm[1] != "" {
					val := ""
					if len(vm) >= 3 {
						val = strings.TrimSpace(vm[2])
					}
					enum.Values = append(enum.Values, CEnumValue{
						Name:  vm[1],
						Value: val,
					})
				}
			}
			b.enums = append(b.enums, enum)
		}
	}
}

// parseStructs extracts struct declarations
func (b *Binder) parseStructs(content string) {
	// Match: struct <name> { ... };
	re := regexp.MustCompile(`(?s)struct\s+(\w+)\s*\{([^}]+)\}\s*;`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			structName := strings.TrimSpace(m[1])
			fieldsStr := m[2]

			struc := CStruct{Name: structName}
			// Parse fields
			fieldRe := regexp.MustCompile(`([\w\s\*\[\]]+)\s+(\w+)\s*;`)
			fieldMatches := fieldRe.FindAllStringSubmatch(fieldsStr, -1)
			for _, fm := range fieldMatches {
				if len(fm) >= 3 {
					struc.Fields = append(struc.Fields, CField{
						Name: strings.TrimSpace(fm[2]),
						Type: parseCType(strings.TrimSpace(fm[1])),
					})
				}
			}
			b.structs = append(b.structs, struc)
		}
	}
}

// parseFunctions extracts function declarations
func (b *Binder) parseFunctions(content string) {
	// Match: <return_type> <name>(<params>);
	// This is a simplified regex - real C parsing is more complex
	re := regexp.MustCompile(`([\w\s\*]+?)\s+(\w+)\s*\(([^)]*)\)\s*;`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 4 {
			returnType := strings.TrimSpace(m[1])
			name := strings.TrimSpace(m[2])
			paramsStr := strings.TrimSpace(m[3])

			// Skip C keywords and common non-function patterns
			if isCKeyword(returnType) || isCKeyword(name) {
				continue
			}

			// Skip if it looks like a typedef or struct
			if strings.Contains(returnType, "typedef") || strings.Contains(returnType, "struct") {
				continue
			}

			fn := CFunction{
				Name:       name,
				ReturnType: parseCType(returnType),
				IsVariadic: strings.Contains(paramsStr, "..."),
			}

			// Parse parameters
			if paramsStr != "void" && paramsStr != "" {
				params := splitParams(paramsStr)
				for _, p := range params {
					p = strings.TrimSpace(p)
					if p == "" || p == "..." {
						continue
					}
					param := parseParam(p)
					if param.Name != "" {
						fn.Params = append(fn.Params, param)
					}
				}
			}

			b.functions = append(b.functions, fn)
		}
	}
}

// parseDefines extracts #define constants
func (b *Binder) parseDefines(content string) {
	re := regexp.MustCompile(`#define\s+(\w+)\s+(.+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			name := strings.TrimSpace(m[1])
			value := strings.TrimSpace(m[2])
			// Skip function-like macros
			if !strings.Contains(name, "(") {
				b.defines[name] = value
			}
		}
	}
}

// parseCType parses a C type string
func parseCType(s string) CType {
	s = strings.TrimSpace(s)

	ct := CType{}

	// Check for const
	if strings.HasPrefix(s, "const ") {
		ct.IsConst = true
		s = strings.TrimPrefix(s, "const ")
	}

	// Check for unsigned
	if strings.HasPrefix(s, "unsigned ") {
		ct.IsUnsigned = true
		s = strings.TrimPrefix(s, "unsigned ")
	}

	// Check for struct
	if strings.HasPrefix(s, "struct ") {
		ct.IsStruct = true
		s = strings.TrimPrefix(s, "struct ")
	}

	// Check for pointer
	if strings.HasSuffix(s, "*") {
		ct.IsPointer = true
		s = strings.TrimSuffix(s, "*")
		s = strings.TrimSpace(s)
	}

	// Check for array
	if strings.Contains(s, "[") {
		ct.IsArray = true
		re := regexp.MustCompile(`\[(\w*)\]`)
		matches := re.FindStringSubmatch(s)
		if len(matches) > 1 {
			ct.ArraySize = matches[1]
		}
		s = re.ReplaceAllString(s, "")
	}

	ct.Name = strings.TrimSpace(s)
	return ct
}

// parseParam parses a function parameter
func parseParam(s string) CParam {
	s = strings.TrimSpace(s)

	// Handle "type name" and "type *name"
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return CParam{}
	}

	// Find where type ends and name begins
	name := parts[len(parts)-1]
	typeParts := parts[:len(parts)-1]

	// Handle pointer in name
	if strings.HasPrefix(name, "*") {
		typeParts = append(typeParts, "*")
		name = strings.TrimPrefix(name, "*")
	}

	// Handle array in name
	if strings.Contains(name, "[") {
		name = strings.Split(name, "[")[0]
	}

	return CParam{
		Name: name,
		Type: parseCType(strings.Join(typeParts, " ")),
	}
}

// splitParams splits parameter list respecting nested parentheses
func splitParams(s string) []string {
	var params []string
	depth := 0
	current := ""

	for _, ch := range s {
		switch ch {
		case '(', '[', '<':
			depth++
			current += string(ch)
		case ')', ']', '>':
			depth--
			current += string(ch)
		case ',':
			if depth == 0 {
				params = append(params, strings.TrimSpace(current))
				current = ""
			} else {
				current += string(ch)
			}
		default:
			current += string(ch)
		}
	}

	if strings.TrimSpace(current) != "" {
		params = append(params, strings.TrimSpace(current))
	}

	return params
}

// isCKeyword checks if string is a C keyword
func isCKeyword(s string) bool {
	keywords := []string{
		"if", "else", "while", "for", "do", "switch", "case", "default",
		"break", "continue", "return", "goto", "sizeof", "typeof",
		"typedef", "extern", "static", "auto", "register",
		"const", "volatile", "restrict",
		"struct", "union", "enum",
		"inline", "noreturn",
	}
	for _, k := range keywords {
		if s == k {
			return true
		}
	}
	return false
}

// GenerateCortex generates Cortex bindings
func (b *Binder) GenerateCortex() string {
	var sb strings.Builder

	sb.WriteString("// Cortex bindings for " + b.libName + "\n")
	sb.WriteString("// Generated by cortex bind\n\n")
	sb.WriteString("#include <" + b.libName + ".h>\n\n")

	// Generate defines as constants
	if len(b.defines) > 0 {
		sb.WriteString("// Constants\n")
		for name, value := range b.defines {
			// Simple heuristic for type
			if strings.Contains(value, "\"") {
				sb.WriteString(fmt.Sprintf("const string %s = %s;\n", name, value))
			} else if strings.ContainsAny(value, ".") {
				sb.WriteString(fmt.Sprintf("const float %s = %s;\n", name, value))
			} else {
				sb.WriteString(fmt.Sprintf("const int %s = %s;\n", name, value))
			}
		}
		sb.WriteString("\n")
	}

	// Generate enums
	for _, enum := range b.enums {
		sb.WriteString(fmt.Sprintf("// Enum: %s\n", enum.Name))
		for _, v := range enum.Values {
			sb.WriteString(fmt.Sprintf("const int %s = %s;\n", v.Name, v.Value))
		}
		sb.WriteString("\n")
	}

	// Generate struct wrappers
	for _, struc := range b.structs {
		sb.WriteString(fmt.Sprintf("// Struct: %s\n", struc.Name))
		sb.WriteString(fmt.Sprintf("struct %s {\n", struc.Name))
		for _, f := range struc.Fields {
			cortexType := cTypeToCortex(f.Type)
			sb.WriteString(fmt.Sprintf("    %s %s;\n", cortexType, f.Name))
		}
		sb.WriteString("}\n\n")
	}

	// Generate function bindings
	sb.WriteString("// Functions\n")
	for _, fn := range b.functions {
		sb.WriteString(b.generateFunctionBinding(fn))
		sb.WriteString("\n")
	}

	return sb.String()
}

// generateFunctionBinding generates a Cortex function binding
func (b *Binder) generateFunctionBinding(fn CFunction) string {
	var sb strings.Builder

	// Determine if cleanup is needed
	cleanup := b.determineCleanup(fn)

	// Generate extern declaration
	returnType := cTypeToCortex(fn.ReturnType)

	sb.WriteString(fmt.Sprintf("extern %s %s(", returnType, fn.Name))

	// Parameters
	for i, p := range fn.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		paramType := cTypeToCortex(p.Type)
		sb.WriteString(fmt.Sprintf("%s %s", paramType, p.Name))
	}

	sb.WriteString(")")

	// Add cleanup annotation if needed
	if cleanup != "" {
		sb.WriteString(fmt.Sprintf(" cleanup(%s)", cleanup))
	}

	sb.WriteString(";\n")

	return sb.String()
}

// determineCleanup determines if a function needs cleanup annotation
func (b *Binder) determineCleanup(fn CFunction) string {
	name := strings.ToLower(fn.Name)

	// Common allocation patterns
	if strings.Contains(name, "alloc") || strings.Contains(name, "create") ||
		strings.Contains(name, "new") || strings.Contains(name, "load") {
		if fn.ReturnType.IsPointer {
			return "free"
		}
	}

	// Library-specific patterns
	switch b.libName {
	case "raylib":
		if strings.HasPrefix(name, "load") {
			return "UnloadTexture"
		}
		if strings.HasPrefix(name, "loadfont") {
			return "UnloadFont"
		}
	}

	return ""
}

// cTypeToCortex converts a C type to Cortex type
func cTypeToCortex(ct CType) string {
	// Handle pointers
	if ct.IsPointer {
		// String detection
		if ct.Name == "char" {
			return "string"
		}
		// Generic pointer
		return "void*" // Cortex handles pointers internally
	}

	// Map C types to Cortex types
	switch ct.Name {
	case "int", "int32_t", "int16_t", "int8_t":
		if ct.IsUnsigned {
			return "int"
		}
		return "int"
	case "uint32_t", "uint16_t", "uint8_t", "unsigned int":
		return "int"
	case "long", "long int", "int64_t":
		return "int"
	case "unsigned long", "uint64_t":
		return "int"
	case "float":
		return "float"
	case "double":
		return "double"
	case "char":
		return "char"
	case "bool", "_Bool":
		return "bool"
	case "void":
		return "void"
	case "size_t":
		return "int"
	default:
		// Struct or typedef name
		return ct.Name
	}
}

// SaveToFile saves the generated bindings to a file
func (b *Binder) SaveToFile(outputPath string) error {
	content := b.GenerateCortex()

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

// Stats returns binding statistics
func (b *Binder) Stats() (functions, structs, enums, defines int) {
	return len(b.functions), len(b.structs), len(b.enums), len(b.defines)
}
