# BEE COMPILER MANIFEST

- **Active Phase:** Phase 5 – Type Checker & Semantic Analysis (`internal/typechecker`)
- **Active Task:** Task 5.1: Implement symbol table resolution and type inference.
- **Last Updated:** 2026-08-28

## Development Methodology
- **TDD Protocol**: All implementation MUST follow Test-Driven Development. 
  1. Define requirement in `/issues/` and `/solution/`.
  2. Formalize in `/spec/`.
  3. Create failing test in `/test/levelX/`.
  4. Implement & Pass.
- **Audit Protocol**: Regular verification of implementation against `/spec/`.

## Roadmap & Phases

### Phase 5: Type Checker & Semantic Analysis (`internal/typechecker`)
- [ ] Task 5.1: Implement symbol table, scoping rules, and static type resolution.
- [ ] Task 5.2: Enforce zero-based indexing validation across array/list node expressions.
- [ ] Task 5.3: Validate precondition (`require`) and postcondition (`ensure`) contract bindings.

### Phase 6: LLVM IR Codegen Engine (`internal/codegen`)
- [ ] Task 6.1: Map AST nodes to LLVM IR module definitions.
- [ ] Task 6.2: Implement basic block generation with explicit terminators.
- [ ] Task 6.3: Implement dynamic type coercion ($Qm.n$, Unicode operators).

### Phase 7: Specification Audit & TDD Validation
- [x] Task 7.1: Audit `/spec/` vs `/web/` documentation.
- [ ] Task 7.2: Create and pass missing edge-case tests identified during audit.
- [ ] Task 7.3: Integrate automated `/spec/` to `/web/` sync scripts.

## Completed Specifications (`/spec`)
- [x] `spec/00-07`: Memory model, Lexical, Statements, Rules, Structure, Types, Objects, Functions.
- [x] `spec/10-14`: Collections, Processing, Concurrency, Graphics, System Library.
