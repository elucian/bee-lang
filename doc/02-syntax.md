# Bee Syntax

## 1. Comments
Comments are inspired by Ada/PL/SQL:
- Single line: `--` (from start of line or EOL)
- Block: `+- ... -+` (unique "missing corner" box notation)

```bee
-- Single line comment
+- Multi-line 
   block comment -+
```

## 2. Keywords
Bee core reserves approximately 72 keywords (e.g., `rule`, `new`, `set`, `apply`, `done`, `repeat`). Reserved keywords cannot be used as identifiers.

## 3. Statements
Statements are imperative or declarative. Multiple statements on a line require `;`.

```bee
new a := 10;
let a := a + 1;
print a;
```

## 4. Control Flow
Bee uses explicit block terminators (`done`, `repeat`, `return`). Indentation is a mandatory syntactic requirement.

```bee
if condition do
  -- statements
done;

cycle:
  -- loop body
repeat;
```

## 5. Declarations & Identifiers
Variables must be declared. Identifiers support Latin, Greek, and Cyrillic.
- **Subscript:** `x₀`, `x₁`, etc.
- **Superscript:** `x¹`, `x²`, `xⁿ` (mapped to `x^n`).

## 6. Expressions & Conditional Selection
Expressions are combinations of identifiers, operators, and literals.
```bee
new kind := ("digit" if x ∈ ['0'..'9'] else "letter");
```

## 7. Error Handling
The `trial` statement is the primary error handling mechanism.
```bee
trial:
  try()
    fail {100, "Error"} if condition;
  case $error.code == 100 do
    resume;
  final
    -- Cleanup logic
done;
```
