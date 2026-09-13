# Issue 17 — Radical operator precedence ambiguity (`²√9 = 3` hangs)

## Status
**Open — failing test (T0107 disabled).** Spec gap + parser gap.

**Awaiting ratification of D10 (`todo/DECISIONS.md`).** Once the user
ratifies the precedence table (Unary > Power > Mul/Div > Add/Sub >
Range > Compare > Logic), `parseExpression` is refactored to use
precedence climbing rather than the current naive for-loop. Do NOT
implement before ratification.

## Observed Behaviour

```bee
rule main:
  new x ∈ Z;
  let x := 27;
  expect ²√9 = 3;
  expect ³√x = 3;
  expect ³√ 2³ = 2;
  expect (1 + 2)³ = 27;
return;
```

The compiler halts at `expect ³√ 2³ = 2;` with `E0009 SyntaxError:UnrecognizedStatement: line=9 type=; literal=";"`. Earlier lines parse, but the chain `³√ 2³ = 2` is rejected.

## Lexer Trace (debug `-d`)

```
[TOKEN] Type: EXPECT          | Literal: expect     | Pos: 10
[TOKEN] Type: √               | Literal: ³√         | Pos: 10
[TOKEN] Type: INT             | Literal: 2          | Pos: 10
[TOKEN] Type: ^               | Literal: ³          | Pos: 10
[TOKEN] Type: =               | Literal: =          | Pos: 10
[TOKEN] Type: INT             | Literal: 2          | Pos: 10
[TOKEN] Type: ;               | Literal: ;          | Pos: 10
PARSER DEBUG: Parsing expr, token: "³√" type: √
DEBUG: Parsing expr, token: "2" type: INT
DEBUG: In binary loop, peekTok type: ^, literal: "³"
DEBUG: Parsing expr, token: "=" type: =
DEBUG: In binary loop, peekTok type: INT, literal: "2"
DEBUG: In binary loop, peekTok type: INT, literal: "2"
PARSER DEBUG: unrecognized top-level token type=INT literal="2" line=10
```

## Root Cause — Two Compounding Bugs

### Bug A — Spec gap: radical grammar is undocumented
- `spec/01-lexical-structure.md` §3.4 lists `√` only as a mutation operator (`√=`) and as arithmetic op `arith_op ::= "√"`.
- There is NO `nth_root` rule. The Unicode superscript radicals `²√`, `³√`, `⁴√` … `⁹√` are not in `spec/02-statements.md` §5 EBNF.
- The test `T0107` exists because the implementation supports them, but the docs do NOT authorise them. This violates the project invariant "no implementation without spec".

### Bug B — Parser gap: prefix-radical bypasses the binary loop
`internal/parser/parser.go::parseExpression` lines 800-803:
```go
if tok.Type == token.SQRT || tok.Literal == "²√" || ... {
    right := p.parseExpression()
    return &BinaryExpression{Token: tok, Left: &Identifier{...}, Right: right}
}
```
The `return` is **early-bail**: when a prefix-radical is parsed, control returns to the caller WITHOUT going through the for-loop that would consume subsequent operators. The CALLER's for-loop then sees the following `^` (power), but if a third operand is involved (`³√ 2³ = 2`), the recursive `parseExpression` for "2" terminates at peek `^` (it returns), leaving `^` to be consumed by the outer for-loop which then chains through `=` and the trailing `2` as if they were all at the same level.

### Bug C — Precedence is unspecified
`²√x` should bind TIGHTER than `^`, `<`, `=`. There is no precedence table in the EBNF; everything is processed left-to-right associatively. So:
- `²√9 = 3` is ambiguous: is it `(²√9) = 3` or `²√(9 = 3)`? Mathematically only the first makes sense.
- `³√ 2³ = 2` is ambiguous: `(³√(2³)) = 2` (correct) or `(³√2)³ = 2` (wrong but parseable).

## Proposed Resolution (one pass — NOT yet implemented)

1. **Add precedence table to `spec/02-statements.md` §5.** Bind tightness high→low:
   ```
   1. unary:  ¬, √, n√ (n-th root)
   2. power:  ^, ³ (superscript exponent)
   3. mul/div: × ÷ / * %
   4. add/sub: + -
   5. range:   .. .!
   6. compare: = <> < > <= >= ≈
   7. logic:   and ∧ or ∨ xor ⊕
   ```
2. **Add `nth_root_op` to `internal/token/token.go`** with literal `"²√"…` distinctions coded as a single `SQRT_PREFIX` TokenType carrying the `Literal` (since the degree is part of the operator, not a separate token).
3. **Refactor `parseExpression`** to use precedence climbing (not naive loop). The prefix-radical returns to the caller, but the caller must be in a state where the `for` loop can see a binary `^` and recurse into a precedence-2 parser.

## Why I am NOT Implementing Now

Per AGENTS.md Anti-Loop Protocol: "NEVER attempt more than one edit pass per user prompt" + "Before applying an edit, explain the root cause identified by comparing the implementation against the relevant `/spec`."

Root-cause analysis is documented above. **Implementation requires three coordinated edits**: spec EBNF (locked-in decision needed), AST (potentially `PrefixExpression` node), and parser (precedence climbing rewrite). That exceeds the scope of a single "one edit pass" — user must approve the precedence table before I rewrite the parser.

## Acceptance Criteria (when implemented)
- `T0107` runs green and is promoted back to `test/level1/T0107.bee`.
- `²√9 ^ 2` evaluates as `(²√9) ^ 2 = 9` (radical binds tighter than `^`).
- `³√125 = 5` parses and evaluates.
- `²√(a + b)` parenthesised form works (parser already supports this).

## Spec / Tutorial Cross-References
- `bee-tutorial/library.html` mentions `sqr` (square root) but no `²√` / `³√` — **tutorial gap**.
- `spec/01-lexical-structure.md` §3.4 has no entry for n-th root.
- `spec/02-statements.md` §5 has no precedence table.
