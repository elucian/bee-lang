# BEE COMPILER ARCHITECT & GENERATOR SYSTEM

## 1. Core Engineering Invariants
* **Go Ecosystem:** Pure standard library only. Zero external dependencies.
* **Allocation Strategy:** Low/zero heap allocation patterns. Pass by pointer only when state mutation is required.
* **Error Handling:** Return explicit `error` types. Never use `panic`/`recover` for grammar, lexical, or type errors.
* **Instrumentation:** Direct all debug instrumentation strictly to `os.Stderr` (or under `-d` flag). Keep `stdout` pure for generated output.
* **File Header Standard:** Every Go file must begin with a concise header comment stating component purpose, responsibility, and AST/runtime boundary.

## 2. Phase Protocols
* **Phase 1 (Specification Mode):**
  - Specs in `/spec` must be machine-parsable EBNF/BNF.
  - Formally resolve operator precedence, associativity, and left-recursion ambiguity before code generation.
  - **Forbidden:** Do not generate Go implementation code during Phase 1 specification turns.
* **Phase 2 (Implementation Mode):**
  - Deterministically map finalized `/spec` EBNF directly into Go recursive-descent parser methods and AST types.

## 3. Token-Optimized Agent Execution
* **Action Only:** Omit conversational fluff, status chatter, and pleasantries.
* **Targeted Emissions:** Emit ONLY modified methods, structs, or targeted diffs. Never re-print full unchanged files.
* **Focused Context:** Rely strictly on explicitly tagged `@file` inputs; do not auto-crawl or index unreferenced workspace directories.

## 4. Testing & Manual Commit Protocol
* **Verification:** Execute `scripts/watch_test.py <test_case>` or `test/test_runner.py` to validate AST/parser correctness prior to completing a task.
* **Git Restrictions:** **NEVER** run `git commit`, `git push`, or alter git history. All commits are performed manually by the developer.
* **Gate Protocol:** End every completed step with this exact JSON payload:
  `{"status": "TASK_COMPLETE", "completed_task": "<TASK_NAME>", "next_task": "<NEXT_TASK>", "git_status": "PENDING_MANUAL_COMMIT"}`
