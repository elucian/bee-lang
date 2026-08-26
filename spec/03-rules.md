# Bee Specification: Rules (03-rules.md)

## 1. Rule Anatomy
```ebnf
rule_decl     ::= "rule" identifier "(" parameters? ")" "=>" "(" results? ")" block "return" ;
parameters    ::= parameter ("," parameter)* ;
parameter     ::= identifier (":" default_val)? "∈" type ;
results       ::= result ("," result)* ;
result        ::= identifier "∈" type ;
```

## 2. Rule Execution
```ebnf
rule_apply    ::= "apply" identifier "(" arguments? ")" ;
rule_call     ::= identifier "(" arguments? ")" ;
```

## 3. Operational Semantics
- **Invocation:** Rules are invoked via `apply` (ignore result) or direct call (capture result).
- **Termination:** `return` is the mandatory terminal statement. `exit` terminates without error.
- **Param Passing:** Primitive types by value; composite types by share (reference).
- **Scope:** Rules have static constructor scope and dynamic object scope (`self`).
- **Hoisting:** No hoisting; private rules must be defined before use. Forward declarations are required for mutual recursion.
