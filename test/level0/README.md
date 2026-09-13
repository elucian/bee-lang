# Level 0: Compiler Self-Bootstrap & Smoke

Boot-level tests that validate the Bee compiler can lex, parse, evaluate, and print outputs without crashing. These tests intentionally cover the minimum surface required to exercise the print/io grammar path and entry-rule header (`rule main: ... return;`).

| CASE     | DESCRIPTION                      | STATUS     |
| -------- | -------------------------------- | ---------- |
| T0001    | Hello World Test                 | PASS       |
| T0002    | Large list of numbers            | PASS       |
| T0003    | Multiple arguments               | PASS       |
| T0004    | Test print with custom separa... | PASS       |
| smoke    | Smoke test validation            | PASS       |

## Spec Coverage

- `spec/02-statements.md` §5 — `io_stmt` EBNF: `print [ "(" expression_list ")" | expression_list ] [ "(" [ named_arg_list ] ")" ]` (Decision 6 curried named-argument postfix; legacy `using` is deprecated — emits non-fatal `E0011`).
- `spec/03-rules.md` §2.1 — canonical rule signature `rule identifier(...): ... block ... return;`.
- `spec/03-rules.md` §2.4 — optional named-parameter slot `(sep: ", " ∈ Str)` for curried kwargs.
- `spec/03-rules.md` §3.1 — curried call-site form `print(a, b)(sep: " | ")`.
- `spec/03-rules.md` §3.2 — mandatory `return;` block terminator alignment (0 relative indentation).

## Decision 6 Migration (2026-09-12)

`T0002` and `T0004` were migrated from the legacy `print(...) using "sep"` / `print(...) using: "sep"` postfix to the canonical curried `(sep: "sep")` call-site form. The legacy lex surface remains accepted but emits `E0011 DeprecatedSymbol 'using'` and is scheduled to harden to `E0009` in Phase 7 audit task 7.2.

## Phase 8.1 Operator Identity (2026-09-12)

The lexer now collapses `is not` (literal sequence `is` + ASCII space + `not`) into a single `IS_NOT` (Decision 7). The evaluator maintains a parallel identity-ID plane so `a is a` returns 1, `a is b` returns 0 across distinct variables, and integer literals never share ephemeral IDs. The lexer additionally emits a non-fatal `E0010 deprecated-symbol` warning for any of `≠`, `==`, `!=` (they remain functional for legacy reasons but track toward `E0009` enforcement in Phase 7.2).

## Known Hazards Surfaced

- `issues/14-parser-silent-token-drop.md` — `parseStatement` returns `nil` for unrecognized keywords without recording a parser error. Legacy `fn main() {}` source previously passed as a false positive; canonical `rule main:` form verified below.

## Verification

```sh
./bin/bee.exe -e test/level0/T0001.bee   # → Hello World
./bin/bee.exe -e test/level0/T0002.bee   # → 10,20,30,40,50
./bin/bee.exe -e test/level0/T0003.bee   # → Hello, World!
./bin/bee.exe -e test/level0/T0004.bee   # → 1 | 2
./bin/bee.exe -e test/level0/smoke.bee   # → Smoke test
```
