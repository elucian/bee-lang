# Bee Memory Model Specification (00-memory-model.md)

## 1. Overview
Bee employs a three-tier hybrid memory management architecture:
- **Reference Counting (RC):** Default for mutable objects.
- **Manual Management (MMM):** Explicit `new`/`zap` for Hot Zones.

## 3. Boundary Management & Thread Safety
- **Boundary Tracing:** Static analysis ensures RC-managed mutable structures pointing to GC-managed strings are correctly traced.
- **Thread Safety:** Atomic RC ensures thread-safe sharing between `begin` spawned threads. Immutable strings (GC-managed) are intrinsically thread-safe.
- **Parallel Error Handling:** Failures in worker threads are captured in `$trial`, surfaced upon `wait` synchronization.
The `zap` keyword explicitly relinquishes control of an object reference.

### Usage Policy
- **Hot Zone Priority:** `zap` MUST be used in performance-critical paths (e.g., hot loops) where Reference Counting (RC) overhead is prohibited.
- **Parallel Error Handling:** Failures in worker threads are captured in `$trial`, surfaced upon `wait` synchronization.
