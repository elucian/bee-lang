# Issue: Lexical Grammar Completeness
- **Status:** Resolved
- **Description:** `spec/01-lexical-structure.md` contains comprehensive lexical definitions aligned with the language manual.
- **Completed Improvements:**
    - **String Literals:** Added EBNF and escape rules for single-quoted `' '`, double-quoted interpolated `" #(expr) "`, and raw backtick `` ` `` strings.
    - **Markup Tags:** Complete EBNF for embedded DSL blocks (`<sql>`, `<html>`, `<code>`, `<text>`, etc.).
    - **Unicode Identifiers:** Full character definitions for Greek (`α-ω`, `Α-Ω`), Cyrillic (`а-я`, `А-Я`), subscript indices (`x₀`, `x₁`), and superscript power expressions (`x²`).
    - **Tokenization Rules:** Explicit "Maximal Munch" strategy disambiguating `.`, `..`, `.!`, `!.`, `!!`, `:`, `:=`, `::`, `--`, and `+-`.
- **Spec Reference:** `spec/01-lexical-structure.md`
