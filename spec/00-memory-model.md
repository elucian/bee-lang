# Bee Memory Model Specification (00-memory-model.md)

## 1. Overview
Bee employs a three-tier hybrid memory management architecture:
- **Reference Counting (RC):** Default for mutable objects.
- **Manual Management (MMM):** Explicit `new`/`zap` for Hot Zones.

## 2. Memory Management Directive: `zap`
The `zap` keyword explicitly relinquishes control of an object reference.

### Usage Policy
- **Hot Zone Priority:** `zap` MUST be used in performance-critical paths (e.g., hot loops) where Reference Counting (RC) overhead is prohibited.
- **Compiler Guarantee:** Subsequent access to a `zap`ped identifier triggers a compile-time diagnostic or runtime `Panic`.
