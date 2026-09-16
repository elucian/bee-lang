# Issue: Level-4 Collections & Processing Implementation Gaps

## Problem
`test/level4/` covers `spec/10-collections.md` and `spec/11-processing.md`. A static-code audit of all 44 cases found the active suite is green (9 active/@FROZEN cases PASS), but **34 cases are `@DISABLED`** because the collections/processing features they exercise are either (A) valid spec features the compiler has not yet implemented, or (B) not aligned with the ratified spec (or misfiled from another level). This file tracks the not-yet-clarified / not-yet-implemented surface so each item can be decided and implemented once.

The `@DISABLED` line-1 reasons are stale (they reference old line-5 `E0009` errors) and do not reflect today's lexer/parser/evaluator behaviour. They should be refreshed together with each item's implementation.

## Category A — Valid spec features, compiler not yet implemented
These are genuine `spec/10`/`spec/11` features with no (or partial) lexer/parser/evaluator support. Each is a self-contained implementation unit.

| Test(s) | Spec | Feature | Current failure |
| :--- | :--- | :--- | :--- |
| T0411, T0425 | 10 §3.2 | Typed fixed array decl `new a ∈ [Z](5)` / range-init `[Z](1..10)` | Parser `E0009` |
| T0413, T0435, T0436 | 11 §5 / builders | Set & array/hash-map builders/comprehensions `{ x \| x ∈ s ∧ p }` | Lexer rejects `\|` (`ILLEGAL`) |
| T0433 | 10 §2.3 | Ranged slice-view binding `::` (`new first :: a[1..$]`) | Parser `E0009` |
| T0434 | 02 §2.1 (D13) | Postfix stepped range in a declaration `(1..5)(2)` | Parser `E0009` |
| T0437 | 11 §3 | List element iteration `for e ∈ l` (range iteration works; list does not) | Evaluator yields 0 elements |
| T0438, T0414, T0415 | 10 §1/§3.1 | List concat `a + b`, append/remove `<+` `+<` `<<`, char literals `'a'` | Lexer/parser `E0009`; concat degrades to arithmetic |
| T0439 | 11 §5.1 / D8 (deferred) | Array deconstruction `new x, y, *rest := list` | Parser `E0009` (D8 deferred) |
| T0440 | 10 §3.3 + 11 §5.2 | Typed matrix `new M ∈ [Z](2,2)`, 2D index `M[1,1]`, row/col slice `M[1,*]` | Parser `E0009` |
| T0422 | 10 §3.4 | Set algebra `∪ ∩ Δ ⊂` (parse OK, evaluate wrong) | Evaluator returns 0 for every op |
| T0424 | 11 §2.1 | Deep-clone declaration `new copied :: a` | Parser `E0009` |
| T0441–T0444 | 10/11 | Set algebra, quantifiers (`∀`/`∃`), casting, queue FIFO | Lexer/parser voids (see per-test run) |

## Category B — Misaligned or misfiled; needs decision or reclassification
| Test(s) | Reason |
| :--- | :--- |
| T0402–T0409 | Contract constructs (`expect`, `assert`, `old`, `@` ref-arg, `=> (r ∈ Z)` result binding) belong to `spec/03-rules.md` (level 3), not level-4 collections. Misfiled; currently disabled. |
| T0417 | Uses `scrap m[1]` for map removal, but spec 10 §3.5 mandates `zap`. Token exists but no parser statement; spec decision needed on `scrap` vs `zap`. |
| T0421 | Set element add/remove via `s += e` / `s -= e` is NOT specified in 10 §3.4 (only `∩ ∪ \ Δ ⊂ ⊃`). Degrades to arithmetic. Either spec a `+=`/`-=` for sets or fix the test. |
| T0423 | ASCII synonyms `union`/`inter`/`subset` have no token/operator in spec or lexer. Either adopt or remove. |
| T0426 | Slice bound `a[0..2]` uses a `0` index, violating 1-based spec §2.1 (E1006). Test intent is error-detection; confirm. |
| T0427 | `for k, v ∈ m` two-index loop header is not in the parser cycle-header grammar (single index only). Decide whether to spec a key/value iteration form. |
| T0419, T0418 | Quantifier/comprehension spelling does not match ratified EBNF (`∀ (i∈S) ∧ (p)` vs `(∀ x ∈ S : cond)`; `{(x:x²) \| x∈..}`). Test syntax is stale vs spec. |

## Questions for decision
1. List concatenation / element mutation operators: adopt `<+`/`+<`/`<<` (as the disabled tests assume) or a different set? `+` on lists currently degrades to arithmetic — needs an explicit operator.
2. Set element add/remove (`+=`/`-=` on sets) — is this a desired feature to spec, or reject in favour of builder/algebra only?
3. Key/value map iteration (`for k, v ∈ m`) — desired syntax?
4. `scrap` vs `zap` for map/collection removal where no allocation lifetime is involved — is `scrap` a separate token worth keeping?

## Priority
**Medium** — Items are individually schedulable and mostly independent. T0417 (`scrap`/`zap`), T0421 (set `+=`), T0427 (map kv iteration), and T0423 (ASCII synonyms) require language decisions before spec changes. Category A parser/evaluator work can proceed once each feature's grammar is locked in `spec/10`/`spec/11`.
