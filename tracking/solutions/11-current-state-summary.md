# Solution: Compiler Implementation Roadmap (State 2026-08-26)

## 1. Memory Management
- **Region-Based:** Each `rule` has a private region (stack+heap). Reclaimed on `return`.
- **Heap vs Stack:** `[]` boxed variables are promoted to Heap only if public/object-bound. Local boxes are region-local.
- **Zap Protocol:** Used in "Hot Zones" (loops). Must be documented via comments.
- **Error Handling:** Deferred via `$trial`. Propagated to parent at `wait` synchronization.

## 2. Execution & Concurrency
- **Hybrid VM/IR:** In-memory evaluator (VM) for development; LLVM IR backend for native binaries.
- **Parallelism:** `begin`/`wait` model. Threads use message-passing queues (lock-free) for shared collection mutation.
- **Purity:** Lambdas are pure (side-effect free, no `rule` calls). Rules are "dirty".

## 3. Parser/Lexer Strategy
- **Lexical:** Maximal Munch. 1-based indexing for collections.
- **Parser:** Recursive Descent.
- **Normalization:** 1-based indices normalized to 0-based at the LLVM IR lowering phase.
