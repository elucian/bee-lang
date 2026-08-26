# Bee Specification: Library (14-library.md)

## 1. Standard API
```ebnf
io_call      ::= identifier "." identifier "(" arguments? ")" ;
introspection ::= "type" "(" identifier ")" | "size" "(" identifier ")" ;
```

## 2. System Operations
- **I/O:** `File` and `Folder` types are treated as opaque handles (`F`).
- **Standard Errors:** Errors are defined as `Object` types in the range `1..200`. 
  - System reserved: `< 200`.
  - Panic (unrecoverable): `≤ -1`.

## 3. Operational Semantics
- **Library Inclusion:** Bee uses a "pay-for-what-you-use" model; the standard library is linked at compile-time to minimize binary footprint.
- **Introspection:** `type()` and `size()` are evaluated at compile-time where possible, or runtime via reflection for dynamic collections.
- **File Lifecycle:** `open` creates a handle in the rule's local region; `close` is mandatory to release the OS descriptor.
