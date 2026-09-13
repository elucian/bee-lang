# Solution 16 — Rule-call tuple destructuring

Approach proposed to close `issues/16-rule-tuple-destructure.md`.

## 1. Spec change (BLOCKER #1)
Update `spec/03-rules.md` §3.1 to formalise the tuple-return EBNF:
```ebnf
(* Tuple-return rule signature *)
rule_signature ::= "rule" identifier [ "(" param_list ")" ] [ "(" named_param_list ")" ] [ "=>" "(" result_list ")" ] ":" ;

result_list     ::= typed_param ( "," typed_param )* ;
(* typed_param ::= identifier "∈" type_specifier is already in §2.1 *)

(* Parallel tuple destructure at call site *)
decl_stmt       ::= "new" ident_list ":=" callable_expr ";" ;
decl_stmt       ::= "new" ident_list callable_expr ";" ;   (* zero-init broadcast *)
decl_stmt       ::= ident_list ":=" callable_expr ";" ;    (* mutation form: let *)
```

The spec should clarify:
- The tuple binds positionally: `new a, b, c := cuple();` requires `cuple() => (_, _, _)`.
- A single-value rule can still be called via the current `x := f();` form (no parens demanded).

## 2. AST
Extend the existing `CallExpression` with a `Returns []string` field (return-type names). For tuple destructuring on the LHS, the RHS expression must produce an `IdentifierListConsumed` flag in the `DeclarationStatement`.

```go
type CallExpression struct {
    Token     token.Token
    Function  Expression   // Identifier node (rule name)
    Arguments []Expression
    Returns   []string     // populated at parse time from cached scope lookup
}

type DeclarationStatement struct {
    Token    token.Token
    Names    []string
    Value    Expression
    Values   []Expression
    Tuple    bool   // new: parallel-destructure flag
}
```

## 3. Parser
In `parseDeclaration` (parser.go:463), when the RHS is a `CallExpression` and the LHS has multiple identifiers:
```go
if isCallExpression(ds.Value) && len(ds.Names) > 1 {
    ds.Tuple = true
}
```
The validator (Phase 5 typechecker, currently postponed) will enforce arity: `len(ds.Names) == len(call.Returns)`.

## 4. Evaluator
When evaluating a `DeclarationStatement` with `Tuple == true`:
1. Call the rule with the empty arg list, capturing the post-`return` slot frame.
2. For each `i` in `[0, len(ds.Names))`, read the i-th named return slot and bind it to `ds.Names[i]`.

The rule's `return;` already propagates results via the rule's `Returns` field on the AST. The evaluator just needs to:
- Push a tuple-return frame (`exitingRule` already captures OVER; need a sibling mechanism for `return;` capture).
- On `return`, freeze the frame and surface the tuple values.

## 5. Verification
- Run `test/debt/T0127-rule-call-result-binding.bee`: `python test/solo.py T0127-rule-call-result-binding`.
- Promote to `test/level1/T0127.bee`.
- Run `sh run.sh test level1`.

## Edge Cases
- **Recursive tuple destructure:** `new x, (y, z) := f();` — out of scope for v1, raise `E0304: DestructuringNestingUnsupported`.
- **Discordant tuple arity:** hard error `E0303`.
- **Zero-tuple return:** a rule that does NOT declare `=> (...)` returns nothing; calling it as `new a := no_ret();` raises `E0305: UnexpectedNonTupleDestructuring`.

## Tutorial Update
`bee-tutorial/rules.html` §3 (Calling Rules) gets a subsection:
> ### Multi-Value Returns
> A rule can return multiple values as a tuple. Use parallel destructuring at the call site to bind each:
> ```bee
> rule mul_div(a, b ∈ Z) => (q ∈ Z, r ∈ Z):
>   let q := a / b;
>   let r := a % b;
> return;
> rule main:
>   new q, r := mul_div(10, 3);   -- (3, 1)
> return;
> ```
