# Bee Compiler Specification Index (spec/readme.md)

This directory contains the machine-parsable grammar and operational semantics.

## Error Code Registry

- `errors.json`: Machine-parsable registry of every diagnostic error code (all modules). It is the single source of truth for code uniqueness. **When adding a new error code to any spec module, register it in `errors.json` (bumping to an unused code on collision) in the same change set.**

## Module Map
- `00-memory-model.md`: Region-based management, `zap` usage, thread-safety boundaries.
- `01-lexical-structure.md`: Maximal Munch rules, Unicode ranges, operator tokenization, markup literals.
- `02-statements.md`: Assignment, declaration, control flow (if/cycle/match), and transfer statements.
- `03-rules.md`: Rule anatomy, signature declaration, and execution rules.
- `04-structure.md`: Module architecture, program structure, and dependency graph.
- `05-types.md`: Primitive definitions (B, A, U, Q, N, Z, R, S), promotion table.
- `06-objects.md`: Constructor rules, traits, inheritance, encapsulation.
- `07-functions.md`: Functions and lambda expressions, purity restrictions, callback semantics.
- `10-collections.md`: Lists, Arrays, Matrices (row-major), Sets, Hash Maps.
- `11-processing.md`: Slicing, spreading, string interpolation, markup literals.
- `12-concurrency.md`: Multi-threading (`begin`/`wait`), Coroutines (`yield`).
- `13-graphics.md`: Cartesian coordinate system, geometric primitives.
- `14-library.md`: Standard API, Error objects, System I/O.
