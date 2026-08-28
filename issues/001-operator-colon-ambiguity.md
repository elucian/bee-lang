# Issue: Underspecified Operator Semantics
The `:` operator is currently ambiguously grouped under "assignment" in documentation, while functioning as a structural "pair-up" binding operator.

## Impact
- Inconsistent parsing of key-value pairs vs. state mutation.
- Potential for invalid memory operations if the parser treats `:` as `:=`.
- Confusion in generated code for data structures.

## Requirement
- Explicitly define `:` as a binding operator in the specification.
- Separate its grammar rules from assignment/mutation (`:=`).
