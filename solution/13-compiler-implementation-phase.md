# Solution 13: Compiler Implementation Phase Plan

## Overview
This document tracks the execution plan for implementing the Bee compiler toolchain features in pure Go (standard library only, zero external dependencies).

Work is executed incrementally, **one task at a time**, followed by verification and user review.

---

## Task Breakdown & Status

### Phase 1: Native Code Beautifier Engine (`internal/beautifier`)
- [x] **Task 1.1:** Implement `internal/beautifier/beautifier.go` with 2-space block indentation enforcement.
- [x] **Task 1.2:** Implement EOL comment alignment and block keyword/colon normalization.
- [x] **Task 1.3:** Implement AST auto-fix for implicit multiplication (`2(a+b)` $\rightarrow$ `2 * (a + b)`).
- [x] **Task 1.4:** Integrate `-b` / `--beautify` CLI flag in `cmd/bee/main.go`.

### Phase 2: Lexer & Tokenizer Enhancement (`internal/lexer` & `internal/token`)
- [ ] **Task 2.1:** Expand `internal/token/token.go` with all spec tokens (`::`, `.!`, `!.`, `!!`, `+>`, `<<`, `λ`, `∈`, `∩`, `∪`, `\`, `≈`, `≠`, etc.).
- [ ] **Task 2.2:** Update `internal/lexer/lexer.go` with Maximal Munch disambiguation rules for `.`, `..`, `.!`, `!.`, `!!`, `:`, `:=`, `::`, `--`, `+-`.
- [ ] **Task 2.3:** Add support for string interpolation `#(expr)`, raw backtick strings `` `...` ``, and embedded markup DSL blocks (`<sql>`, `<html>`).

### Phase 3: Recursive Descent Parser & AST Expansion (`internal/parser`)
- [ ] **Task 3.1:** Implement EBNF AST nodes for `require`/`ensure` contracts, `if`/`match`/`cycle`/`while`/`for` blocks, and `trial` error handling.
- [ ] **Task 3.2:** Implement 2-space indentation verification and error reporting (`E0201: IndentationMismatch`).
- [ ] **Task 3.3:** Implement deconstruction assignment parsing (`new x, y, *rest := coll;`).

### Phase 4: AST Evaluator & Execution Engine (`internal/evaluator`)
- [ ] **Task 4.1:** Implement universal `entity.type()` introspection in `internal/evaluator`.
- [ ] **Task 4.2:** Support collections (Lists, Arrays, Matrices, Sets, Maps) and 1-based indexing with `$` anchor.
- [ ] **Task 4.3:** Support quantifiers (`∀`, `∃`), data pipelines (`>>`), and coroutines (`yield`).

---

## Verification Strategy
Every completed task is verified by:
1. Compiling the compiler binary via `python build.py`.
2. Running the multi-level test runner via `python test/test_runner.py`.
3. Pausing for developer review.
