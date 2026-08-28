# Solution: Structural Binding Operator (:)
Define `:` exclusively as a binding operator for structural associations.

## Design
1.  **Lexer/Parser**: Treat `:` as a `TOKEN_PAIR_UP` separate from `TOKEN_ASSIGN` (`:=`) and `TOKEN_CLONE` (`::`).
2.  **Grammar**: `:` is restricted to:
    - Map/Dictionary key-value pairs (`key: value`).
    - Named parameter binding (`func(n: 5)`).
    - Struct field initialization.
3.  **Safety**: Prohibit any state modification when encountering `:`.
