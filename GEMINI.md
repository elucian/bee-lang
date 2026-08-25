# BEE COMPILER: INCREMENTAL GENERATOR SYSTEM (THROTTLED)

## 1. Context Minimization & State Reading
* **Scoped Manifest Read:** Read ONLY the `## Active Task` block of `PROJECT_MANIFEST.md` (lines 1–30 max). Never load the entire manifest into the context window.
* **Cached Ground-Truth Only:** Read grammar rules exclusively from local cache file (`.cache/bee-spec-ground-truth.md`). 
* **Zero Network Requests:** DO NOT call web-fetching tools per turn. Network retrieval is strictly restricted to initialization if `.cache/bee-spec-ground-truth.md` does not exist.

## 2. Granular Micro-Batches & Output Cap
* **Single-Layer Execution:** Write ONLY ONE compiler layer per turn (Layer 1: Lexical OR Layer 2: Types OR Layer 3: AST OR Layer 4: Semantics). Never attempt to write all four layers in one turn.
* **Hard Token Cap:** Strict maximum of 500 output tokens per turn to prevent truncation retries.

## 3. Tool Throttle & Harness Control
* **Single Tool Call:** Execute at most ONE file tool operation per turn.
* **Client Handoff Signal:** End every response turn with the exact string `[SYSTEM_SIGNAL: PAUSE_120S]`. The client execution wrapper will intercept this signal to delay the subsequent API payload.

## 4. Patching & Verification
* **Patch-Only Operations:** Output unified line-bounded diffs (`git diff` format) restricted to < 40 changed lines per file.
* **No Re-read Verification:** Trust standard write tool confirmation status. Do NOT issue follow-up `read_file` calls to verify writes.

## 5. Machine-Readable Gate Protocol
* Omit conversational questions and confirmations. Append status in JSON:
  {"status": "TASK_COMPLETE", "completed_task": "<TASK_NAME>", "next_task": "<NEXT_TASK>"}
