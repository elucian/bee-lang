# Solution: Formalize Colon (:) as Binding Operator in Specification
Define `:` as a non-mutation operator to be incorporated into the next revision of `spec/01-lexical-structure.md` and subsequent documentation updates.

## Design
1. **Definition**: `:` is defined as the "Structural Binding Operator."
2. **Grammar**: Strictly distinct from `:=` (Assignment) and `::` (Clone).
3. **Specification Alignment**: Update `spec/01-lexical-structure.md` to explicitly forbid `:` from modifying state.
4. **Documentation Sync**: Future manual updates to `web/operators.html` will align with this specification.
