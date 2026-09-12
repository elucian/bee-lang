# Bee Specification: Rules & Subroutines (03-rules.md)

## 1. Architectural Philosophy & Purpose

In Bee, all subroutines, procedures, functions, constructors, and methods are unified under a single foundational primitive: the **`rule`**.

$$\text{Rule}: \quad (\mathbb{P}_1 \times \mathbb{P}_2 \times \dots \times \mathbb{P}_n) \longrightarrow (\mathbb{R}_1 \times \mathbb{R}_2 \times \dots \times \mathbb{R}_m)$$

- **Unified Primitive:** A `rule` acts as a pure mathematical function when stateless, a procedure when side-effecting, a constructor when initializing state, or an object method when attached to a closure frame.
- **Main Entry Orchestrator:** Every executable Bee main module MUST define a top-level `rule main:` which serves as the application entry point.
- **Explicit Result Naming:** Rules explicitly declare parameter names and result variable names in their signature, providing self-documenting signatures and zero-overhead result initialization.

![Bee Rule Architecture](img/bee-rule.svg)

---

## 2. Rule Anatomy & Signature Layout

### 2.1 Complete Signature Syntax
```bee
rule identifier(param_list) [ (named_param_list) ] => (result_list):
  -- Preconditions (Contracts)
  assert condition;
  
  -- Body statements (indented 2 spaces)
  
  -- Postconditions / Invariants (Contracts)
  expect condition;
return;
```

> **Curried signatures (Decision 6, 2026-09-12):** A rule may declare an optional *named-parameter slot* — a second parenthesised parameter list immediately following the primary parameter list. Each named parameter is bound by name at the call site via curried application, providing a self-documenting, keyword-argument style distinct from the positional primary list. See §2.4 for full semantics and §3.1 for call-site syntax.

### 2.2 Parameter Passing Conventions
- **Primitive Types ($\mathbb{Z}, \mathbb{N}, \mathbb{R}, \mathbb{Q}, \mathbb{C}, \mathbb{B}$):** Transferred **by value** (copy on call).
- **Composite Types (`array`, `list`, `map`, `set`, `object`):** Transferred **by share** (reference-counted Region pointer).
- **Optional Parameters:** Parameter declarations may specify default values using `:` (e.g. `param: default_val ∈ Type`).
- **Variadic Parameters (`*varargs`):** The final parameter in a list may be prefixed with `*` to accept a variable number of arguments as an array (`*args ∈ [Z]`).
- **Named-Parameter Slot** *(Decision 6)*: A rule may declare an optional second parameter list enclosed in its own delimiters (see §2.4). Each parameter in this slot is normally require-bound by name at the call site via curried application; the slot participates in no positional binding. Parameters in the named slot may themselves specify `:` defaults to permit omission.

### 2.3 Result Declarations & Deconstruction
- **Named Results:** Results are explicitly declared with identifier and type `=> (y ∈ N)` or `=> (s, d ∈ Z)`.
- **Default Result Initialization:** Result variables are automatically initialized to their zero-values upon entering the rule body.
- **Single Capture:** `new r := compute(x);`
- **Multi-Result Deconstruction:**
  $$(s, d) \leftarrow \text{compute\_both}(x, y)$$
  ```bee
  new s, d := compute_both(x, y);
  ```
- **Result Wildcard (`_`):** Unwanted results can be suppressed using `_` (e.g. `new s, _ := compute_both(x, y);`).
- **Expression Restriction:** Rules returning multiple results ($> 1$) CANNOT be embedded directly within nested arithmetic expression trees; they must be evaluated via deconstruction assignment.

### 2.4 Named-Parameter Slots & Curried Signatures *(Decision 6)*

A rule signature may include **at most one** secondary parameter list — the *named-parameter slot* — rendered directly after the primary parameter list and before the optional result list:

```bee
rule identifier(param_list) (named_param_list) => (result_list):
  -- signatures may declare both, named, neither
return;
```

- **Slot lifecycle:** The named slot is purely declarative. No positional binding occurs between the primary parameter list and the named slot — the two lists are independent binders.
- **Slot declaration grammar:** Each entry in the named slot is a `named_parameter`, identical syntactically to a primary-list parameter except that it MUST be standalone (no `*` variadic prefix): `named_parameter ::= identifier [ ":" expression ] ( "∈" | "in" ) type_specifier`.
- **Slot call site:** The named slot is invoked by appending `(named_arg_list)` after the primary call's closing `)`. The named-argument list consists only of named arguments (`identifier ":" expression`); positional entries are not permitted.
- **Order independence:** Named arguments may appear in any order. The compiler resolves them by name, not position.
- **Default fall-back:** A named parameter may declare a default expression following Decision 6 syntax (e.g. `sep: ", "  ∈  Str`). If the call site omits the name, the default binds. If the parameter declares no default and the caller omits it, the compiler emits `E0307 MissingNamedArgument` and halts resolution.
- **Empty slot:** Either list may be empty. `rule ping()(steady ∈ B):` and `rule ping():` are both valid forms; the named slot is reserved at compile time even when empty.

```bee
-- Canonical named-slot rule
rule print(*items ∈ [Any]) (sep: ", "  ∈  Str, end: "."  ∈  Str) => (void ∈ Void):
  -- body: emit items joined by sep, terminated by end
return;

-- Canonical call sites
apply print(1, 2, 3) (sep: " | ", end: ";");
new x := print("a", "b") ();  -- empty slot, both defaults used
new y := print("a", "b");     -- ERROR if slot is required and no defaults supplied
```

**Rationale.** The named-slot pattern unifies keyword-argument calling (Haskell records / Python kwargs / Ruby hashes) with call-site readability. It replaces the legacy `using` keyword on `print`/`write`/formatting primitives with a structural language feature reusable by every rule, not just I/O directives.

---

## 3. Invocations & Early Termination

### 3.1 Invocation Syntax
- **Direct Assignment / Expression Call:** Captures return values into variables.
  ```bee
  new result := fib(n: 5);
  ```
- **Apply Directive (`apply`):** Executes a rule for side-effects, ignoring any returned results.
  ```bee
  apply log_message("Processing complete");
  ```
- **Curried Application** *(Decision 6)*: When the invoked rule declares a named-parameter slot (§2.4), the caller applies the slot with a *second* parenthesised argument list:
  ```bee
  print("Hello", "World") (sep: " | ");
  apply write_row(record, fields) (mode: "tsv");
  ```
  The named-slot list contains only **named arguments** of the form `identifier ":" expression`. The positional primary list is unaffected. Named-argument naming and order are independent: any permutation is accepted, and omitted named parameters fall back to their declared `:` defaults.

  > **Deprecated syntax:** The legacy `using` / `using:` postfix in `io_stmt` is **deprecated** as of Decision 6. Author-targeted forms such as `print(a, b) using " | ";` are removed from normative examples; the canonical form is `print(a, b) (sep: " | ");` where the `print` rule declares a named slot `(sep ∈ Str)`. The lexer emits a non-fatal `E0011 deprecated-symbol: 'using' — use named-argument `(name: ...)`` until Phase 7 audit task 7.2 hardens it to `E0009`.

### 3.2 Terminal Directives
- **`return` (Mandatory Block Terminator):** Closes the rule block and returns execution to the caller. Must align horizontally with the `rule` header (0 relative indentation).
- **`exit` (Early Successful Return):** Instantly terminates rule execution cleanly, returning current values of result variables without raising an error:
  ```bee
  if count = 0 do
    let result := 0;
    exit;
  done;
  ```

---

## 4. Design by Contract (`assert` / `expect`)

Bee enforces formal contract assertions directly within rule definitions:

$$\text{Contract}(\text{divide}) = \begin{cases} \text{assert}\big(b \;\text{<>}\; 0\big) \;\textit{[Precondition Warning]} \\ \text{expect}\big(\text{ratio} \cdot b \approx a\big) \;\textit{[Invariant Guarantee]} \end{cases}$$

> **Operator syntax note (Decision 3, 2026-09-12):** The canonical value-inequality operator is `<>`. The Unicode form `≠` is deprecated. When authors write `≠` in source, the lexer emits a non-fatal `E0010 deprecated-symbol: '≠' — use '<>'`; after Phase 7 audit task 7.2 the diagnostic upgrades to a hard `E0009` syntax error.

```bee
rule divide(a ∈ R, b ∈ R) => (ratio ∈ R):
  assert b <> 0;         -- Precondition / Warning check: emits diagnostic warning to stderr if false (canonical inequality — Decision 3 deprecates ≠)
  let ratio := a / b;
  expect ratio * b ≈ a; -- Postcondition / Invariant check: raises runtime error if false
return;
```

- **Assertion Warning (`assert`):** Evaluates a condition. If it fails ($0$ or `false`), a diagnostic warning (`W0301: AssertionWarning`) is printed to standard error (`stderr`), but program execution continues uninterrupted.
- **Expectation Enforcement (`expect`):** Evaluates an invariant condition. If it fails ($0$ or `false`), a runtime error (`E0303: ExpectationFailed`) is raised and propagated up the call stack. If not caught by an enclosing `trial` block, the application halts immediately with a failure exit status.

---

## 5. Advanced Rule Architectures

### 5.1 Companion & Singleton Rules
![Companion Rule](img/companion-rule.svg)
![Singleton Rule](img/singleton-rule.svg)

- **Companion Rule:** Binds state and helper logic to an existing module or data type.
- **Singleton Rule:** Encapsulates globally isolated state within a single instance generator.

### 5.2 Forward Declarations (No Hoisting)
Bee does NOT use compiler hoisting. Identifiers must be declared before use. For mutual or cyclic rule dependencies, explicit forward declarations (signature ending with `;`) MUST be provided:

```bee
-- Forward declaration signature
rule even(n ∈ N) => (b ∈ B);
rule odd(n ∈ N) => (b ∈ B);

-- Implementation
rule even(n ∈ N) => (b ∈ B):
  if n = 0 do
    let b := true;
    exit;
  done;
  let b := odd(n - 1);
return;

rule odd(n ∈ N) => (b ∈ B):
  if n = 0 do
    let b := false;
    exit;
  done;
  let b := even(n - 1);
return;
```

### 5.3 Tail Call Optimization (TCO)
A rule invocation in final return position (`let r := tail_rule(...)` followed immediately by `return`) is optimized by the compiler via Tail Call Optimization:

$$T(n) = \mathcal{O}(1) \text{ Stack Frames}$$

- Overwrites the caller's Region Arena stack frame instead of pushing a new frame.
- Re-executes rule body with updated argument values, converting recursion into iteration with $O(1)$ stack space.

### 5.4 Closures & State Generators
Rules can encapsulate persistent state and nested method rules (`closures`). 

- **State Mutability**: Variables stored in the rule closure must be explicitly boxed using the **`[]`** operator (e.g., `set .count := [start];`) to be heap-allocated and mutable. Unboxed variables defined via `set` within a rule are immutable.

```bee
rule counter_generator(start ∈ Z) => (next ∈ Rule):
  set .count := [start];  -- Boxed for mutability
  
  rule .next() => (val ∈ Z):
    let .count += 1;      -- Mutation allowed due to boxing
    let val := .count;
  return;
return;
```

---

## 6. Formal EBNF Grammar

```ebnf
(* Rule Definition *)
rule_def          ::= "rule" [ member_access ] identifier "(" [ param_list ] ")" [ "(" [ named_param_list ] ")" ] [ "=>" "(" result_list ")" ] ":" [ contract_clause ] block "return" ";" ;
forward_decl      ::= "rule" identifier "(" [ param_list ] ")" [ "(" [ named_param_list ] ")" ] [ "=>" "(" result_list ")" ] ";" ;

(* Parameters & Results *)
param_list        ::= parameter ( "," parameter )* ;
parameter         ::= [ "*" ] identifier [ ":" expression ] ( "∈" | "in" ) type_specifier ;

named_param_list  ::= named_parameter ( "," named_parameter )* ;
named_parameter   ::= identifier [ ":" expression ] ( "∈" | "in" ) type_specifier ;
(* named_parameters MUST NOT be prefixed with "*" (no variadic named slot per Decision 6) *)

result_list       ::= result_item ( "," result_item )* ;
result_item       ::= identifier [ ":" expression ] [ ( "∈" | "in" ) type_specifier ] ;

(* Contracts *)
contract_clause   ::= ( "assert" condition ";" )* ( "expect" condition ";" )* ;

(* Invocations *)
rule_apply        ::= "apply" identifier "(" [ arg_list ] ")" [ "(" [ named_arg_list ] ")" ] ";" ;
rule_call         ::= identifier "(" [ arg_list ] ")" [ "(" [ named_arg_list ] ")" ] ;
arg_list          ::= argument ( "," argument )* ;
argument          ::= [ identifier ":" ] expression ;
named_arg_list    ::= named_argument ( "," named_argument )* ;
named_argument    ::= identifier ":" expression ;
(* named_arguments are order-independent; positional entries are not permitted in the named slot *)
```

---

## 7. Diagnostic Codes

| Error Code | Error Condition | Description |
| :--- | :--- | :--- |
| `E0301` | `UndeclaredRule` | Rule invoked prior to definition or forward declaration |
| `W0301` | `AssertionWarning` | Assertion failed in `assert` contract clause (warning emitted to stderr) |
| `E0303` | `ExpectationFailed` | Expectation failed in `expect` contract clause or body statement (fatal runtime error if unhandled) |
| `E0304` | `MultiResultInExpression` | Multi-result rule used inside arithmetic expression tree |
| `E0305` | `MissingReturnTerminator` | Rule block does not end with aligned `return;` |
| `E0306` | `SignatureMismatch` | Forward declaration does not match rule implementation signature |
| `E0307` | `MissingNamedArgument` | Call site omits a required named-parameter slot binding (no default supplied per Decision 6) |
| `W0308` | `NamedSlotIgnored` | Curried `(...)` argument list appended where the rule declares no named slot (silently tolerated, warning emitted) |
| `E0309` | `UnknownNamedArgument` | Curried call site names an identifier not present in the rule's declared named slot |
| `E0010` | `DeprecatedSymbol '≠'` | Legacy value-inequality operator encountered; canonical `<>` should be used (Phase 7 audit pre-`E0009`) |
| `E0011` | `DeprecatedSymbol 'using'` | Legacy `using` / `using:` postfix encountered on a call site; canonical is named-argument `(name: ...)` per Decision 6 (Phase 7 audit pre-`E0009`) |

---

## 8. Alignment Status

- **Issues Addressed:** Formalized rule semantics, contracts, TCO, closures, and forward declarations.
- **Manifest Tracking:** Updated `MANIFEST.md` to reflect completion of `spec/03-rules.md`.
- **Decision 3 Deprecation Sweep (2026-09-12):** All normative examples in this spec now use the canonical value-inequality operator `<>`. The single residual `≠` reference is an intentional deprecation annotation in a code comment per Decision 3.
- **Spec Audit Locked (Phase 7 prep):** `spec/01-lexical-structure.md`, `spec/02-statements.md`, and `spec/03-rules.md` are now harmonized with Decisions 1–6.
- **Decision 6 — Curried Rule Signatures (2026-09-12):** This pass introduces §2.4 (named-parameter slots), updates §2.1/§2.2/§3.1, and extends the §6 EBNF with `named_param_list`, `named_parameter`, `named_arg_list`, and `named_argument` productions. The legacy `using` keyword is deprecated; canonical I/O call sites use curried named arguments (`print(a, b)(sep: " | ")`). New diagnostic codes `E0307`, `W0308`, `E0309`, and `E0011` are added. Implementer (`internal/lexer/` & `internal/parser/`) is gated by Anti-Loop Gate until `spec/02-statements.md` is also harmonized with the new `io_stmt` EBNF (next pass).
- **Remaining Hurdles:** `spec/02-statements.md` §5 `io_stmt` EBNF still references the legacy `using` postfix; harmonization is sequenced for the immediate next pass under the Anti-Loop Gate. `internal/lexer/` and `internal/parser/` updates are deferred until both specs are locked.
