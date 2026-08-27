# GEMINI.md - BEE COMPILER ARCHITECT & GENERATOR SYSTEM

## 1. Engineering Invariants
* Pure Go standard library. Zero external dependencies.
* Explicit `error` return values only. Never call `panic` or `recover` for compiler errors.
* Enforce zero-based array indexing across parser, AST, type-checker, and codegen passes.
* Output debug traces strictly to `os.Stderr`. Keep `stdout` clear for generated code or execution output.

## 2. Test Lifecycle & Freeze Protocol (.bee Test Cases)
* **Spec-Driven Generation:** Generate new test files (`.bee`) strictly under `test/levelX/` derived directly from `/spec/`.
* **Locking Created Tests:** Every newly generated `.bee` test file MUST include this header tag on line 1:
  `// @FROZEN: Generated from /spec/. Immutable ground truth.`
* **Read-Only Enforcement:** Test files in `test/levelX/*.bee` (especially those marked `@FROZEN`) are strictly immutable. NEVER modify `.bee` test inputs, assertions, or expected outputs to force a failing compiler build to pass.
* **Disable, Never Delete:** If a test fails persistently across fix attempts, NEVER delete the file. Disable it by prepending `// @DISABLED: <reason>` on line 1 of the `.bee` file.

## 3. Execution & Workflow Rules
- **Zero Preamble:** Omit conversational chatter. Output actionable Go code, unified diffs, or terminal commands directly.
- **Targeted Emissions:** Edits are restricted strictly to compiler packages inside `/internal/`. Emit ONLY modified functions, structs, or unified diffs.
- **Verification Harness:**
  - Full Test Suite: `python test.py`
  - Single Test Execution: `python test/solo.py test/levelX/TXXYY.bee`
  - Benchmarks / CLI Verification: `python test/bench.py` / `python test/dryrun.py`
- **Git Operations:** NEVER run `git commit` or `git push`.
- **Strict Anti-Loop Rule:** Never attempt more than one edit pass per user prompt. If a test fails after one fix attempt, halt immediately, write error analysis to `os.Stderr`, and yield back without modifying `.bee` files.

## Final message
When you finish send this message: "Task Completed in <runtime>"
