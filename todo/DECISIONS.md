# Bee Compiler — Decisions Backlog

This file is the canonical backlog of **user-locked decisions** governing the
Bee language design. A *decision* is a normative choice that resolves an open
question in `/spec` or `/issues/`. Implementing agents MUST consult this
backlog before introducing any new operator, statement, keyword, or grammar
production.

> **Resolution protocol:** The user resolves one decision at a time. Each entry
> has a `Status` field (`✅ Ratified`, `🟡 Pending ratification`,
> `🔴 Open debate`, `🟢 Deferred`). Items below `D0` are ratified and live in
> `MANIFEST.md` § "User-Locked Decisions" by reference.

---

## Index of Decisions

| ID    | Topic                                       | Status            | Date        |
| ----- | ------------------------------------------- | ----------------- | ----------- |
| D0    | Architectural constraints (GEMINI.md §4)    | ✅ Ratified       | 2026-09-12  |
| D1    | 1-based indexing at source/parser/AST       | ✅ Ratified       | 2026-09-12  |
| D2    | Identity vs mutation (`is`/`is not`, `+=`)  | ✅ Ratified       | 2026-09-12  |
| D3    | Operator canonicalisation (`<>` vs `≠`)     | ✅ Ratified       | 2026-09-12  |
| D4    | Typechecker posture (postpoend)             | ✅ Ratified       | 2026-09-12  |
| D5    | Tutorial sync (`/bee-tutorial/` live)       | ✅ Ratified       | 2026-09-12  |
| D6    | Curried rule signatures `(sep: ...)`        | ✅ Ratified       | 2026-09-12  |
| D7    | Logic operator synonymy + `is not` token    | ✅ Ratified       | 2026-09-12  |
| D8    | Rule-call result destructuring (T0127 park) | 🟢 Deferred       | 2026-09-12  |
| D9    | Colon `:` pair-up semantics                 | ✅ Ratified       | 2026-09-13  |
| D10   | Radical (`²√`, `³√`) precedence table      | ✅ Ratified       | 2026-09-12 |
| D11   | Parallel colon-initialisation (T0121/0122) | 🟡 Pending ratification | 2026-09-12 |
| D12   | Inequality/NOT operator refactoring        | ✅ Ratified       | 2026-09-13  |
| D13   | Range operator + postfix-step refactor     | 🟡 Pending ratification | 2026-09-13 |

---

## D0 — Architectural constraints

**Status:** ✅ Ratified 2026-09-12.
**Source of truth:** `GEMINI.md` §4.

Idiomatic Go throughout. No `panic` in compiler error flows; explicit
`error` returns. 1-based indexing at source, lowered to 0-based only in
`internal/compiler/`. Every LLVM basic block MUST terminate explicitly.
Debug logs go to `os.Stderr`, `stdout` reserved for clean build outputs.

---

## D1 — Indexing basis

**Status:** ✅ Ratified 2026-09-12.
**Source of truth:** MANIFEST.md "Decision 1".

Bee source is **1-based** for lists, arrays, matrices, and string indices.
`$` is the dynamic **end anchor** (index of the last element), not a
zero terminator. All compiler passes (lexer, parser, typechecker,
evaluator) use 1-based semantics. The bridging pass in
`internal/compiler/` lowers with the invariant
`ir_index = source_index - 1` and `$` → `len(c) - 1`.

---

## D2 — Identity vs mutation

**Status:** ✅ Ratified 2026-09-12.
**Source of truth:** MANIFEST.md "Decision 2".

* `{is, is not}` are pointer-identity operators (Decision 7 confirms
  `is not` is a single lexeme, not a sequence).
* `+=`, `-=` are **lex-level shorthand** for `+:=`, `-:=` — they map
  to the same AST node and produce identical semantics.
* `+is`, `-is` are NOT operators; proposals to introduce them are
  rejected as nonsense (literal-vs-pointer confusion).

---

## D3 — Operator canonicalisation

**Status:** ⛔ SUPERSEDED by D12 (2026-09-13). Retained for history.
**Source of truth:** MANIFEST.md "Decision 3" (historical).

| Legacy            | Canonical   | Status                                    |
| ----------------- | ----------- | ----------------------------------------- |
| `≠` (U+2260)      | `<>`        | Deprecated; lexer emits E0010 → E0009     |
| `==`              | `=`         | Reserved; emitted as warning              |
| `!=`              | `<>`        | Removed; replaced by `<>`                 |
| `≡` (U+2261)      | `≡`         | Reserved for graphics congruence (§13)    |

> D12 replaces the canonical inequality form `<>` with `¬` and
> repoints every deprecated alias (`≠`, `!=`, `<>`) at `¬`. The `≡`
> reservation for graphics congruence is unchanged.

---

## D4 — Typechecker posture

**Status:** ✅ Ratified 2026-09-12.
**Source of truth:** MANIFEST.md "Decision 4".

* Phase 5 deliverable reduced to **symbol table + scope lookup only**.
* Full type inference is **deferred until Phase 6 IR lowers 1-based
  indexing cleanly**. We do NOT implement inference against runtime
  values — TS-AST declarations register a `Kind` and a `TypeHint`,
  but no flow analysis.

---

## D5 — Tutorial sync — `/bee-tutorial/` is live

**Status:** ✅ Ratified 2026-09-12.
**Source of truth:** MANIFEST.md "Decision 5".

* `/web/` is a legacy local copy (will be discarded).
* `/bee-tutorial/` is the canonical **live** repository target, with
  direct git push propagation to the user's public site.
* `.github/` exclusions ensure no GitHub-action workflow auto-mirrors
  the legacy copy.
* All future documentation edits land in `/bee-tutorial/` only.

---

## D6 — Curried rule signatures

**Status:** ✅ Ratified 2026-09-12.
**Source of truth:** `spec/03-rules.md` §2.4 + §3.1.

Rules MAY declare an optional *named-parameter slot* after the
positional list. The slot form is

```
rule foo(*args)(sep: ", " ∈ Str);
```

Call sites apply the slot via a second parenthesised named-argument
list:

```
print(a, b)(sep: " | ");
```

* Named arguments are **order-independent**.
* Defaults fall through if the slot is omitted.
* Legacy `print a using sep ":"` is **deprecated**; lexer emits E0011
  until Phase 7 audit 7.2 hardens it to E0009.

---

## D7 — Logic-operator synonymy and `is not` token

**Status:** ✅ Ratified 2026-09-12.
**Source of truth:** `spec/01-lexical-structure.md` §3.6 /
`internal/token/token.go` / `internal/lexer/lexer.go`.

* `is not` is **one token**. The lexer performs Maximal Munch
  collapsing `is` followed by whitespace + `not` into `IS_NOT`.
* Logical operators: `and` / `or` / `xor` / `not` are accepted as
  ASCII synonyms for `∧` / `∨` / `⊕` / `¬`. The **canonical citation
  form** is the ASCII keyword (Decision 7 mandate). The Unicode
  glyphs remain functional and pass through the same parser path.
* Legacy `≠`, `==`, `!=` are tolerated but emit **non-fatal E0010**
  deprecation warnings; Phase 7.2 hardens them to E0009.

---

## D8 — Rule-call result destructuring

**Status:** 🟢 Deferred to Phase 8.4.
**Source of truth:** `issues/16-rule-tuple-destructure.md`, parked
debt test `test/debt/T0127-rule-call-result-binding.bee`.

Grammar for `new a, b, c := zeros();` where `zeros()` returns a tuple
is unresolved. Awaiting a user decision on **tuple-return rule
syntax** (`=> (a ∈ Z, b ∈ R)`) before the AST change (`RuleCall`
+ multi-binding destructuring) can be implemented.

---

## D9 — Colon `:` is the pair-up operator

**Status:** ✅ Ratified 2026-09-13.
**Affects tests:** T0121 (promoted to active level1); T0122 remains
disabled pending D11 + a user-side string-literal typo fix.

The user has proposed and we are ratifying the following semantics for `:`:

```
:`  is the canonical "pair-up" or "binding" operator.
   It associates a left-side key/name with a right-side value
   WITHOUT triggering type inference. It is purely structural.
```

**Semantics by context (all 1-based semantics apply):**

| Syntactic context                             | Left operand           | Right operand        | Result                                        |
|----------------------------------------------|-----------------------|----------------------|-----------------------------------------------|
| `rule foo:` / `if a > 0 do` / `cycle for:`   | block-init head keyword | (nothing)            | Block-head delimiter — NO value binding.       |
| `func(n: Z)` / `rule print(*xs)(sep: U)`     | formal parameter name  | type annotation      | Parameter → declared type.                    |
| `print(a, b)(sep: " | ")`                    | named-argument name    | expression value     | Binding a value to a named parameter slot.    |
| `{ "k": "v" }` (map literal)                 | key expression         | value expression     | Entry in a map literal.                        |
| `new name: "Alice" ∈ U;`                     | identifier (new)       | expression literal   | Single-var parallel assignment, typed literal.|
| `new a: 1, b: 2 ∈ Z;`                        | identifier list        | expression list      | Parallel typed declaration (Issue 18).        |
| `let x: 0 ∈ N;`                              | identifier (let)       | typed literal         | Re-binding to a typed literal (no inference).  |
| `label:` block_head_keyword                  | label identifier       | (nothing)            | Block-label scope anchor.                     |

**Comparison to `=` and `:=`:**

* `=`  — value equality (logic operation). Replaces legacy `==`.
* `:=` — type-inferred assignment. Compiler infers type from the RHS
         expression and allocates the variable accordingly.
* `:`  — **structural pair-up**. NO type inference, EVER. The RHS is
         a literal of the type supplied by an explicit type context:
         in a parameter slot the function's signature governs; in a
         `new`/`let` declaration the trailing `∈ Type` is REQUIRED.
         If no type is specified, `:` CANNOT be used to set the value
         of a variable — use `:=` (type inference) instead. Writing
         `new a: 1;` without `∈ Type` is a compile-time error
         (E0009), not an invitation to guess the type from the
         literal shape. Map-literal entries (`{ "k": "v" }`) are not
         variable bindings, so this rule does not apply to them.

**Consequence for D8:** the `:=` new+let pair-up with `:` collapses
into a single family once D9 is ratified; the two forms are
mutually exclusive rather than ambiguous: `new a := 1` (type
inferred from the RHS) is the ONLY way to declare without a type,
while `new a: 1 ∈ Z;` (explicit type) is the ONLY way to use `:`
in a declaration. Bare `new a: 1;` is rejected.

**When ratified:** Issue 18 unlocks; T0121 + T0122 are promoted back
to active level1; `parseDeclaration` `:` branch is rewritten to
build a proper `DeclarationStatement` with N parallel bindings +
REQUIRED trailing `∈ Type` (per the `Issue 18` resolution path);
a `:` binding without `∈ Type` raises E0009 directing the user
to `:=`.

---

## D10 — Radical operator precedence (USER-CLARIFIED 2026-09-12)

**Status:** ✅ Ratified 2026-09-12.
**Affects tests:** T0107 (was disabled, now enabled and passing).

> **User statement (verbatim pivot):** *"Radical has high precedence
> but behind power. Power is executed first. The radical applies to
> its first term, but if the term has a power operator, the power is
> executed first then the sqrt. But radical has higher precedence
> than multiplication, division, or addition. The first term can be
> an expression starting with `()` — in that case priority changes
> and what is in parenthesis comes first. If the parenthesis has a
> power, the power executes first, then the sqrt over the result."*

### Precedence table (highest → lowest)

| Level | Token family                       | Binding                          |
|------:|------------------------------------|----------------------------------|
|   1   | `( … )` — matched grouping         | always resolves first            |
|   2   | power (`^`, `²`, `³`, `ⁿ`)        | binary, right-associative        |
|   3   | radical (`√`, `²√`, `³√`, … `ⁿ√`) | unary **prefix**, left-assoc.    |
|   4   | mul/div (`×`, `÷`, `/`, `*`, `%`) | binary                          |
|   5   | add/sub (`+`, `-`)                 | binary                          |
|   6   | range (`..`, `.!`)                 | binary                          |
|   7   | compare (`=`, `<>`, `<`, `>`, `<=`, `>=`, `≈`) | binary        |
|   8   | logic (`and`, `or`, `xor`, `not`, `∧`, `∨`, `⊕`, `¬`) | binary |

### Read-out of the rule

A radical `√` or `ⁿ√` **binds tighter than `* / + -` / range /
compare / logic** ("radical has higher power than multiplication,
division, or addition").

A radical **binds LOOSER than `^`** ("radical has high precedence
but **behind** power"). The radical's operand (the "first term") is
therefore evaluated as if `^` were inside it, and only THEN the
radical is applied to the result.

Parentheses inside the radical's operand override the rule: `(…)`
always resolves first, and any `^` *inside* the parens executes
before the radical wraps the result.

### Worked examples (anchors for the parser)

| Expression                | Parse                                              | Value     |
|---------------------------|----------------------------------------------------|-----------|
| `²√9`                     | `radical(9, 2)`                                    | `3`       |
| `³√x`  (with `x := 27`)   | `radical(x, 3)`                                    | `3`       |
| `³√ 2³`                   | `radical((2 ^ 3), 3)` — power binds inside radical | `2`       |
| `(1 + 2)³`                | power is suffix; behaves as `((1+2) ^ 3)`           | `27`      |
| `2³ √ 9`                  | `(2^3) op √9` — power first, then mul-tight radical | `8 * 3 = 24` |
| `²√144`                   | `radical(144, 2)`                                  | `12`      |
| `²√(x + y)`               | parens first inside radical; literal evaluation    | `²√(x+y)` |

### Implementation consequence

When ratified:

* **AST:** Introduce a `PrefixExpression` node (`Token`,
  `Operator` ∈ {`√`, `ⁿ√`}, `Right` Expression). The current
  hack `BinaryExpression{Left: &Identifier{"0"}, …}` in
  `internal/parser/parser.go::parseExpression` is replaced.
* **Parser:** Refactor `parseExpression` into a precedence-climbing
  parser. Levels are numbered per the table above. Prefix-radical
  is a **level-3 nullary prefix** that consumes one sub-expression
  parsed at the *radical's* level (≥ 3), which means the inner
  parser stops at level 4 boundaries: any `^` token inside the
  radical's operand is part of that operand because it binds at
  level 2 (tighter than level 3). This is exactly the precedence
  rule the user is ratifying.
* **Evaluator:** Translate `PrefixExpression{Operator: "²√"}` to
  `pow(operand, 1/2)` (or `pow(operand, 1/n)` for `ⁿ√`).
* **Test:** T0107 becomes the acceptance test. All eight
  assertions in the current disabled file must evaluate under the
  new precedence rules.

### Why we are NOT implementing now

The user is in the middle of clarifying D10. We capture the
clarification above, then yield back. When the user replies with
"`D10 ratified`" — or words to that effect — the implementation
team will:

1. Add a `PrefixExpression` node to `internal/ast/ast.go` (or
   `internal/parser/ast_nodes.go`).
2. Refactor `parseExpression` to a precedence-climbing loop using
   the table above.
3. Update the evaluator to dispatch `PrefixExpression`.
4. Enable T0107 and verify green via `sh run.sh solo T0107`.
5. Run `sh run.sh smoke` for system-wide health check.

---

## D11 — Parallel colon-initialisation

**Status:** 🟡 Pending ratification 2026-09-12.
**Affects tests:** T0121, T0122 (currently disabled).

The grammar extension to `decl_stmt` in
`spec/02-statements.md` §2.1:

```ebnf
decl_stmt ::= "new" ident_list ":" expr_list ( "∈" | "in" ) type_specifier ";"
            | "set" ident_list ":" expr_list ( "∈" | "in" ) type_specifier ";"
ident_list ::= identifier ( "," identifier )* ;
expr_list  ::= expression   ( "," expression  )* ;
```

When the length of `ident_list` (N) equals the length of `expr_list`
(N), each identifier is bound to the corresponding expression, and
all share the trailing `∈ Type` qualifier. The `:` form is **not**
type-inferring — the RHS values are treated as literals of the
specified type.

When ratified:
* `parseDeclaration` `:` branch is rewritten to read the parallel
  expression list, then expect `∈` / `in` + a type identifier.
* T0121 (`new a: 1, b: 2 ∈ Z;` simple typed pair-up) is promoted
  back. T0122 has a separate string-literal typo that requires
  user-side correction (the frozen test source has unterminated
  quotes).

---

## D12 — Inequality / logical-NOT operator refactoring

**Status:** ✅ Ratified 2026-09-13.
**Supersedes:** D3 (canonical inequality form).
**Source of truth:** `spec/01-lexical-structure.md` §3.3 + §3.6.

### User directive (verbatim)

> *"The symbol = is good, it makes a comparison of value. The symbol <>
> is opposite, represents not =, like in ADA, but I want to change. ¬
> represents now the 'not' but '!' also represents the logic 'not' so we
> duplicated the symbol not. I want to use ¬ instead of ≠ and <> — it
> will represent a binary operator between left and right and means
> different values. In Bee we do not have other operators for difference
> of references. We use @a = @b to check if a is b. For example
> a = True or a ¬ True will work. @a = True will fail because @a is
> always a reference and True is a constant. @a = @b will fail even if
> a = b if a is a different reference than b is."*

### Ratified semantics

| Token  | Arity        | Class              | Semantics                                             |
| ------ | ------------ | ------------------ | ----------------------------------------------------- |
| `=`    | binary       | Value equality     | Structural value comparison.                          |
| `¬`    | **binary**   | Value inequality   | Canonical inequality; logical negation of `=` over values. |
| `!`    | **unary**    | Logical NOT        | Canonical prefix negation; synonym: keyword `not`.    |
| `is` / `is not` | binary | Pointer identity | Allocation-identity comparison (D2, D7 unchanged).    |
| `@`    | unary prefix | Reference-of       | `@a = @b` ⟺ `a is b`; `@a = value` is always false.   |

### Deprecation mapping (lexer MUST emit E0010, map to canonical token)

| Legacy form | Emitted token | Warning                                              |
| ----------- | ------------- | ---------------------------------------------------- |
| `≠`         | `NEQ` (`¬`)   | `E0010 deprecated-symbol: '≠' — use '¬' (Decision 12)` |
| `!=`        | `NEQ` (`¬`)   | `E0010 deprecated-symbol: '!=' — use '¬' (Decision 12)` |
| `<>`        | `NEQ` (`¬`)   | `E0010 deprecated-symbol: '<>' — use '¬' (Decision 12)` |
| `==`        | `EQ` (`=`)    | `E0010 deprecated-symbol: '==' — use '=' (Decision 12)` |

Phase 7.2 audit will harden all four to `E0009`.

### Lexical constraints

* `¬` is now exclusively a **binary** operator. It MUST NOT be parsed as
  a prefix operator; prefix logical negation is `!` (or keyword `not`).
* `!` remains subject to Maximal Munch: `!.`, `!!`, `!∈` keep their
  range/set meanings; a standalone `!` lexes as `LOGICAL_NOT`.
* `!=` is matched before standalone `!` so the deprecated form still
  tokenizes as one `NEQ` lexeme.

### Implementation consequence (completed 2026-09-13)

* `internal/token/token.go`: added `NEQ = "¬"`; `LOGICAL_NOT` changed
  from `"¬"` to `"!"`; `NOT_EQ` / `NEQ_UNICODE` retained as deprecated
  literal constants.
* `internal/lexer/lexer.go`: `¬` → `NEQ`; `≠`/`!=`/`<>` → `NEQ` with
  E0010; standalone `!` → `LOGICAL_NOT`.
* `internal/parser/parser.go`: `NEQ` registered in `isBinaryOp` and
  `opPrecedence` at `precCompare`; `parsePrimary` LOGICAL_NOT prefix
  branch now handles `!`.
* `internal/evaluator/`: `evalComparison` and the binary-expression
  dispatch recognise `NEQ`; the prefix NOT dispatch recognises `!`.

---

## D13 — Range operator and step-postfix refactor

**Status:** ✅ Ratified 2026-09-13, implemented 2026-09-13.
**Supersedes:** the pre-D13 range-token set `.!` / `!.` / `!!`, and the
legacy colon-postfix `(min..max : step)` form. Both legacy families still
lex with E0010 deprecation warnings and map to their canonical
counterparts during the deprecation window.
**Source of truth (final):** `spec/01-lexical-structure.md` §2.1, §5
(`range_op`); `spec/05-types.md` §3.2, §3.3, §6 (`range_sep`,
`domain_expr`); `spec/02-statements.md` §3.4, §5 (`range_expr`).

### User directive (verbatim)

> *"The range say (1.!x) is valid syntax. But I want to discard the idea
> of .! and !. operators and use something else. Also the (min..max:step)
> is wrong. I will use (min..max)(step) instead for example (1..5)(0.1)
> will include (1, 1.1, 1.2 ...etc) and (1..5)(0.1)[3] will be valid
> and returning number 1.3. Instead of (a.!b) we use (a..<b) to exclude
> b or we can use (a>..b) to exclude a and (a>..<b) to exclude both."*

### Ratified semantics (design-of-record; pending ratification)

| Source syntax | Notation           | Endpoint inclusion | Replaces  |
| :------------ | :----------------- | :----------------- | :-------- |
| `a..b`        | $[a, b]$           | both inclusive     | unchanged |
| `a..<b`       | $[a, b)$           | left-inclusive, right-exclusive | `.!` |
| `a>..b`       | $(a, b]$           | left-exclusive, right-inclusive | `!.` |
| `a>..<b`      | $(a, b)$           | fully exclusive    | `!!` |

The postfix step operator changes from colon-postfix to function-call
postfix:

* Legacy: `(min..max : step)` — colon postfix inside the parens.
* New:    `(min..max)(step)`  — *postfix application* using a second
  parenthesised list after the range expression.

The new postfix `(min..max)(step)` is uniform across all four
endpoint-inclusion variants above and is independent of the
range-separator choice. Indexing into a stepped range
(`(1..5)(0.1)[3]` → `1.3`) is normative and synchronous with the
1-based indexing convention (Decision 1).

### Lexical constraints (Maximal Munch)

The new range separators are multi-rune ASCII constructions and MUST
be tokenised as single tokens for the largest match:

* `..`  → `RANGE_INCL`     (lex when not followed by `<`).
* `..<` → `RANGE_LEFT_INC` (lex preferentially over `..` when `<` follows).
* `>..` → `RANGE_RGHT_INC` (lex when `>` directly precedes `..`).
* `>..<`→ `RANGE_EXCL`     (lex preferentially over `>..` when `<` follows).

Because `>..` and `>..<` start with `>`, the lexer MUST consult
lookahead before emitting a `GT` (`>`) token in any context where a
range operator is grammatically legal. The same applies to `<`:
when `<` follows `..`, the lexer MUST NOT emit a separate `LT` token.

The postfix `(min..max)(step)` is not a new lexical token; it is
parsed at the parser level as a postfix application of a rule-call
to a range expression — exactly mirroring the curried named-argument
postfix defined in Decision 6 (`(min..max)(step)`). The lexer
MUST keep its current handling of `(`, `)`, `..`, `<`, `>`,
`!` (D12: standalone `!` → `LOGICAL_NOT`), independent of the
new postfix semantics.

### Parsing constraints

* The new range operators are infix binary operators at the same
  precedence tier as the previous `..` (D10: precedence slot
  *Range*, between additive and comparison).
* `..<`, `>..`, `>..<` have **left-to-right associativity** and
  bind with the same power as `..`. The parser's
  precedence-climbing refactor (Decision 10) requires only the
  new TokenType registrations; no new precedence tier is needed.
* The postfix step `(min..max)(step)` uses the same parse
  production as the curried named-argument postfix
  (Decision 6: `apply_expr ::= primary ( "(" expr_list ")" )*`)
  but accepts a *positional* expression list, not a named
  argument list. The parser MUST distinguish the two by
  inspecting the first token of the trailing paren list: a
  leading `identifier ":" …` is a named-argument slot; a
  leading literal or expression is a positional step.

### Implementation consequence (✅ COMPLETED 2026-09-13)

* `internal/token/token.go`: range separators updated to
  `RANGE_LEFT_INC = "..<"`, `RANGE_RGHT_INC = ">.."`,
  `RANGE_EXCL = ">..<"`. `RANGE_INCL` (`..`) unchanged.
* `internal/lexer/lexer.go`: `NextToken` performs Maximal Munch over the
  new three-/two-rune sequences; `PeekCharN` added for the 3-rune
  lookahead. Legacy `.!`, `!.`, `!!` forms still lex with E0010
  deprecation, mapping to the canonical token types.
* `internal/parser/parser.go`: `NOT` keyword registered as a `not`
  prefix-synonym (D7); new `isRangeSeparatorToken` helper;
  `SteppedRangeExpression` AST; postfix `(step)` parser-side on a
  parenthesised range expr. IDENT-peekChar bug also fixed.
* `internal/evaluator/evaluator.go`: `isRangeSeparatorToken` and
  `inRangePerOp` route `in (range)` membership through the four
  endpoint-inclusion variants. `steppedRanges` map binds
  SteppedRangeExpression values to identifiers; `steppedRangeValue`
  materialises `stepped[i]` under 1-based semantics (Decision 1).
* Test cases: `test/level1/T0131` (`..<`), `T0132` (`>..`), `T0133`
  (`>..<`), `T0134` (`(1..5)(1)[3]`), `T0135` (`!!` legacy) all
  `@FROZEN` and PASS.

* `internal/token/token.go`: replace `RANGE_LEFT_INC = ".!"`,
  `RANGE_RGHT_INC = "!."`, `RANGE_EXCL = "!!"` with
  `RANGE_LEFT_INC = "..<"`, `RANGE_RGHT_INC = ">.."`,
  `RANGE_EXCL = ">..<"`. `RANGE_INCL` (`..`) unchanged.
* `internal/lexer/lexer.go`: `NextToken` must be updated to
  perform Maximal Munch over the new three-rune sequences
  (`>..<` precedes `>..` precedes `>`, and `..<` precedes `..`
  precedes `<`).
* `internal/parser/parser.go`: register the new `RANGE_LEFT_INC`,
  `RANGE_RGHT_INC`, `RANGE_EXCL` TokenTypes in `isBinaryOp` and
  `opPrecedence`; add a `parseStepPostfix` pass that handles
  `(step_expr)` after any range expression.
* `internal/evaluator/evaluator.go`: dispatch the new
  RangeExpression variants with the requested endpoint
  semantics; `applyStep` lowers the postfix step into a
  stepped range, with `[i]` indexing returning the i-th
  element under Decision 1 (1-based).
* Test cases: create `test/levelX/T0131-…`, `T0132-…`, `T0133-…`
  in `test/levelX/` with `@FROZEN` header tags covering the four
  endpoint variants and the postfix-step semantics.

### Why we are NOT implementing now

Decision 13 has not yet been drafted into the normative `/spec/`
documents; per the anti-loop gate (`MANIFEST.md` line 18), no
lexer/parser/evaluator edits may begin until `/spec/` is
100% harmonised with this decision. The next step is **formalise
the spec** (this entry + spec edits) and **request user
ratification**.

---

## Resolution Workflow

When the user picks a decision to resolve next, the AI MUST:

1. **ONE EDIT PASS.** Do not iterate silently.
2. Update `/spec/` to formalise the grammar.
3. Update `MANIFEST.md` to reflect the ratified status.
4. Update `/bee-tutorial/` HTML files if the surface area changed.
5. Park the corresponding debt test or unblock it as appropriate.
6. Run `sh run.sh smoke` for system-wide health check before yielding.
7. **Yield back** with status report.

If the user's chosen decision has dependent decisions (e.g. D9 unlocks
D11), surface that in the status report. Do NOT auto-resolve
downstream decisions.

---

*Last updated: 2026-09-13 — D1–D7, D9, D10, D12 ratified (D3 superseded by
D12), D8 deferred, D11, D13 pending user ratification.*
