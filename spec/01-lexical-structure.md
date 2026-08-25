# Bee Specification: Lexical Structure (01-lexical-structure.md)

## 1. Character Encoding
- Bee source files: **Strict UTF-8**.

## 2. Token Grammar
```ebnf
identifier ::= [a-zA-Zλ-ωБ-Я][a-zA-Z0-9_]* ;
comment_single ::= "--" [^\n]* ;
comment_block ::= "+-" .* "-+" ;
integer ::= [0-9]+ ;
real ::= [0-9]+ "." [0-9]+ ;
```
