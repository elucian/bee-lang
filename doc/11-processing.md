# Data Processing

Bee provides extensive tools for manipulating collections and strings.

## 1. Boxed Values
Boxing is the process of converting a primitive type to a reference.

```bee
new n ∈ Z;  -- Primitive
new k ∈ [Z]; -- Boxed integer
let k := [n]; -- Explicit boxing
```

## 2. Array Operations
Arrays support fast indexing and contiguous memory.

```bee
new test ∈ [R](10);
print test[1]; -- First element
print test[$]; -- Last element
```

## 3. Slicing & Spreading
Slices create a view of the parent array.

```bee
new a := [0,1,2,3,4,5,6,7,8,9];
new slice := a[2..5];
```

## 4. Collection Builders
Use builder syntax for sets and maps.

```bee
new s := {1, 2, 3};         -- Set
new m := {'key': "value"};  -- Hash Map
```

## 5. String Interpolation
Used for dynamic message formatting.

```bee
print ("User: #(name)" ? user);
```
