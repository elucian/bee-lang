# Issue 18 — Parallel colon-initialization (`new a: 1, b: 2 ∈ Z;`) is unimplemented

## Status
**Open — failing test (T0121+T0122 disabled).** Spec ambiguity + parser gap.

**Awaiting ratification of D9 + D11 (`todo/DECISIONS.md`).** Once the user
ratifies the colon `:` pair-up semantics (D9) and the parallel
colon-initialisation grammar (D11), the `parseDeclaration` rewrite
becomes a single coordinated change. Do NOT implement before
ratification.

## Related
- Issue #1 (`issues/001-operator-colon-ambiguity.md`) covers the general `:` vs `:=` vs `=` semantics. This issue is the SPECIFIC parallel-destructure grammar extension.

## Observed Behaviour

```bee
rule main:
  new a: 1, b: 2 ∈ Z;
  let a, b := b, a;
  expect a = 2;
  expect b = 1;
  print ("success switching a,b);
return;
```

Compiler halts at `new a: 1, b: 2 ∈ Z;` with `E0009 SyntaxError:UnrecognizedStatement: line=5 type=; literal=";"`.

## Compiler Trace (debug `-d`)

```
[TOKEN] Type: NEW             | Literal: new        | Pos: 5
[TOKEN] Type: IDENT           | Literal: a          | Pos: 5
[TOKEN] Type: :               | Literal: :          | Pos: 5
[TOKEN] Type: INT             | Literal: 1          | Pos: 5
[TOKEN] Type: ,               | Literal: ,          | Pos: 5
[TOKEN] Type: IDENT           | Literal: b          | Pos: 5
[TOKEN] Type: :               | Literal: :          | Pos: 5
[TOKEN] Type: INT             | Literal: 2          | Pos: 5
[TOKEN] Type: ∈               | Literal: ∈          | Pos: 5
[TOKEN] Type: IDENT           | Literal: Z          | Pos: 5
[TOKEN] Type: ;               | Literal: ;          | Pos: 5
```

## Root Cause — Two Compounding Issues

### Bug A — Spec ambiguity: the `:` operator has multiple meanings

`spec/02-statements.md` §2.1 lists:
- `:=` for type inference
- `:` for structural pair-up
- `= with ∈ Type` for explicit-type initialization

The phrase "structural pair-up" is NOT defined elsewhere in the spec. The user-intended semantics for `new a: 1, b: 2 ∈ Z;` is "declare `a` initialised to `1` AND `b` to `2`, both of type `Z`". This is a **typed pair-up parallel declaration**. The spec EBNF does NOT express this grammar.

### Bug B — Parser gap: `parseDeclaration` has `:` branch but consumes raw tokens blindly

`internal/parser/parser.go::parseDeclaration` lines 498-503:
```go
} else if nextTok.Literal == ":" {
    // e.g. new a: 1, b: 2 ∈ Z;
    _ = p.parseExpression() // consume initial val
    for p.l.PeekToken().Type != token.SEMICOLON && p.l.PeekToken().Type != token.EOF {
        p.l.NextToken()
    }
}
```

This branch fires when `nextTok` is the `:`. It calls `parseExpression` once (consumes `1`), then **iterator-eats all remaining tokens until `;`**. That correctly skips the comma-separated list but DOES NOT build a typed AST. The evaluation phase then sees a half-baked `DeclarationStatement` with `Value` set to "1" and no record of the second variable, the `∈ Z` qualifier, or even that the `:` operator ever existed.

### Bug C — T0122 has its own typo (`print("` unterminated)
The `.bee` source line 9 reads:
```
print ("success switching a,b);
```
— the opening `"` is never closed. This makes the test fail at a different layer than the underlying grammar gap. **Two problems in one test**: the parser rejects the syntax AND the source has an unbalanced string literal.

## Proposed Resolution (one pass — NOT yet implemented)

1. **Spec update**: extend `spec/02-statements.md` §2.1 with a formal pair-up declaration grammar:
   ```ebnf
   decl_stmt ::= "new" ident_list ":" expr_list ( "∈" | "in" ) type_specifier ";" ;
   decl_stmt ::= "set" ident_list ":" expr_list ( "∈" | "in" ) type_specifier ";" ;
   ```
   Add prose: "When `ident_list` length N matches `expr_list` length N, each identifier is bound to the corresponding expression, and all share the type qualifier."
2. **Fix T0122 test source** — replace `print ("success switching a,b);` with the correct `print("success switching a, b");`. (Human modification; AI is forbidden to touch frozen tests without user intervention.)
3. **Parser fix** in `parseDeclaration`:
   - Detect `:` branch.
   - Read first value via `parseExpression`.
   - Loop on `,` to collect parallel initial-value expressions into `Values`.
   - Expect `∈` or `in` next.
   - Read type identifier into the new `TypeAnnotation` field.
   - Pop `;`.

## Why I am NOT Implementing Now

Same anti-loop posture as Issue 17: the spec needs to be locked first, then the parser change requires confirmation that the `:` operator here does NOT collide with the `:` in `for i := 1..10:` (range notation), nor with the `:` used in `match targets :` or in Named-arg `sep: "..."`. Deciding that these are different contexts is a decision, not a code change.

## Acceptance Criteria (when implemented)
- `test/level1/T0122.bee` runs green.
- `new a: 1, b: 2 ∈ Z; let a, b := b, a;` correctly swaps them.
- `new name: "Alice" ∈ U;` declares a single string variable initialised to `"Alice"` of type `U`.

## Anti-Loop Note
T0122 has a string-literal typo. AI must NOT silently fix that. The user must modify the test source manually before the grammar fix can be verified.
