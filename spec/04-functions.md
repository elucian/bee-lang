# Bee Specification: Functions (04-functions.md)

## 1. Lambda Expressions
```ebnf
lambda_decl ::= "new" identifier ":=" "λ" "(" parameters? ")" "=>" "(" expression ")" "∈" type ;
```

## 2. Operational Semantics
- **Purity:** Lambda expressions are pure (side-effect free, deterministic, no internal state).
- **Restrictions:** Cannot call `rule` entities or use `let`/`new` (no side effects allowed).
- **Concurrency:** Single-threaded execution; runtime-only GPU/SIMD parallelization for math-heavy operations.
- **Callback:** Passed as `L` (reference) type, used in rules for functional transformations.
