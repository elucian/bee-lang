# Technical Debt & TODO List

## 1. Compiler & Memory Management
- [ ] `$trial` Object Lifecycle: Define memory management rules for `$trial.messages` in multi-threaded contexts.
- [ ] `$trial` Concurrency: Define interaction between `$trial` objects and concurrent threads.
- [ ] Memory Model: Explicitly document stack vs. heap allocation rules for boxed types (`[]`) and object instances (`new`).

## 2. Language Design
- [ ] Hoisting: Implement compiler logic to allow `main` and other rules to be defined anywhere in a module (excluding local variables).
- [ ] Unicode Specification: Mandate UTF-8 for all `.bee` source files; eliminate ambiguous ASCII-to-Unicode mappings (e.g., `¬` is the only `NOT`).
- [ ] Trial/Error System: Refactor the `trial` block specification to be cleaner and less complex.

## 3. Specification
- [ ] Finalize termination syntax for all blocks (enforce `repeat` for cycles, `done` for conditionals/blocks).
- [ ] Resolve ambiguity: `!` as range/exclusion operator vs. logic/negation (strictly exclude logic use).

---

## 4. Compiler Implementation
- [ ] AST Evaluator: Implement symbol table lookups for local variables in `evaluator.go`.
- [ ] Parser: Formalize assignment and multi-result return parsing.
- [ ] LLVM Lowering: Begin LLVM IR generation for basic arithmetic expressions.
- [ ] 1-Based Normalization: Verify all `n-1` logic in the lowering phase for matrices/arrays.

## 5. Specification Debt
- [ ] Unicode Identifiers: Finalize explicit Unicode range regex in `spec/01-lexical-structure.md`.
- [ ] Type Promotion: Document the implicit casting rules for `A` to `U` and `N` to `Z` in `spec/05-types.md`.

## 6. Concurrency/Runtime
- [ ] Message Queue: Implement the lock-free message-passing queue for cross-region data passing.
- [ ] Coroutine Checkpointing: Formalize the register-state save mechanism for `yield` in the LLVM backend.

## 7. Tooling (bee-ed)
- [ ] Report: `sh run.sh ed edit` verified working for single-line atomic replacement
      (`clean.py` and `.temp` clean integration, 2026-09-17). No regression found; spooling
      multi-line content via `.temp/` + `@path` continues to be the recommended pattern.
      (Manual §9 appended to `manual/DEVELOPER.md` via `append` + `balance`, 2026-09-17: balance OK 29 tags after edit; workflow confirmed.)
