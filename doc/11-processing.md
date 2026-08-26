# Bee Data Processing

Collections are data structures that group a limited number of values together. You can access individual values using different methods, depending on the collection type.

## 1. Boxed Values
A boxed value is a reference to a primitive type, converted via boxing.

```bee
-- define boxed values
new int ∈ [Z]; -- boxed integer
new flt ∈ [R]; -- boxed double float

-- Explicit boxing
let k := [n]; 

-- Explicit unboxing
let n := Z(r);
```

## 2. Share vs Copy
- `:=` creates a shared reference binding (aliasing).
- `::` creates a deep clone.

```bee
new a := [1];
new b := a;   -- Shared reference
new c :: a;   -- Deep clone
```

## 3. Array Operations
Arrays support fast direct access. **Bee uses 1-based indexing.**

```bee
new test ∈ [R](10);
print test[1]; -- First element
print test[$]; -- Last element
```

## 4. Spreading & Decomposition
```bee
new array := [1, 2, 3, 4, 5];
new x, y, *other := array;
-- x = 1, y = 2, other = [3, 4, 5]
```

## 5. Matrix Operations
Matrices are multi-dimensional arrays, row-major order.

```bee
new M: [Z](2, 2);
let M[1,1] := 100;
let M[1,*] := 0; -- Modify entire row
```

## 6. Collections & String Literals
- **Sets:** Sorted unique collections.
- **Maps:** Key-value pairs.
- **String Literals:** Unicode by default; supports markup tags (`<text>`, `<sql>`, etc.) for literals.

```bee
new s := {1, 2, 3};
new sqlQuery ∈ S := <sql> select * from users; </sql>;
```
