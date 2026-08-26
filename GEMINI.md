# BEE COMPILER: GENERATOR SYSTEM (TOKEN-OPTIMIZED)

## 1. Autonomous Execution & Token Efficiency
* **Action-Oriented:** Solve tasks directly. Omit conversational filler, status chatter, and unnecessary explanations.
* **Token Conservation:** Keep responses short and focused. Do not echo full files, large diffs, or redundant output.
* **Code Standard:** Every source file requires a header comment (purpose, responsibility, architecture strategy). Functions require concise input/output comments.
* **Stderr Debugging:** Direct instrumentation to `os.Stderr` (`fmt.Fprintf(os.Stderr, ...)` or `-d` flag) to keep `stdout` clean.

## 2. Specification & Testing Standards
* **High Density:** Spec files in `/spec` must be complete and machine-parsable references without stub implementations.
* **Automated Testing:** Run `scripts/watch_test.py <test_case>` or `test/test_runner.py` for verification before task completion.

## 3. Workflow & Gate Protocol
* **Single Step Focus:** Execute one logical modification per turn.
* **Completion Protocol:** End completed tasks with the gate protocol JSON:
  `{"status": "TASK_COMPLETE", "completed_task": "<TASK_NAME>", "next_task": "<NEXT_TASK>"}`
