package binder

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cc "modernc.org/cc/v3"
)

// ParseOptions configures C header binding (AST path).
type ParseOptions struct {
	LegacyBind     bool
	IncludeDirs    []string
	Defines        []string
	IncludeHeader  string
	SkipPreprocess bool
}

func joinDefineSource(defines []string) string {
	var b strings.Builder
	for _, d := range defines {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		d = strings.TrimPrefix(d, "-D")
		b.WriteString("#define ")
		if idx := strings.IndexByte(d, '='); idx >= 0 {
			b.WriteString(strings.TrimSpace(d[:idx]))
			b.WriteByte(' ')
			b.WriteString(strings.TrimSpace(d[idx+1:]))
		} else {
			b.WriteString(d)
			b.WriteString(" 1")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func acquireHostConfig() (predef string, includePaths, sysIncludePaths []string, err error) {
	candidates := []string{}
	if c := os.Getenv("CC"); strings.TrimSpace(c) != "" {
		candidates = append(candidates, c)
	}
	candidates = append(candidates, "gcc", "clang", "cpp")
	for _, c := range candidates {
		if c == "" {
			continue
		}
		p, i, s, e := cc.HostConfig(c)
		if e == nil {
			return p, i, s, nil
		}
	}
	return "", nil, nil, fmt.Errorf("no C preprocessor found for AST bind (install gcc/clang or set CC)")
}

func absIncludeDirs(dirs []string) ([]string, error) {
	out := make([]string, 0, len(dirs)+1)
	for _, d := range dirs {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		a, err := filepath.Abs(d)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func formatOperand(o cc.Operand) string {
	if o == nil {
		return ""
	}
	v := o.Value()
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case cc.Int64Value:
		return fmt.Sprintf("%d", int64(x))
	case cc.Uint64Value:
		return fmt.Sprintf("%d", uint64(x))
	case cc.Float32Value:
		return fmt.Sprintf("%g", float32(x))
	case cc.Float64Value:
		return fmt.Sprintf("%g", float64(x))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func ccTypeToBinderType(t cc.Type) CType {
	if t == nil {
		return CType{Name: "void"}
	}
	t = t.Alias().Decay()
	ptrs := 0
	for t.Kind() == cc.Ptr {
		ptrs++
		t = t.Elem().Alias().Decay()
	}
	if t.Kind() == cc.Function {
		return CType{Name: "void", IsPointer: ptrs >= 1, IsFunctionPointer: true}
	}
	var ct CType
	if t.Kind() == cc.Array {
		ct = baseCCType(t.Elem().Alias().Decay())
		ct.IsArray = true
		if t.Len() > 0 {
			ct.ArraySize = fmt.Sprintf("%d", t.Len())
		}
	} else {
		ct = baseCCType(t)
	}
	if ptrs > 1 {
		return CType{Name: "void", IsPointer: true}
	}
	if ptrs == 1 {
		ct.IsPointer = true
	}
	return ct
}

func baseCCType(t cc.Type) CType {
	if t == nil {
		return CType{Name: "void"}
	}
	t = t.Alias()
	switch t.Kind() {
	case cc.Void:
		return CType{Name: "void"}
	case cc.Char, cc.SChar, cc.UChar:
		ct := CType{Name: "char"}
		if t.Kind() == cc.UChar {
			ct.IsUnsigned = true
		}
		return ct
	case cc.Short, cc.UShort:
		ct := CType{Name: "short"}
		if t.Kind() == cc.UShort {
			ct.IsUnsigned = true
		}
		return ct
	case cc.Int, cc.UInt, cc.Long, cc.ULong, cc.LongLong, cc.ULongLong:
		if t.Kind() == cc.UInt || t.Kind() == cc.ULong || t.Kind() == cc.ULongLong {
			return CType{Name: "int", IsUnsigned: true}
		}
		return CType{Name: "int"}
	case cc.Float:
		return CType{Name: "float"}
	case cc.Double, cc.LongDouble:
		return CType{Name: "double"}
	case cc.Bool:
		return CType{Name: "bool"}
	case cc.Struct:
		tag := t.Tag().String()
		if tag == "" {
			tag = "struct"
		}
		return CType{Name: tag, IsStruct: true}
	case cc.Union:
		tag := t.Tag().String()
		if tag == "" {
			tag = "union"
		}
		return CType{Name: tag, IsStruct: true}
	case cc.Enum:
		return CType{Name: "int"}
	case cc.TypedefName:
		return CType{Name: t.Name().String()}
	default:
		return CType{Name: t.String()}
	}
}

func collectEnumsFromTU(tu *cc.TranslationUnit) []CEnum {
	var out []CEnum
	for tu != nil {
		ed := tu.ExternalDeclaration
		if ed != nil && ed.Case == cc.ExternalDeclarationDecl && ed.Declaration != nil {
			walkDeclSpecsEnum(ed.Declaration.DeclarationSpecifiers, &out)
		}
		tu = tu.TranslationUnit
	}
	return out
}

func walkDeclSpecsEnum(ds *cc.DeclarationSpecifiers, out *[]CEnum) {
	for ds != nil {
		if ts := ds.TypeSpecifier; ts != nil && ts.Case == cc.TypeSpecifierEnum && ts.EnumSpecifier != nil {
			if e := enumFromSpecifier(ts.EnumSpecifier); e.Name != "" && len(e.Values) > 0 {
				*out = append(*out, e)
			}
		}
		ds = ds.DeclarationSpecifiers
	}
}

func enumFromSpecifier(es *cc.EnumSpecifier) CEnum {
	if es == nil || es.EnumeratorList == nil {
		return CEnum{}
	}
	name := ""
	switch es.Case {
	case cc.EnumSpecifierDef:
		name = es.Token2.Value.String()
	case cc.EnumSpecifierTag:
		name = es.Token2.Value.String()
	default:
		return CEnum{}
	}
	if name == "" {
		return CEnum{}
	}
	var vals []CEnumValue
	for list := es.EnumeratorList; list != nil; list = list.EnumeratorList {
		if en := list.Enumerator; en != nil {
			nm := en.Token.Value.String()
			if nm == "" {
				continue
			}
			val := formatOperand(en.Operand)
			vals = append(vals, CEnumValue{Name: nm, Value: val})
		}
	}
	return CEnum{Name: name, Values: vals}
}

func (b *Binder) parseFromCC(headerAbs string, preprocessed []byte, includeDirs, defines []string) error {
	predef, hostInc, hostSys, err := acquireHostConfig()
	if err != nil {
		return fmt.Errorf("%w\nhint: use -legacy-bind for regex-only parsing without a host C toolchain", err)
	}

	abi, err := cc.NewABIFromEnv()
	if err != nil {
		return err
	}

	userInc, err := absIncludeDirs(includeDirs)
	if err != nil {
		return err
	}

	includePaths := make([]string, 0, len(userInc)+len(hostInc)+2)
	includePaths = append(includePaths, userInc...)
	includePaths = append(includePaths, "@")
	dir := filepath.Dir(headerAbs)
	includePaths = append(includePaths, dir)
	for _, p := range hostInc {
		if filepath.IsAbs(p) {
			includePaths = append(includePaths, p)
		}
	}

	sysPaths := append([]string(nil), hostSys...)

	cfg := &cc.Config{
		Config3: cc.Config3{
			WorkingDir: dir,
		},
		ABI:       abi,
		MaxErrors: -1,
	}

	sources := []cc.Source{
		{Name: "<predefined>", Value: predef, DoNotCache: true},
		{Name: "<built-in>", Value: builtinPrelude, DoNotCache: true},
	}

	if len(preprocessed) == 0 {
		if s := joinDefineSource(defines); s != "" {
			sources = append(sources, cc.Source{Name: "<bind-defs>", Value: s, DoNotCache: true})
		}
		sources = append(sources, cc.Source{Name: headerAbs, DoNotCache: true})
	} else {
		sources = append(sources, cc.Source{Name: filepath.Base(headerAbs) + ".i", Value: string(preprocessed), DoNotCache: true})
	}

	ast, err := cc.Translate(cfg, includePaths, sysPaths, sources)
	if err != nil {
		return fmt.Errorf("cc translate: %w\nhint: try -I/-D, -no-preprocess, or -legacy-bind", err)
	}
	if ast == nil {
		return fmt.Errorf("cc translate returned nil AST")
	}

	seenFn := map[string]struct{}{}
	var dlist []*cc.Declarator
	for d := range ast.TLD {
		dlist = append(dlist, d)
	}
	sort.Slice(dlist, func(i, j int) bool {
		return dlist[i].Name().String() < dlist[j].Name().String()
	})

	for _, d := range dlist {
		if d.IsTypedefName {
			nm := d.Name().String()
			if nm == "" || strings.HasPrefix(nm, "__") {
				continue
			}
			b.typedefs[nm] = ccTypeToBinderType(d.Type())
			continue
		}
		if !d.IsFunctionPrototype() {
			continue
		}
		nm := d.Name().String()
		if nm == "" || strings.HasPrefix(nm, "__builtin_") {
			continue
		}
		if _, ok := seenFn[nm]; ok {
			continue
		}
		seenFn[nm] = struct{}{}

		ft := d.Type().Alias()
		if ft.Kind() != cc.Function {
			continue
		}
		fn := CFunction{
			Name:       nm,
			ReturnType: ccTypeToBinderType(ft.Result()),
			IsVariadic: ft.IsVariadic(),
		}
		for i, p := range ft.Parameters() {
			pt := p.Type()
			paramName := ""
			if decl := p.Declarator(); decl != nil {
				paramName = decl.Name().String()
			}
			if paramName == "" {
				paramName = fmt.Sprintf("arg%d", i)
			}
			ctp := ccTypeToBinderType(pt)
			fn.Params = append(fn.Params, CParam{Name: paramName, Type: ctp})
		}
		b.functions = append(b.functions, fn)
	}

	tags := make([]cc.StringID, 0, len(ast.StructTypes))
	for id := range ast.StructTypes {
		tags = append(tags, id)
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i].String() < tags[j].String() })
	for _, sid := range tags {
		st := ast.StructTypes[sid]
		t := st.Alias()
		if t.Kind() != cc.Struct {
			continue
		}
		if t.IsIncomplete() {
			continue
		}
		name := sid.String()
		if name == "" {
			continue
		}
		cs := CStruct{Name: name}
		nf := t.NumField()
		for i := 0; i < nf; i++ {
			fld := t.FieldByIndex([]int{i})
			if fld == nil {
				continue
			}
			fnm := fld.Name().String()
			if fnm == "" {
				continue
			}
			cs.Fields = append(cs.Fields, CField{
				Name: fnm,
				Type: ccTypeToBinderType(fld.Type()),
			})
		}
		if len(cs.Fields) > 0 {
			b.structs = append(b.structs, cs)
		}
	}

	b.enums = append(b.enums, collectEnumsFromTU(ast.TranslationUnit)...)

	return nil
}

// ParseHeaderWithOptions parses a header using the modernc.org/cc/v3 AST pipeline
// unless LegacyBind is set (then it uses regex ParseHeader).
func (b *Binder) ParseHeaderWithOptions(path string, opt ParseOptions) error {
	if opt.LegacyBind {
		return b.ParseHeader(path)
	}

	b.includeHeader = opt.IncludeHeader
	b.functions = nil
	b.structs = nil
	b.enums = nil
	b.typedefs = make(map[string]CType)
	b.defines = make(map[string]string)

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	b.parseDefinesContent(string(raw))

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	var pre []byte
	if !opt.SkipPreprocess {
		pr := RunCPP(abs, opt.IncludeDirs, opt.Defines)
		if pr.Warning != "" {
			fmt.Fprintln(os.Stderr, pr.Warning)
		}
		pre = pr.Output
	}

	return b.parseFromCC(abs, pre, opt.IncludeDirs, opt.Defines)
}
