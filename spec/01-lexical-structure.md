# Bee Specification: Lexical Structure (01-lexical-structure.md)

## 1. Character Encoding & Source Representation

- **Encoding Standard:** All Bee source code files MUST be encoded in valid **UTF-8** (RFC 3629). Byte Order Marks (BOM) are disallowed and trigger a lexical error `E0101: UnexpectedBOM`.
- **Source Units & Rune Handling:** Source streams consist of decoded UTF-8 sequences (runes). All lexing and parsing MUST treat source inputs strictly as decoded UTF-8 sequences via `[]rune` slices or `bufio.Reader.ReadRune()`. Never use raw byte indexing (`s[i]`) for tokenization.
- **Operator Syntax:** Bee operators can consist of single Unicode symbols (e.g., `≠`), standard ASCII symbols, mixed combinations of Unicode and ASCII characters, and Unicode superscript/subscript ranges.
- **Lookahead Matching:** Implement lexing with rune lookahead (`peekRune()`) rather than fixed-size assumptions to correctly parse multi-character operators containing mixed scripts or modifiers.
- **Token Definitions:** Store operators as `[]rune` slices in token lookup tables or transition tries to support arbitrary multi-rune Unicode operators.
- **Whitespace & Line Terminology:**
  - Horizontal Whitespace: Space (`U+0020`), Tab (`U+0009`).
  - Line Breaks: Line Feed (`U+000A` / `\n`), Carriage Return + Line Feed (`U+000D U+000A` / `\r\n`).
  - Physical line breaks act as token separators except within string literals, multi-line string backticks, or markup block payloads.

---

## 2. Tokenization Strategy & Maximal Munch Disambiguation

The Bee lexer employs a strict **Maximal Munch (Longest Match)** rule: at any position, the lexer consumes the longest sequence of characters that forms a valid token.

### 2.1 Ambiguous Character Resolution
1. **Dot (`.`) Disambiguation:**
   - Single Dot (`.`): Member access operator (e.g., `object.member`).
   - Double Dot (`..`): Inclusive range operator (e.g., `1..10`).
   - Dot-Exclamation (`.!`): Left-inclusive, right-exclusive range (e.g., `1.!10`).
   - Exclamation-Dot (`!.`): Left-exclusive, right-inclusive range (e.g., `1!.10`).
   - Double Exclamation (`!!`): Fully exclusive range (e.g., `1!!10`).
   - Numeric Decimal (`1.5`): Disambiguated by lookahead: if digits follow `.`, it is parsed as a Real literal unless followed immediately by another `.`.

2. **Colon (`:`) Disambiguation:**
   - Single Colon (`:`): Type signature / named parameter separator (e.g., `x: Z`).
   - Colon-Equals (`:=`): Variable assignment operator (e.g., `let x := 10;`).
   - Double-Colon (`::`): Deep clone assignment operator (e.g., `let copy :: original;`).

3. **Comment Markers:**
   - `--`: Single-line comment; consumes characters up to `\n` or `EOF`.
   - `+-`: Block comment header; consumes all code points until matching closing `-+` delimiter. Supports nested block comments `+- ... +- ... -+ ... -+`.
   - `(: ... :)`: Expression comment; delimited by `(:` and `:)`. Supports nesting and can be placed inside expressions or span regions containing other comments.

---

## 3. Identifiers & Reserved Names

Identifiers follow strict casing and character set conventions to preserve mathematical readability:

- **Variable & Routine Identifiers:** Must start with a lowercase Latin letter (`a-z`), lowercase Greek letter (`α-ω`), or lowercase Cyrillic letter (`а-я`). May be followed by letters, digits (`0-9`), underscores (`_`), or subscript digits (`₀-₉`).
- **User-Defined Type / Constant Identifiers:** Must start with an uppercase letter (`A-Z`, `Α-Ω`, `А-Я`) and contain 2 or more characters.
- **Single-Letter Built-in Types (Strictly Reserved):**
  - `Z` (Integer), `N` (Natural/Unsigned), `R` (Real/Float), `Q` (Rational), `C` (Complex)
  - `B` (Boolean), `S` (String), `A` (Array), `M` (Map), `L` (List), `G` (Graph)
- **Math & Domain Symbols:**
  - `λ`: Lambda expression marker.
  - `π`: Constant pi (`3.141592653589793...`).
  - `ε`: Tolerance epsilon for floating-point comparison (`≈`).
  - `α`, `β`: Reserved for angle quantities (`∠`).

---

## 4. Literals & Escape Sequences

### 4.1 String Literals
1. **Single-Quoted Strings (`'...'`):**
   - Interpret standard backslash escape sequences. Cannot span physical line breaks.
2. **Double-Quoted Strings (`"..."`):**
   - Interpolated strings. Expression interpolation uses `#(expression)` syntax (e.g., `"val = #(a + b)"`).
3. **Raw Multi-Line Strings (``` `...` ```):**
   - Enclosed in backticks. Preserves raw newlines and characters without escape processing.

### 4.2 Standard Escape Sequences
| Sequence | Value / Character |
| :--- | :--- |
| `\n` | Line feed (`U+000A`) |
| `\r` | Carriage return (`U+000D`) |
| `\t` | Horizontal tab (`U+0009`) |
| `\\` | Backslash (`U+005C`) |
| `\'` | Single quote (`U+0027`) |
| `\"` | Double quote (`U+0022`) |
| `\0` | Null character (`U+0000`) |
| `\uXXXX` | 16-bit Unicode character (4 hex digits) |
| `\UXXXXXXXX` | 32-bit Unicode character (8 hex digits) |

### 4.3 Embedded Markup Blocks
Embedded domain-specific text blocks are delimited by opening `<tag>` and closing `</tag>` markers:
- **Supported Tags:** `<text>`, `<sql>`, `<html>`, `<xml>`, `<json>`, `<code>`.
- Payloads inside markup blocks are preserved verbatim as raw multiline string slices for runtime engines or preprocessors.

---

## 5. Formal Lexical EBNF Grammar

```ebnf
(* Character Sets *)
latin_lower    ::= [a-z] ;
latin_upper    ::= [A-Z] ;
greek_lower    ::= [α-ω] ;
greek_upper    ::= [Α-Ω] ;
cyrillic_lower ::= [а-я] ;
cyrillic_upper ::= [А-Я] ;
digit          ::= [0-9] ;
subscript_digit::= [₀-₉] ;
superscript    ::= [⁰-⁹ᵃ-ᶻ⁺⁻] ;

(* Identifiers *)
variable_ident ::= ( latin_lower | greek_lower | cyrillic_lower ) ( latin_lower | latin_upper | greek_lower | greek_upper | cyrillic_lower | cyrillic_upper | digit | "_" | subscript_digit )* ;
type_ident     ::= ( latin_upper | greek_upper | cyrillic_upper ) ( latin_lower | latin_upper | greek_lower | greek_upper | cyrillic_lower | cyrillic_upper | digit | "_" )+ ;
builtin_type   ::= "Z" | "N" | "R" | "Q" | "C" | "B" | "S" | "A" | "M" | "L" | "G" ;

(* Numerics *)
integer_lit    ::= digit+ ;
real_lit       ::= digit+ "." digit+ [ ( "e" | "E" ) [ "+" | "-" ] digit+ ] ;

(* Operators & Delimiters *)
assign_op      ::= ":" | ":=" | "::" | "+=" | "-=" | "*=" | "/=" | "%=" | "^=" ;
range_op       ::= ".." | ".!" | "!." | "!!" ;
comment_single ::= "--" [^\n]* ;
	comment_block  ::= "+-" ( [^+] | "+" [^-] | comment_block )* "-+" ;
	expr_comment   ::= "(:" ( [^:] | ":" [^)] )* ":)" ;

(* Strings & Markup *)
single_string  ::= "'" ( escape_seq | [^'\\] )* "'" ;
interp_string  ::= '"' ( escape_seq | "#(" expression ")" | [^"\\] )* '"' ;
raw_string     ::= "`" [^`]* "`" ;
markup_block   ::= "<" tag_name ( attribute )* ">" markup_payload "</" tag_name ">" ;
tag_name       ::= "text" | "sql" | "html" | "xml" | "json" | "code" ;
attribute      ::= variable_ident "=" '"' [^"]* '"' ;
markup_payload ::= [^<]* ;
```

---

## 6. Diagnostic Lexical Codes

| Error Code | Error Condition | Resolution |
| :--- | :--- | :--- |
| `E0101` | Unexpected UTF-8 BOM marker | Strip BOM from source header |
| `E0102` | Unterminated block comment (`+-`) | Ensure closing `-+` exists |
| `E0103` | Invalid escape sequence in string | Use valid `\n`, `\t`, `\uXXXX` escapes |
| `E0104` | Single-letter uppercase variable identifier | Use lowercase start or multi-letter type identifier |
| `E0105` | Unterminated string or markup block | Check closing quotes or matching tag |

---

## 7. Issue Alignment

- **Resolves:** `issues/07-lexical-completeness.md` (Fully defines string interpolation, raw backtick strings, markup EBNF, Unicode sets, escape codes, and Maximal Munch disambiguation rules).
