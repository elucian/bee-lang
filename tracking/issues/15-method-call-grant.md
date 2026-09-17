# Issue 15 — Method-call dispatch (`x.type()`) is unimplemented

## Status
**Open — debt** (test parked in `test/debt/T0126-method-call.bee`).

## Observed Behaviour
Running any `.bee` source that contains a dotted method call:

```bee
rule main:
  new a := 10;
  expect a.type() is Z;
return;
```

triggers `E0009 SyntaxError:UnrecognizedStatement` at the line of the `expect`. The compiler reports an unrecognized token, leaving `a.type()` undigested.

## Root Cause
The expression grammar in `internal/parser/parser.go::parseExpression` (see § 5 EBNF) recognises only `expression "(" [ arg_list ] ")"` as a callable. There is no AST node for `expression "." identifier "(" [ arg_list ] ")"`. The lexer also lacks a `.` token type binding.

## Spec Reference
- `spec/07-functions.md` §6 — proposed `method_call ::= expression "." identifier "(" [ arg_list ] ")" ;`
- `spec/06-objects.md` §5 — method descriptor table for each type.

## Decision Required
1. **Token grammar.** Add `DOT` to `internal/token/token.go`. The lexer must treat `.` as a structural token only when followed by an identifier (i.e. `.beep` is a public-export prefix; `a.type` is a method receiver). A backslash escape is needed for floating-point literals (`3.14` must not be mislexed).
2. **AST shape.** Add `MethodCallExpression { Receiver Expression, Method string, Arguments []Expression }` in `internal/ast/`.
3. **Resolution.** Method tables live on the type descriptor (Decision 4: typechecker postponed; for now, evaluator pulls methods from a hardcoded map keyed by `Type`).
4. **Backwards compatibility.** The `.beep` export prefix (per `spec/04-structure.md` §4.1) must continue to work at the top-level grammar — `member_decl ::= [ "." ] ( decl_stmt | ... )` already covers this. No conflict.

## Acceptance Criteria
- `test/debt/T0126-method-call.bee` runs green when promoted back to `test/level1/`.
- `a.type()` returns the type tag string (`"Z"`, `"R"`, `"U"`, `"B"`, `[T]`, etc.).
- `pip.add(x)` style invocations on collections parse without error.

## Anti-Loop Note
Do NOT touch the test cases in `test/level1/` directly. Promote `T0126-method-call.bee` only after the AST node + resolver are wired and the spec/07 §6.1 grammar is locked.
