# Solution: Formalize Rule State Mutability via Boxing
Formally integrate the box operator `[]` into the `spec/` to define the mutability invariant.

## Design
1. **Definition**: Boxing (`[x]`) allocates a value on the heap, allowing it to be treated as a mutable reference within a rule closure.
2. **Grammar Update**: Explicitly mention the `[]` operator in `spec/03-rules.md` (Section 5.3) and `spec/05-types.md`.
3. **Safety**: Enforce that unboxed variables in closures are constant/immutable by default.
