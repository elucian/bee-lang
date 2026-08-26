# Bee Specification: Type System Architecture (05-types.md)

## 1. Executive Type System Architecture

Bee utilizes a **strongly typed, static type system with global type inference**. Types represent domain constraints over values.

- **Universal Entity Model & Introspection:** Every value or variable in Bee is an **`Entity`** that exposes the `.type()` introspection method (e.g. `10.type()`, `"hello".type()`). The legacy `kind()` function is removed in favor of `entity.type()`.
- **Mathematical Primitive Designator:** Single uppercase Latin letters are strictly reserved for primitive data types (e.g., `Z` for Integer, `R` for Real, `Q` for Rational).
- **Custom & Subtype Identifiers:** User-defined types must use title-cased identifiers (`TypeIdentifier`).

---

## 2. Primitive & Built-In Type Catalogue

### 2.1 Single-Letter Primitive Types
| Type Symbol | Alias | Representation / Width | Alignment | Default Value | Description |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **`B`** | Boolean | 8-bit unsigned integer | 1 byte | `0b0` (`false`) | Logic boolean: `0` = False, `≥1` = True |
| **`A`** | Alpha | 8-bit E-ASCII character | 1 byte | `'0'` | ASCII character literal (`'a'`, `'Z'`, `'0'`) |
| **`U`** | Unicode | 32-bit unsigned code point | 4 bytes | `U+0000` | UTF-32 Rune code point (`'Ω'`, `U+0041`) |
| **`N`** | Natural | 64-bit unsigned integer | 8 bytes | `0` | Non-negative integer range $[0 \dots 2^{64}-1]$ |
| **`Z`** | Integer | 64-bit signed integer | 8 bytes | `0` | Signed 2's complement $[-2^{63} \dots 2^{63}-1]$ |
| **`R`** | Real | 64-bit IEEE 754 Float | 8 bytes | `0.0` | Double precision floating point |
| **`Q`** | Rational | Fixed-point fraction / $Qm.n$ | 4–16 bytes | `0\1` | Fixed-point fraction $p/q$ or $Q(14.17)$ |

### 2.2 Special Reference & Domain Types
| Type Symbol | Alias | Heap Structure | Description |
| :--- | :--- | :--- | :--- |
| **`C`** | Complex | Float Pair $(a + bj)$ | 128-bit pair of double precision Reals |
| **`S`** | String | UTF-8 Slice / Rope | GC-managed immutable string payload |
| **`D`** | Date | Struct `(day, month, year)` | Calendar date representation |
| **`T`** | Time | Struct `(h, m, s, ms)` | Time of day representation |
| **`L`** | Lambda | Function Pointer Descriptor | Callable closure or rule reference |
| **`G`** | Angular | Fixed 16-bit float | Angular range $[1^\circ \dots 360^\circ]$ |

---

## 3. Subtypes (`<:`), Ranges & Domain Constraints

### 3.1 Type Aliasing & Subtype Inheritance
Subtypes constrain existing types to specific value domains using the `<:` operator:

```bee
-- Custom type declarations
type Digit: ('0'..'9') <: Z;
type Capital: ('A'..'Z') <: A;
type LatinRune: (U+0041..U+FB02) <: U;
```

### 3.2 Range Notation & Limits
Ranges define numeric or character bounds:
- `(min..max)`: Fully inclusive range $[min, max]$.
- `(min.!max)`: Left-inclusive, right-exclusive $[min, max)$.
- `(min!.max)`: Left-exclusive, right-inclusive $(min, max]$.
- `(min!!max)`: Fully open / exclusive $(min, max)$.
- Unbounded open limits use `-` or `+` (e.g. `(0..+)` for non-negative numbers).

### 3.3 Domain Types with Step Ratio
Domains extend ranges by specifying a discretization step ratio `(min..max:step)`:
```bee
-- Domain producing rational step 1\4
new q_domain := (0..1: 1\4); -- 0\4, 1\4, 2\4, 3\4, 1\1

-- Domain producing float step 0.25
new r_domain := (0..1: 0.25); -- 0.00, 0.25, 0.50, 0.75, 1.00
```

---

## 4. Rational Fixed-Point Arithmetic & Approximate Comparison (`≈`)

### 4.1 $Q$ Notation & Fixed-Point Representation
Fixed-point rationals are expressed as $Qm.n$ where $m$ is integer bits, $n$ is fractional bits, and step resolution is $2^{-n}$:
- **Default Rational Format:** $Q(14.17)$ stored in a 32-bit container with resolution $2^{-17} \approx 0.0000076$.
- **Literal Notation:** Numerator backslash denominator $p \backslash q$ (e.g., `1\2`, `1\4`, `3\16`).

### 4.2 Approximate Equality (`≈`) & Tolerance (`±`)
Because rationals and reals represent real-world approximations, Bee provides native approximate equality testing:
- **Default Comparison:** `a ≈ b` checks if $|a - b| \le \$precision$ (where `$precision` defaults to $10^{-5} = 0.00001$).
- **Explicit Tolerance Modifier:** `a ≈ b ± tolerance` overrides the default precision threshold:
  ```bee
  set $precision := 0.01;
  new a := 0.25; -- Real
  new b := 1\3;  -- Rational (0.333...)
  
  pass if (a ≈ b ± 0.10); -- True because |0.25 - 0.333| = 0.083 <= 0.10
  ```

### 4.3 Type Cast Operator (`:>`)
- **Implicit Promotion:** $B \rightarrow N \rightarrow Z \rightarrow R \rightarrow Q \rightarrow C$.
- **Explicit Casting:** Conversion from Float to Rational requires explicit cast operator `:>`:
  ```bee
  new real_val := 0.25;
  new rat_val := real_val :> Q; -- Explicit conversion to Rational
  expect rat_val = 1\4;
  ```

---

## 5. Type Promotion Hierarchy

When operators combine operands of different primitive types, Bee applies deterministic type promotion:

```
  B (Boolean)
   │
   ▼
  N (Natural) ──► A (Alpha)
   │
   ▼
  Z (Integer) ──► U (Unicode)
   │
   ▼
  R (Real)
   │
   ▼
  Q (Rational)
   │
   ▼
  C (Complex)
```

---

## 6. Formal EBNF Grammar

```ebnf
(* Type Declarations *)
type_decl         ::= "type" type_ident ":" type_descriptor [ "<:" super_type ] ";" ;
type_descriptor   ::= primitive_type | range_expr | domain_expr | collection_type ;

primitive_type    ::= "B" | "A" | "U" | "N" | "Z" | "R" | "Q" | "C" | "S" | "D" | "T" | "L" | "G"
                    | "Q" "(" integer_lit "." integer_lit ")" ;

(* Ranges & Domains *)
range_expr        ::= "(" limit range_sep limit ")" ;
domain_expr       ::= "(" limit range_sep limit ":" step_expr ")" ;
limit             ::= [ "-" | "+" ] ( literal | identifier ) ;
range_sep         ::= ".." | ".!" | "!." | "!!" ;
step_expr         ::= expression ;

(* Type Casting & Universal Introspection *)
type_cast         ::= expression ":>" type_specifier ;
type_check        ::= expression ( "∈" | "in" ) type_specifier ;
type_intro        ::= expression "." "type" "(" ")" ;
```

---

## 7. Diagnostic Error Codes

| Error Code | Error Condition | Description |
| :--- | :--- | :--- |
| `E0501` | `TypeMismatch` | Assigned value violates static type boundary |
| `E0502` | `InvalidTypeCast` | Incompatible explicit cast with `:>` operator |
| `E0503` | `DomainOutOfBounds` | Assigned literal falls outside constrained range/domain |
| `E0504` | `RationalOverflow` | Value exceeds maximum capacity of specified $Qm.n$ container |
| `E0505` | `SingleLetterTypeReserved` | User type identifier uses reserved single uppercase Latin letter |
| `E0506` | `PrecisionUndefined` | Approximate comparison `≈` executed with uninitialized `$precision` |

---

## 8. Alignment Status

- **Issues Addressed:** Formalized mathematical primitives, fixed-point $Qm.n$ rationals, subtyping (`<:`), ranges/domains, approximate equality (`≈`), type casting (`:>`), and universal `.type()` introspection (`kind()` removed).
- **Manifest Tracking:** Updated `MANIFEST.md` to reflect completion of `spec/05-types.md`.
