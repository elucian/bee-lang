# BEE COMPILER MANIFEST

## Current Status
- **Current Phase:** Architectural Specification Review & Compiler Formalization
- **Active Task:** Formalizing `spec/03-rules.md`
- **Last Updated:** 2026-08-26

## Specification Review Progress
- [x] `spec/00-memory-model.md`: Formal 3-tier memory model (ARC + Region Arenas + String GC), EBNF & Diagnostics (`E0401`-`E0404`).
- [x] `spec/01-lexical-structure.md`: UTF-8, Maximal Munch disambiguation, string interpolations `#(expr)`, markup DSL blocks, Unicode identifiers, and Diagnostics (`E0101`-`E0105`).
- [x] `spec/02-statements.md`: Declarations (`new`), mutations (`let`/`alter`), clone (`::`), control flow (`if`/`match`/`cycle`/`while`/`for`), error `trial`, and block alignment (`E0201`-`E0205`).
- [/] `spec/03-rules.md`: Rule definitions, parameter modes, pre/postconditions (`require`/`ensure`).
- [ ] `spec/04-structure.md`: Module architecture, qualification, `#using`/`#define`/`use`.
- [ ] `spec/05-types.md`: Primitive mathematical types (`Z`, `N`, `R`, `Q`, `C`), fixed precision, custom types.
- [ ] `spec/06-objects.md`: OOP structure, traits, encapsulation, method dispatch.
- [ ] `spec/07-functions.md`: Lambda expressions `λ`, pure functions, higher-order routines.
- [ ] `spec/10-collections.md`: Arrays, Lists, Maps, Sets, 1-based indexing, set algebra (`∩`, `∪`, `\`).
- [ ] `spec/11-processing.md`: Quantifiers (`∀`, `∃`), pipelines, map-reduce operations.
- [ ] `spec/12-concurrency.md`: Coroutines, `begin`/`wait` parallel task execution, worker channels.
- [ ] `spec/13-graphics.md`: 2D geometry primitives (`∠`, `⊡`, `◷`).
- [ ] `spec/14-library.md`: System standard library API.

## Compiler & Toolchain Implementation Plan
1. [x] Index project and ingest documentation (`/manual`).
2. [x] Establish solution & issue tracking (`/issues`, `/solution`).
3. [x] Level 1 - Level 5 test runner integration (`test/test_runner.py`).
4. [x] Lexer implementation (`internal/lexer`).
5. [x] Parser implementation (`internal/parser`).
6. [ ] AST Evaluator enhancement (`internal/evaluator`).
7. [ ] LLVM IR Native Lowering (`internal/compiler`).
