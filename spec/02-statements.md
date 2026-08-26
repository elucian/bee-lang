# Bee Specification: Statements & Execution Control (02-statements.md)

## 1. Executive Statement Taxonomy

Bee divides statements into five distinct syntactic categories:

1. **Declarative Statements:** Allocate variable names and bind initial static/dynamic types (`new`, `const`).
2. **Mutation Statements:** Rebind values, clone memory structures, or apply in-place mathematical updates (`let`, `alter`, `:=`, `::`, compound operators).
3. **Control Flow Statements:** Direct branching, pattern matching, and iteration loops (`if`, `match`, `cycle`, `while`, `for`, `with`).
4. **Error Handling & Trial Statements:** Transactional exception handling and structured cleanup (`trial`, `try`, `fail`, `case`, `miss`, `final`).
5. **Transfer & Termination Statements:** Direct jump and routine completion (`return`, `stop`, `next`, `yield`, `raise`, `retry`).

---

## 2. Declarations, Mutations & Assignment Operators

### 2.1 Variable Declaration (`new`)
- **Explicit Type Declaration:** `new identifier ∈ Type;` (e.g. `new count ∈ Z;`)
- **Inferred Type Declaration:** `new identifier := expression;` (e.g. `new total := 0;`)
- **Constant Declaration:** `const CONSTANT_NAME := expression;` (e.g. `const MaxBuffer := 1024;`)

### 2.2 Mutation Semantics (`let` vs `alter`)
- **Reassignment (`let`):** `let identifier := expression;` modifies an existing variable in the current scope.
- **In-Place Structural Mutation (`alter`):** `alter identifier := expression;` forces in-place mutation of fields, collections, or boxed variables without creating a new binding.
- **Deep Clone Assignment (`::`):** `let target :: source;` performs a deep copy of nested collection/object structures, disconnecting ARC references.
- **Compound Mutation Operators:**
  `+=` (add), `-=` (subtract), `*=` (multiply), `/=` (divide), `%=` (modulo), `^=` (power), `√=` (square root).

### 2.3 Memory Directives
- **Explicit Deallocation:** `zap identifier;` invalidates the target identifier in Hot Zone performance paths.

---

## 3. Control Flow Mechanics

### 3.1 Conditional Execution (`if / else`)
- **Block Structure:**
  ```bee
  if condition then:
    statement_1;
    statement_2;
  else:
    statement_3;
  done;
  ```
- **Conditional Expression Selector (Ternary Alternative):**
  ```bee
  new value := (expr_true if condition else expr_false);
  ```

### 3.2 Pattern & Decision Matching (`match`)
- Supports **First Match** (`match_first` / `match`) and **Match Every** (`match_every`) modes.
  ```bee
  match mode:
    when cond_1 do:
      -- branch 1
    when cond_2 do:
      -- branch 2
    other:
      -- fallback default
  done;
  ```

### 3.3 Iteration & Cycle Blocks (`cycle`, `while`, `for`)
1. **Infinite Cycle / Stop Condition:**
   ```bee
   cycle loop_label:
     if exit_condition then:
       stop loop_label;
     done;
   repeat loop_label;
   ```
2. **While Cycle:**
   ```bee
   while condition do:
     -- body
   repeat;
   ```
3. **Collection / Range For Cycle:**
   ```bee
   for element ∈ collection_or_range do:
     -- body
   repeat;
   ```

### 3.4 Scoped Qualifier Block (`with`)
- Simplifies module and object qualifier access without repeating prefixes:
  ```bee
  with object_or_module do:
    method_1();
    method_2();
  done;
  ```

---

## 4. Error Handling & Trial Semantics (`trial`)

The `trial` block is Bee's unified transactional exception management framework:

```bee
trial trial_label:
  try()
  fail {code: 100, message: "Buffer Overflow"} if check_failed;
case $error.code = 100 do:
  -- Handle specific code
  resume;
miss:
  -- Fallback handler for unhandled errors
final:
  -- Mandatory cleanup block executed unconditionally
done trial_label;
```

### Transfer Directives in Errors:
- `resume`: Suppresses the error and continues execution after the failing statement.
- `retry`: Re-executes the enclosing `trial` block from the beginning.
- `raise`: Re-raises the captured error up the call stack.

---

## 5. Formal EBNF Statement Grammar

```ebnf
(* Statements *)
statement         ::= decl_stmt | mutation_stmt | memory_stmt | io_stmt | control_stmt | trial_stmt | transfer_stmt ";" ;

decl_stmt         ::= "new" identifier ( "∈" | "in" ) type_specifier [ ":=" expression ]
                    | "new" identifier ":=" expression
                    | "const" type_ident ":=" expression ;

mutation_stmt     ::= ( "let" | "alter" ) identifier assign_op expression ;
assign_op         ::= ":=" | "::" | "+=" | "-=" | "*=" | "/=" | "%=" | "^=" | "√=" ;
memory_stmt       ::= "zap" identifier ;
io_stmt           ::= "print" expression ( "," expression )*
                    | "write" expression ;

(* Control Flow *)
if_stmt           ::= "if" condition "then" ":" block [ "else" ":" block ] "done" ;
match_stmt        ::= "match" [ match_mode ] ":" ( "when" condition "do" ":" block )+ [ "other" ":" block ] "done" ;
match_mode        ::= "first" | "every" | "total" ;

cycle_stmt        ::= "cycle" [ label ] ":" block "repeat" [ label ]
                    | "while" condition "do" ":" block "repeat"
                    | "for" identifier ( "∈" | "in" ) expression "do" ":" block "repeat" ;

with_stmt         ::= "with" expression "do" ":" block "done" ;

(* Error Handling *)
trial_stmt        ::= "trial" [ label ] ":" block [ "case" condition "do" ":" block ]* [ "miss" ":" block ] [ "final" ":" block ] "done" [ label ] ;

(* Transfers *)
transfer_stmt     ::= "return" [ expression ]
                    | "stop" [ label ]
                    | "next" [ label ]
                    | "yield" [ expression ]
                    | "raise" [ expression ]
                    | "retry" ;
```

---

## 6. Block Indentation & Alignment Rules

1. **Mandatory 2-Space Indentation:** All statements inside a block body MUST be indented by exactly 2 spaces relative to the block header statement (`rule`, `if`, `cycle`, `while`, `for`, `with`, `trial`).
2. **Block Terminator Alignment:** Block terminators (`done`, `repeat`, `return`) MUST align horizontally with the indentation level of their corresponding block header (0 relative indentation).
3. **Nested Blocks:** Each nested level adds +2 spaces of indentation.

```bee
rule main:
  if x > 0 then:
    print "positive";
  else:
    print "non-positive";
  done;
return;
```

---

## 7. Diagnostic Error Codes

| Error Code | Violation | Description |
| :--- | :--- | :--- |
| `E0201` | `IndentationMismatch` | Statement is not aligned to 2-space offset boundary |
| `E0202` | `UnboundVariable` | Attempt to mutate variable without prior `new` declaration |
| `E0203` | `UnterminatedBlock` | Missing `done`, `repeat`, or `return` terminator |
| `E0204` | `InvalidCloneOperation` | Using `::` clone operator on non-clonable primitive |
| `E0205` | `LabelMismatch` | Closing label on `repeat` or `done` does not match opening header label |

---

## 8. Alignment with Solution & Issues

- **Issues Addressed:** Resolves `issues/01-syntax-design.md` statement specification tasks.
- **Manifest Tracking:** Updated `MANIFEST.md` to reflect completion of `spec/02-statements.md`.
