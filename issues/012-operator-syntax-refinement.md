# Issue: Operator Syntax Refinement (Inequality and Logic)

The current specification lacks formalization for inequality and logical operators, relying on ad-hoc symbols. We must align with Bee's design goals: using readable keywords for logic and standard mathematical/structural symbols for relations.

## Impact
- **Symbolic Inconsistency:** Mixing `!=` (common but non-Bee) with potential `≠` or `<>` creates parsing confusion.
- **Logical Verbosity:** The language lacks a standardized set of logical keywords (`and`, `or`, `xor`, `not`).
- **Range Ambiguity:** The `!` operator is currently overloading range notation, necessitating a clear definition.

## Requirements
1. **Inequality Operator:** Replace all forms of "not equals" with `<>`.
2. **Identity Negation:** Formalize `a is not b` instead of `!=`.
3. **Logical Keywords:** Standardize `and`, `or`, `xor`, `not` as the official logical operators.
4. **Range Notation:** Explicitly reserve `!` for range-based notation (e.g., `1!10`).
5. **Spec Update:** Update `spec/02-statements.md` grammar and documentation.
