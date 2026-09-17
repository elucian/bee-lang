# Solution: `next` Canonicalization as Loop-Jump Keyword (D15)

## Design

### 1. Synonymy-Based Migration (same mechanism as D7/D12/D13)

Rather than removing `repeat` from the token dispatch (which would break the `@FROZEN` suite), the migration reuses the established deprecation pattern: the lexer collapses the legacy spelling onto the canonical token and emits a non-fatal warning.

```
source `repeat`  ──lexer──▶  token.NEXT (literal "next")  +  E0010 warning (stderr)
source `next`    ──lexer──▶  token.NEXT (literal "next")
```

Consequences:
- **Parser:** unchanged — `parseTransferStatement` already dispatches `token.NEXT` with optional label and `if` guard. The defensive `token.REPEAT` case is retained so any direct construction of the legacy token still parses.
- **Evaluator:** unchanged — it dispatches on `NEXT` transfer semantics (jump-to-header / advance-iterator).
- **Tests:** `@FROZEN` files spelling `repeat` lex to `NEXT` and pass unmodified; stderr carries the deprecation warning without failing the run (warnings are non-fatal per the existing `cmd/bee/main.go` surface).

### 2. Canonical Grammar

```ebnf
transfer_stmt ::= ( "return" [ expression_list ]
                  | "stop" [ label ]
                  | "redo" [ label ]
                  | "next" [ label ]        (* legacy synonym: "repeat" — deprecated, E0010 *)
                  | "pass"
                  | "raise" [ expression ]
                  | "resume"
                  | "retry"
                  | "fail" expression ) [ "if" expression ] ;
```

### 3. Jump Table (ratified)

| Keyword | While-cycle behaviour | For-cycle behaviour |
| :--- | :--- | :--- |
| `next [label]` | Jump to loop header; re-evaluate condition | Advance iterator to next domain element; re-evaluate domain boundaries |
| `stop [label]` | Exit loop block immediately past `done` | Exit loop block immediately past `done` |
| `redo [label]` | Restart current iteration body without re-evaluating condition | Re-run body for current iterator value without advancing domain |

### 4. Hardening Path
- **Now:** `E0010 deprecated-keyword: 'repeat' — use 'next'` — non-fatal, stderr only.
- **Phase 7.2:** diagnostic upgrades to `E0009` (hard syntax error); `"repeat": REPEAT` is then removed from `internal/token/token.go` Keywords and the parser's defensive `token.REPEAT` case is dropped.

## Verification
1. Lex a file containing both `next;` and `repeat;` — both must produce `token.NEXT`; only `repeat` emits E0010.
2. `sh run.sh smoke` — all active `@FROZEN` levels green (repeat-spelling tests pass via synonymy).
3. Tutorial grep: no `repeat` keyword usage remains in `tutorial/*.html` code examples (prose uses of the English word "repeat" are unaffected).
