# Bee Specification: Functions & Lambda Expressions (07-functions.md)

## 1. Executive Architectural Philosophy

Bee enforces a strict separation between **`rule`** (stateful/imperative subprograms) and **`function` / lambda** (pure mathematical transformations):

- **Pure Lambda Expressions (`λ`):** Represent stateless, deterministic, side-effect free mathematical operations.
- **Symbolic Designator:** Delimited by the Greek letter **`λ`** (`U+03BB`) or ASCII shorthand `\`.
- **Lambda Type Descriptor (`L`):** Lambda expressions are first-class values typed as **`L`** (or `Lambda`). They can be stored in variables, passed as callback parameters, or returned from rules.

---

## 2. Lambda Syntax & Anatomy

### 2.1 Explicit Lambda Declaration
```bee
-- Explicit declaration with full parameter and result types
new power := λ(x, y ∈ Z) => (x ^ y) ∈ Z;

-- Using shorthand type inference
new double := λ(x ∈ Z) => (x * 2);
```

### 2.2 Inline Callback Shorthand
When passed directly as arguments to rules or processing operators, lambdas permit lightweight parameter syntax:
```bee
-- Inline callback passed to a rule
new result := apply_transform(10, 20, (x, y) => x + y);
```

### 2.3 Lambda Type Signatures (`L`)
Variables and parameters representing functions use the `L` type descriptor:
```bee
-- Parameter taking a lambda callback
rule compute(a, b ∈ R, fn: λ(x, y ∈ R) => R) => (r ∈ R):
  let r := fn(a, b);
return;
```

---

## 3. Purity Invariants & Restrictions

To guarantee mathematical referential transparency, compiler SIMD vectorization, and thread safety, lambda expressions operate under four strict invariants:

1. **Statelessness:** Lambdas CANNOT declare local state variables (`new`, `let`, `alter`) or maintain persistent static memory.
2. **Referential Transparency:** Given the same arguments, a lambda MUST return the exact same output without relying on global mutable state or system clocks.
3. **Side-Effect Isolation:** Lambdas CANNOT perform I/O operations (`print`, `write`, file access) or mutate outer variables.
4. **Rule Isolation:** A `rule` CAN call a `lambda`; a `lambda` CANNOT call a `rule`. This rule prevents side-effects from creeping into mathematical expressions.

---

## 4. Higher-Order Functions & Collection Storage

First-class lambda references (`L`) can be stored in arrays, lists, or maps to form expression dictionaries:

```bee
-- Expression Dictionary
new gt := λ(x, y ∈ Z) => (x > y) ∈ B;
new lt := λ(x, y ∈ Z) => (x < y) ∈ B;
new eq := λ(x, y ∈ Z) => (x = y) ∈ B;

type LambdaMap: {[S]: L} <: Map;
new ops := {"gt": gt, "lt": lt, "eq": eq} ∈ LambdaMap;

-- Indirect lambda evaluation
new is_greater := ops["gt"](10, 5); -- Evaluates to true (0b1)
```

---

## 5. SIMD & GPU Auto-Vectorization

Because lambdas are guaranteed to be pure and side-effect free, the Bee compiler automatically parallelizes collection map operations over vector domains:

- **SIMD Auto-Vectorization:** Single Instruction Multiple Data instructions (e.g. AVX-512) process array items in parallel chunks.
- **GPU Kernel Offloading:** High-throughput lambda operations applied over large collection domains are automatically lowered into native OpenCL / GPU compute pipelines.

---

## 6. Formal EBNF Grammar

```ebnf
(* Lambda Expressions *)
lambda_decl       ::= "new" identifier ":=" lambda_expr ";" ;
lambda_expr       ::= ( "λ" | "\" ) "(" [ param_list ] ")" "=>" "(" expression ")" [ ( "∈" | "in" ) type_specifier ] ;
short_lambda      ::= "(" [ ident_list ] ")" "=>" expression ;

(* Type Descriptor *)
lambda_type       ::= "L" | "λ" "(" [ type_list ] ")" "=>" type_specifier ;
type_list         ::= type_specifier ( "," type_specifier )* ;

(* Invocations *)
lambda_call       ::= expression "(" [ arg_list ] ")" ;
```

---

## 7. Diagnostic Error Codes

| Error Code | Error Condition | Description |
| :--- | :--- | :--- |
| `E0701` | `RuleCallInLambda` | Lambda expression attempts to invoke a stateful `rule` |
| `E0702` | `SideEffectInLambda` | Lambda contains I/O statement or variable mutation |
| `E0703` | `LambdaTypeMismatch` | Passed lambda signature does not match expected parameter `L` |
| `E0704` | `UnboundLambdaVariable` | Lambda references outer variable that is not in scope |
| `E0705` | `StateDeclarationInLambda` | Attempt to use `new`, `let`, or `alter` inside a lambda |

---

## 8. Alignment Status

- **Issues Addressed:** Formalized `λ` lambda syntax, callback shorthand, purity invariants, `L` type descriptor, expression maps, SIMD/GPU vectorization, and diagnostic codes.
- **Manifest Tracking:** Updated `MANIFEST.md` to reflect completion of `spec/07-functions.md`.
