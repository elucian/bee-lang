# GEMINI.md - BEE COMPILER ARCHITECT & GENERATOR SYSTEM

## 1. Engineering Invariants & Go Standards
* **Ecosystem:** Pure Go standard library. Zero external dependencies.
* **Allocation Strategy:** Low/zero heap allocation patterns. Pass by pointer exclusively when mutating state; otherwise use value semantics.
* **Error Model:** Explicit `error` return values only. Never call `panic` or `recover` for lexical, parsing, or type-checking errors.
* **Instrumentation:** Output debug traces strictly to `os.Stderr` (or under explicit debug flags). Keep `stdout` clear for generated code or AST output.
* **File Header Standard:** Every Go file MUST begin with a concise comment declaring component scope, responsibilities, and runtime/AST boundary.

## 2. Token-Optimized Execution & Anti-Loop Protocol
- **Zero Preamble:** Omit conversational chatter, introductions, status updates, and closing summaries. Output actionable code or diffs directly.
- **Targeted Emissions:** Emit ONLY modified functions, structs, or unified diffs. Never output full unchanged source files.
- **Strict Anti-Loop Guard:** If a test or diagnostic failure repeats twice after a fix, **HALT IMMEDIATELY** and stop to prevent unproductive iteration loops. Maximum 2 attempts per issue; if failure persists on the 3rd check, cut it off, halt, and explain the blocker to the user rather than looping endlessly.

## 3. Implementation Phase & Defect Triaging
* **Implementation Mapping:** Deterministically map finalized `/spec` EBNF to recursive-descent parser routines and AST types.
* **Root-Cause Classification:** On test failures or parser mismatches, analyze whether the root cause lies in:
  1. **Compiler Code:** Implementation logic, state mutation, or AST construction bugs.
  2. **Specification Defect:** Contradictory EBNF, ambiguous precedence/grammar, or incomplete operator rules in `/spec`.
  3. **Test Suite Defect:** Invalid test vectors, inaccurate expected AST outputs, or broken runner assertion logic.
* **Integrity Guard:** **NEVER** modify Go code to force passing tests against a defective EBNF specification or invalid test expectation.

## 4. Testing, Repair & Defect Signaling Protocol
* **Test Verification:** Run `python test/watch.py <test_case>` or `python test.py` to evaluate AST/parser state.
* **Repair & Signal Logic:**
  - **If Code Bug:** Perform a single-pass batch fix addressing all reported errors across tests simultaneously.
  - **If Spec or Test Bug:** Do NOT edit code or attempt fixes. Halt execution and explicitly signal the defect target.
* **Circuit Breaker (Loop Hard-Stop):**
  - Execute post-fix verification.
  - If post-fix tests yield identical error signatures, or if a second failure occurs on the same component, **HALT IMMEDIATELY**.
* **Git Operations:** **NEVER** run `git commit`, `git push`, or modify git history. All commits are performed manually by the developer.

## Final message
When you finish send this message: "Task Completed in <runtime>
