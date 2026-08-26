# Bee Specification: Statements & Expressions (02-statements.md)

## 1. Statement Grammar
Statements are categorized by their imperative (action) or declarative (definition) nature.

```ebnf
statement       ::= declaration | assignment | call | control_flow | transfer ;
declaration     ::= ("new" | "set") identifier (":=" expression | "∈" type) ;
assignment      ::= "let" identifier (":=" | "+=" | "-=" | "&=" | "*=" | "/=") expression ;
control_flow    ::= if_stmt | cycle_stmt | trial_stmt | match_stmt ;
transfer        ::= "panic" | "over" | "exit" | "yield" | "redo" | "next" | "pass" | "raise" ;
```

## 2. Control Flow EBNF
```ebnf
if_stmt         ::= "if" condition "do" block ("else" block)? "done" ;
cycle_stmt      ::= "cycle" label? (while_clause | for_clause)? "do" block "repeat" label? ;
while_clause    ::= "while" condition ;
for_clause      ::= "for" "∀" identifier "∈" range ;
match_stmt      ::= "match" identifier ("all" | "one") "when" branch+ "other" block "done" ;
trial_stmt      ::= "trial" label? block "case" case_block* "miss" block "final" block "done" label? ;
branch          ::= "when" expression_list "do" block ;
case_block      ::= "case" condition "do" block ;
```

## 3. Expression Grammar
```ebnf
expression      ::= primary | binary_op | conditional ;
primary         ::= identifier | constant | "(" expression ")" | call ;
binary_op       ::= expression operator expression ;
conditional     ::= expression "if" condition ("else" expression)? ;
condition       ::= expression ;
expression_list ::= expression ("," expression)* ;
```

## 4. Operational Semantics
- **Statement Sequencing:** Statements are evaluated in order; `;` is mandatory.
- **Scope:** Every block introduces a new lexical region.
- **Indentation:** Mandatory 2-space rule; `IndentationMismatch` triggers a compile-time error.
- **Transfer:** Transfer statements (`exit`, `redo`, `next`) interrupt the current block's control flow.
