# Issue: Multi-Variable Declaration and Initialization Syntax Underspecified

The specification (`spec/02-statements.md`) only documents single-variable declaration patterns (`new count ∈ Z := 0;` and `new name := "Bee";`). It lacks formal definitions for multi-variable declarations, parallel/destructured bindings, and gradual typing rules distinguishing type inference (`:=`) from structural pair-up initialization (`:`).

## Impact
- **Incomplete Grammar:** The EBNF in `spec/02-statements.md` does not support comma-separated identifier lists (`IdentList`) or expression lists (`ExprList`) in declaration statements.
- **Operator Role Ambiguity:** The exact semantics of `:=` (type inference without explicit type) versus `:` (pair-up initializer requiring an explicit type specification) are not formally contrasted in the declaration specification.
- **Parallel & Uniform Initializer Semantics:** Parallel multi-variable initializations (e.g., `new x, y, z := 1, 2, 3;`) and uniform broadcast pair-up initializations (e.g., `new xo, yo, zo ∈ Z:10;`) lack formal semantic rules.

## Requirements
1. Formalize comma-separated variable declarations (`new a, b, c ∈ Type;`) with default zero-value initialization for the specified type.
2. Formalize type inference using `:=` with `new` and `set` (`new x, y, z := 1, 2, 3;`) where types are inferred from expressions without explicit type specification.
3. Formalize the pair-up operator `:` for initializing variables with explicit type specification (`new xo, yo, zo ∈ Z:10;`), enforcing gradual typing invariants where `:` requires an explicit type.
4. Update `spec/02-statements.md` narrative, mathematical representations, and EBNF grammar (`decl_stmt`, `ident_list`, `expr_list`).
