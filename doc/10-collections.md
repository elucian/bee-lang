# Bee Collections

Collections are data structures that group a limited number of values together. You can access individual values using different methods, depending on the collection type.

## 1. Core Collections

### Ordinal Type
An `Ordinal` is an ordered small set of identifiers. Each identifier maps to an integer value starting from a specified number.

```bee
type OrdinalName: (1){name1, name2, name3} <: Ordinal;

rule main:
  new a, b, c ∈ OrdinalName;
  let a := .name1; -- 1
  let b := .name2; -- 2
  return;
```

### Lists
A `List` is a dynamic collection implemented as a doubly-linked structure (`prior`, `next` pointers).
- **Head:** First element, accessed via `[1]`.
- **Tail:** Last element, accessed via `[$]`.
- **Properties:** Unlimited capacity, ordered, same-type elements.

```bee
new myList ∈ List(Z) := (0, 1, 2, 3);
for ∀ x ∈ myList do
  print x;
repeat;
```

### Arrays
Arrays are contiguous memory blocks with fixed capacity.

```bee
-- Declare array with 10 integers
new numbers ∈ [Z](10) := [0, 1, 2, 3, 4, 5, 6, 7, 8, 9];

-- Access index 1 (first element)
print numbers[1]; 
```

### Matrix
Matrices are multidimensional arrays stored in row-major order. Indexing is 1-based `[row, col]`.

```bee
new M ∈ [Z](2, 2) := [[1, 2], [3, 4]];
let M[1, 1] := 100;
```

## 2. Symbolic Data

### Sets & Hash Maps
- **Sets:** Sorted collection of unique values.
- **Hash Maps:** Key-value pairs sorted by key hash.

```bee
new myMap ∈ {A:S} := {'key1': "value1"};
let myMap['key2'] := "value2";
```

### Text Markup
For large literals, Bee uses opaque tags.

```bee
new sqlQuery ∈ S := <sql> select * from users; </sql>;
```

---
**Read next:** [Data Processing](/projects/bee/processing/)


* * *

**Read next:** [Data Processing](/projects/bee/processing/)
