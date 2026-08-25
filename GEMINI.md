# BEE COMPILER: INCREMENTAL GENERATOR SYSTEM (THROTTLED)

## 1. Context Minimization & State Reading
* **Source of Truth:** ALWAYS start by reading `MANIFEST.md` to identify the `Active Task`.
* **Scoped Read:** Never load entire documentation files; use `start_line` / `end_line` for incremental reads.
* **State Preservation:** Rely on `MANIFEST.md` for task state; do not re-summarize past work unless explicitly asked.

## 2. Granular Micro-Batches
* **Output Cap:** Maximum 500 output tokens per turn. 
* **One-Step Rule:** Only ONE file operation (write/edit/move) per turn.
* **Commit Protocol:** Commit after every logical change. Keep commits small and atomic.

## 3. Workflow Throttling
* **Throttle:** Execute only one tool call per turn. 
* **Mandatory Pause:** End every turn with `[SYSTEM_SIGNAL: PAUSE_120S]` to enforce the 120s cool-down.
* **No Speculation:** Never infer language rules. Use local ground-truth files or explicit documentation reads.

## 4. Patching & Verification
* **Patch-Only Operations:** Output unified line-bounded diffs (`git diff` format) restricted to < 40 changed lines per file.
* **No Re-read Verification:** Trust standard write tool confirmation status. Do NOT issue follow-up `read_file` calls to verify writes.

## 5. Machine-Readable Gate Protocol
* Omit conversational questions and confirmations. Append status in JSON:
  {"status": "TASK_COMPLETE", "completed_task": "<TASK_NAME>", "next_task": "<NEXT_TASK>"}
