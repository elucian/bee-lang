# Bee Specification: Module Architecture & Program Structure (04-structure.md)

## 1. Executive Program Architecture

A Bee application is built as a graph of isolated, single-responsibility **modules**. Each module resides in its own `.bee` source file and defines a strict lexical scope (namespace).

```text
$pro_home/             # Project Root Directory
├── bin/               # Compiled binary executables
├── src/               # Application secondary modules
│   ├── math_utils.bee # Secondary module (src/math_utils.bee)
│   └── data_store.bee # Secondary module (src/data_store.bee)
├── lib/               # Project-local library modules
│   └── logger.bee     # Library module (lib/logger.bee)
├── main.bee           # Main module application entry point
└── config.json        # Environment configuration
```

---

## 2. Module Classification & Rules

### 2.1 Main Module (`main.bee`)
- **Role:** Application orchestrator and executable entry point.
- **Invariants:**
  1. MUST contain exactly one `rule main(*params ∈ S) => (code ∈ Z):` or `rule main:`.
  2. MUST NOT be imported or loaded by any secondary or library module.
  3. Serves as the top-level parent scope for global system configuration directives.

### 2.2 Secondary Modules (`src/`)
- **Role:** Project-specific domain logic, data models, and sub-orchestrators.
- **Invariants:**
  1. MUST NOT contain a `rule main`.
  2. Loaded via local import directive: `use module_name;` or `use qualifier:module_name;`.
  3. Accessible via explicit module qualifier dot-notation (`module_name.member`).

### 2.4 Module Lifecycle Persistence Invariant
- **Lifecycle**: All modules (Secondary and Library) are loaded as singleton instances upon the first `use` directive.
- **Persistence**: Loaded modules are persistent in memory. Explicit dynamic unloading or reloading of modules is NOT supported. Module state exists for the entire execution lifecycle of the program and is finalized only upon program termination.

---

## 3. System Variables & Compiler Directives

System paths, compiler settings, and environment variables are protected behind the **`$`** sigil to prevent accidental shadowing by user code.

### 3.1 Standard System Path Variables
| Sigil Identifier | Description | Default Resolution |
| :--- | :--- | :--- |
| `$bee_home` | Bee compiler runtime installation directory | OS System Installation Path |
| `$bee_lib` | Global system library root directory | `$bee_home/lib/` |
| `$pro_home` | Current application project root directory | Location of `main.bee` |
| `$pro_lib` | Project-local library directory | `$pro_home/lib/` |
| `$pro_mod` | Secondary module search path list | `$pro_home/src/` |
| `$pro_log` | Diagnostic and execution log folder | `$pro_home/log/` |

### 3.2 Compiler Directives & Execution Limits
Global compiler control parameters are defined at top-of-file scope or in configuration:
- `$max_precision`: Precision threshold for rational arithmetic (Default: `0.00001`).
- `$max_recursion`: Maximum allowed call-stack call depth (Default: `10000`).
- `$log_debug`: Toggles debug instrumentation output (`"On"` / `"Off"`).
- `$platform`: Target OS runtime build platform (`"Windows"`, `"Linux"`, `"Darwin"`).

---

## 4. Name Space & Encapsulation Rules

### 4.1 Member Visibility
Module encapsulation is governed by a strict prefix convention:

- **Public Members (`.` Prefix):** Members prefixed with `.` are exported and accessible outside the module scope.
  ```bee
  #module math_utils
  set .pi := 3.14159265; -- Public constant
  
  rule .add(a, b ∈ R) => (r ∈ R): -- Public rule
    let r := a + b;
  return;
  ```
- **Private Members (No Prefix):** Members without `.` are strictly private to the defining module file.
  ```bee
  new helper_cache ∈ [R]; -- Private module variable
  
  rule compute_internal(x ∈ R) => (r ∈ R): -- Private rule
    let r := x * 2;
  return;
  ```

### 4.2 Qualifier Suppression (`with`) & Symbol Aliasing
- **Qualifier Suppression (`with`):** Inside a `with` block, exported members of the specified module/object can be invoked directly without repeating the qualifier prefix:
  ```bee
  use qualifier:math_utils;
  
  with math_utils do:
    print .add(10, 20); -- Invokes math_utils.add
  done;
  ```
- **Symbol Aliasing (`alias`):** Binds a module export to a local alias:
  ```bee
  alias sum: math_utils.add;
  print sum(10, 20);
  ```

---

## 5. Execution Primitives & Scope Boundaries

Bee separates execution modes into distinct primitives:

1. **Synchronous Execution (`apply` / Call):**
   - Executed inline on the calling thread within the active Region Arena frame.
   - `apply module.rule(args);` executes for side effects, discarding return values.

2. **Asynchronous Parallel Spawning (`begin`):**
   - `begin module.rule(args);` spawns a new concurrent task thread.
   - The task receives an isolated Region Arena and executes concurrently with the parent thread.

3. **Barrier Synchronization (`wait`):**
   - `wait;` acts as a synchronization barrier, blocking caller execution until ALL asynchronous tasks spawned in the current scope complete.
   - Any worker thread errors or panics captured in worker `$trial` state are re-raised at the `wait` barrier.

---

## 6. Formal EBNF Grammar

```ebnf
(* Module Header & Directives *)
module_file       ::= [ module_header ] ( import_stmt | directive_stmt )* ( member_decl )* ;
module_header     ::= ( "#module" | "#define" ) identifier ;
import_stmt       ::= "use" [ "qualifier:" ] module_path [ "as" identifier ] ";" ;
module_path       ::= [ "$" identifier "." ] identifier ( "." identifier )* ;
directive_stmt    ::= "$" identifier ":=" expression ";" ;

(* Member Visibility & Declarations *)
member_decl       ::= [ "." ] ( variable_decl | rule_def | object_def | const_decl ) ;
alias_stmt        ::= "alias" identifier ":" module_path "." identifier ";" ;

(* Scoped Blocks & Execution *)
with_stmt         ::= "with" expression "do" ":" block "done" ;
sync_execution    ::= "apply" module_path "." identifier "(" [ arg_list ] ")" ";" ;
async_execution   ::= "begin" module_path "." identifier "(" [ arg_list ] ")" ";" ;
wait_barrier      ::= "wait" ";" ;
```

---

## 7. Diagnostic Error Codes

| Error Code | Violation | Description |
| :--- | :--- | :--- |
| `E0401` | `MainModuleRedefinition` | Secondary or library module contains `rule main` |
| `E0402` | `PrivateMemberAccess` | Attempt to access unexported member (missing `.` prefix) |
| `E0403` | `ModuleNotFound` | Module file missing from search path (`$pro_mod`, `$pro_lib`, `$bee_lib`) |
| `E0404` | `CyclicImportError` | Circular dependency detected between imported modules |
| `E0405` | `MainImportForbidden` | Attempt to load `main.bee` via `use` directive |
| `E0406` | `UnsynchronizedWorker` | Scope exited with pending `begin` threads prior to `wait` barrier |

---

## 8. Alignment Status

- **Issues Addressed:** Formalized module types, system paths (`$`), encapsulation rules (`.`), `#using`/`use`, `with` blocks, and `begin`/`wait` primitives.
- **Manifest Tracking:** Updated `MANIFEST.md` to reflect completion of `spec/04-structure.md`.
