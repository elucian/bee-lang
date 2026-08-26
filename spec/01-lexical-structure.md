# Bee Specification: Lexical Structure (01-lexical-structure.md)

## 1. Character Encoding & Sets
- Encoding: **Strict UTF-8** (RFC 3629).
- Identifier Base: 
  - Latin: `A-Z`, `a-z`
  - Greek: `λ-ω`, `Σ-Ω`
  - Cyrillic: `Б-Я`
- Digits: `0-9`
- Whitespace: Space (U+0020), Tab (U+0009). Physical line breaks act as whitespace except within literals.

## 2. Identifier & Reserved Symbol Rules
- **Type Identifiers:** Single uppercase Latin letter (e.g., `Z`, `N`, `R`).
- **User-Defined Types:** Uppercase start, >1 letter (e.g., `MyType`).
- **Variables:** Lowercase start (e.g., `myVar`, `x`, `count`).
- **User-Defined Constants:** Uppercase start, >1 letter (e.g., `MaxBuffer`).
- **Reserved Symbols (Strict):**
    - `λ`: Lambda-definition operator.
    - `π`: Mathematical constant (3.1415...).
    - `ε`: Tolerance constant (Used in approximate comparisons `≈`).
- **Reserved Semantic Symbols:**
    - `α`, `β`: Reserved for Angle-type variables.
- **Identifiers:** Cannot be single-letter uppercase (Reserved for Types) nor reserved symbols.

## 3. Operator & Delimiter Map
| Class | Symbols |
| :--- | :--- |
| **Range** | `..`, `.!`, `!.`, `!!` |
| **Logic** | `¬`, `∧`, `∨`, `⊕`, `↓`, `↑` |
| **Arithmetic**| `+`, `-`, `*`, `/`, `×`, `÷`, `%`, `√`, `^` |
| **Relation** | `=`, `≠`, `≡`, `!≡`, `≈`, `>`, `<`, `≥`, `≤`, `∈`, `!∈` |
| **Assignment**| `:`, `:=`, `::`, `+=`, `-=`, `*=`, `/=`, `%=`, `^=`, `√=` |
| **Collection**| `∩`, `∪`, `⊂`, `⊃`, `Δ`, `«`, `»` |
| **Delimiter**| `(`, `)`, `[`, `]`, `{`, `}`, `,`, `;` |

## 4. Formal EBNF Grammar
```ebnf
(* Tokens *)
identifier  ::= [a-zA-Zλ-ωБ-Я][a-zA-Z0-9_]* ;
integer     ::= [0-9]+ ;
real        ::= [0-9]+ "." [0-9]+ ;
unicode_lit ::= "U+" [0-9A-Fa-f]{4} | "U-" [0-9A-Fa-f]{8} ;
range_op    ::= ".." | ".!" | "!." | "!!" ;
string_lit  ::= "'" (char_esc | [^'\\])* "'" 
              | '"' (char_esc | [^"\\])* '"' 
              | "`" (char_esc | [^`\\])* "`" ;
char_esc    ::= "\\" ("n" | "t" | '"' | "'" | "\\") ;
markup_block   ::= "<" tag_name (attribute)* ">" (markup_content | markup_nested)* "</" tag_name ">" ;
tag_name       ::= identifier | "text" | "sql" | "html" | "xml" | "json" | "code" ;
attribute      ::= identifier "=" '"' identifier '"' ;
markup_content ::= [^<]* ;
markup_nested  ::= markup_block ;

(* Disambiguation *)
member_acc  ::= "." ; 
```

## 5. Indentation & Scoping
- **Indentation:** Mandatory 2-space indentation.
- **Blocks:** Explicitly terminated by `done`, `repeat`, or `return`.
- **Physical Structure:** Newlines are ignored unless inside string or markup literals.
