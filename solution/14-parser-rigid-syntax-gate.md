# Solution 14: Parser Rigid Syntax Gate — Fail-Hard on Unknown Tokens

**Status:** Proposed (counters Issue 14)
**Date:** 2026-09-12
**Source Issue:** `issues/14-parser-silent-token-drop.md`
**Spec Reference:** `spec/02-statements.md` §1 (executive statement taxonomy), `spec/02-statements.md` §5 (EBNF), `spec/03-rules.md` §1 (canonical `rule main:` form).

## Root Cause

`internal/parser/parser.go::parseStatement` (line 88) only handles a closed switch over a small keyword set (`NEW`, `LET`, `PRINT`, `ASSERT`, `EXPECT`, `IF`). Any tokenized keyword outside this set — including `RULE` (the canonical entry-block keyword!), `IS`, `MATCH`, `CYCLE`, `WHILE`, `FOR`, `START`, `WITH`, `TRIAL`, `TRY`, `CASE`, `MISS`, `FINAL`, `RETURN`, `STOP`, `REDO`, `NEXT`, `PASS`, `RAISE`, `RESUME`, `RETRY`, `FAIL`, `ZAP`, `WRITE`, `READ`, `SET` — falls through to `return nil`. The `ParseProgram` loop at line 80–84 only appends non-nil statements, so unrecognized tokens are silently swallowed with **no parser error recorded**.

## Strategy

1. **Error Emission (`p.errors`)** — convert silent nil-drop into explicit `E0009 SyntaxError:UnrecognizedStatement` diagnostic with token Type, Literal, and source position.
2. **Switch-extension to the canonical taxonomy** — extend `parseStatement` to dispatch the full statement taxonomy from `spec/02-statements.md` §1 and §5:
   - `NEW, SET` → declaration handler (canonical)
   - `LET` → mutation assignment handler (extended for Decision 2 harden)
   - `ZAP` → memory invalidation handler (new)
   - `ASSERT, EXPECT` → contract handlers (existing, retained)
   - `PRINT, WRITE, READ` → io_stmt handler (existing, retained)
   - `IF` → conditional handler (existing, retained)
   - `MATCH` → pattern matching handler (new; grammar-mapped skeleton)
   - `START, WITH` → scope block handler (new; grammar-mapped skeleton)
   - `CYCLE, FOR, WHILE` → loop handler (new; grammar-mapped skeleton)
   - `TRIAL, TRY, CASE, MISS, FINAL` → transactional handler (new; grammar-mapped skeleton)
   - `RETURN, STOP, REDO, NEXT, PASS, RAISE, RESUME, RETRY, FAIL` → transfer handler (new; grammar-mapped skeleton)
   - `RULE` → entry-block handler (new; canonical form)
3. **Adding `RULE` to the dispatch** — the canonical `rule main:` form is recognized; without this, the parser silently drops the keyword, leaving the entry block body orphaned.
4. **Fatal `E0009` only when zero statements are produced** — soft warning `W0901 UnrecognizedStatement` is attached to `p.errors`, but the parser still continues to surface later errors and emit zero exit-1 when any `E0009` is recorded. This matches the spec's intent: a single unknown token cannot make the build "PASS" silently, but partial token streams from incomplete authoring are still parseable enough to enumerate remaining errors.

## Implementation Notes

- The parser must produce **both** a `p.errors` slice entry AND continue to surface later tokens to allow multi-error diagnosis.
- Errors are formatted as `E0009 SyntaxError:UnrecognizedStatement: line=N type=<Type> literal=<Literal>` to match existing Bee diagnostic code conventions.
- Position tracking uses `token.Pos` (already present on every `Token`); the `Parser` does not need a new lexer field.

## Validation

- `test/level0/T0001.bee` (canonical `rule main:` form) → still emits `Syntax OK` exit 0 (PASS).
- `test/level0/T0002.bee` (Decision 6 curried form) → still emits `Syntax OK` exit 0 (PASS).
- A new traceability test (T0005) is added: a file containing the legacy `fn main() { print(...) }` form must emit `E0009` for the `fn` token and exit non-zero (FAIL).
- `sh run.sh smoke` exits 0 (compiler build and smoke pass).

## Risks Mitigated

- **False-positive "PASS" hazard:** resolved — unknown tokens now raise non-zero exit codes.
- **Decision 2/3 deprecation sweep:** deprecation warnings (`E0010`, `E0011`) flow into `p.errors` cleanly.
- **Phase 7.2 `E0009` hardening:** groundwork laid — once each grammar form has a real handler, switching the policy from soft-warning to hard-fail is a constant-time change.
