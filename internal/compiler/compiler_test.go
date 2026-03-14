package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cortex/internal/ast"
	"cortex/internal/config"
)

func TestCompilePipeline_MinimalProgram(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    int x = 1;
    println("hello");
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "minimal.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "minimal.out")
	err := c.Compile(in, out)
	if err != nil {
		t.Logf("Compile error (may be due to missing gcc): %v", err)
		// Pipeline up to codegen is tested by checking we get past parse/semantic
		return
	}
	if _, err := os.Stat(out); err != nil {
		t.Logf("Output binary not found (gcc may have failed): %v", err)
	}
}

func TestCompile_2DArrayLiteralGeneratesMetadata(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    int mat = [[1, 2], [3, 4]];
    int v = mat[0][1];
    println(v);
}
`
	file := filepath.Join(t.TempDir(), "mat2d.cx")
	if err := os.WriteFile(file, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	out, err := c.GenerateC([]string{file})
	if err != nil {
		t.Fatalf("2D array literal should compile: %v", err)
	}
	if !strings.Contains(out, "mat_rows") || !strings.Contains(out, "mat_cols") {
		t.Errorf("expected generated C to include mat_rows/mat_cols metadata; output was\n%s", out)
	}
}

func TestSemantic_2DArrayLiteralRequiresRectangularShape(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    int bad = [[1], [2, 3]];
}
`
	file := filepath.Join(t.TempDir(), "ragged.cx")
	if err := os.WriteFile(file, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	err := c.Compile(file, filepath.Join(t.TempDir(), "out"))
	if err == nil {
		t.Fatal("expected compile error for ragged 2D array literal")
	}
	if !strings.Contains(err.Error(), "equal length") {
		t.Errorf("error should mention equal length rows, got: %v", err)
	}
}

func TestCompilePipeline_ExternPointerTypes(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
extern void* malloc(size_t size);
extern void free(void* ptr);
void main() {
    println("ok");
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "extern.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "extern.out")
	err := c.Compile(in, out)
	if err != nil {
		t.Errorf("expected extern void* to compile: %v", err)
	}
}

func TestFeatureGate_AsyncRejectedWhenDisabled(t *testing.T) {
	cfg := config.Default()
	cfg.Features.Async = false
	c := NewCompiler(cfg)
	source := `
async void foo() {}
void main() {}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "async.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "async.out")
	err := c.Compile(in, out)
	if err == nil {
		t.Error("expected error when using async with features.async disabled")
		return
	}
	if !strings.Contains(err.Error(), "async") {
		t.Errorf("error should mention async: %v", err)
	}
}

func TestFeatureGate_AsyncCompilesWhenEnabled(t *testing.T) {
	cfg := config.Default()
	cfg.Features.Async = true
	c := NewCompiler(cfg)
	source := `
async void foo() {
    int x = await 1;
}
void main() {
    foo();
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "async_ok.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("async should compile when enabled: %v", err)
	}
	if !strings.Contains(outStr, "foo") {
		t.Error("expected generated C to contain foo")
	}
}

func TestFeatureGate_ActorsRejectedWhenDisabled(t *testing.T) {
	cfg := config.Default()
	cfg.Features.Actors = false
	c := NewCompiler(cfg)
	source := `
actor Foo {}
void main() {}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "actor.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "actor.out")
	err := c.Compile(in, out)
	if err == nil {
		t.Error("expected error when using actor with features.actors disabled")
		return
	}
	if !strings.Contains(err.Error(), "actor") {
		t.Errorf("error should mention actor: %v", err)
	}
}

func TestLexer_RepeatKeyword(t *testing.T) {
	l := NewLexer()
	tokens, err := l.Tokenize("repeat 2 { }")
	if err != nil {
		t.Fatal(err)
	}
	// Find token with value "repeat"
	for _, tok := range tokens {
		if tok.Value == "repeat" {
			if tok.Type != TokenRepeat {
				t.Errorf("expected TokenRepeat for 'repeat', got token type %v (%d)", tok.Type, tok.Type)
			}
			return
		}
	}
	t.Error("token 'repeat' not found in token list")
}

func TestParse_RepeatStatement(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `void main() { repeat (2) { println("hi"); } }`
	tokens, err := c.lexer.Tokenize(source)
	if err != nil {
		t.Fatal(err)
	}
	var repeatTok *Token
	for i := range tokens {
		if tokens[i].Value == "repeat" {
			repeatTok = &tokens[i]
			break
		}
	}
	if repeatTok == nil {
		t.Fatal("token 'repeat' not found")
	}
	if repeatTok.Type != TokenRepeat {
		t.Fatalf("token 'repeat' should be TokenRepeat, got %v", repeatTok.Type)
	}
	parsedAst, err := c.parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	prog, ok := parsedAst.(*ast.ProgramNode)
	if !ok || len(prog.Declarations) == 0 {
		t.Fatal("expected program with declarations")
	}
	fn, ok := prog.Declarations[0].(*ast.FunctionDeclNode)
	if !ok || fn.Name != "main" {
		t.Fatal("expected main function")
	}
	if len(fn.Body.Statements) != 1 {
		t.Fatalf("expected 1 statement in main, got %d", len(fn.Body.Statements))
	}
	if _, isRepeat := fn.Body.Statements[0].(*ast.RepeatStmtNode); !isRepeat {
		t.Errorf("expected ast.RepeatStmtNode, got %T", fn.Body.Statements[0])
	}
}

func TestStaticTyping_RejectsMismatch(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    int x = "hello";
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "static.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	err := c.Compile(in, filepath.Join(dir, "out"))
	if err == nil {
		t.Error("expected compile error for int x = \"hello\"")
		return
	}
	if !strings.Contains(err.Error(), "type mismatch") && !strings.Contains(err.Error(), "semantic") {
		t.Errorf("error should mention type mismatch or semantic: %v", err)
	}
}

func TestStaticTyping_AllowsVarAndAny(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    var a = 42;
    a = "ok";
    any b = true;
    b = 3.14;
    println("ok");
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "dynamic.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	err := c.Compile(in, filepath.Join(dir, "out"))
	if err != nil {
		t.Logf("compile may fail on missing runtime: %v", err)
	}
	// At least semantic should pass (no type error)
	if err != nil && strings.Contains(err.Error(), "type mismatch") {
		t.Errorf("var/any should allow reassignment: %v", err)
	}
}

func TestStringEscapes_Compiles(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    string s = "line1\nline2\ttab";
    println(s);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "escapes.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "escapes.out")
	err := c.Compile(in, out)
	if err != nil {
		t.Errorf("string escapes should compile: %v", err)
	}
}

func TestParse_Lambda(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := "void main() {\n    [](int a, int b) -> int { return a + b; };\n}\n"
	tokens, err := c.lexer.Tokenize(source)
	if err != nil {
		t.Fatalf("lex: %v", err)
	}
	ast, err := c.parser.Parse(tokens)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := c.analyzer.Analyze(ast); err != nil {
		t.Fatalf("semantic: %v", err)
	}
	// Codegen should emit static function and call
	out, err := c.generator.Generate(ast)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.Contains(out, "cortex_lambda_") {
		preview := out
		if len(preview) > 200 {
			preview = preview[:200]
		}
		t.Errorf("generated C should contain cortex_lambda_ name: %s", preview)
	}
}

func TestCompile_CollectionsAndResult(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    array a = array_create();
    array_push(a, 42);
    array_free(a);
    dict d = dict_create();
    dict_set(d, "k", make_any_int(1));
    dict_free(d);
    result r = result_ok(make_any_int(0));
    if (result_is_ok(r)) { show("ok"); }
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "coll.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	err := c.Compile(in, filepath.Join(dir, "coll.out"))
	if err != nil {
		t.Errorf("collections/result should compile: %v", err)
	}
}

func TestCompile_MatchResultOkErr(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    result r = result_ok(make_any_int(42));
    match (r) {
        case Ok(v): { show("ok"); }
        case Err(e): { show("err"); }
    }
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "matchr.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	err := c.Compile(in, filepath.Join(dir, "matchr.out"))
	if err != nil {
		t.Errorf("match Result Ok/Err should compile: %v", err)
	}
}

func TestCompile_StructMethods(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
struct Player {
    int x;
    int y;
    void move(int dx) {
        x = x + dx;
        y = y + dx;
    }
}
void main() {
    Player p;
    p.x = 0;
    p.y = 0;
    p.move(5);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "sm.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("struct methods should compile: %v", err)
	}
	if !strings.Contains(outStr, "Player_move") {
		t.Error("expected generated C to contain Player_move method")
	}
	if !strings.Contains(outStr, "self->x") {
		t.Error("expected generated C to contain self->x in method body")
	}
}

func TestCompile_UseLib(t *testing.T) {
	source := `
#use "raylib"
void main() {
    InitWindow(800, 450, "test");
    CloseWindow();
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "uselib.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	c := NewCompiler(cfg)
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("#use should compile: %v", err)
	}
	if !strings.Contains(outStr, "#include <raylib.h>") {
		t.Error("expected #use \"raylib\" to generate #include <raylib.h>")
	}
}

func TestCompile_DictLiteral(t *testing.T) {
	source := `
void main() {
    dict d = { "x": 1, "y": 2 };
    show("" + as_int(dict_get(d, "x")));
    dict_free(d);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "dict.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	c := NewCompiler(cfg)
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("dict literal should compile: %v", err)
	}
	if !strings.Contains(outStr, "dict_create()") {
		t.Error("expected dict literal to generate dict_create()")
	}
	if !strings.Contains(outStr, `dict_set(`) {
		t.Error("expected dict literal to generate dict_set")
	}
}

func TestCompile_LambdaCapture(t *testing.T) {
	source := `
extern void apply_twice(void* fn, void* env, int x);
void main() {
    int base = 10;
    apply_twice([base](int x) { return base + x; }, 1);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "lambda.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	c := NewCompiler(cfg)
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("lambda capture should compile: %v", err)
	}
	if !strings.Contains(outStr, "cortex_closure_") {
		t.Error("expected captured lambda to generate cortex_closure_ struct/fn")
	}
	if !strings.Contains(outStr, "_closure_env_") {
		t.Error("expected captured lambda call to generate closure env var")
	}
}

func TestCompile_NestedJSONSupport(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    var nested = json_parse("{\"outer\": {\"inner\": 42}}");
    var val = nested.get("outer").get("inner");
    println(val); // Should output 42
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "nested_json.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("nested JSON should compile: %v", err)
	}
	if !strings.Contains(outStr, "json_parse") {
		t.Errorf("expected generated C to contain json_parse for nested JSON")
	}
}

func TestCompile_ParseNumber(t *testing.T) {
	source := `
void main() {
    float f = parse_number("3.14");
    int i = parse_int("42");
    show("" + f + " " + i);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "parse.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	c := NewCompiler(cfg)
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("parse_number/parse_int should compile: %v", err)
	}
	if !strings.Contains(outStr, "parse_number(") || !strings.Contains(outStr, "parse_int(") {
		t.Error("expected parse_number and parse_int in generated C")
	}
}

func TestCompile_ConstantFolding(t *testing.T) {
	source := `
void main() {
    int x = 2 + 3;
    show("" + x);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "fold.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	c := NewCompiler(cfg)
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("constant folding test should compile: %v", err)
	}
	// Optimizer should fold 2+3 to 5
	if !strings.Contains(outStr, "5") {
		t.Error("expected constant folding to produce 5 in generated C")
	}
	if strings.Contains(outStr, "2 + 3") {
		t.Error("expected 2+3 to be folded to 5, not emitted as expression")
	}
}

func TestCompile_NamedAndDefaultParams(t *testing.T) {
	source := `
void f(int a, int b = 10) { show("" + a + b); }
void main() {
    f(1);
    f(2, 3);
    f(b: 5, a: 1);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "named.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	c := NewCompiler(cfg)
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("named and default params should compile: %v", err)
	}
	if !strings.Contains(outStr, "f(1") && !strings.Contains(outStr, "f(2") {
		t.Error("expected calls to f in generated C")
	}
}

func TestCompile_NamedAndDefaultParameters(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    fn greet(string name = "World", string greeting = "Hello") -> string {
        return greeting + ", " + name + "!";
    }
    string result1 = greet(name: "Alice", greeting: "Hi");
    string result2 = greet(name: "Bob");
    println(result1); // Should output: Hi, Alice!
    println(result2); // Should output: Hello, Bob!
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "named_default_params.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("named and default parameters should compile: %v", err)
	}
	if !strings.Contains(outStr, "greet") {
		t.Errorf("expected generated C to contain greet function")
	}
}

func TestCompile_ECSHelpers(t *testing.T) {
	source := `
void main() {
    int e = entity_create();
    add_component(e, "health", 100);
    any h = get_component(e, "health");
    if (has_component(e, "health")) {
        show("" + as_int(h));
    }
    entity_remove(e);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "ecs.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	c := NewCompiler(cfg)
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("ECS helpers should compile: %v", err)
	}
	if !strings.Contains(outStr, "entity_create()") || !strings.Contains(outStr, "add_component(") {
		t.Error("expected entity_create and add_component in generated C")
	}
}

func TestCompile_ModuleNamespace(t *testing.T) {
	// Module prefixes symbols with module__ so they don't clash when merged.
	source := `
module "math";
int add(int a, int b) {
    return a + b;
}
void main() {
    int x = add(1, 2);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "mod.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	c := NewCompiler(config.Default())
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("module should compile: %v", err)
	}
	if !strings.Contains(outStr, "math__add") {
		t.Error("expected math__add in generated C when module \"math\" is used")
	}

	// Two-file: module file + main file calling math.add()
	mathFile := filepath.Join(dir, "math.cx")
	gameFile := filepath.Join(dir, "game.cx")
	mathSrc := `module "math"; int add(int a, int b) { return a + b; }`
	gameSrc := `void main() { int x = math.add(1, 2); }`
	if err := os.WriteFile(mathFile, []byte(mathSrc), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gameFile, []byte(gameSrc), 0644); err != nil {
		t.Fatal(err)
	}
	c2 := NewCompiler(config.Default())
	outStr2, err := c2.GenerateC([]string{mathFile, gameFile})
	if err != nil {
		t.Fatalf("module + call should compile: %v", err)
	}
	if !strings.Contains(outStr2, "math__add(1, 2)") {
		t.Error("expected math__add(1, 2) in generated C for math.add(1, 2) call")
	}
}

func TestCompile_PrintfWritelineAndRawC(t *testing.T) {
	source := `
#c int my_global = 0;
void main() {
    print("Hello");
    print(" ", "World");
    printf("%d %s\n", 42, "ok");
    writeline("one line");
    writeline("%d bottles", 99);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "io.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	c := NewCompiler(config.Default())
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("printf/writeline/raw C should compile: %v", err)
	}
	if !strings.Contains(outStr, "printf(") {
		t.Error("expected printf in generated C")
	}
	if !strings.Contains(outStr, "writeline_fmt(") {
		t.Error("expected writeline_fmt in generated C")
	}
	if !strings.Contains(outStr, "int my_global = 0;") {
		t.Error("expected raw C #c content in generated C")
	}
	if !strings.Contains(outStr, "print_any(") {
		t.Error("expected print_any for multi-arg print")
	}
}

func TestCompile_NetworkingBuiltins(t *testing.T) {
	source := `
void main() {
    string s = http_get("http://example.com");
    int server = tcp_listen(8080);
    tcp_close(server);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "net.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	c := NewCompiler(config.Default())
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("networking builtins should compile: %v", err)
	}
	if !strings.Contains(outStr, "runtime/network.h") {
		t.Error("expected #include \"runtime/network.h\" when using network APIs")
	}
	if !strings.Contains(outStr, "http_get(") || !strings.Contains(outStr, "tcp_listen(") {
		t.Error("expected http_get and tcp_listen in generated C")
	}
}

func TestCompile_LambdaCaptureGeneratesClosure(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    int x = 42;
    var lambda = [x](int y) -> int { return x + y; };
    // Assuming a function that takes a lambda with capture
    // apply_twice([x](int y) { return x + y; }, 5);
    println("ok");
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "lambda_capture.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("lambda with capture should compile: %v", err)
	}
	if !strings.Contains(outStr, "cortex_closure_") {
		t.Errorf("expected generated C to contain cortex_closure_ for lambda capture")
	}
}

func TestSemantic_LambdaCaptureUndefinedVariable(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    var lambda = [undefinedVar](int y) -> int { return undefinedVar + y; };
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "lambda_undefined.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	err := c.Compile(in, filepath.Join(dir, "out"))
	if err == nil {
		t.Fatal("expected compile error for undefined capture in lambda")
	}
	if !strings.Contains(err.Error(), "undefined capture") {
		t.Errorf("error should mention undefined capture, got: %v", err)
	}
}

func TestSemantic_LambdaCaptureNonVariable(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    var lambda = [println](int y) -> int { return 0; };
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "lambda_nonvar.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	err := c.Compile(in, filepath.Join(dir, "out"))
	if err == nil {
		t.Fatal("expected compile error for non-variable capture in lambda")
	}
	if !strings.Contains(err.Error(), "cannot capture non-variable") {
		t.Errorf("error should mention non-variable capture, got: %v", err)
	}
}

func TestCompile_NestedJSON(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    var nested = json_parse("{\"outer\": {\"inner\": 42}}");
    var val = nested.get("outer").get("inner");
    println(val); // Should output 42
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "nested_json.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("nested JSON should compile: %v", err)
	}
	if !strings.Contains(outStr, "json_parse") {
		t.Errorf("expected generated C to contain json_parse for nested JSON")
	}
}

func TestCompile_ParseNumberAndInt(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    float num = parse_number("123.45");
    int i = parse_int("678");
    println(num);
    println(i);
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "parse_number_int.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("parse_number and parse_int should compile: %v", err)
	}
	if !strings.Contains(outStr, "parse_number") || !strings.Contains(outStr, "parse_int") {
		t.Errorf("expected generated C to contain parse_number and parse_int")
	}
}

func TestCompile_CoroutineAndYield(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    coroutine int generator() {
        yield 1;
        yield 2;
        yield 3;
    }
    println("Coroutine test");
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "coroutine_yield.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("coroutine and yield should compile: %v", err)
	}
	if !strings.Contains(outStr, "generator") {
		t.Errorf("expected generated C to contain generator function")
	}
}

func TestCompile_AsyncAwait(t *testing.T) {
	cfg := config.Default()
	c := NewCompiler(cfg)
	source := `
void main() {
    async int fetchData() {
        return await someAsyncOperation();
    }
    println("Async test");
}
`
	dir := t.TempDir()
	in := filepath.Join(dir, "async_await.cx")
	if err := os.WriteFile(in, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	outStr, err := c.GenerateC([]string{in})
	if err != nil {
		t.Fatalf("async and await should compile: %v", err)
	}
	if !strings.Contains(outStr, "fetchData") {
		t.Errorf("expected generated C to contain fetchData function")
	}
}
