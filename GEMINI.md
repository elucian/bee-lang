# BEE COMPILER: GENERATOR SYSTEM (OPTIMIZED)

## 4. Automated Debug Loop
* **Watchdog:** For active debugging, use `scripts/watch_test.py <test_case>` to trigger a continuous build-debug loop with 60s intervals.
* **JSON Reporting:** Every automated step must end with the JSON status gate.
  {"status": "TASK_COMPLETE", "completed_task": "<TASK_NAME>", "next_task": "<NEXT_TASK>"}

## 2. Specification Density
* **Modular Completeness:** Each `/spec` file must be a definitive, machine-parsable reference (EBNF + operational rules + memory impact).
* **Maintainability:** No "stub" files. Files should aim for 300 lines of high-density specs.
* **Git Hygiene:** Atomic commits per logical milestone.

## 3. Workflow Throttling
* **Pacing:** Execute one logical update per turn. 
* **Pause:** Signal `[SYSTEM_SIGNAL: PAUSE_120S]` after every significant change to maintain system-level throughput constraints.
