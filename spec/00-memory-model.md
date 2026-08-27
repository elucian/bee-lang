# Bee Specification: Memory Model Architecture (00-memory-model.md)

## 1. Executive Architectural Overview

Bee utilizes a deterministic, three-tier hybrid memory management model engineered for zero-overhead performance in critical paths, safe parallel execution, and low latency:

1. **Atomic Reference Counting (ARC):** Default management for mutable dynamic structures, object instances, and system resources. Provides immediate, deterministic resource cleanup upon last reference release.
2. **Region-Based Arena Allocation & Manual Memory Control (MMM / `zap`):** Local arena allocations scoped to rule boundaries with zero-allocation-overhead performance. Explicit deallocation using `zap` bypasses ARC overhead in Hot Zones.
3. **Compacting Generational Garbage Collection (GC):** Special-purpose garbage collector specifically dedicated to immutable strings, string slices, and rope buffers, guaranteeing efficient interning, zero fragmentation, and lock-free thread-safe sharing.

---

## 2. Memory Tier Invariants & Memory Layouts

### 2.1 Tier 1: Atomic Reference Counting (ARC)
- **Target Allocation Scope:** All mutable heap allocations including objects (`object`), dynamic collections (`array`, `map`, `list`, `set`), and system OS handles (files, channels, sockets).
- **Header Layout:** Every ARC-allocated object contains a standardized 16-byte header:
  ```
  +-------------------+-------------------+-----------------------+
  |  RefCount (64bit) |   TypeID (32bit)  | Flags / Metadata(32) |
  +-------------------+-------------------+-----------------------+
  |                   Payload Data (Variable Width)               |
  +---------------------------------------------------------------+
  ```
- **Operational Semantics:**
  - Upon assignment/copy: `AtomicInc(RefHeader.RefCount)`
  - Upon scope exit/re-assignment: `if AtomicDec(RefHeader.RefCount) == 0 then DestructAndFree(Object)`
  - Destructor execution is immediate and synchronous, releasing backing native resources (e.g. file descriptors) on the calling thread.

### 2.2 Tier 2: Region-Based Arena Allocation & `zap` Directive
- **Region Boundary:** Every `rule` or nested block execution frame allocates a light stack-bound Region Arena (2KB initial chunk).
- **Transient Objects:** Short-lived local variables allocate directly from the active Region Arena offset without individual heap allocations (`ptr = arena.top; arena.top += size`).
- **Region Cleanup:** Upon reaching block termination (`return`, `done`, `repeat`), the entire Region Arena offset is unwound in a single instruction (`arena.top = arena.base`), instantly reclaiming all region-allocated memory.
- **Manual Deallocation (`zap`):**
  - Syntax: `zap identifier;`
  - In Hot Zones (performance-critical loops), `zap` explicitly invalidates the pointer and resets the target memory slot immediately.
  - **Static Analysis Invariant:** The compiler performs lifetime tracking. Reading or writing an identifier post-`zap` within the same control flow graph triggers a compile-time error `E0401: AccessAfterZap`.
  - **Runtime Safety Guard:** In non-optimized debug builds (`-d`), `zap` zero-fills the target pointer slot; subsequent dereference triggers a runtime `Panic: UseAfterZap`.

### 2.3 Tier 3: Compacting Generational GC (Immutable Strings & Ropes)
- **Target Allocation Scope:** Immutable string literals, dynamic string concatenations, and rope nodes.
- **GC Roots:** Thread execution stacks, Region Arenas, and ARC object payload pointers pointing into the String Heap.
- **Compaction Phase:** Generational GC operates independently on background threads. Because string payloads are strictly immutable, compaction uses two-space copying without requiring write barriers on reader threads.

---

## 3. Concurrency & Cross-Thread Memory Boundaries

### 3.1 Thread Isolation & Ownership Transfer
- **Spawned Threads (`begin`):** Spawned routine tasks operate within isolated Region Arenas.
- **Immutable Sharing:** GC-managed immutable strings and ropes can be passed across thread boundaries without locks or reference copies.
- **Mutable ARC Transfer:** Passing a mutable ARC object to a spawned thread increments its atomic reference count (`LOCK XADD`).
- **Parallel Error Isolation:** Uncaught exceptions or panics within worker threads freeze the thread's local Region Arena and capture diagnostic context into the thread-local `$trial` handle, preventing corruption of parent state.

---

## 4. Formal EBNF Grammar

```ebnf
(* Memory Management Statements *)
zap_statement     ::= "zap" identifier ";" ;
new_statement     ::= "new" identifier [ ":" type_specifier ] [ ( ":=" | "::" ) expression ] ";" ;
let_statement     ::= "let" identifier ( ":=" | "::" ) expression ";" ;

### Assignment (`:=`) vs Clone Assignment (`::`)
- **Operator `:=` Evaluation:**
  - With `new` (`new x := expr;`), `:=` allocates a new storage location.
  - With `let` (`let x := expr;`), `:=` performs an update action on the existing value of the variable.
- **Value Modification Rules (`let` with `:=`)**: 
  - If the variable is native, it changes its value in place.
  - If the variable is boxed, it changes its boxed value in place.
  - If a reference is on the right side, the value is transferred via deep copy.
- **Clone Assignment (`::`)**: For boxed variables or objects, `::` performs a deep copy/structural clone allocating an independent payload on the heap. For primitive native variables, `::` and `:=` are equivalent.

(* Variable Mutability Modifiers *)
variable_decl     ::= "new" identifier ( "∈" | "in" ) type_specifier ";" ;
mutability_prefix ::= "let" | "alter" ;
```

---

## 5. Diagnostics & Safety Protocols

| Error Code | Violation Description | Mitigation / Diagnostic Action |
| :--- | :--- | :--- |
| `E0401` | Use of identifier after `zap` statement | Compile-time fatal error with source line location |
| `E0402` | Pointer escape from Region Arena to outer scope | Compiler automatic promotion from Region to ARC Heap |
| `E0403` | Unhandled cycle in ARC structure | Static lifetime analysis or runtime leak warning under `-d` flag |
| `E0404` | Invalid cross-thread mutation of unshared reference | Compile-time data race restriction |

---

## 6. Issue & Solution Alignment Status

- **Issues Addressed:** Fully resolves `issues/MEMORY_MODEL.md` by formalizing Region Arenas, `zap` static analysis guarantees, and error diagnostics.
- **Solution Verification:** Aligned with `solution/01-memory-management.md` and `solution/03-region-based-memory.md`.
