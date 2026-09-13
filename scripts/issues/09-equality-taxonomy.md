# Issue: Operator Taxonomy for Equality and Identity

## Description
The current evaluator implementation treats `=`, `==`, and `token.EQ` interchangeably as a single equality check. This contradicts the requested operator semantics where:
- `=` (equals): Logic/Value Comparison (Deep equality).
- `==` (double-equals): Reference/Identity Comparison (Shallow equality).
- `≠` (not-equals): Value inequality.
- `!=` (bang-equals): Reference inequality.
- `≡` (equivalence): Equivalent to `==` (Reference equality).
- `<>` (not-equivalent): Equivalent to `≠` (Value inequality).

## Solution
1. Update `internal/parser` to distinguish between value-comparison tokens and reference-comparison tokens.
2. Update `internal/evaluator` to implement separate logic for value-based comparisons vs. reference-based comparisons.
3. Align `spec/01-lexical-structure.md` to formalize this distinction.
4. Formalize `≡` as an alias for `==` and `<>` as an alias for `≠`.
