# Solution: Formalize Geometric Congruence Operator (≡)

Define the `≡` operator exclusively for geometric congruence (similarity) in the Bee language's geometric processing system.

## Design

1. **Semantic Definition:**
   - Define `A ≡ B` as "A is congruent (similar) to B."
   - Criteria: Equal number of sides, equal interior/exterior angles, and same geometric shape class, regardless of scale (size).

2. **Grammar & Lexer:**
   - Ensure the lexer treats `≡` as a distinct `TOKEN_GEOMETRIC_CONGRUENCE`.

3. **Specification Updates:**
   - Update `spec/13-graphics.md` (or the relevant geometry specification) to include the formal definition of `≡`.
   - Update `spec/02-statements.md` to note that `≡` is reserved for geometric congruence in graphics/geometry contexts.
