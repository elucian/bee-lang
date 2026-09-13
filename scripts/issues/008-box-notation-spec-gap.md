# Issue: Box Notation for Mutable Rule State Undefined
The specification lacks a formal definition of the box operator `[]` as a requirement for mutable state management within rule closures.

## Impact
- Developers may be confused about why mutable state requires boxing (e.g., `set .count := [start];`).
- Inconsistent terminology between the specification and the web documentation.

## Requirements
- Formally define the box operator `[]` in the specification as the mechanism for explicit heap-allocation of mutable state within rule scopes.
