# Issue: `≡` Still Registered as a Value-Comparison Operator

The triple-bar operator (`≡`) was ratified as **geometric congruence only**
(same shape class, same interior/exterior angles, same side count, any scale),
reserved for the graphics engine per `spec/13-graphics.md` §5.1 and Decision 3
(the reservation survives Decision 12). The tutorial now follows this rule:
`=` is value equality, `is` / `is not` and `@a = @b` are reference identity,
and `≡` appears only in geometry.

The compiler front-end, however, still treats `≡` as a general-purpose binary
comparison operator rather than a geometry-only congruence test.

## Current Implementation

| Layer | File | Behavior |
| :--- | :--- | :--- |
| Token | `internal/token/token.go` | Defines `EQUIV = "≡"`. |
| Lexer | `internal/lexer/lexer.go` | `case '≡': token.EQUIV`. |
| Parser | `internal/parser/parser.go` | `≡` in `isBinaryOp` (both `token.EQUIV` and literal `"≡"`) and `opPrecedence` → `precCompare`. |
| Evaluator | `internal/evaluator/comparison.go` | `evalComparison` has **no** `EQUIV`/`≡` branch; the literal falls through to `switch op` (which has no `"≡"` case) and silently returns `0`. |

## Impact

- `≡` is accepted anywhere a binary comparison is accepted, not gated to
  geometric types/contexts as the spec requires.
- Any `poly_a ≡ poly_b` expression evaluates to `0` unconditionally, because
  no congruence semantics exist — silently wrong, with no diagnostic.

## Requirements

1. Decide whether `≡` should:
   - be removed from the general comparison operator set (lexer/parser) and
     re-introduced only inside the graphics engine, or
   - remain tokenizable but be rejected (non-fatal/hard) outside geometric
     type contexts until congruence evaluation is implemented.
2. Implement congruence semantics (`≡` → true iff same shape class / angles /
   side count, regardless of scale) under `spec/13-graphics.md` or defer with
   an explicit diagnostic rather than silent `0`.
3. Keep lexer, parser, and evaluator in agreement with `spec/01-lexical-structure.md`
   §5 (`cmp_val_op`, `cmp_ref_op`) so `≡` is not a value-op in any normative path.

## Notes

- Documentation is already harmonized (tutorial + spec); this issue is scoped
  to the compiler front-end only.
- No edits to `internal/lexer/`, `internal/parser/`, or `internal/evaluator/`
  begin until this issue is triaged and the target behavior is ratified.
