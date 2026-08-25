# Issue: Lexical Grammar Completeness
- **Status:** Open
- **Description:** `spec/01-lexical-structure.md` is currently a stub that lacks critical lexical definitions.
- **Technical Debt:**
    - **String Literals:** Missing definitions for `' '`, `" "`, and backquoted `` ` `` literals, including escape sequence rules.
    - **Markup Tags:** Missing EBNF for `<text>`, `<sql>`, etc.
    - **Unicode Identifiers:** Regex is currently too restrictive; needs to support the full range of Greek/Cyrillic identifiers described in `11-processing.md`.
    - **Tokenization Rules:** Missing "Maximal Munch" definition to disambiguate operators like `.` vs `..` vs `.` (member).
- **Plan:** Refine `spec/01-lexical-structure.md` with full EBNF and tokenization strategy.
