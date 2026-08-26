# Bee Specification: Types (05-types.md)

## 1. Type Declarations
```ebnf
type_decl     ::= "type" identifier ":" type_descriptor ("<:" super_type)? ";" ;
type_desc     ::= primitive_type | custom_type ;
primitive     ::= "B" | "A" | "U" | "Q" | "N" | "Z" | "R" | "S" ;
```

## 2. Type Promotion & Memory Layout
| Type | Bit-Width | Alignment | Promotion Path |
| :--- | :--- | :--- | :--- |
| B (Boolean) | 8-bit | 1 | -> N |
| A (Alpha) | 8-bit | 1 | -> U |
| U (Unicode) | 32-bit | 4 | -> S |
| N/Z (Int) | 64-bit | 8 | -> R |
| R (Real) | 64-bit | 8 | -> Q |
| Q (Rational) | 128-bit | 16 | N/A |
