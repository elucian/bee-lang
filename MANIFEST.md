# BEE COMPILER MANIFEST

- **Active Phase:** Phase 7 – Specification Audit & TDD Validation (re-opened to harmonize `/spec` after Decisions 1–6)
- **Active Task:** Task 7.7 (2026-09-12, post-handoff): **Parser implementation — `rule main:` grammar & curried `print(...)(sep: ...)` — PASS.** Identity vs warnings split in `internal/parser/parser.go`: `p.errors []string` reserved for hard `E0009` SyntaxErrors, `p.warnings []string` introduced for soft `W0901` stubs (signature grammar heuristic, decision-6 phase 7.2 gate) and `E0011 DeprecatedSymbol 'using'`. `cmd/bee/main.go` now distinguishes the two (exits 1 only on `p.Errors()`, surfaces `p.Warnings()` to stderr). `parseRuleEntry` rewritten to consume the canonical signature `rule identifier (param_list) [(named_param_list)] [=> (result_list)] :`, parse indented body statements, and emit aligned `return;` terminator (spec/03 §2.1 / §2.3 / §5.2; spec/02 §6.3). `RuleStatement` AST node extended with `Body *BlockStatement`, `ForwardDecl bool`, `SignatureGrammarPartial bool`. `parsePrintStatement` extended with the Decision 6 curried `(name: expression [, name: expression]*)` postfix; legacy `using` postfix still tolerated with `E0011` soft warning. `eval.Eval` now descends into `RuleStatement.Body` so statements under `rule main:` actually execute. **Verification:** `sh run.sh test level0` → 5 / 5 PASS (T0001 `Hello World`, T0002 `10,20,30,40,50`, T0003 `Hello, World!`, T0004 `1 | 2`, `smoke` `Smoke test`); legacy `fn main() {}` source now hard-fails with 6 × `E0009 SyntaxError:UnrecognizedStatement` (Issue 14 silent-drop hazard closed). **Next gate:** level1 — write tests under `test/level1/` for `new` / `let` declarations, `:=` assignment, `<>` / `<` / `>` comparisons, range operators.
- **Tutorial Workflow (post-handoff, 2026-09-12):** `/web/` is the legacy local copy and will be discarded by the user. `/bee-tutorial/` is the canonical live-repository target linked via direct push. All future doc edits must be applied to `/bee-tutorial/` only — do not re-mirror to `/web/`.
- **Last Updated:** 2026-09-12 (Phase 7.7 parser implementation + Issue 14 closure)

## Development Methodology
- **TDD Protocol**: All implementation MUST follow Test-Driven Development. 
  1. Define requirement in `/issues/` and `/solution/`.
  2. Formalize in `/spec/`.
  3. Create failing test in `/test/levelX/`.
  4. Implement & Pass.
- **Audit Protocol**: Regular verification of implementation against `/spec/`.
- **Anti-Loop Gate (ratified 2026-09-12):** No edits to `internal/lexer/`, `internal/parser/`, `internal/evaluator/`, or `internal/compiler/` may begin until the corresponding `/spec` sections are 100% harmonized with user-locked decisions. This rule prevents implementer from chasing documentation that isn't yet decided.

## User-Locked Decisions (2026-09-12)
- **Decision 1 – Indexing basis:** 1-based throughout source/parser/AST/evaluator/typechecker; IR lowering (`internal/compiler/`) is the only place 0-based appears. `$` = `len(c)-1`.
- **Decision 2 – Identity vs mutation:** `{is, is not}` are pointer-identity operators. `+=` / `-=` are pure lex-level shorthand for `+:=` / `-:=` (single AST node). `+is`/`-is` are NOT operators.
- **Decision 3 – Operator syntax canonicalization:** `≠` deprecated, canonical form `<>`. `==` / `!=` removed from lexer canonicalization. `≡` reserved for graphics congruence (per `solutions/013-geometric-congruence-operator.md`). Lexer emits non-fatal `E0010` for `≠`; Phase 7.2 hardens to `E0009`.
- **Decision 4 – Typechecker posture:** Postponed. Phase 5 deliverable reduced to symbol table + scope lookup; full type inference deferred until after Phase 6 IR codegen lowers 1-based indexing cleanly.
- **Decision 5 – Tutorial sync:** Host-side `web/` folder contains the canonical Bee tutorial referenced from the public site; the local `web/` copy (this repo) is the design-of-record and is symlinked into `C:\Users\eluci\sage-code\scl\projects\bee\bee-tutorial` for update propagation. Excluded from `.github/` workflows per user instruction.
- **Decision 6 – Curried rule signatures:** Rules MAY declare an optional *named-parameter slot* (`rule foo(*args)(sep: ", "  ∈  Str)`). Call sites apply the slot via a second parenthesised named-argument list (`print(a, b)(sep: " | ");`). The legacy `using` / `using:` postfix on `io_stmt` is deprecated; the lexer emits non-fatal `E0011` until Phase 7 audit task 7.2 hardens it to `E0009`. Named-arguments are order-independent; defaults fall through.

## Roadmap & Phases

### Phase 5: Type Checker & Semantic Analysis (`internal/typechecker`)
- [ ] Task 5.0: Create `internal/typechecker/` package skeleton with minimal scope — symbol table + scope lookup only; full type inference deferred until Phase 6 IR lowers 1-based indexing cleanly. (Decision 4, 2026-09-12)
- [ ] Task 5.1: Implement symbol table, scoping rules, and minimal type resolution (declaration site only).
- [x] Task 5.2: ~~Enforce zero-based indexing validation~~ **Resolved as Decision 1, 2026-09-12**: indexing is **1-based at source/parser/AST/evaluator/typechecker**, lowered to **0-based at IR in `internal/compiler/`** with rule `ir_index = source_index - 1` and `$` → `len(c)-1`. Ratified in `GEMINI.md` §4.
- [ ] Task 5.3: Validate precondition warning (`assert`) and invariant enforcement (`expect`) contract bindings.

### Phase 6: LLVM IR Codegen Engine (`internal/codegen`)
- [ ] Task 6.1: Map AST nodes to LLVM IR module definitions.
- [ ] Task 6.2: Implement basic block generation with explicit terminators.
- [ ] Task 6.3: Implement dynamic type coercion ($Qm.n$, Unicode operators).

### Phase 7: Specification Audit & TDD Validation
- [x] Task 7.1: Audit `/spec/` vs `/web/` documentation.
- [ ] Task 7.2: Create and pass missing edge-case tests identified during audit.
- [ ] Task 7.3: Integrate automated `/spec/` to `/web/` sync scripts.
- [x] Task 7.4 (2026-09-12): **Spec harmonization with Decisions 1–6.** spec/01 ✓, spec/02 ✓, spec/03 ✓ (Decision 6 curried signatures added §2.4 + §6 EBNF extensions), spec/00 ✓, spec/04 ✓, spec/05 ✓, spec/06 ✓, spec/07 ✓, spec/10 ✓, spec/11 ✓, spec/12 ✓, spec/13 ✓, spec/14 ✓. All normative legacy operators swept; `≠` references now confined to intentional deprecation annotations in spec/01 (line 54–56), spec/02 (lines 48, 53, 90), spec/03 (lines 124, 132, 247, 256). `==` references now confined to intentional reservations in spec/01 (line 56, line 182) and pseudocode in spec/00 (line 27). All `!=` references removed. New diagnostic codes E0307/W0308/E0309/E0011 added to spec/03 §7 per Decision 6.
- [x] Task 7.5 (2026-09-12): **level0 test harmonization** — T0001 normalized to spec/03 `rule main:` syntax. T0002 / T0004 migrated to Decision 6 curried `(sep: ...)` call-site form; `test/level0/README.md` updated.
- [x] Task 7.7 (2026-09-12): **Parser implementation gate (Issue 14 closed).** `internal/parser/Parser` now separates `errors` (E0009 hard) from `warnings` (W0901 / E0011 soft). `cmd/bee/main.go` ignores warnings so the build does not exit on stubs. `parseRuleEntry` consumes signature + parses indented body + emits aligned `return;` (spec/03 §2.1). `parsePrintStatement` supports curried `(name: expr [, name: expr]*)` per Decision 6. `RuleStatement.{Body, ForwardDecl, SignatureGrammarPartial}` and `PrintStatement.NamedArgs` added to AST. `eval.Eval` descends into `RuleStatement.Body`. Verification: `sh run.sh test level0` → 5 / 5 PASS, 0 fail.

## Completed Specifications (`/spec`)
- [x] `spec/00-07`: Memory model, Lexical, Statements, Rules, Structure, Types, Objects, Functions.
- [x] `spec/10-14`: Collections, Processing, Concurrency, Graphics, System Library.

## Spec Audit Sweep Log (2026-09-12)
| File | Old `≠` instances | Old `==`/`!=` instances | Status |
| :--- | :---: | :---: | :--- |
| `spec/01-lexical-structure.md` | 0 (only doc-comment refs) | 0 (line 182 notes them as legacy) | ✓ Harmonized |
| `spec/02-statements.md` | 0 normative (deprecation note retained) | 0 | ✓ Harmonized |
| `spec/03-rules.md` | 0 normative (deprecation note retained in comment) | 0 | ✓ Harmonized (this pass) |
| `spec/00-memory-model.md` | 0 (only pseudocode `== 0` reference) | 0 | ✓ Compliant |
| `spec/04-07`, `spec/10-14` | 0 | 0 (only `== 0` pseudocode in spec/00 §2.1) | ✓ Harmonized |
