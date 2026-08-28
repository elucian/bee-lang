# Solution: Synchronize Web Documentation
Update `web/syntax.html` to reflect the binding vs. assignment operator definitions.

## Design
1. **Audit**: Locate the operator table in `web/syntax.html`.
2. **Implementation**:
    - Add/update the binding operator `:` to the operator table (Category: Structural Binding).
    - Update `:=` description to strictly "State Mutation/Assignment".
    - Mark the operator table as "Reflecting Spec v0.1.2".
3. **Future**: Design a simple script in `/scripts/` to generate `web/syntax.html` from `spec/` files to automate sync.
