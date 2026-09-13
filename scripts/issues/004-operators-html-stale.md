# Issue: Ambiguous Colon (:) Operator in Operators Documentation
The current documentation for the `:` operator in `web/operators.html` is ambiguous, listing it both as a block initiator and a pair-up operator, without defining it as a structural binding operator.

## Impact
- Developers may confuse `:` (structural binding) with `:=` (state mutation).
- Inconsistent terminology between Specification and Public Documentation.

## Requirement
- Clarify in the specification that `:` is the structural binding (pair-up) operator.
- Clearly delineate its scope (parameters, map literals) versus mutation operators.
