# Bee Specification: Lexical Structure (01-lexical-structure.md)

## 1. Character Encoding
- Bee source files: **Strict UTF-8**.

## 2. Tokenization Rules (Maximal Munch)
- **Longest Match:** The Lexer always consumes the longest valid token sequence.
- **Ambiguity Resolution:**
  - `..` (Range) takes precedence over `.` (Member Access).
  - `!!` (Excluded Range) takes precedence over `!` (Range negation/limit).
  - `!.` (Range, first limit missing) takes precedence over `!` (Range negation).
  - `.!` (Range, last limit missing) takes precedence over `.` (Member Access).

## 3. Token Grammar
```ebnf
identifier ::= [a-zA-Zλ-ωБ-Я][a-zA-Z0-9_]* ;
integer    ::= [0-9]+ ;
real       ::= [0-9]+ "." [0-9]+ ;
range_op   ::= ".." | ".!" | "!." | "!!" ;
```

## 4. Operator Map
| Symbol | Range Behavior | Note |
|--------|----------------|------|
| ..     | [min..max]     | Full inclusion |
| .!     | [min..max)     | Exclude upper |
| !.     | (min..max]     | Exclude lower |
| !!     | (min..max)     | Exclude both |

## 5. Lexical Semantics
- **Ranges:** "!" in context of range operators designates missing limits.
- **Indentation:** Mandatory 2-space indentation; physical line breaks are ignored unless within string literals.
- **Public Members:** Start with dot `.` prefix.
- **Identifiers:** Restricted set of Unicode (Greek/Cyrillic) defined in docs.
