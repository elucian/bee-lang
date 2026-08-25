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

## 3. Operator Set
The following operators are strictly reserved. Non-defined Unicode symbols are invalid.

| Category | Operators |
|----------|-----------|
| Range | "!", "..", ".!", "!.", "!!" |
| Logic | "¬", "∧", "∨", "⊕", "↓", "↑" |
| Arithmetic | "+", "-", "*", "/", "×", "÷", "%", "√", "^" |
| Relation | "=", "≠", "≡", "!≡", "≈", ">", "<", "≥", "≤", "∈", "!∈" |
| Assignment | ":", ":=", "::", "+=", "-=", "*=", "/=", "%=", "^=", "√=" |
| Collection | "∩", "∪", "⊂", "⊃", "Δ", "«", "»" |

## 4. Operational Semantics
- **Ranges:** "!" is reserved for range negation.
- **Precedence:** Highest to lowest: Unary {¬, -}, Multiplicative {*, /, ×, ÷, %, √}, Additive {+, -}, Relational {=, ≠, ≡, ∈, <, >}, Logical {∧, ∨}.
- **Indentation:** Mandatory 2-space indentation; physical line breaks are ignored unless within string literals.
