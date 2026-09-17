# Solution: Unify Initialization with `=` and Formalize Equality Taxonomy

Standardize initialization syntax to use `=` for all declaration initializations and refine the equality operator taxonomy to separate value comparison (`==`) from identity check (`is`).

## Design

1. **Initialization Syntax (`=`):**
   - Use `=` for all variable initialization in `new` and `set` declarations. This removes structural pair-up ambiguity by replacing `:` with `=`.
   - Update `decl_stmt` EBNF:
     ```ebnf
     decl_stmt ::= "set" ident_list ":=" expr_list
                 | "new" ident_list ( "∈" | "in" ) type_specifier [ "=" ( expression | expr_list ) ]
                 | "new" ident_list ":=" expr_list ;
     ```

2. **Equality Taxonomy:**
   - **Value Comparison (`==`)**: Formally replace `=` with `==` for all value comparisons in `expect`, `assert`, and conditional expressions.
   - **Identity Comparison (`is`)**: Introduce `is` to check for pointer/identity equality (e.g., `a is b`).

3. **Specification Updates:**
   - Update `spec/02-statements.md` to reflect the new `==` and `is` operators.
   - Update all `expect` and `assert` examples in the specification and test suite to use `==`.
