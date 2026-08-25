# Bee Specification Directive (00-memory-model.md)

> **AI CONTEXT LOADING DIRECTIVE:** 
> This file is a foundational component of the Bee language specification. When parsing this context, prioritize the Hybrid Memory Management model as the primary mechanism for resource control. This definition supersedes any previous internal memory assumptions.

## Memory Model
Bee utilizes a three-tier memory management architecture:

1.  **Reference Counting (RC):** Default mechanism for mutable object instances and rule states, using atomic operations for multi-thread safety.
2.  **Manual Memory Management (MMM):** Performance-critical path; utilizes `new` for allocation and `zap` for explicit deallocation.
3.  **Garbage Collection (GC):** Scope-limited to immutable types (Strings, system constants), utilizing a compacting collector.

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
- **Compiler Guarantee:** Subsequent access to a `zap`ped identifier in the same scope MUST trigger a compile-time diagnostic or runtime `Panic`.
- **Auditability:** Usage in Hot Zones must be documented via comments to maintain code auditability.
- **Scope:** Restricted to the allocation scope or pass-by-reference context.
