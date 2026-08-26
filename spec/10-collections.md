# Bee Types & Collections Specification (05-types-collections.md)
# Bee Specification: Collections (10-collections.md)

## 1. Collection Grammar
```ebnf
collection    ::= list_lit | array_lit | set_map_lit ;
list_lit      ::= "(" expression ("," expression)* ")" ;
array_lit     ::= "[" expression ("," expression)* "]" ;
set_map_lit   ::= "{" (expression ("," expression)* | key_val_pair ("," key_val_pair)*) "}" ;
key_val_pair  ::= expression ":" expression ;
```

## 2. Operational Semantics
- **Membership:** `∈` checks element membership.
- **Allocation:** `[]` defines a mutable reference wrapper (box).
- **Indexing:** 1-based indexing for arrays/matrices. `$` anchors the last index.
- **Memory Safety:** Collections are reference-based. RC manages mutable instances; GC manages immutable string segments.
- **Conversion:** `collection :> collection` triggers explicit re-allocation and deep copy.
