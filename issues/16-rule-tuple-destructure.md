# Issue 16 — Rule-call tuple destructuring (`new a, b, c := zeros();`) is unimplemented

## Status
**Open — debt** (test parked in `test/debt/T0127-rule-call-result-binding.bee`).

## Observed Behaviour

```bee
rule zeros() => (z ∈ Z, r ∈ R, f ∈ B):
  expect True;
return;
rule main:
  new a, b, c := zeros();
  print (a, b, c);
return;
```

triggers `E0009 SyntaxError:UnrecognizedStatement` at the line `new a, b, c := zeros();`. The compiler fails to recognise that `zeros()` is a rule invocation returning a tuple of three values.

## Root Cause
- `parseDeclaration` (parser.go:463) supports `new` followed by a comma-separated identifier list and `:=` followed by a single `parseExpression()`. It does NOT support `new a, b, c := f();` where the RHS is a function call returning a tuple.
- The call `zeros()` itself parses fine, but the binding logic still expects a single value where the AST is actually a `CallExpression` returning a tuple.

## Spec Reference
- `spec/03-rules.md` §2.4 — Curried rule signatures. The decision was that the FIRST parameter slot can be a vararg, and the second slot (the named-argument group) holds optional named parameters. **This is the binding shape, not the return shape.**
- `spec/03-rules.md` §3.1 — Call-site form. No explicit tuple-return grammar yet — must be defined.

## Decision Required

1. **Return-type annotation.** A rule whose signature ends in `=> (a ∈ T1, b ∈ T2)` declares an n-tuple return. The compiler must propagate the tuple arity through the symbol table (postponed to Phase 5 per Decision 4).
2. **Parallel destructuring grammar.** Extend `parseDeclaration` to recognise:
   ```ebnf
   decl_stmt ::= "new" ident_list ":=" callable_expr ";"  (* parallel tuple-destructure *)
   ```
   The arity of the LHS ident list must match the arity of the call site's declared return tuple. Mismatch → `E0303: TupleArityMismatch`.
3. **Evaluator integration.** The evaluator must invoke the called rule, capture its tuple-framed environment at the `return;` point, and broadcast each slot to the corresponding LHS identifier.
4. **Tutorial update.** The tutorial (`bee-tutorial/rules.html`) currently does NOT document tuple-return rules. Add § "Multi-Value Returns" with a running example.

## Acceptance Criteria
- `test/debt/T0127-rule-call-result-binding.bee` runs green when promoted back to `test/level1/`.
- `print (a, b, c);` outputs three distinct integer values (e.g., `0,0,0` initially — note that the test as-is declares `new a, b, c := zeros();` WITHOUT explicit `∈ Z`, so infer-from-call semantics apply).
- A 1-tuple rule `() => (r ∈ Z)` can be destructured as `new x := get_value();`.

## Anti-Loop Note
Do NOT promote the debt test until both:
- (a) `spec/03-rules.md` §3.1 has a formal tuple-return EBNF, AND
- (b) the evaluator's return-capture logic supports multi-frame teardown.
