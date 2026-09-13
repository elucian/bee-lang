# Issue: Control Flow Syntax Refactoring (D14)

## Problem
The `repeat` keyword is overloaded as both a block terminator (`repeat;` closes `cycle`/`for`) and an inline loop-continuation modifier (`repeat if condition;`). This creates parser ambiguity and grammar inconsistency. Additionally, anonymous cycles cannot access the stable outer scope prologue, forcing developers to manufacture dummy labels solely for persistent variable declarations.

## Root Cause
- **Grammar asymmetry:** `cycle`/`for` terminate with `repeat [label];` while all other blocks (`if`, `match`, `start`, `with`, `trial`) terminate with `done [label];`. The parser must maintain block-type-specific state to choose the correct terminator.
- **Scope-label coupling:** The `:` (declaration prologue) is bound to the label (`cycle name:`), so anonymous cycles (`cycle do`) have no prologue. There is no way to get a stable scope without inventing a label.

## Requirements (Decision 14)

### 1. Uniform Block Termination
- All control blocks (`cycle`, `for`, `if`, `match`, `start`, `with`, `trial`) terminate with **`done [label];`**.
- The `repeat` keyword is **removed as a block terminator** entirely.

### 2. Anonymous Scope Prologue (`cycle:`)
- **`cycle:`** (colon, no label) opens a stable declaration prologue without requiring a label identifier. Variables persist across all iterations.
- **`cycle name:`** (label + colon) opens a labeled stable prologue (same as before).
- **`cycle do`** / **`cycle while cond do`** (no colon) is the volatile-body form — no prologue. A label may still prefix the body header (`cycle name do`) purely as a jump target; it creates no scope.
- Bare `cycle:` is now **valid** (previously invalid). Label and colon are **independent options**: `cycle [label] [:]`.

### 3. `repeat` Repurposed as Inline Jump
- **`repeat [label] [if condition];`** is now an inline transfer statement (like `stop`, `redo`, `next`), equivalent to `continue` in C-family languages.
- **`next`** is retired as a keyword (absorbed by `repeat`).

### 4. Loop-Jump Consolidation
| Keyword | Role | Context |
| :--- | :--- | :--- |
| `repeat` | Re-evaluate / advance to next iteration | `cycle`, `for` |
| `stop` | Terminate loop immediately | `cycle`, `for` |
| `redo` | Restart current iteration without advancing | `cycle` (non-`for`) |

### 5. `then` Semantics
- The optional `then` block becomes a **post-loop epilogue**: it executes exactly once after the loop exits (normally via `stop` or condition exhaustion, NOT via `repeat`/`redo`/`next`).
- The `then` block runs in the enclosing scope of the `cycle` (i.e., it sees the prologue variables but the volatile body scope is already popped).

## Impact Surface
| File | Sections | Action |
| :--- | :--- | :--- |
| `spec/02-statements.md` | §3.4, §5, §6, §7 | Rewrite loop EBNF, terminator rules, jump semantics |
| `spec/06-objects.md` | §5 trait example | `repeat;` → `done;` |
| `spec/11-processing.md` | §3.1 quantifier example | `repeat;` → `done;` |
| `spec/12-concurrency.md` | §2.1, §3, §4.1, §4.2 examples | `repeat;` → `done;` |
| `spec/14-library.md` | §4.2 example | `repeat;` → `done;` |
| `tutorial/control.html` | §cycle, §for, §nested-cycles, §while-condition | All examples + notes |
| `tutorial/collections.html` | 3 code examples | `repeat;` → `done;` |
| `tutorial/concurrency.html` | 4 code examples | `repeat;` → `done;` |
| `tutorial/objects.html` | 1 code example | `repeat;` → `done;` |
| `tutorial/processing.html` | 4 code examples | `repeat;` → `done;` |
| `tutorial/rules.html` | 2 code examples | `repeat;` → `done;` |
| `tutorial/structure.html` | 1 code example | `repeat;` → `done;` |
| `tutorial/syntax.html` | Keyword table | Update `repeat` description |
| `todo/DECISIONS.md` | Index + D14 entry | Add D14 ratification record |
| `MANIFEST.md` | D14 mirror | Add D14 to User-Locked Decisions |

## Acceptance Criteria
1. `spec/02-statements.md` §3.4, §5, §6, §7 fully reflect D14 grammar.
2. All tutorial `.html` examples use `done;` terminators for loops.
3. No `repeat` block terminators remain in any spec or tutorial example.
4. `next` keyword is removed from EBNF and tutorial examples.
5. D14 entry added to `todo/DECISIONS.md` with full ratified semantics.
6. MANIFEST.md updated with D14 mirror entry.
7. `sh run.sh smoke` passes (existing tests are legacy-syntax `@FROZEN`; the smoke run confirms no parser regressions from doc-only changes).

## Compiler Implementation (Deferred)
Compiler changes (`internal/lexer/`, `internal/parser/`, `internal/evaluator/`) are **not** in this pass. They will follow in a separate implementation phase after this spec/tutorial harmonization is ratified.
