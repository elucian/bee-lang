# Bee Specification: Object Oriented Model (06-objects.md)

## 1. Executive Object Model & Root Entity Philosophy

Bee employs a unified **Universal Entity Model**:
- **Root Object Hierarchy:** All values in Bee—including single-letter mathematical primitives (`Z`, `N`, `R`, `S`), collections (`array`, `list`, `map`, `set`), and user-defined instances—derive from the root **`Object`** (or `O`).
- **Universal `.type()` Introspection:** Because every entity derives from `Object`, every identifier or literal expression responds to the `.type()` introspection method without grammar ambiguity:
  ```bee
  new x := 42;
  print x.type(); -- Prints "Z"
  
  new str := "hello";
  print str.type(); -- Prints "S"
  
  new obj := Foo(1, 2);
  print obj.type(); -- Prints "Foo"
  ```
  *Note:* The global function `kind(entity)` serves as an alias for `entity.type()`.

---

## 2. Object Definition & Constructor Rules

An Object is created by a constructor **`rule`** whose return result is designated as **`self`**.

### 2.1 Constructor Rule Anatomy
```bee
-- Object constructor definition
rule MyObject(p1 ∈ Z, p2 ∈ R) => (self ∈ MyObject):
  -- Public properties (prefixed with self.)
  new self.prop1 := p1;
  new self.prop2 := p2;
  
  -- Private encapsulation (no prefix)
  new internal_cache := 0;
  
  -- Public method definition
  rule .compute(self ∈ MyObject) => (r ∈ R):
    let r := self.prop1 + self.prop2;
  return;
return;
```

### 2.2 Anonymous JSON & Dictionary Objects
Objects are internally backed by high-performance key-value maps:
```bee
-- Anonymous object instantiation via JSON literal
new obj := {name: "Cleopatra", age: 15};

-- Dynamic member property access
print obj.name;     -- Dot-notation: "Cleopatra"
print obj["age"];   -- Index-notation: 15

-- Dynamic property deletion
del obj["age"];
```

---

## 3. Encapsulation & Member Visibility

- **Public Members (`.` Prefix):** Any property or method prefixed with `.` (e.g. `self.value`, `.log()`) is exported on the `self` instance and accessible via dot-notation (`instance.log()`).
- **Private Encapsulation (No Prefix):** Local variables declared within the constructor rule without `self.` (e.g., `new count := 0;`) are strictly private to the constructor's static closure frame.
- **Dynamic Scope (`self`):** The `self` reference holds the dynamic instance heap context allocated when `new` invokes the constructor.

---

## 4. Inheritance (`<:`), `super.` & Abstraction

### 4.1 Subtype Inheritance (`<:`)
Subclasses declare inheritance using `<: SuperType` and initialize base state using `super(...)` or `super.method()`:

```bee
-- Base type
type Foo: {a ∈ N, b ∈ N} <: Object;

rule Foo(p1, p2 ∈ N) => (self ∈ Foo):
  let self := {a: p1, b: p2};
  
  rule .log(self ∈ Foo):
    print "a = ", self.a, " b = ", self.b;
  return;
return;

-- Derived subtype
type Bar: {a ∈ N, b ∈ N, c ∈ R} <: Foo;

rule Bar(p1, p2 ∈ N, p3 ∈ R) => (self ∈ Bar):
  -- Invoke supertype constructor
  let self := super(p1, p2);
  new self.c := p3;
  
  -- Override method with super call
  rule .log(self ∈ Bar):
    apply super.log();
    print "c = ", self.c;
  return;
return;
```

### 4.2 Abstract Types & Forward Method Signatures
An abstract constructor rule contains forward declared public method signatures without statement bodies. Derived types MUST override and supply implementations for all abstract methods:

```bee
-- Abstract interface rule
rule Shape() => (self ∈ Shape):
  rule .area(self ∈ Shape) => (a ∈ R); -- Abstract signature
return;

-- Concrete implementation
rule Circle(radius ∈ R) => (self ∈ Circle <: Shape):
  new self.r := radius;
  
  rule .area(self ∈ Circle) => (a ∈ R):
    let a := 3.14159265 * self.r ^ 2;
  return;
return;
```

---

## 5. Composition & Traits (`+`)

Traits provide reusable behavior composition (Multiple Inheritance via composition):

```bee
-- Trait rule definition
rule Printable(self ∈ Object):
  rule .print_summary(self ∈ Object):
    for k ∈ self.keys() do:
      print k, " => ", self[k];
    repeat;
  return;
return;

-- Applying trait composition
rule Item(id ∈ Z) => (self ∈ Item + Printable):
  new self.id := id;
  
  -- Augment self with trait methods
  apply Printable(self);
return;
```

---

## 6. Formal EBNF Grammar

```ebnf
(* Object Type Declarations *)
object_type       ::= "type" type_ident ":" "{" prop_list "}" "<:" super_type [ "with" "(" trait_list ")" ] ";" ;
prop_list         ::= prop_item ( "," prop_item )* ;
prop_item         ::= identifier [ ":" expression ] [ ( "∈" | "in" ) type_specifier ] ;

(* Constructors & Methods *)
constructor_def   ::= "rule" type_ident "(" [ param_list ] ")" "=>" "(" "self" [ "∈" type_specifier ] ")" ":" block "return" ";" ;
method_def        ::= "rule" "." identifier "(" [ param_list ] ")" [ "=>" "(" result_list ")" ] ":" block "return" ";" ;

(* Instantiation & Member Access *)
instantiation     ::= "new" identifier ":=" ( type_ident "(" [ arg_list ] ")" | "{" [ json_pairs ] "}" ) ;
member_access     ::= expression "." identifier [ "(" [ arg_list ] ")" ]
                    | expression "[" expression "]" ;
super_call        ::= "super" [ "." identifier ] "(" [ arg_list ] ")" ;
```

---

## 7. Diagnostic Error Codes

| Error Code | Error Condition | Description |
| :--- | :--- | :--- |
| `E0601` | `AbstractMethodUnimplemented` | Concrete class fails to supply body for inherited abstract method |
| `E0602` | `PrivatePropertyAccess` | Attempt to access property without `.` or `self.` export prefix |
| `E0603` | `SuperConstructorMissing` | Derived constructor fails to execute `super(...)` call |
| `E0604` | `CyclicInheritance` | Class hierarchy contains circular inheritance loop |
| `E0605` | `UnknownMemberKey` | Accessing dynamic key that does not exist in object dictionary |
| `E0606` | `InvalidTraitApplication` | Trait applied to incompatible target type |

---

## 8. Alignment Status

- **Issues Addressed:** Formalized Universal Entity Model (`entity.type()`), Object constructors, public/private member scoping, inheritance (`<:`), `super.`, abstract methods, traits (`+`), and dictionary keys.
- **Manifest Tracking:** Updated `MANIFEST.md` to reflect completion of `spec/06-objects.md`.
