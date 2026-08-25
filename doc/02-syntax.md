# Bee Syntax

## Comments
Comments use "--" for single lines and "+-" ... "-+" for blocks.

```bee
-- Single line comment
+- Multi-line 
   block comment -+
```

## Keywords & Statements
Bee uses imperative keywords. Statements are terminated by ";".

```bee
new a := 10;     -- Declaration
let a := 20;     -- Assignment
print a;         -- Execution
```

## Declarations
Explicit typing via `∈`.

```bee
new x ∈ Z;       -- Signed integer
set y := 10 ∈ N; -- Static natural
```

## Expressions & Control
Control flow uses explicit block terminators.

```bee
if x > 0 do
  print x;
done;

cycle:
  let x := x - 1;
repeat;
```

## Error Handling
The `trial` block is the primary error handling mechanism.

```bee
trial:
  try()
    fail {100, "Error"} if x < 0;
  case $error.code == 100 do
    resume;
done;
```
