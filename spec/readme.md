# Bee Compiler Specification Index (spec/readme.md)

This directory contains the machine-parsable grammar and operational semantics.

## Module Map
- `00-memory-model.md`: Region-based management, `zap` usage, thread-safety boundaries.
- `01-lexical-structure.md`: Maximal Munch rules, Unicode ranges, operator tokenization, markup literals.
- `02-statements.md`: Assignment, declaration, control flow (if/cycle/match), and transfer statements.
- `03-rules.md`: Rule anatomy, signature declaration, and execution rules.
- `04-functions.md`: Lambda (L) type, purity restrictions, and callback semantics.
- `05-types.md`: Primitive definitions (B, A, U, Q, N, Z, R, S), promotion table.
- `06-objects.md`: Constructor rules, traits, inheritance, encapsulation.
- `10-collections.md`: Lists, Arrays, Matrices (row-major), Sets, Hash Maps.
- `11-processing.md`: Slicing, spreading, string interpolation, markup literals.
- `12-concurrency.md`: Multi-threading (`begin`/`wait`), Coroutines (`yield`).
- `13-graphics.md`: Cartesian coordinate system, geometric primitives.
- `14-library.md`: Standard API, Error objects, System I/O.
