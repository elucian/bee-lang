# Bee Specification: Statements & Expressions (02-statements.md)

## 2.1 Statement Grammar
```ebnf
statement     ::= declaration | assignment | call | control_flow ;
declaration   ::= ("new" | "set") identifier (":=" expression | "∈" type) ;
assignment    ::= "let" identifier (":=" | "+=" | "-=" | "&=") expression ;
control_flow  ::= if_stmt | cycle_stmt | trial_stmt | match_stmt ;
```

## 2.2 Control Flow EBNF
```ebnf
if_stmt       ::= "if" condition "do" block ("else" block)? "done" ;
cycle_stmt    ::= "cycle" label? "do" block "repeat" label? ;
match_stmt    ::= "match" identifier ("all" | "one") "when" branch+ "other" block "done" ;
trial_stmt    ::= "trial" label? block "case" case_block* "miss" block "final" block "done" label? ;
```

## 3. Operational Semantics
- **Statement Sequencing:** Statements are evaluated in order; `;` is mandatory.
- **Scope:** Every block introduces a new lexical region.
- **Indentation:** Mandatory 2-space rule; `IndentationMismatch` triggers a compile-time error.
