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
- Statements are evaluated in the order they appear.
- Declarations must precede usage within the same lexical scope.
