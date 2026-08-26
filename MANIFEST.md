# BEE COMPILER MANIFEST

## Current Status
- **Current Phase:** Phase 3 – Recursive Descent Parser & AST Expansion (`internal/parser`)
- **Active Task:** Task 3.1: Implement EBNF AST nodes and recursive descent parsing rules.
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

## Issues & Architectural Solutions
- [x] `issues/08-code-beautifier.md`: Native compiler code beautifier (`-b` / `--beautify`).
- [x] `solution/12-code-beautifier.md`: In-place formatting, 2-space block alignment, EOL comment alignment, implicit multiplication auto-fix (`2(a+b)` $\rightarrow$ `2 * (a + b)`).
- [x] `solution/13-compiler-implementation-phase.md`: Compiler implementation phase tracking document.

## Compiler Implementation Phase Roadmap
1. [x] **Phase 1: Native Code Beautifier Engine (`internal/beautifier`)**
   - [x] Task 1.1: Implement `internal/beautifier/beautifier.go` with 2-space block indentation enforcement.
   - [x] Task 1.2: Implement EOL comment alignment and block keyword/colon normalization.
   - [x] Task 1.3: Implement AST auto-fix for implicit multiplication (`2(a+b)` $\rightarrow$ `2 * (a + b)`).
   - [x] Task 1.4: Integrate `-b` / `--beautify` CLI flag in `cmd/bee/main.go`.
- [x] **Phase 2: Lexer & Tokenizer Enhancement (`internal/lexer` & `internal/token`)**
   - [x] Task 2.1: Expand `internal/token/token.go` with all spec tokens (`::`, `.!`, `!.`, `!!`, `+>`, `<<`, `λ`, `∈`, `∩`, `∪`, `\`, `≈`, `≠`, etc.).
   - [x] Task 2.2: Update `internal/lexer/lexer.go` with Maximal Munch disambiguation rules for `.`, `..`, `.!`, `!.`, `!!`, `:`, `:=`, `::`, `--`, `+-`.
   - [x] Task 2.3: Add support for string interpolation `#(expr)`, raw backtick strings `` `...` ``, and embedded markup DSL blocks (`<sql>`, `<html>`).
3. [ ] **Phase 3: Recursive Descent Parser & AST Expansion (`internal/parser`)**
4. [ ] **Phase 4: AST Evaluator & Execution Engine (`internal/evaluator`)**
