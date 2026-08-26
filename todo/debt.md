# Technical Debt & Todo Tracker

## 1. Compiler Implementation
- [ ] **AST Evaluator:** Implement symbol table lookups for local variables in `evaluator.go`.
- [ ] **Parser:** Formalize assignment and multi-result return parsing.
- [ ] **LLVM Lowering:** Begin LLVM IR generation for basic arithmetic expressions.
- [ ] **1-Based Normalization:** Verify all `n-1` logic in the lowering phase for matrices/arrays.

## 2. Specification Debt
- [ ] **Unicode Identifiers:** Finalize explicit Unicode range regex in `spec/01-lexical-structure.md`.
- [ ] **Type Promotion:** Document the implicit casting rules for `A` to `U` and `N` to `Z` in `spec/05-types.md`.

## 3. Concurrency/Runtime
- [ ] **Message Queue:** Implement the lock-free message-passing queue for cross-region data passing.
- [ ] **Coroutine Checkpointing:** Formalize the register-state save mechanism for `yield` in the LLVM backend.
