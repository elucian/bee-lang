# Bee Specification: Statements & Expressions (02-statements.md)

## 1. Statement Grammar
Statements are categorized by their imperative (action) or declarative (definition) nature.

```ebnf
statement ::= declaration | assignment | call | control_flow ;
declaration ::= ("new" | "set") identifier (":=" expression | "∈" type) ;
assignment ::= "let" identifier (":=" | "+=" | "-=") expression ;
```

## 2. Expression Grammar
```ebnf
expression ::= primary | binary_op | conditional ;
primary ::= identifier | constant | "(" expression ")" | call ;
binary_op ::= expression operator expression ;
conditional ::= expression "if" condition ("else" expression)? ;
```

## 3. Operational Semantics
- **Statement Sequencing:** Statements are evaluated in the order they appear. Semicolons (`;`) are mandatory statement terminators.
- **Declarative Order:** All identifiers must be declared via `new` or `set` prior to their usage in `let` or expression contexts within the same scope.
- **Scoping:** Every `do`, `cycle`, and `rule` block introduces a new local lexical scope. Identifier shadowing is permitted but triggers a compiler diagnostic.
- **Memory Safety:** Local allocations (via `new`) are scoped to the rule/block region and are automatically reclaimed at the termination of the block.
- **Indentation:** The parser strictly enforces 2-space indentation. Deviation results in a `SyntaxError: IndentationMismatch`.
