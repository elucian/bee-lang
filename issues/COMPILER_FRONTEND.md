# Issue: Compiler Frontend Development
- **Goal:** Track Lexer/Parser development and translation logic.
- **Indexing:** Map Bee's `1`-based user-facing indexing to internal `0`-based LLVM offsets during the lowering phase.
- **Validation:** Ensure AST construction handles explicit rule-result binding (pre-allocated containers).
