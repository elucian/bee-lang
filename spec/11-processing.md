# Bee Specification: Processing (11-processing.md)

## 1. Boxing Semantics
```ebnf
boxing_expr    ::= "[" identifier "]" ;
unboxing_expr  ::= type_identifier "(" identifier ")" ;
```

## 2. Array/Matrix Slicing
```ebnf
slice_expr     ::= identifier "[" range "]" ;
range          ::= (integer | "$") ".." (integer | "$") ;
```

## 3. Markup Literal Semantics
- **Lexical State:** Tags (e.g., `<sql>`) transition the lexer to a literal capture state until the closing tag `</tag>` is encountered.
- **Assignment:** The captured content is treated as a single `Rope` (S) literal.
- **Nesting:** Markup blocks support recursive nesting of identical or different tags.

## 4. Operational Semantics
- **Mutation:** `let` is the only mechanism for mutating boxed variables.
- **Copying:** `::` operator performs a deep recursive clone of the reference hierarchy.
- **Indexing:** All indexed structures (Arrays, Matrices) strictly implement 1-based indexing.
