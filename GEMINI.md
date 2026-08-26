# BEE COMPILER: GENERATOR SYSTEM (OPTIMIZED)

## 1. Expert-Driven Autonomy
* **Act, Don't Ask:** You are a Bee expert. Use established patterns from `doc/` and `spec/`. Ask only for architectural conflicts.
* **Debug-First Implementation:**
    * Use `-d` flag for lexical/parsing debugging.
    * Use `fmt.Fprintf(os.Stderr, ...)` for debug instrumentation to keep `stdout` clean for actual code output.
    * Every function must have a block comment describing purpose, inputs, and outputs.
    * Inline "spy" comments are mandatory for complex state transitions.
* **Report Completion:** Omit conversational fillers. End with JSON gate protocol.

## 2. Specification Density
* **Modular Completeness:** Each `/spec` file must be a definitive, machine-parsable reference (EBNF + operational rules + memory impact).
* **Maintainability:** No "stub" files. Aim for 300 lines of high-density specs per file.
* **Git Hygiene:** Atomic commits per logical milestone.

## 3. Workflow Throttling
* **Pacing:** Execute one logical update per turn. 
* **Pause:** Signal `[SYSTEM_SIGNAL: PAUSE_120S]` after every significant change.

## 4. Automated Debug Loop
* **Watchdog:** Use `scripts/watch_test.py <test_case>` for continuous build-debug.
* **Gate Protocol:** 
  {"status": "TASK_COMPLETE", "completed_task": "<TASK_NAME>", "next_task": "<NEXT_TASK>"}
