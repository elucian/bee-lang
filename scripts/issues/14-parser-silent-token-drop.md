# Issue 14: Parser Silently Drops Unrecognized Tokens — False PASS Risk

**Status:** Resolved (Phase 7.7, 2026-09-12)
**Detected:** 2026-09-12 — Phase 7 TDD Audit, level0 harmonization pass
**Spec Reference:** `spec/03-rules.md` §1, `spec/02-statements.md` §1 (exhaustive statement taxonomy)

## Problem

`internal/parser/parser.go::parseStatement` (line 88) only handles a closed switch over a small keyword set (`NEW`, `LET`, `PRINT`, `ASSERT`, `EXPECT`, `IF`). Any tokenized keyword outside this set — including `RULE` (the canonical entry-block keyword!), `IS`, `MATCH`, `CYCLE`, `WHILE`, `FOR`, `START`, `WITH`, `TRIAL`, `RETURN`, `USING`, etc. — falls through to `return nil`. The `ParseProgram` loop at line 80–84 only appends non-nil statements, so unrecognized tokens are silently swallowed with **no parser error recorded**.

## Demonstrable Failure Mode

The legacy `fn main() { print("Hello World") }` source lexes as `[IDENT fn, IDENT main, LPAREN, RPAREN, LBRACE, PRINT, LPAREN, STRING, RPAREN, RBRACE]`. In the parser loop:

1. `fn` (IDENT) → `parseStatement` returns `nil` → token silently dropped.
2. `main`, `(`, `)`, `{` → all silently dropped.
3. `print` (PRINT) → `parsePrintStatement` parses the string argument correctly.
4. `)`, `}` → silently dropped.

Result: program has exactly one statement (`print("Hello World")`) and prints "Hello World" successfully — exit 0, status "PASS" — even though the file violates `spec/03-rules.md` §1 ("Every executable main module MUST define a top-level `rule main:`"). The legacy `fn` keyword is **not** in `internal/token/token.go::keywords` nor in any spec file.

## Impact

- False positives in TDD tests: authoring a test with **any** unsupported syntax produces a "PASS" by accident, masking real spec violations.
- Decision 2/3 deprecation sweep (`≠`, `==`, `!=`, `is`) cannot be observed in level0 because the tokenizer tolerates them — but neither can the compiler diagnose their rejection.
- Hardening Decisions 2/3 to `E0009` (hard syntax error) becomes impossible until every parser path emits errors for unknown top-level tokens.

## Required Resolution (Solution C-)

Refactor `internal/parser/parser.go::ParseProgram` to:

1. **Detect non-nil `stmt == nil`** on a non-EOF token → append a diagnostic entry to `p.errors` with token type, literal, and source line (`token.Pos`).
2. **Extend `parseStatement`** to handle the full statement taxonomy enumerated in `spec/02-statements.md` §1: `set`, `new`, `let`, `zap`, `assert`, `expect`, `print`, `write`, `read`, `if`, `match`, `start`, `with`, `cycle`, `for`, `trial`, `try`, `case`, `miss`, `final`, `return`, `stop`, `redo`, `next`, `pass`, `raise`, `resume`, `retry`, `fail`.
3. **Add `RULE` to the dispatch** so the canonical entry-block form `rule main:` is recognized (currently requires manual keyword handling).
4. **Defer full E0009 enforcement** until the canonical forms are wired; emit a soft warning `W0901: UnrecognizedStatement` for fallback cases that are not yet grammar-mapped.

## Test Evidence (Pre-Fix)

- `test/level0/T0001.bee`: previously contained `fn main() { print("Hello World") }`; exited 0 with stdout "Hello World" — false PASS.
- After harmonization (this fix): T0001 normalized to `rule main: print("Hello World"); return;` — still passes, but now matches spec/03 §2.1 and spec/02 §3.5.
- The parser-silent-drop hazard remains unaddressed for other legacy forms users may still author.

## Severity

**High** — blocks Phase 7.2 "Edge case tests identified during audit" because edge-case failure modes cannot be observed until the parser records errors for unknown statements.

## Owner

Implementation (post-spec-harmonization gate per `MANIFEST.md` Anti-Loop Gate).
## Resolution (Phase 7.7, 2026-09-12)

1. **Errors / Warnings split.** `internal/parser/parser.go::Parser` now carries `errors []string` (hard E0009) and `warnings []string` (soft W0901 + E0011). `Warnings()` accessor added.
2. **Hard-error path unchanged.** `ParseProgram` still records `E0009 SyntaxError:UnrecognizedStatement` when `parseStatement` returns `nil` on a non-EOF token (issue 14 §1 requirement).
3. **Soft-warning path added.** Every former-W0901 emission in `parseRuleEntry` and `parseMatchStatement` is now written to `p.warnings` (not `p.errors`).
4. **`cmd/bee/main.go` ignores warnings.** The post-parse check now distinguishes the two slices — warnings are surfaced to stderr but do not trigger `os.Exit(1)`.
5. **Demonstrable proof (post-fix):** the legacy `fn main() { print("Hello World") }` source now hard-fails with six `E0009 SyntaxError:UnrecognizedStatement: line=1 type=IDENT literal="fn"` (and five siblings) — exit status 1. The previously false-positive PASS path is closed.
6. **Verification artefacts:** `sh run.sh test level0` reports `Total Passed: 5, Total Failed: 0` (T0001–T0004 + smoke). All four .bee files execute via the new `rule main:` parser path and return their expected stdout.

**Next gate.** Phase 8 — extend the warning surface to include missing-keyword structural checks (`expect` without `;`, `if` without `do`, etc.) so the entire executor grammar receives active coverage.
