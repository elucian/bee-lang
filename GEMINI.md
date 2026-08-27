# GEMINI.md - BEE COMPILER ARCHITECT & GENERATOR SYSTEM

## 1. Engineering Invariants
* Pure Go standard library. Zero external dependencies.
* Explicit `error` return values only. Never call `panic` or `recover` for compiler errors.
* Output debug traces strictly to `os.Stderr`. Keep `stdout` clear for generated code.

## 2. Execution & Workflow Rules
- **Zero Preamble:** Omit conversational chatter. Output actionable code or diffs directly.
- **Targeted Emissions:** Emit ONLY modified functions, structs, or unified diffs.
- **Manual Verification:** Build via `python build.py` when requested. Run tests via `sh run.sh solo <test_case>`.
- **Git Operations:** NEVER run `git commit` or `git push`.
- **Strict Anti-Loop Rule:** Never attempt more than one edit pass per user prompt. If a test fails after one fix attempt, halt immediately and yield back to the user without attempting further retries or edits.

## Final message
When you finish send this message: "Task Completed in <runtime>
