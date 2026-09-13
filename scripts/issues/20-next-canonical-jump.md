# Issue: `next` Canonicalization as Loop-Jump Keyword (D15)

## Problem
Decision 14 repurposed `repeat` as the inline loop-jump statement (continue semantics) and retired `next`. In review, `repeat` proved the worse canonical spelling:

1. **Length & clarity:** `repeat` is a long keyword for the most frequently used loop jump. `next` is a single short word that reads as exactly what it does — advance to the next iteration.
2. **Linguistic consistency:** the jump trio reads naturally as `next` / `stop` / `redo` (advance / exit / restart). `repeat` collides conceptually with `redo` (both suggest "do again"), while `next` is unambiguous.
3. **Legacy weight:** pre-D14 Bee code used `next` for continue semantics; retiring it broke intuition for existing authors while gaining nothing semantically.

## Root Cause
D14 optimized for grammar uniformity (demoting `repeat` from block terminator) but chose the wrong keyword identity for the jump role. The keyword set was not evaluated for semantic distinctness across the jump table.

## Requirements (Decision 15)

### 1. `next` Canonical Jump
- **`next [label] [if condition];`** is the canonical loop-jump statement.
- While-cycle: jump directly to the loop header; re-evaluate the condition.
- For-cycle: advance the iterator to the next domain element and re-evaluate domain boundaries.

### 2. `repeat` Deprecated Synonym
- `repeat` remains accepted by the lexer but maps to the `NEXT` token.
- Every occurrence emits a non-fatal `E0010 deprecated-keyword: 'repeat' — use 'next'` warning (stderr).
- Phase 7 audit task 7.2 hardens the diagnostic to a hard `E0009` syntax error, at which point `repeat` leaves the language.

### 3. Unchanged Jump Table
| Keyword | While-cycle | For-cycle |
| :--- | :--- | :--- |
| `next [label]` | Jump to loop header; re-evaluate condition | Advance iterator to next domain element; re-evaluate domain boundaries |
| `stop [label]` | Exit loop block immediately past `done` | Exit loop block immediately past `done` |
| `redo [label]` | Restart current iteration body without re-evaluating condition | Re-run body for current iterator value without advancing domain |

### 4. Frozen-Suite Preservation
- `@FROZEN` tests in `test/levelX/` spelling the jump `repeat` MUST NOT be modified. The synonymy mechanism keeps them green: the parser only ever sees `NEXT`.

## Impact Surface
| File | Sections | Action |
| :--- | :--- | :--- |
| `internal/lexer/lexer.go` | `NextToken` ident dispatch | Map `repeat` → `NEXT` + E0010 warning (✅ done) |
| `spec/02-statements.md` | §1 taxonomy, §3.4, §5 EBNF, §7 diagnostics | `next` canonical; E0010 row; E0206 wording |
| `spec/00-memory-model.md` | §2.2 region cleanup | Keyword mention `repeat` → `next` |
| `tutorial/control.html` | §cycle notes, §nested, §for examples | `repeat` → `next` in code/keyword mentions |
| `tutorial/syntax.html` | Keyword tables | `repeat` → `next` |
| `tutorial/js/bee.js` | Highlighter | `repeat` recategorized control → interruption |
| `todo/DECISIONS.md` | D15 entry | Ratification record |
| `MANIFEST.md` | D15 mirror + Phase 8.6 delta | Status update |

## Acceptance Criteria
1. `spec/02-statements.md` EBNF lists `"next" [ label ]` in `transfer_stmt` with a deprecation note for `repeat`.
2. Tutorial code examples spell the jump `next`; no keyword-table row presents `repeat` as canonical.
3. Lexer emits E0010 on `repeat` and produces identical parse results for `repeat` and `next`.
4. D15 recorded in `todo/DECISIONS.md` and mirrored in `MANIFEST.md`.
5. `sh run.sh smoke` passes — all `@FROZEN` levels green without modifying any frozen test.
