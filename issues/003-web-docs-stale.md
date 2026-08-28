# Issue: Stale Web Documentation (syntax.html)
The `web/syntax.html` file is out of sync with the updated `spec/01-lexical-structure.md`.

## Impact
- Inaccurate documentation for binding operator `:` (documented as assignment).
- Confusing developer reference for Bee syntax.
- Mismatched terminology between specification and public documentation.

## Requirements
- Audit `web/syntax.html` for operator terminology.
- Sync `web/syntax.html` tables with the updated EBNF taxonomy.
