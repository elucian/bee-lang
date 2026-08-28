# Issue: Type Inference Invariant Undefined
The specification lacks a formal definition of Bee's type inference mechanism, specifically the "first-assignment" binding rule and the handling of divergent type assignments through variant (union) types.

## Impact
- Compiler implementation lacks a unified semantic rule for type propagation.
- Divergent assignments in different control flow paths are currently not formalized.
- Developers lack a clear definition of Bee's support for union/variant types generated via inference.

## Requirements
- Define the "First-Assignment Binding" rule.
- Define the behavior for divergent assignments in control flow paths (Variant/Union type promotion).
- Update the type system specification to reflect gradual typing and inference invariants.
