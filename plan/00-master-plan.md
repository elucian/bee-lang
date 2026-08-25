# Bee Specification Creation Plan

## Objective
To develop a complete, compiler-grade EBNF specification for the Bee language and map it to LLVM IR implementations.

## Phased Approach

### Phase 1: Formal Grammar Definition (EBNF & Lexical)
- Create individual spec files in `spec/` for each language construct.
- Define EBNF rules for lexical structure, statements, rules, and types.

### Phase 2: Compiler Implementation
- Implement Lexer/Parser in `internal/`.
- Lower Bee AST to LLVM IR.
- Implement Region-Based Memory Management.

### Phase 3: Validation & Verification
- Verify spec against `doc/` ground truth.
- Validate compiler output using `test/` cases.

## Tasks
1. [x] Index project and ingest documentation.
2. [x] Establish solution/issue tracking.
3. [x] Refactor and reset Memory Model Specification (`spec/00-memory-model.md`).
4. [ ] Incrementally expand Memory Model Specification sections (zap, error handling).
5. [ ] Formalize Lexical Grammar (`spec/01-lexical-structure.md`).
6. [ ] Formalize Statement Grammar (`spec/02-statements.md`).
7. [ ] Map AST Nodes and implement Lexer/Parser in `internal/`.
