# Bee Compiler — Decisions Backlog

This file is the canonical backlog of **user-locked decisions** governing the
Bee language design. A *decision* is a normative choice that resolves an open
question in `/spec` or `/tracking/issues/`. Implementing agents MUST consult this
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
| D0    | Architectural constraints (config/AGENTS.md §2) | ✅ Ratified       | 2026-09-12  |
| D1    | 1-based indexing at source/parser/AST       | ✅ Ratified       | 2026-09-12  |
| D2    | Identity vs mutation (`is`/`is not`, `+=`)  | ✅ Ratified       | 2026-09-12  |
| D3    | Operator canonicalisation (`<>` vs `≠`)     | ✅ Ratified       | 2026-09-12  |
| D4    | Typechecker posture (postpoend)             | ✅ Ratified       | 2026-09-12  |
| D5    | Tutorial sync (`/tutorial/` versioned local)| ✅ Ratified       | 2026-09-12  |
| D6    | Curried rule signatures `(sep: ...)`        | ✅ Ratified       | 2026-09-12  |
| D7    | Logic operator synonymy + `is not` token    | ✅ Ratified       | 2026-09-12  |
| D8    | Rule-call result destructuring (T0127 park) | 🟢 Deferred       | 2026-09-12  |
| D9    | Colon `:` pair-up semantics                 | ✅ Ratified       | 2026-09-13  |
| D10   | Radical (`²√`, `³√`) precedence table      | ✅ Ratified       | 2026-09-14  |
| D11   | Parallel colon-init `new (a,b):(1,2)`      | ✅ Ratified       | 2026-09-14  |
| D12   | Inequality/NOT operator refactoring        | ✅ Ratified       | 2026-09-13  |
| D13   | Range operator + postfix-step refactor     | ✅ Ratified       | 2026-09-13  |
| D14   | Uniform `done` terminator + `repeat` jump  | ✅ Ratified       | 2026-09-13  |
| D15   | `next` canonical loop-jump; `repeat` deprecated | ✅ Ratified   | 2026-09-13  |
| D16   | Diagnostic code registry (`registry/diagnostics.json`) + `E04xx` de-collision | ✅ Ratified | 2026-09-14 |

---

## D0 — Architectural constraints

**Status:** ✅ Ratified 2026-09-12.
**Source of truth:** `config/AGENTS.md` §2.

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

## D5 — Tutorial sync — `/tutorial/` is a versioned local mirror

**Status:** ✅ Ratified 2026-09-12. De-linked from junction 2026-09-14.
**Source of truth:** MANIFEST.md "Decision 5".

* `/web/` has been removed; there is no legacy local copy.
* `/tutorial/` is a **versioned local directory** in this repo — the
  source of truth for the rendered HTML pages — *not* a symlink/junction.
* It is mirrored one-way into the external SCL site directory
  (`C:\Users\eluci\sage-code\scl\projects\bee`) via `sh run.sh sync`
  (backing script `scripts/sync_tutorial.py`).
* Push (local → SCL) is the default; `sh run.sh sync pull` reverses direction.
* `.github/` exclusions ensure no GitHub-action workflow auto-mirrors
  the legacy copy.
* All future documentation edits land in `/tutorial/` only.

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
**Source of truth:** `tracking/issues/16-rule-tuple-destructure.md`, parked
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

## D10 — Radical operator precedence (USER-CLARIFIED 2026-09-12, ratified 2026-09-14)

**Status:** ✅ Ratified 2026-09-14.
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

### Implementation consequence (✅ COMPLETED 2026-09-14)

* **AST:** `PrefixExpression` node (`Token`, `Operator`, `Right`)
  added — the canonical nullary-prefix form for radicals (`√`, `ⁿ√`)
  and logical NOT (`¬`, `not`). The old `BinaryExpression{Left:
  &Identifier{"0"}, …}` hack is gone.
* **Parser:** `parseExpression` rewritten as a precedence-climbing
  loop (`parseExpressionClimb`) with the level table above; the
  bare `√` / suffix `²√` `³√` `ⁿ√` are consumed in `parsePrimary` at
  `minPrec=precPower`, so `³√ 2³ ≡ ³√(2³)` yields
  `Prefix(radical, Binary(2,^,3))`.
* **Evaluator:** `evalIntExpressionWithID` dispatches radical
  operators to `pow(operand, 1/n)`.
* **Test:** T0107 enabled and PASS (all eight radical assertions,
  including `³√ 2³ = 2`).
* **Verification:** `sh run.sh solo T0107` green; smoke PASS.

---

## D11 — Parallel colon-initialisation

**Status:** ✅ Ratified 2026-09-14.
**Affects tests:** T0122 (currently disabled).

### User directive (verbatim, 2026-09-13)

> *"new a, b: 1, 2 ∈ Z should be new (a, b) : (1, 2) ∈ Z or
> new a:1, b:2 ∈ Z"*

### Rejected form

The bare parallel list form is **rejected**:

```bee
new a, b: 1, 2 ∈ Z;   -- INVALID
```

Rationale: `:` is the pair-up operator (D9) binding **one** name to
**one** value. Two bare comma-lists around a single `:` hide the
pairing and read ambiguously against `expr_list` elsewhere in the
grammar. The previously proposed `set` parallel form is dropped with
it.

### Accepted forms

1. **Parenthesised parallel form** (new grammar — the substance of D11):

```bee
new (a, b) : (1, 2) ∈ Z;
```

```ebnf
decl_stmt ::= "new" "(" ident_list ")" ":" "(" expr_list ")" ( "∈" | "in" ) type_specifier ";"
```

The parenthesised `ident_list` (N names) and `expr_list` (M values)
MUST have equal length (N = M); binding is positional (`a : 1`,
`b : 2`), all sharing the trailing `∈ Type` qualifier. The `:` form
is **not** type-inferring — the RHS values are treated as literals
of the specified type.

2. **Repeated pair-up form** (already ratified under D9 — no new
grammar needed):

```bee
new a: 1, b: 2 ∈ Z;
```

This is the D9 `pair_up_list` production, already implemented and
verified by active test T0121.

When ratified:
* `parseDeclaration` gains a paren-detect branch: after `new`, a
  `(` token enters the parallel form — parse `ident_list`, expect
  `)`, `:`, `(`, `expr_list`, `)`, then `∈` / `in` + type identifier.
* A static arity check (E0009-class diagnostic) enforces N = M.

### Implementation consequence (✅ COMPLETED 2026-09-14)

* Parallel colon-initialisation `new (a, b) : (1, 2) ∈ Z;`
  implemented; acceptance test T0139 ("Decision 11: parenthesised
  parallel form") enabled and PASS in level1.
* The D9 repeated pair-up form `new a: 1, b: 2 ∈ Z;` (T0121) and the
  parallel-assignment form `let a, b := b, a` (T0122) are separate,
  already-active PASS cases and are unaffected by D11.

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

### Ratified semantics (design-of-record)

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

## D14 — Uniform `done` terminator + `repeat` repurposed as jump

**Status:** ✅ Ratified 2026-09-13, spec harmonized 2026-09-13.
**Supersedes:** the pre-D14 overloaded `repeat` keyword that served as
both block terminator and post-conditioned loop-continuation modifier.
**Source of truth (final):** `spec/02-statements.md` §3.4, §5, §6, §7.

### User directive (verbatim)

> *"I have decided again about the semantic of control statements: Uniform
> Block Termination (`done`): Eliminating `repeat` as a block closer creates
> a single closure rule across all AST control nodes. Repurposing `repeat`:
> Demoting `repeat` to an inline jump statement inside the loop body gives
> it an unambiguous operational meaning (re-evaluate / step to next
> iteration) equivalent to `continue` in C-family languages. Anonymous
> Scope Prologue (`cycle:`): Decoupling stable scope allocation from
> obligatory label identifiers removes arbitrary naming overhead while
> preserving two-tier scope lifetime."*

### Ratified semantics

#### 1. Uniform Block Termination (`done`)
All control blocks (`cycle`, `for`, `if`, `match`, `start`, `with`,
`trial`) terminate with **`done [label];`**. The `repeat` keyword is
removed as a block terminator entirely.

| Control Block | Previous Terminator | Updated Terminator |
| :--- | :--- | :--- |
| `start` / `with` | `done` | `done` (unchanged) |
| `if` / `else` / `ladder` | `done` | `done` (unchanged) |
| `match` | `done` | `done` (unchanged) |
| `trial` | `done` | `done` (unchanged) |
| `cycle` | `repeat [label]` | **`done [label]`** |
| `for` | `repeat [label]` | **`done [label]`** |

#### 2. Optional Label, Optional Scope (colon = prologue marker)
The `:` (colon) is the **prologue marker**, decoupled from label binding.
Label and colon are independent options — `cycle [label] [:]`:

| Form | Label? | Prologue? | Semantics |
| :--- | :--- | :--- | :--- |
| `cycle name:` decls `do` body `done name;` | ✅ | ✅ | Labeled with stable prologue |
| `cycle:` decls `do` body `done;` | ❌ | ✅ | Anonymous with stable prologue |
| `cycle name` `do` body `done name;` | ✅ | ❌ | Labeled volatile body (label is jump target only) |
| `cycle do` body `done;` | ❌ | ❌ | Anonymous volatile body |
| `cycle while cond do` body `done;` | ❌ | ❌ | While-loop volatile body |

Variables declared in the prologue persist across all iterations.
Variables declared in the `do` body are volatile (re-created per pass).

#### 3. `repeat` Repurposed as Inline Jump
`repeat` is demoted from block terminator to inline transfer statement:

```ebnf
jump_stmt ::= ( "repeat" | "stop" | "redo" ) [ label ] [ "if" condition ] ";" ;
```

- **`repeat`** — skip remainder of current iteration body; jump to
  loop-header re-evaluation. In a `for` loop, advance to next element.
- **`stop`** — terminate the loop immediately; transfer execution past
  `done [label];`.
- **`redo`** — restart current iteration without advancing the iterator
  (non-`for` cycles only).

#### 4. `next` Retirement
The `next` keyword is retired; its semantic (advance to next iteration)
is absorbed by `repeat`. No grammar production retains `next`.

#### 5. `then` Post-Loop Epilogue
The optional `then` block executes exactly once after the loop exits
(via `stop` or condition exhaustion), before the prologue scope is
popped. It does NOT execute between iterations.

### EBNF changes (spec/02-statements.md §5)

```ebnf
cycle_stmt  ::= "cycle" [ label ] [ ":" decl_block ]
                ( "do" | "while" expression "do"
                | "for" [ "∀" ] identifier ( "∈" | "in" ) expression "do" )
                block
                [ "then" block ]
                "done" [ label ] ";"
              | "for" [ "∀" ] identifier ( "∈" | "in" ) expression "do"
                block "done" ";" ;

(* colon ⇒ prologue, with or without label; label alone ⇒ jump target only *)
decl_block  ::= { declaration_stmt } ;
```

```ebnf
transfer_stmt ::= ( "return" [ expression_list ]
                  | "stop" [ label ]
                  | "redo" [ label ]
                  | "repeat" [ label ]
                  | "pass"
                  | "raise" [ expression ]
                  | "resume"
                  | "retry"
                  | "fail" expression ) [ "if" expression ] ;
```

### Diagnostic codes
- `E0203` — updated: "Missing `done` or `return` terminator" (removed
  `repeat`)
- `E0205` — updated: "Closing label on `done` does not match opening
  header label"
- `E0206` — new: `InvalidJumpContext` — `repeat`/`stop`/`redo` outside
  a loop body

### Implementation consequence (deferred)
Compiler changes (`internal/lexer/`, `internal/parser/`,
`internal/evaluator/`) are **not** in this pass. The spec and tutorial
are now 100% harmonized with D14. The next implementation phase must:
1. Update `internal/parser/parser.go` to enforce `done` terminators.
2. Treat `repeat` as a `TransferStatement` (jump), not a block closer.
3. Remove `next` from the token dispatch.
4. Update `internal/evaluator/evaluator.go` for the new loop-jump
   semantics and `then` epilogue.

> **Superseded (2026-09-13):** Items 2–3 are reversed by **Decision 15**
> below — `next` is canonical; `repeat` is retained as a deprecated
> synonym lexing to `NEXT`.

---

## D15 — `next` canonicalized as the loop-jump keyword; `repeat` deprecated

### User directive (verbatim)

> *"The repeat keyword should be replaced by 'next' — this will make the
> language more consistent. A single word instead of continue is shorter.
> `next [label]` jumps directly to loop header / re-evaluates while
> condition; advances iterator to next domain element and re-evaluates
> domain boundaries. `stop [label]` exits the loop block immediately past
> done. `redo [label]` (optional) restarts the current iteration body
> without re-evaluating the condition; re-runs the body for the current
> iterator value without advancing the domain."*

### Ratified semantics

1. **`next` is canonical.** The loop-jump statement (continue semantics)
   is spelled `next [label] [if condition];`. This reverses the D14
   retirement of `next`.
2. **`repeat` is a deprecated synonym.** The keyword remains in the
   token dispatch; the lexer maps it to the `NEXT` token and emits a
   non-fatal `E0010 deprecated-keyword: 'repeat' — use 'next'` warning
   (identical mechanism to the D7/D12/D13 deprecation surfaces). It
   hardens to `E0009` in the Phase 7.2 sweep.
3. **Synonymy preserves the frozen suite.** `@FROZEN` tests in
   `test/levelX/` that spell the jump `repeat` keep passing — the parser
   sees only `NEXT`. No frozen test is modified.
4. **`stop` and `redo` are unchanged.** The jump table is now:

| Keyword | While-cycle behaviour | For-cycle behaviour |
| :--- | :--- | :--- |
| `next [label]` | Jump to loop header; re-evaluate condition | Advance iterator to next domain element; re-evaluate domain boundaries |
| `stop [label]` | Exit loop block immediately past `done` | Exit loop block immediately past `done` |
| `redo [label]` | Restart current iteration body without re-evaluating the condition | Re-run body for current iterator value without advancing the domain |

### Implementation consequence (✅ lexer COMPLETED 2026-09-13)

* `internal/lexer/lexer.go`: `NextToken` intercepts the identifier
  `repeat`, emits `E0010` via the lexer warning surface, and returns
  `token.NEXT` with literal `"next"`. The `"repeat": REPEAT` entry in
  `internal/token/token.go` Keywords is retained for lookup stability.
* `internal/parser/parser.go`: `parseTransferStatement` already
  dispatches `token.NEXT` (and `token.REPEAT` defensively); both accept
  the optional label and `if` guard.
* Docs pass (this entry): `spec/02-statements.md` §1/§3.4/§5/§7,
  `spec/00-memory-model.md` §2.2, `tutorial/control.html`,
  `tutorial/syntax.html`, `tutorial/js/bee.js`, `MANIFEST.md`,
  `tracking/issues/20-next-canonical-jump.md`, `tracking/solutions/20-next-canonical-jump.md`.

## D16 — Central diagnostic-code registry (JSON) + `E04xx` de-collision

### User directive (verbatim, 2026-09-14)

> *"create a special file registry in spec with all error codes in a json
> file. Then if the error code collide, bump the error code, register new
> codes."* (operator clarification in the same session: `=` is the only equal
> operator — value equality; `@a = @b` compares two references; `≡` is
> reserved strictly for geometric congruence.)

### Ratified semantics

1. **`registry/diagnostics.json` is the single source of truth** for every
   `E`/`W` diagnostic code referenced in `/spec` and `/tutorial`.
2. **Register first, then use.** No diagnostic table is injected into a spec
   module or tutorial page before its codes are registered.
3. **Collision policy.** If a proposed code would collide with a registered
   one, bump the newer / less-canonical use to the next free slot in its
   module block and re-register. Never silently reuse an occupied code with a
   different meaning.
4. **`E04xx` de-collision (applied).** `spec/04-structure.md` retains the
   canonical `E04xx` block. `spec/00-memory-model.md` (module `00`, which has
   no natural `E00xx` block) was bumped from `E0401`–`E0404` to the free
   `E08xx` block (`E0801`–`E0804`).
5. **Operator taxonomy (reaffirmed).** `=` is the single value-equality
   operator; `@a = @b` is reference/pointer identity; `≡` is geometric
   congruence only and never value equality.

### Implementation consequence (registry created 2026-09-14)

* Created `registry/diagnostics.json` capturing all global codes (`E0009`,
  `E0010`, `E0011`, `W0901`) and every module block `E01xx`–`E14xx` (with
  memory on `E08xx`), plus `registry/README.md` documenting the allocation
  table and collision policy.
* Updated `spec/00-memory-model.md` (§2.2 in-text `E0801`; §5 table) and
  `tutorial/memory.html` (diagnostic table + resolved-note replacing the
  former "E04xx overloaded" warning).
* Audit trail in `tracking/todo/TUTORIAL_TODO.md` §3.3.

---

## Resolution Workflow

When the user picks a decision to resolve next, the AI MUST:

1. **ONE EDIT PASS.** Do not iterate silently.
2. Update `/spec/` to formalise the grammar.
3. Update `MANIFEST.md` to reflect the ratified status.
4. Update `/tutorial/` HTML files if the surface area changed.
5. Park the corresponding debt test or unblock it as appropriate.
6. Run `sh run.sh smoke` for system-wide health check before yielding.
7. **Yield back** with status report.

If the user's chosen decision has dependent decisions (e.g. D9 unlocks
D11), surface that in the status report. Do NOT auto-resolve
downstream decisions.

---

*Last updated: 2026-09-14 — D1–D7, D9, D10, D11, D12, D13, D14, D15, D16
ratified (D3 superseded by D12; D14 items 2–3 superseded by D15); D8
is the sole deferred decision.*
