# Solution: Formalize Type Inference and Variant Promotion
Integrate gradual typing rules into the specification to clarify how the compiler handles variable declarations and divergent path assignments.

## Design
1. **Binding Rule**: Formalize that the variable type is bound at the first assignment (`new x := expr;`).
2. **Variant Promotion**: When a variable is assigned different types across branches (e.g., `if` branches), promote the type to a `Variant` (Union) type (e.g., `Z | R`).
3. **Specification Update**: Add a section on "Type Inference & Variant Promotion" to `spec/05-types.md`.
