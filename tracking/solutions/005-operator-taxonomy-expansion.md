# Solution: Expand Operator Taxonomy in Specification
Systematically categorize and document all missing operators identified in the implementation.

## Design
1. **Taxonomy Expansion**:
   - Numeric Modifiers: `+=`, `-=`, `*=`, `/=`, `^=`, `√=`, `%=`.
   - Collection/Concurrency: `+>`, `<+`, `++`, `-=`.
   - Arithmetic: `√` (Radical), `^` (Power).
2. **Grammar Update**: Add these to the `assign_op` and generic `op` definitions in `spec/01-lexical-structure.md`.
3. **Documentation**: Ensure formal definition of arity and usage context for each.
