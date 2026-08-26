# Bee Specification: Collections & Data Structures (10-collections.md)

## 1. Executive Collection Model

Bee provides five primary collection types distinguished by syntax delimiters and internal memory representations:

| Collection Type | Delimiter | Memory Structure | Indexing / Access | Mutability |
| :--- | :--- | :--- | :--- | :--- |
| **List** | `( ... )` | Dynamic Doubly-Linked Chain | 1-Based `list[i]` | Dynamic Append / Remove |
| **Array** | `[ ... ]` | Contiguous Memory Block | 1-Based `array[i]` | Fixed / Size Bound |
| **Matrix** | `[ ... ](r, c)` | Row-Major 2D Memory Chunk | 1-Based 2D `matrix[r, c]` | Element Mutability |
| **Set** | `{ ... }` | Hash Table / Red-Black Tree | Unique Membership `∈` | Set Algebra (`∩`, `∪`, `\`) |
| **Map** | `{ k: v }` | Hash Map Dictionary | Key Indexing `map[key]` | Dynamic Keys |
| **Ordinal** | `(start){ ... }` | Integer Enumeration | Member Access `.id` | Constant Enum Set |

---

## 2. Indexing Conventions & The `$` Anchor

### 2.1 Mandatory 1-Based Indexing
All indexed collections in Bee (Lists, Arrays, Matrices, Strings, Slices) enforce **1-based indexing**. The first element is always at index `1`. Accessing index `0` triggers a compile-time or runtime error `E1001: IndexOutOfBounds`.

### 2.2 The Last-Element Anchor (`$`)
- **End Anchor (`$`):** The `$` token represents the dynamic length of the collection (`count`).
  ```bee
  new list := (10, 20, 30, 40);
  
  print list[1];     -- First element: 10
  print list[$];     -- Last element: 40
  print list[$ - 1]; -- Second to last element: 30
  ```

### 2.3 Slicing Syntax
Sub-collections are extracted using range operators:
```bee
new slice := list[2..$ - 1]; -- Extract from index 2 to second-to-last item
```

---

## 3. Detailed Collection Specifications

### 3.1 Ordinal Enums
An `Ordinal` maps identifiers to sequential integer values starting at a specified base:
```bee
-- Ordinal type starting at base index 1
type Days: (1){Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday} <: Ordinal;

new today ∈ Days := .Monday; -- Value = 1
```

### 3.2 Lists `(...)` & Arrays `[...]`
- **Dynamic List Declaration:** `new L ∈ List(Z) := (10, 20, 30);`
- **Fixed Array Declaration:** `new A ∈ [Z](10) := [0, 1, 2, 3, 4, 5, 6, 7, 8, 9];`
- **Matrix Declaration:**
  ```bee
  new M ∈ [Z](2, 3) := [[1, 2, 3], [4, 5, 6]];
  let M[1, 2] := 100; -- Row 1, Column 2 updated to 100
  ```

### 3.3 Set Algebra Operators
Sets are unordered collections of unique elements supporting native mathematical set algebra:

| Operation | Operator Symbol | Example Expression | Description |
| :--- | :--- | :--- | :--- |
| **Intersection** | `∩` | `s1 ∩ s2` | Elements present in both `s1` and `s2` |
| **Union** | `∪` | `s1 ∪ s2` | Combined unique elements of `s1` and `s2` |
| **Difference** | `\` | `s1 \ s2` | Elements in `s1` not present in `s2` |
| **Sym Difference** | `Δ` | `s1 Δ s2` | Elements in `s1` or `s2` but not both |
| **Subset Test** | `⊂` | `s1 ⊂ s2` | Evaluates to `true` if `s1` is strict subset |
| **Superset Test** | `⊃` | `s1 ⊃ s2` | Evaluates to `true` if `s1` is strict superset |

```bee
new s1 := {1, 2, 3};
new s2 := {3, 4, 5};

new common := s1 ∩ s2; -- {3}
new total := s1 ∪ s2;  -- {1, 2, 3, 4, 5}
```

### 3.4 Hash Maps `{:}`
Key-value dictionaries associate unique keys with values:
```bee
new userMap ∈ {S: Z} := {"alice": 100, "bob": 200};

let userMap["charlie"] := 300; -- Insert new key
del userMap["alice"];           -- Remove key
```

---

## 4. Formal EBNF Grammar

```ebnf
(* Collection Literals *)
collection_lit    ::= list_lit | array_lit | matrix_lit | set_lit | map_lit | ordinal_lit ;

list_lit          ::= "(" expression ( "," expression )* ")" ;
array_lit         ::= "[" expression ( "," expression )* "]" ;
matrix_lit        ::= "[" array_lit ( "," array_lit )* "]" ;
set_lit           ::= "{" expression ( "," expression )* "}" ;
map_lit           ::= "{" key_value_pair ( "," key_value_pair )* "}" ;
key_value_pair    ::= expression ":" expression ;
ordinal_lit       ::= "(" integer_lit ")" "{" identifier ( "," identifier )* "}" ;

(* Indexing & Set Algebra *)
indexing          ::= expression "[" index_expr ( "," index_expr )* "]" ;
index_expr        ::= expression | "$" | range_expr ;

set_algebra_op    ::= "∩" | "∪" | "\" | "Δ" | "⊂" | "⊃" | "∈" | "!∈" ;
set_expr          ::= expression set_algebra_op expression ;
```

---

## 5. Diagnostic Error Codes

| Error Code | Error Condition | Description |
| :--- | :--- | :--- |
| `E1001` | `IndexOutOfBounds` | Index is less than 1 or exceeds collection length `$"` |
| `E1002` | `DuplicateSetKey` | Attempt to insert duplicate key/element into Set/Map |
| `E1003` | `MatrixDimensionMismatch` | Incompatible row/column dimension during assignment |
| `E1004` | `TypeIncompatibleCollection` | Value type does not match declared collection element type |
| `E1005` | `InvalidSetAlgebra` | Set operator (`∩`, `∪`) applied to non-set types |
| `E1006` | `ZeroBasedIndexAttempt` | Attempting to use index 0 in Bee 1-based indexing |

---

## 6. Alignment Status

- **Issues Addressed:** Formalized Lists `()`, Arrays `[]`, Matrices `[](r, c)`, Sets `{}` (with set algebra `∩`, `∪`, `\`), Maps `{:}` (with dynamic keys), Ordinals, strict 1-based indexing, and the `$` anchor.
- **Manifest Tracking:** Updated `MANIFEST.md` to reflect completion of `spec/10-collections.md`.
