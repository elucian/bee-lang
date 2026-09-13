# Technical Debt Tests (`test/debt/`)

This folder contains `.bee` test cases that **represent known unimplemented features** or unresolved grammatical gaps in the Bee compiler. Tests are moved here by a human decision (not auto-disabled by `test/test.py`); they are **excluded from the active test levels** (level0–level8) so the build can stay green while the corresponding features are still in design or upstream-spec limbo.

## Lifecycle

1. **Promoted from `test/levelX/`.** A passing level test (PASS) is never moved here. A failing test that surfaces a grammar gap or unimplemented feature MAY be moved here manually once a `solution/` note documents the intended resolution.
2. **Read-only ground truth.** `@FROZEN` headers are preserved. The `@DISABLED` marker is replaced with `@DEBT` to signal "parked, awaiting design decision".
3. **Re-promotion.** When the underlying feature is implemented, the test is moved back to the active level and the `@DEBT` marker is removed.

## Current Inventory

| CASE | Topic                                  | Spec Reference                          | Status                                                       |
| ---- | -------------------------------------- | --------------------------------------- | ------------------------------------------------------------ |
| T0126 | Method-call dispatch (`x.type()`)     | `spec/07-functions.md` §6 (lambda_call) | Awaiting decision: extend `Expression` AST with `MethodCall`; bind to spec/06-objects.md method table. |
| T0127 | Rule call → parallel destructuring     | `spec/03-rules.md` §3.1 (call sites)    | Awaiting decision: tuple-return rule grammar; multi-binding destructuring. |

## Conventions

* Filename prefix `T0XYZ-` preserves the original level test ID for traceability.
* Each file keeps its original `-- @DESC:` header so the issue tracker can grep for context.
* Header `<!-- @DEBT: <reason> -->` is mandatory on line 2 (replaces `@DISABLED`).
* These tests are NOT executed by `sh run.sh test` — they require an explicit `sh run.sh solo <name>` from `test/debt/`.
