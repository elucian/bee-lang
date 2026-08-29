# Issue: Equality Operator and Initialization Syntax Ambiguity

The current use of `=` for value comparison is overloaded, creating ambiguity between equality checks, initialization, and assignment. Furthermore, the use of `:` for pair-up initialization is inconsistent with Bee's assignment philosophy.

## Impact
- **Operator Overload:** `=` performs both value equality checks (in `expect`) and is implied in assignments, confusing the language grammar.
- **Initialization Inconsistency:** `:` serves as a structural pair-up operator, but using it for initialization obscures whether a value is an assignment or a structural mapping.
- **Identity Clarity:** Bee lacks a distinct operator for identity checking versus value comparison.

## Proposed Resolution
1. **Redefine Equality Operators:**
   - Change `==` to be the exclusive value comparison operator (e.g., `expect x == y`).
   - Introduce `is` as the identity check operator (e.g., `x is y`).
2. **Standardize Initialization:**
   - Replace the structural `:` operator with `=` for initialization in declarations to avoid inference-based confusion.
   - Example: `new xo, yo, zo ∈ Z = 10;` instead of `new xo, yo, zo ∈ Z : 10;`.
3. **Grammar Impact:**
   - `assignment_op` in mutations will remain `:=` for type inference or update semantics.
   - `comparison_op` will use `==` and `is`.
