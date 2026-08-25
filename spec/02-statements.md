# Bee Specification: Statements & Expressions

## 1. Syntax Overview
Statements consist of imperative/declarative keywords followed by expressions.

## 2. EBNF
```ebnf
statement ::= declaration | assignment | call | control_flow ;
declaration ::= ("new" | "set") identifier (":=" expression | "∈" type) ;
assignment ::= "let" identifier (":=" | "+=" | "-=") expression ;
```

## 3. Operational Semantics

## 4. Expressions
## 5. Control Flow
```ebnf
control_flow ::= "if" condition "do" block ("else" block)? "done" 
               | "cycle" label? "do" block "repeat" label?
               | "for" "∀" identifier "∈" range "do" block "repeat" ;
```
