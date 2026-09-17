# Solution: Refine Operator Taxonomy (Inequality, Logic, and Ranges)

Standardize the operator set to align with Ada-inspired readability for logic and standard mathematical notation for relations.

## Design

1. **Relation Operators:**
   - **Inequality:** Use `<>` for value inequality.
   - **Identity Negation:** Use `is not` for identity negation.
   - **Ranges:** Reserve `!` strictly for range boundaries (e.g., `start!end`).

2. **Logical Operators:**
   - Standardize keywords: `and`, `or`, `xor`, `not`.

3. **Grammar Impact:**
   - Update `comparison_expr` to include the new relational operators.
   - Update logical expressions.

4. **Specification Updates:**
   - Update `spec/02-statements.md` grammar section.
