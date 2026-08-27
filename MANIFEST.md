# BEE COMPILER MANIFEST

## Current Status
- **Active Phase:** Phase 5 – Type Checker & Semantic Analysis (`internal/typechecker`)
- **Active Task:** Task 5.1: Implement symbol table resolution and type inference.
- **Last Updated:** 2026-08-27

## Manifest Maintenance Protocol
* **On-Demand Test Execution:** Do NOT run test scripts (`python test.py` or `python test/solo.py`) automatically after code modifications. Execute tests ONLY when explicitly commanded.
* **Sync Protocol:** When a test run is requested and succeeds:
  1. Toggle the completed task checkbox to `[x]`.
  2. Advance `Active Task` (and `Active Phase` if shifting passes) in the header.
  3. Update `Last Updated` to current date (`YYYY-MM-DD`).

## Completed Specifications (`/spec`)
- [x] `spec/00-07`: Memory model, Lexical, Statements, Rules, Structure, Types, Objects, Functions.
- [x] `spec/10-14`: Collections, Processing, Concurrency, Graphics, System Library.
- [x] `issues/08`, `solution/12-13`: Beautifier & Phase tracking architecture.

## Completed Phases
- [x] **Phase 1: Beautifier Engine** (`internal/beautifier`) – Formatting, 2-space alignment, auto-fix `2(a+b)`.
- [x] **Phase 2: Lexer & Tokens** (`internal/lexer`, `internal/token`) – Spec tokens, Maximal Munch, `#(expr)`.
- [x] **Phase 3: Recursive Descent Parser** (`internal/parser`) – AST generation for rules, decls, control flow.
- [x] **Phase 4: AST Evaluator** (`internal/evaluator`) – Tree-walking execution engine.

## Pending Roadmap

### Phase 5: Type Checker & Semantic Analysis (`internal/typechecker`)
- [ ] Task 5.1: Implement symbol table, scoping rules, and static type resolution.
- [ ] Task 5.2: Enforce zero-based indexing validation across array/list node expressions.
- [ ] Task 5.3: Validate precondition (`require`) and postcondition (`ensure`) contract bindings.

### Phase 6: LLVM IR Codegen Engine (`internal/codegen`)
- [ ] Task 6.1: Map AST nodes to LLVM IR module definitions via Go LLVM bindings.
- [ ] Task 6.2: Implement basic block generation with explicit terminators (`CreateBr`, `CreateRet`, `CreateCondBr`).
- [ ] Task 6.3: Implement dynamic type coercion and mathematical operation emission ($Qm.n$, Unicode operators).

### Phase 7: Test Suite Validation (`test/`) [On-Demand Only]
- [ ] Task 7.1: Level 1 Lexical & Declaration (`test/level1/`)
- [ ] Task 7.2: Level 2 Control Flow & I/O (`test/level2/`)
- [ ] Task 7.3: Level 3 Rules & Contracts (`test/level3/`)
- [ ] Task 7.4: Level 4 Collections & Pipeline (`test/level4/`)
- [ ] Task 7.5: Level 5 Concurrency & Objects (`test/level5/`)
