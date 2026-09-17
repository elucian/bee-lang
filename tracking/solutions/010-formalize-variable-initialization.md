# Solution: Formalize Variable Initialization, Type Inference, and Pair-Up Semantics

Update `spec/02-statements.md` and `spec/05-types.md` to formalize the dual mechanisms of declaration in Bee's gradual typing system:
1. **Type Inference (`:=`):** Used with `new` and `set` to define and initialize variables without type specifications.
2. **Structural Pair-Up (`:`):** Pairs variables with initial values while requiring explicit type specification (`∈ Type`).

## Design

1. **Type Inference via `:=` (`new` and `set`):**
   - Syntax: `new id_1, id_2, ..., id_n := expr_1, expr_2, ..., expr_n;` or `set id := expr;`
   - Semantics: No explicit type is specified. The compiler infers static types from the evaluated expressions element-wise ($\text{alloc}(x_i) \in \text{typeof}(v_i)$).

2. **Explicit Type Zero-Initialization:**
   - Syntax: `new id_1, id_2, ..., id_n ∈ Type;`
   - Semantics: Allocates each identifier $id_i$ with the default zero value of $Type$ ($\text{alloc}(x_i) \in \mathbb{T}, x_i \leftarrow \text{zero}(\mathbb{T})$).

3. **Structural Pair-Up Initializer via `:` (Gradual Typing):**
   - Syntax: `new id_1, id_2, ..., id_n ∈ Type:init_expr;` (or per-identifier binding)
   - Semantics: Pairs variable(s) with an initial value while mandating explicit type specification $Type$. Broadcasts `init_expr` to all identifiers or binds pairwise.

4. **EBNF Updates in `spec/02-statements.md`:**
   ```ebnf
   decl_stmt         ::= "set" ident_list ":=" expr_list
                       | "new" ident_list ( "∈" | "in" ) type_specifier [ ":" ( expression | expr_list ) | ":=" expr_list ]
                       | "new" ident_list ":=" expr_list ;
   ident_list        ::= identifier ( "," identifier )* ;
   expr_list         ::= expression ( "," expression )* ;
   ```
