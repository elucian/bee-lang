# Bee Specification: Rules & Functions

## 1. Rule Anatomy
```ebnf
rule_decl ::= "rule" identifier "(" parameters? ")" "=>" "(" results? ")" block "return" ;
parameters ::= parameter ("," parameter)* ;
parameter ::= identifier (":" default_val)? "∈" type ;
results ::= result ("," result)* ;
result ::= identifier "∈" type ;
```

## 2. Rule Execution
```ebnf
rule_apply ::= "apply" identifier "(" arguments? ")" ;
rule_call ::= identifier "(" arguments? ")" ;
```

## 3. Lambda Expressions
```ebnf
lambda_decl ::= "new" identifier ":=" "λ" "(" parameters? ")" "=>" "(" expression ")" "∈" type ;
```
