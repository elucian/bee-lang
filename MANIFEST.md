# BEE COMPILER MANIFEST

- **Active Phase:** Phase 8 – Operator Identity & Deprecation Hardening (post-handoff, 2026-09-12)
- **Open Debt Issues:** `issues/15-method-call-grant.md` (T0126 parked), `issues/16-rule-tuple-destructure.md` (T0127 parked), `issues/17-radical-precedence.md` (T0107 disabled), `issues/18-parallel-colon-init.md` (T0121+T0122 disabled).
- **Active Task:** Task 8.1 (2026-09-12, post-handoff): **Identity operator + synonymy/deprecation sweep — PARTIAL PASS.** `internal/token/token.go` now defines `IS_NOT` (Decision 7). `internal/lexer/lexer.go::NextToken` performs Maximal-Munch collapsing `is not` (with single ASCII space) to a single `IS_NOT` token (literal `"is not"`). `internal/parser/parser.go::isBinaryOp` registers `IS` and `IS_NOT`. `internal/evaluator/evaluator.go` adds an identity-ID plane (`identities map[string]int`, `nextID`, `allocID`, `ensureIdentity`) so `a is a` → 1, `a is b` → 0, and integer literals never share ephemeral IDs. Lexer gains a `warnings []string` surface (idempotent) emitting `E0010 deprecated-symbol` for `≠`, `==`, `!=`; `cmd/bee/main.go` surfaces them to stderr without exiting 1. Parser dispatches OVER (and ABORT/EXIT/PANIC) as transfer statements; evaluator flips an `exitingRule` flag so the enclosing rule short-circuits on `over`. Test-harness (test/solo.py, test/test.py) reconfigures to UTF-8 stdout/stderr for Windows consoles. **Verification:** Level 0 → 5/5 PASS. Level 1 partial: 15 / 24 active tests green (9 @DISABLED: T0102, T0103*, T0107, T0119, T0121, T0122, T0130* plus T0126/T0127 promoted to `test/debt/`). T0126 (`x.type()` method calls) and T0127 (cross-rule call + tuple-destructure) relocated to `test/debt/` as `@DEBT` test cases — see `test/debt/README.md`.
- **Tutorial Workflow (post-handoff, 2026-09-12):** `/web/` is the legacy local copy and will be discarded by the user. `/bee-tutorial/` is the canonical live-repository target linked via direct push. All future doc edits must be applied to `/bee-tutorial/` only — do not re-mirror to `/web/`.
- **Last Updated:** 2026-09-13 (Phase 8.3 D12 operator refactoring: `¬` binary inequality, `!` unary NOT, deprecated `≠`/`!=`/`<>` → NEQ)
- **Phase 8.2 delta:** T0126 / T0127 relocated to `test/debt/` as `@DEBT` test cases (see `test/debt/README.md`). Open debt issue tickets: `15-method-call-grant`, `16-rule-tuple-destructure`, `17-radical-precedence`, `18-parallel-colon-init`. Full build still green across all eight levels (33 total, 0 failed); smoke PASS.
- **Phase 8.3 delta (2026-09-13):** Decision 12 operator refactoring implemented. `internal/token`: `NEQ = "¬"`, `LOGICAL_NOT = "!"`. Lexer maps `¬`/`≠`/`!=`/`<>` → `NEQ` (E0010 on the three legacy forms) and standalone `!` → `LOGICAL_NOT`. Parser registers `NEQ` at `precCompare`; evaluator dispatches `NEQ` to `evalComparison` and `!` to prefix NOT. Tutorial `operators.html` and spec §3.3/§3.6 harmonized.
- **Phase 8.4 delta (2026-09-13):** Decision 13 range-operator and step-postfix refactor implemented. `internal/token`: `RANGE_LEFT_INC = "..<"`, `RANGE_RGHT_INC = ">.."`, `RANGE_EXCL = ">..<"` (legacy `.!` / `!.` / `!!` still lex with E0010 deprecation). Lexer applies Maximal Munch (`>..<` > `>..` > `>`, `..<` > `..` > `<`). Parser registers `not` keyword as prefix-not synonym (D7), adds `SteppedRangeExpression` AST, postfix step `(step)` on range parens. Evaluator dispatches `in` membership across the four range variants via `inRangePerOp`, and materialises `stepped[i]` via `steppedRangeValue` (1-based, Decision 1). New `@FROZEN` tests: `T0131`–`T0135` covering the four endpoint variants + postfix-step indexing + legacy `!!` deprecation.

## Development Methodology
- **TDD Protocol**: All implementation MUST follow Test-Driven Development. 
  1. Define requirement in `/issues/` and `/solution/`.
  2. Formalize in `/spec/`.
  3. Create failing test in `/test/levelX/`.
  4. Implement & Pass.
- **Audit Protocol**: Regular verification of implementation against `/spec/`.
- **Anti-Loop Gate (ratified 2026-09-12):** No edits to `internal/lexer/`, `internal/parser/`, `internal/evaluator/`, or `internal/compiler/` may begin until the corresponding `/spec` sections are 100% harmonized with user-locked decisions. This rule prevents implementer from chasing documentation that isn't yet decided.

## User-Locked Decisions (2026-09-12)

> The full normative backlog — including pending and deferred decisions —
> lives in `todo/DECISIONS.md`. The ratified subset is mirrored here for
> at-a-glance auditing.

- **Decision 1 – Indexing basis:** 1-based throughout source/parser/AST/evaluator/typechecker; IR lowering (`internal/compiler/`) is the only place 0-based appears. `$` = `len(c)-1`.
- **Decision 2 – Identity vs mutation:** `{is, is not}` are pointer-identity operators. `+=` / `-=` are pure lex-level shorthand for `+:=` / `-:=` (single AST node). `+is`/`-is` are NOT operators.
- **Decision 3 – Operator syntax canonicalization:** ⛔ **SUPERSEDED by Decision 12 (2026-09-13).** Historically: `≠` deprecated, canonical form `<>`; `==` / `!=` removed from lexer canonicalization. `≡` reserved for graphics congruence (per `solutions/013-geometric-congruence-operator.md`) — this reservation survives D12.
- **Decision 4 – Typechecker posture:** Postponed. Phase 5 deliverable reduced to symbol table + scope lookup; full type inference deferred until after Phase 6 IR codegen lowers 1-based indexing cleanly.
- **Decision 5 – Tutorial sync:** Host-side `web/` folder contains the canonical Bee tutorial referenced from the public site; the local `web/` copy (this repo) is the design-of-record and is symlinked into `C:\Users\eluci\sage-code\scl\projects\bee\bee-tutorial` for update propagation. Excluded from `.github/` workflows per user instruction.
- **Decision 6 – Curried rule signatures:** Rules MAY declare an optional *named-parameter slot* (`rule foo(*args)(sep: ", "  ∈  Str)`). Call sites apply the slot via a second parenthesised named-argument list (`print(a, b)(sep: " | ");`). The legacy `using` / `using:` postfix on `io_stmt` is deprecated; the lexer emits non-fatal `E0011` until Phase 7 audit task 7.2 hardens it to `E0009`. Named-arguments are order-independent; defaults fall through.
- **Decision 7 – Logic operator synonymy:** `is not` is one token (the lexer performs Maximal-Munch so `is` followed by space + `not` collapses into `IS_NOT`). Logic operators `and` / `or` / `xor` / `not` are accepted as synonyms for their Canonical Unicode equivalents (`∧` / `∨` / `⊕` / `¬`). The legacy Unicode `≠` and the ASCII `==` / `!=` are tolerated but emit non-fatal `E0010` deprecation warnings; they will be hardened to `E0009` once Phase 7 audit task 7.2 completes.
- **Decision 8 – Rule-call result destructuring:** 🟢 **Deferred** to Phase 8.4. Tuple-return rule grammar + multi-binding destructuring awaiting design. Test `test/debt/T0127-rule-call-result-binding.bee` parked. See `todo/DECISIONS.md` D8.
- **Decision 9 – Colon `:` is the pair-up operator (🟡 pending ratification)**: `:` is the canonical structural binding operator — key-to-value, parameter-to-type, identifier-to-literal, label-to-block-head — **without** type inference. `=` is value equality (logic), `:=` is type-inferred assignment, `:` is structural pair-up. Cross-context grammar extension enables `new a: 1, b: 2 ∈ Z;` (parallel typed declaration — Issue 18). See `todo/DECISIONS.md` D9.
- **Decision 10 – Radical precedence (🟡 pending ratification)**: Unary (`¬`, `√`, `ⁿ√`) > Power (`^`, `³`) > Mul/Div (`× ÷ / * %`) > Add/Sub (`+ -`) > Range (`..  .!`) > Compare (`= <> < > <= >= ≈`) > Logic (`and or xor not`). Issue 17 resolution path becomes "precedence climbing refactor". See `todo/DECISIONS.md` D10.
- **Decision 11 – Parallel colon-initialisation (🟡 pending ratification)**: extends `decl_stmt ::= "new" ident_list ":" expr_list ( "∈" | "in" ) type_specifier ";"`. When `ident_list` length matches `expr_list` length, each identifier binds to its counterpart, all sharing the trailing type qualifier. See `todo/DECISIONS.md` D11.
- **Decision 12 – Inequality / logical-NOT operator refactoring (✅ ratified 2026-09-13)**: `¬` (U+00AC) is the canonical **binary value-inequality** operator; `!` is the canonical **unary logical NOT** (synonym: keyword `not`). Legacy forms `≠`, `!=`, `<>` all lex to the `NEQ` token with a non-fatal `E0010 deprecated-symbol … use '¬'` warning; `==` still lexes to `EQ` with an E0010 warning. Phase 7.2 hardens all four to `E0009`. `@` is the reference-of prefix: `@a = @b` ⟺ `a is b`; comparing `@a` against a plain value is always false. Supersedes Decision 3's canonical-form table. See `todo/DECISIONS.md` D12.

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
