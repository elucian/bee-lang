# Technical Debt & TODO List

## Compiler & Memory Management
- [ ] `$trial` Object Lifecycle: Define memory management rules for `$trial.messages` in multi-threaded contexts.
- [ ] `$trial` Concurrency: Define interaction between `$trial` objects and concurrent threads.
- [ ] Memory Model: Explicitly document stack vs. heap allocation rules for boxed types (`[]`) and object instances (`new`).

## Language Design
- [ ] Hoisting: Implement compiler logic to allow `main` and other rules to be defined anywhere in a module (excluding local variables).
- [ ] Unicode Specification: Mandate UTF-8 for all `.bee` source files; eliminate ambiguous ASCII-to-Unicode mappings (e.g., `¬` is the only `NOT`).
- [ ] Trial/Error System: Refactor the `trial` block specification to be cleaner and less complex.

## Specification
- [ ] Finalize termination syntax for all blocks (enforce `repeat` for cycles, `done` for conditionals/blocks).
- [ ] Resolve ambiguity: `!` as range/exclusion operator vs. logic/negation (strictly exclude logic use).
