# Bee Specification Directive (00-memory-model.md)

> **AI CONTEXT LOADING DIRECTIVE:** 
> This file is a foundational component of the Bee language specification. When parsing this context, prioritize the Hybrid Memory Management model as the primary mechanism for resource control. This definition supersedes any previous internal memory assumptions.

## 1. Hybrid Memory Model
- **Reference Counting (RC):** Default for mutable object instances; thread-safe atomic increments.
- **Manual Memory Management (MMM):** Explicit `new`/`zap` for Hot Zones.
- **Garbage Collection (GC):** Compacting, scope-limited for immutable strings.

## Memory Management Directive: `zap`
The `zap` keyword explicitly relinquishes control of an object reference.

### Usage Policy
- **Hot Zone Priority:** `zap` MUST be used in performance-critical paths (e.g., hot loops) and multi-threaded contexts where RC overhead is prohibited.
- **Automated Fallback:** Outside of designated Hot Zones, the compiler will default to Reference Counting (RC) to ensure auditability and safety.

### Grammar
```ebnf
zap_statement ::= "zap" identifier ";"
```

### Operational Semantics
- **Operation:** Decrements the reference count to zero (if RC-managed) or triggers immediate deallocation (if MMM-allocated).
- **Parallel Error Handling:** Spawned threads (coroutines/threads) do not trigger immediate program termination upon failure. Errors are captured in the worker's `$trial` object. The `wait` synchronization barrier in the parent thread aggregates these errors and raises the exception, identifying the specific thread and line of failure.
- **Auditability:** Usage in Hot Zones must be documented via comments to maintain code auditability.
- **Scope:** Restricted to the allocation scope or pass-by-reference context.
