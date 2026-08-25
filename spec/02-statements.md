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
```ebnf
expression ::= primary | binary_op | conditional ;
primary ::= identifier | constant | "(" expression ")" | call ;
binary_op ::= expression operator expression ;
conditional ::= expression "if" condition ("else" expression)? ;
```
