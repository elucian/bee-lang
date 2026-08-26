# Bee Lexical Specification (01-lexical-structure.md)

## 1. Character Encoding & Sets
- Encoding: **Strict UTF-8** (RFC 3629).
- Identifier Base: 
  - Latin: `A-Z`, `a-z`
  - Greek: `λ-ω`, `Σ-Ω`
  - Cyrillic: `Б-Я`
- Digits: `0-9`
- Whitespace: Space (U+0020), Tab (U+0009). Physical line breaks act as whitespace except within literals.

## 2. Lexical Tokenization (Maximal Munch)
The lexer always consumes the longest valid token sequence. Ambiguity resolution:
- Range Operators (`..`, `.!`, `!.`, `!!`) take precedence over member access (`.`).
- Assignment operators (`:=`, `::`) take precedence over single-character operators.

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
markup_tag  ::= "<" identifier ">" .* "</" identifier ">" ;

(* Disambiguation *)
member_acc  ::= "." ; 
```

## 5. Escape Sequences
- `\n`: New line
- `\t`: Tab
- `\"`: Double quote
- `\'`: Single quote
- `\\`: Backslash

## 6. Indentation & Scoping
- **Indentation:** Mandatory 2-space indentation.
- **Blocks:** Explicitly terminated by `done`, `repeat`, or `return`.
- **Physical Structure:** Newlines are ignored unless inside string or markup literals.
