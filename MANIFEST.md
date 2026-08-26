# BEE COMPILER MANIFEST

## Current Status
- **Current Phase:** Specification Phase Complete / Architectural Baseline Finalized
- **Active Task:** Final Architectural Review & Preparation for Compiler Implementation Phase
- **Last Updated:** 2026-08-26

## Formal Specification Baseline (`/spec`)
- [x] `spec/00-memory-model.md`: Formal 3-tier memory model (ARC + Region Arenas + String GC), EBNF & Diagnostics (`E0401`-`E0404`).
- [x] `spec/01-lexical-structure.md`: UTF-8, Maximal Munch disambiguation, string interpolations `#(expr)`, markup DSL blocks, Unicode identifiers, and Diagnostics (`E0101`-`E0105`).
- [x] `spec/02-statements.md`: Declarations (`new`), mutations (`let`/`alter`), clone (`::`), control flow (`if`/`match`/`cycle`/`while`/`for`), error `trial`, and block alignment (`E0201`-`E0205`).
- [x] `spec/03-rules.md`: Rule definitions, parameters, variadics (`*`), named results, contracts (`require`/`ensure`), TCO, closures, forward declarations, and Diagnostics (`E0301`-`E0306`).
- [x] `spec/04-structure.md`: Module architecture, secondary/library modules, system path variables (`$`), public (`.`) vs private visibility, `with` blocks, `begin`/`wait` thread barriers, and Diagnostics (`E0401`-`E0406`).
- [x] `spec/05-types.md`: Single-letter mathematical types (`Z`, `N`, `R`, `Q`, `C`), fixed-point $Qm.n$ rationals, ranges/domains, subtyping (`<:`), approximate equality (`≈`), type cast (`:>`), universal `entity.type()`, and Diagnostics (`E0501`-`E0506`).
- [x] `spec/06-objects.md`: Universal Entity Model (`entity.type()`), constructors (`self`), public (`.`) vs private properties, inheritance (`<:`), `super.`, abstract methods, traits (`+`), and Diagnostics (`E0601`-`E0606`).
- [x] `spec/07-functions.md`: Pure lambda expressions (`λ`), callback shorthand, purity invariants, lambda type `L`, expression maps, SIMD/GPU vectorization, and Diagnostics (`E0701`-`E0705`).
- [x] `spec/10-collections.md`: Lists `()`, Arrays `[]`, Matrices `[](r, c)`, Sets `{}` (with set algebra `∩`, `∪`, `\`), Maps `{:}` (with keys), Ordinals, 1-based indexing, `$` end anchor, and Diagnostics (`E1001`-`E1006`).
- [x] `spec/11-processing.md`: Primitive boxing `[x]`, unboxing `Type(boxed)`, quantifiers (`∀`, `∃`), pipelines (`>>`), map-reduce, deconstruction (`*`), matrix row/col slicing (`M[1, *]`), deep clone (`::`), and Diagnostics (`E1101`-`E1106`). Recaptured original `manual/11-processing.md` from `web/processing.html`.
- [x] `spec/12-concurrency.md`: Multithreading (`begin`/`wait`), reduction channels (`+>`), coroutines (`yield`), channel extraction (`<<`), worker exception isolation, and Diagnostics (`E1201`-`E1205`). Updated all `/demo/concurrency` examples to modern syntax.
- [x] `spec/13-graphics.md`: Angular type `G` (`°`, `′`, `″`), 2D geometry primitives (`CRT`, `POL`, `VEC`, `CRC`, `SQR`, `PLG`), scene graph (`Canvas`, `Layer`, `Shape`, `Label`), drawing commands (`draw`, `wipe`, `show`/`hide`), and Diagnostics (`E1301`-`E1305`).
- [x] `spec/14-library.md`: System standard library API (`$bee.sys`), tree-shaking static linkage, file handles (`F`), `$error` codes (`1..199` vs `200+`), and Diagnostics (`E1401`-`E1405`).

## Compiler & Toolchain Implementation Roadmap
1. [x] Ingest documentation & harmonize `/manual` with `/spec`.
2. [x] Complete formal specifications across all `/spec` modules (`00` through `14`).
3. [x] Update all `/demo/concurrency` files to modern language conventions.
4. [x] Level 1 - Level 5 test runner integration (`test/test_runner.py`).
5. [ ] Lexer implementation (`internal/lexer`) alignment with new tokens and Maximal Munch rules.
6. [ ] Parser implementation (`internal/parser`) alignment with EBNF grammars.
7. [ ] AST Evaluator enhancement (`internal/evaluator`).
8. [ ] LLVM IR Native Lowering (`internal/compiler`).
