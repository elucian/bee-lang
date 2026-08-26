# Bee Specification: Objects (06-objects.md)

## 1. Object Anatomy
Objects are high-order rule instances. A rule named `TName` acting as a constructor returns a `self` context, which becomes the object instance.

```ebnf
object_decl     ::= "rule" identifier "(" parameters? ")" "=>" "(" "self" "∈" identifier ")" block "return" ;
object_member   ::= "." identifier "(" parameters? ")" block "return" ;
```

## 2. Inheritance & Abstraction
- **Inheritance:** Derived types use `<:` to specify the supertype and `super.` to invoke the base constructor.
- **Abstraction:** Abstract types contain forward declarations (signatures without bodies). Concrete types must implement all abstract methods.
- **Traits:** Augmented types use the `+` symbol for composition (Multiple Inheritance via composition).

## 3. Operational Semantics
- **Context:** The `self` pointer holds the dynamic object scope (instance data).
- **Public Members:** Indicated by the `.` prefix. Private members have no prefix.
- **Allocation:** Objects reside in the heap-backed memory region of the calling context until `zap` is called.
- **Instantiation:** Keyword `new` triggers the constructor rule.
