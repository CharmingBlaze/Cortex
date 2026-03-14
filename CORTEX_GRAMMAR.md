# Cortex Grammar & Operator Precedence

## Operator Precedence (High → Low)

| Level | Operators | Description | Associativity |
| --- | --- | --- | --- |
| 1 | `()` `[]` `.` `::` `await` postfix `++ --` | Call, index, member, namespace scope, postfix, await | Left-to-right |
| 2 | prefix `++ --` `+ -` `!` `~` `await` `spawn` `async` | Unary/prefix ops | Right-to-left |
| 3 | `* / %` | Multiplicative | Left-to-right |
| 4 | `+ -` | Additive | Left-to-right |
| 5 | `<< >>` | Shifts | Left-to-right |
| 6 | `< <= > >=` | Relational | Left-to-right |
| 7 | `== !=` | Equality | Left-to-right |
| 8 | `&` | Bitwise AND | Left-to-right |
| 9 | `^` | Bitwise XOR | Left-to-right |
| 10 | `|` | Bitwise OR | Left-to-right |
| 11 | `&&` | Logical AND | Left-to-right |
| 12 | `||` | Logical OR | Left-to-right |
| 13 | `|>` | Pipeline operator | Left-to-right |
| 14 | Assignment ops (`=` `+=` `-=` `*=` `/=` `%=` `<<=` `>>=` `&=` `^=` `|=`) | Assignment | Right-to-left |
| 15 | `if` `match` `while` `for` `defer` `try` `async` blocks | Statement keywords | — |

## Lexical Elements

```
letter        ::= "A".."Z" | "a".."z" | "_"
digit         ::= "0".."9"
identifier    ::= letter { letter | digit }
integer_lit   ::= digit { digit }
float_lit     ::= digit { digit } "." digit { digit }
string_lit    ::= '"' { character | escape } '"'
char_lit      ::= '\'' ( character | escape ) '\''
escape        ::= "\\n" | "\\t" | "\\r" | "\\\"" | "\\'" | "\\0" | "\\x" hex hex
comment       ::= "//" { not-newline } newline | "/*" { any } "*/"
```

## Grammar (EBNF excerpt)

```
module        ::= [ "module" identifier ("::" identifier)* ";" ] { import_decl | decl }
import_decl   ::= "import" identifier ("::" identifier)* [ "as" identifier ] ";"
decl          ::= func_decl | var_decl ";" | struct_decl | enum_decl | actor_decl | test_decl | wrapper_decl | extern_block

func_decl     ::= attr_list? async_kw? type identifier param_list return_clause? block
async_kw      ::= "async"
attr_list     ::= { "@" "[" attr ("," attr)* "]" }
attr          ::= identifier [ "(" attr_args? ")" ]
attr_args     ::= expr { "," expr }
param_list    ::= "(" [ param { "," param } ] ")"
param         ::= modifier? type identifier [ "=" expr ]
modifier      ::= "mut" | "ref" | "out"
return_clause ::= "->" type_list
type_list     ::= type { "," type }

var_decl      ::= ("let" | "var") identifier [ ":" type ] [ "=" expr ]
struct_decl   ::= "struct" identifier struct_body
struct_body   ::= "{" { field_decl } "}"
field_decl    ::= attr_list? type identifier ";"
enum_decl     ::= "enum" identifier "{" enum_member { "," enum_member } "}"
enum_member   ::= identifier [ "=" expr ]

actor_decl    ::= "actor" identifier block
test_decl     ::= "test" string_lit block
wrapper_decl  ::= "wrapper" string_lit block
extern_block  ::= "extern" "{" { extern_decl } "}"
extern_decl   ::= type identifier param_list ";"

type          ::= identifier generics? array_type? | "vec2" | "vec3" | "any" | tuple_type | func_type
generics      ::= "<" type_list ">"
array_type    ::= "[" expr? "]"
tuple_type    ::= "(" type_list ")"
func_type     ::= "fn" param_list return_clause?

expr          ::= assignment
assignment    ::= pipeline { assign_op pipeline }
assign_op     ::= "=" | "+=" | "-=" | "*=" | "/=" | "%=" | "<<=" | ">>=" | "&=" | "^=" | "|="
pipeline      ::= logic_or { "|>" logic_or }
logic_or      ::= logic_and { "||" logic_and }
logic_and     ::= bit_or { "&&" bit_or }
bit_or        ::= bit_xor { "|" bit_xor }
bit_xor       ::= bit_and { "^" bit_and }
bit_and       ::= equality { "&" equality }
equality      ::= comparison { ("==" | "!=") comparison }
comparison    ::= shift { ("<" | "<=" | ">" | ">=") shift }
shift         ::= term { ("<<" | ">>") term }
term          ::= factor { ("+" | "-") factor }
factor        ::= unary { ("*" | "/" | "%") unary }
unary         ::= ("!" | "~" | "-" | "++" | "--" | "await" | "spawn") unary | postfix
postfix       ::= primary { call | index | member | postfix_op }
call          ::= "(" [ argument_list ] ")"
argument_list ::= argument { "," argument }
argument      ::= identifier ":" expr | expr
index         ::= "[" expr "]"
member        ::= ("." | "::") identifier
postfix_op    ::= "++" | "--"
primary       ::= literal | identifier | tuple_expr | block_expr | match_expr | lambda_expr | "(" expr ")"

literal       ::= integer_lit | float_lit | string_lit | char_lit | "true" | "false" | "null"
tuple_expr    ::= "(" expr { "," expr } ")"
block_expr    ::= block
lambda_expr   ::= "[" capture_list? "]" param_list return_clause? "=>" (expr | block)
capture_list  ::= capture { "," capture }
capture       ::= identifier | "&" identifier

block         ::= "{" { statement } "}"
statement     ::= attr_list? (var_decl ";" | expr ";" | if_stmt | while_stmt | for_stmt | match_stmt | return_stmt | defer_stmt | block | try_stmt | async_block)
if_stmt       ::= "if" "(" expr ")" block [ "else" (block | if_stmt) ]
while_stmt    ::= "while" "(" expr ")" block
for_stmt      ::= "for" "(" for_init? ";" for_cond? ";" for_iter? ")" block
for_init      ::= var_decl | expr
for_cond      ::= expr
for_iter      ::= expr

match_stmt    ::= "match" "(" expr ")" "{" match_arm+ "}"
match_arm     ::= pattern [ "when" expr ] "=>" (block | expr) ","
pattern       ::= identifier | literal | "_" | tuple_pattern | struct_pattern

return_stmt   ::= "return" expr? ";"
defer_stmt    ::= "defer" block
try_stmt      ::= "try" block [ "catch" "(" identifier ")" block ] [ "finally" block ]
async_block   ::= "async" block
```

This standalone document complements the main language specification and can be referenced by parser, compiler, and tooling authors. 
