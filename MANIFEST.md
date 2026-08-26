# BEE COMPILER MANIFEST

## Current Status
- **Current Phase:** Architectural Specification Review & Compiler Formalization
- **Active Task:** Formalizing `spec/11-processing.md`
- **Last Updated:** 2026-08-26

## Specification Review Progress
- [x] `spec/00-memory-model.md`: Formal 3-tier memory model (ARC + Region Arenas + String GC), EBNF & Diagnostics (`E0401`-`E0404`).
- [x] `spec/01-lexical-structure.md`: UTF-8, Maximal Munch disambiguation, string interpolations `#(expr)`, markup DSL blocks, Unicode identifiers, and Diagnostics (`E0101`-`E0105`).
- [x] `spec/02-statements.md`: Declarations (`new`), mutations (`let`/`alter`), clone (`::`), control flow (`if`/`match`/`cycle`/`while`/`for`), error `trial`, and block alignment (`E0201`-`E0205`).
- [x] `spec/03-rules.md`: Rule definitions, parameters, variadics (`*`), named results, contracts (`require`/`ensure`), TCO, closures, forward declarations, and Diagnostics (`E0301`-`E0306`).
- [x] `spec/04-structure.md`: Module architecture, secondary/library modules, system path variables (`$`), public (`.`) vs private visibility, `with` blocks, `begin`/`wait` thread barriers, and Diagnostics (`E0401`-`E0406`).
- [x] `spec/05-types.md`: Single-letter mathematical types (`Z`, `N`, `R`, `Q`, `C`), fixed-point $Qm.n$ rationals, ranges/domains, subtyping (`<:`), approximate equality (`≈`), type cast (`:>`), universal `entity.type()`, and Diagnostics (`E0501`-`E0506`).
- [x] `spec/06-objects.md`: Universal Entity Model (`entity.type()`), constructors (`self`), public (`.`) vs private properties, inheritance (`<:`), `super.`, abstract methods, traits (`+`), and Diagnostics (`E0601`-`E0606`).
- [x] `spec/07-functions.md`: Pure lambda expressions (`λ`), callback shorthand, purity invariants, lambda type `L`, expression maps, SIMD/GPU vectorization, and Diagnostics (`E0701`-`E0705`).
- [x] `spec/10-collections.md`: Lists `()`, Arrays `[]`, Matrices `[](r, c)`, Sets `{}` (with set algebra `∩`, `∪`, `\`), Maps `{:}` (with keys), Ordinals, 1-based indexing, `$` end anchor, and Diagnostics (`E1001`-`E1006`).
- [/] `spec/11-processing.md`: Quantifiers (`∀`, `∃`), pipelines, map-reduce operations.
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
