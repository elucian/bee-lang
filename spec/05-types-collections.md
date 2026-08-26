# Bee Types & Collections Specification (05-types-collections.md)

## 1. Type Declarations
```ebnf
type_decl     ::= "type" identifier ":" type_descriptor ("<:" super_type)? ";" ;
type_desc     ::= primitive_type | collection_type | custom_type ;
primitive     ::= "B" | "A" | "U" | "Q" | "N" | "Z" | "R" | "S" ;
```

## 2. Collection Grammar
```ebnf
collection    ::= list_lit | array_lit | set_map_lit ;
list_lit      ::= "(" expression ("," expression)* ")" ;
array_lit     ::= "[" expression ("," expression)* "]" ;
set_map_lit   ::= "{" (expression ("," expression)* | key_val_pair ("," key_val_pair)*) "}" ;
key_val_pair  ::= expression ":" expression ;
```

## 3. Operational Semantics
- **Membership:** `∈` checks element membership.
- **Allocation:** `[]` defines a mutable reference wrapper (box).
- **Indexing:** 1-based indexing for arrays/matrices. `$` anchors the last index.
- **Memory Safety:** Collections are reference-based. RC manages mutable instances; GC manages immutable string segments.
- **Conversion:** `collection :> collection` triggers explicit re-allocation and deep copy.
