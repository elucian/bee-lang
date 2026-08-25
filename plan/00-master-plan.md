# Bee Specification Creation Plan

## Objective
To develop a complete, compiler-grade EBNF specification for the Bee language and map it to Go AST implementations.

## Phased Approach

### Phase 1: Formal Grammar Definition (BNF/EBNF)
- Create individual spec files in `bee/spec/` for each language construct.
- Define EBNF rules for lexical structure, operators, control flow, and rules.
- Resolve ambiguities (e.g., `cycle` termination, `match` semantics).

### Phase 2: Structural Mapping
- Map EBNF productions to Go structs in `bee/internal/ast/` (or equivalent).
- Define Type Matrix and Operational Semantics for each AST node.

### Phase 3: Validation & Verification
- Cross-reference retrieved documentation with generated EBNF.
- Implement tests to verify syntax compliance.

## Tasks
1. [ ] Define standard block termination rules (e.g., verify `cycle` termination: `repeat` vs `done`).
2. [ ] Formalize `match` vs `if-else` usage patterns in `spec/06-control-flow.md`.
3. [ ] Generate `spec/01-lexical-structure.md` (Lexical Grammar).
4. [ ] Generate `spec/02-syntax-statements.md` (Statements & Expressions).
5. [ ] Generate `spec/03-rules-functions.md` (Rules & Lambdas).
6. [ ] Generate `spec/04-concurrency-types.md` (Concurrency & Types).
