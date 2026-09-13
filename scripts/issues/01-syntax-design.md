# Language Specification Issues

- [x] **Ambiguous Block Terminations:** Resolved in `spec/02-statements.md`. Block terminators are strictly defined: `done` for `if`/`with`/`match`/`trial`, `repeat` for `cycle`/`while`/`for`, and `return` for `rule`/routines.
- [x] **`match` keyword vs `if-else` ladder:** Formalized in `spec/02-statements.md`. `if-else` is used for binary/boolean branches; `match` supports `first`, `every`, and `total` matching modes.
- [x] **Formal Grammar Definition:** EBNF grammar fully defined in `spec/01-lexical-structure.md` and `spec/02-statements.md`.
- [x] **Indexing Basis (Decision 1, user-confirmed 2026-09-12):** Bee source language uses **1-based indexing** with `$` as the dynamic end-anchor (last-element index). All compiler passes (lexer, parser, AST, typechecker, evaluator) operate in 1-based coordinates. The 1-based-to-0-based translation is performed **once**, at the IR boundary in `internal/compiler/`, using the rule `ir_index = source_index - 1` and `ir_index($) = len(collection) - 1`. Ratified in `GEMINI.md` §4 (Array Semantics + IR Conversion). Downstream: `MANIFEST.md` Task 5.2 closure (done), `spec/10-collections.md` §2.1 confirmed, `test/level4/T0403…T0408` reactivation review pending.
- [x] **Modifier shorthands + identity operators (Decision 2, 2026-09-12):** `+=` and `-=` are pure lex-level shortcuts normalized to `+:=` / `-:=` (single AST node produced). The forms `+is` / `-is` are NOT operators — `is` is a closed-class relational token. Identity operators `{is, is not}` compare pointer identity, NOT value equality: `a is b` is false across distinct allocations even when `a = b`. `spec/01-lexical-structure.md` §3.3 added.
- [x] **Operator canonicalization (Decision 3, 2026-09-12):** Unicode `≠` is deprecated due to canonical-equivalence defects in non-UTF-8-tokenizer code paths. Canonical form is `<>`. The lexer emits diagnostic `E0010 deprecated-symbol: '≠' — use '<>'` until Phase 7 task 7.2 migrates authors; thereafter it converts to hard `E0009`. Tutorial is the source-of-truth for operator semantics; spec follows tutorial; implementation follows spec. `spec/01-lexical-structure.md` §2 paragraph 2 + §3.3 added.
- [x] **Spec harmonization follow-ups (Decisions 2-3) — RESOLVED 2026-09-12:**
  1. ~~EBNF `assign_op` "sugar"~~ → resolved: `spec/01-lexical-structure.md` §5 now declares `mutate_op ::= "+:=" | "-:="` and `mutate_op_sugar ::= "+=" | "-="` as the canonical lex-level shorthand alias. `spec/02-statements.md` §5 mirrors via cross-reference.
  2. ~~EBNF `cmp_val_op` enumerating `is`/`is not`~~ → resolved: `spec/01-lexical-structure.md` §5 now declares `cmp_val_op ::= "=" | "<>"` and `cmp_ref_op ::= "is" | "is not"`. `spec/02-statements.md` §5 points to these via the "canonical home" comment.
  3. ~~Remove `==`/`!=` from `cmp_ref_op`~~ → resolved: `spec/01-lexical-structure.md` §5 line 182 explicitly notes that `==` / `!=` / `≡` are **reserved for future graphics congruence** and are NOT value-ops.
  4. ~~Sweep `≠` from normative examples~~ → resolved: `spec/02-statements.md:90` and `spec/03-rules.md:84/80` now use `<>` as the canonical inequality. Residual `≠` glyphs in those files appear ONLY in deprecation-notice comments, never in normative code blocks.

## Spec Audit Status (after 2026-09-12 sweep)

| Spec file | Status | Notes |
| :--- | :---: | :--- |
| `spec/01-lexical-structure.md` | ✓ Harmonized | EBNF canonical home for `cmp_val_op`, `cmp_ref_op`, `mutate_op_sugar`; `≠`/`==`/`!=` reserved/legacy explained |
| `spec/02-statements.md` | ✓ Harmonized | EBNF points to canonical home; normative examples use `<>` |
| `spec/03-rules.md` | ✓ Harmonized | `Contract(divide)` display math now expresses inequality as `b <> 0`; code example updated; deprecation annotation retained in code comment + new boxed note |
| `spec/00-memory-model.md` | ✓ Compliant | Pseudocode `==` is C-ARC algorithm, not Bee syntax |
| `spec/04-07`, `spec/10-14` | ✓ Harmonized | spec/00, 04, 05, 06, 07, 10, 11, 12, 13, 14 verified 2026-09-12; zero normative `≠`, `==`, `!=` references. Only intentional `== 0` ARC pseudocode in `spec/00 §2.1` remains (out of Bee grammar scope). |

## Implementation Drift (gate for Phase 5/6)

Even though `/spec` is now harmonized, the Go lexer at `internal/lexer/lexer.go` still emits deprecated operators and lacks support for Decisions 2–3:
- Lines 189, 196-198, 228: `≠`, `==`, `!=` are tokenized (no deprecation diagnostic, no canonicalization).
- The `is` / `is not` / `+:=` / `-:=` tokens do not exist in `internal/token/token.go`.
- The evaluator at `internal/evaluator/comparison.go` and `internal/evaluator/evaluator.go` still treats `==` (literal form) as value equality and accepts `!=` / `<>` / `≠` interchangeably.

This drift is tracked in `issues/02-implementation-compiler.md` and **MUST** be resolved before Phase 5 Task 5.1 or any test-level validation can succeed honestly. Records `T0001.bee` PASS as a **false positive** caused by legacy parser logic permissive enough to accept `fn main()`-style headers — will require test-file normalization per frozen TDD protocol.
