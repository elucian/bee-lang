# Solution: LLVM IR Lowering Strategy
- **Frontend:** Lexer/Parser generates an intermediate AST.
- **1-based Indexing:** The parser handles the `1`-based index syntax (arrays/matrices) and performs a `-1` normalization during the lowering phase before emitting LLVM IR offsets.
- **Reference Binding:** Rules do not allocate return types. The compiler emits IR that passes a pointer to a pre-allocated stack/region space into the rule, which the rule then populates.
