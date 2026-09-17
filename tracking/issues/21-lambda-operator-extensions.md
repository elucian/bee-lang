# Issue: Lambda Operator Extensions — Compose `∘`, Pipe `|>`, Partial `?` (Bucket C)

## Problem
Three level-3 tests exercise higher-order lambda operators that have no parser or evaluator support:

| Test | Operator | Description | Current Status |
| :--- | :--- | :--- | :--- |
| T0317 | `∘` (compose) | `f ∘ g` composes two lambdas into a new lambda | Moved to `test/debt/` (`@DEBT`) — parser rejects `∘` |
| T0318 | `|>` (pipe) | `value |> f |> g` chains value through lambdas left-to-right | Moved to `test/debt/` (`@DEBT`) — parser rejects `|>` |
| T0319 | `?` (partial) | `add(5, ?)` creates a curried lambda with one argument bound | Implemented; `T0319.bee` PASS (returned to active level3) |

All three fail with `E0009 SyntaxError:UnrecognizedStatement` at the operator token. These features are **not blocking** — they are sugar-level extensions that don't gate any other level-3 or level-4 work.

## Root Cause
The operators `∘`, `\|>`, and `?` (as placeholder) are absent from:
1. **Lexer:** No token types defined (`COMPOSE`, `PIPE`, `PLACEHOLDER`).
2. **Parser:** No precedence entries in `parseExpressionClimb`; no special-call handling for `?` placeholder.
3. **Evaluator:** No runtime logic for lambda composition, pipe chaining, or partial application.
4. **Spec:** `spec/07-functions.md` has no grammar productions for these operators. Test descriptions reference "§5" and "§6" but those sections cover SIMD/vectorization and EBNF grammar respectively — the operator semantics were never formally specified.

## Prerequisites
Per project protocol, implementation requires ratification **before** any code changes:

1. **Spec update:** Add grammar productions and semantics to `spec/07-functions.md`:
   - `∘` as an infix operator with defined precedence (likely between comparison and arithmetic).
   - `|>` as a left-associative infix operator (pipeline/application).
   - `?` as a placeholder expression valid only in call argument position.
2. **Decision ratification:** Record in `manual/DECISIONS.md` as D16 (or next available).

## Design Sketch

### 1. Compose `∘` (T0317)
```bee
new h := f ∘ g;    -- h(x) = f(g(x))
```
- **Type:** `(L, L) → L`
- **Semantics:** Returns a new lambda that applies right operand first, then left.
- **Precedence:** Higher than comparison, lower than arithmetic.

### 2. Pipe `|>` (T0318)
```bee
new result := 5 |> double |> inc;    -- inc(double(5))
```
- **Type:** `(any, L) → any`
- **Semantics:** Left-associative; `x |> f` desugars to `f(x)`.
- **Precedence:** Lowest of all binary operators (below comparison).

### 3. Partial `?` (T0319)
```bee
new add5 := add(5, ?);    -- λ(y) => add(5, y)
```
- **Type:** Placeholder `?` marks an unbound parameter slot.
- **Semantics:** When `?` appears in a call argument, the entire call expression becomes a lambda with `?` positions as parameters (left-to-right ordering).
- **Constraint:** `?` is only valid in call argument position; bare `?` is a syntax error.

## Impact Surface
| File | Sections | Action |
| :--- | :--- | :--- |
| `internal/token/token.go` | Token type constants | Add `COMPOSE`, `PIPE`, `PLACEHOLDER` |
| `internal/lexer/lexer.go` | Operator dispatch | Lex `∘`, `\|>`, `?` (placeholder context) |
| `internal/parser/` | Precedence table, expression parser | Wire new operators, `?` call-site detection |
| `internal/evaluator/` | Expression evaluator | Implement compose, pipe, partial logic |
| `spec/07-functions.md` | §5 or new §5a | Add grammar productions, semantics, examples |
| `manual/DECISIONS.md` | D16 entry | Ratification record |
| `manual/MANIFEST.md` | D16 mirror | Status update |
| `tutorial/functions.html` | New section or §advanced | Add contract rules and examples for all three operators |

## Acceptance Criteria
1. `spec/07-functions.md` EBNF includes `∘`, `|>`, `?` productions.
2. D16 recorded in `manual/DECISIONS.md` and mirrored in `manual/MANIFEST.md`.
3. T0317, T0318, T0319 pass with `@DISABLED` removed.
4. Tutorial `functions.html` documents all three operators with contract rules and examples.
5. `sh run.sh smoke` passes — no regressions in existing frozen tests.

## Priority
**Low** — These are convenience operators. No other tests or spec sections depend on them. Defer until core level-4 work stabilizes.
