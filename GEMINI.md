# BEE COMPILER: INCREMENTAL GENERATOR SYSTEM

## 1. Session Initialization (Mandatory)
* **Always start by reading `PROJECT_MANIFEST.md`** to load current state and verify the next pending task.
* **Never proceed** until the manifest is acknowledged and the current phase is identified.

## 2. Pacing & Interaction
* **Micro-Batches:** Perform only ONE task per turn.
* **Interactive Gates:** At the end of every turn, summarize the action taken and ask: *"Ready to proceed to [Next Task]?"*
* **Pause:** Observe 120s between tool calls as defined in Resource Constraints.

## 3. File I/O & Anti-Destruction Protocols
* **Zero Whole-File Overwrites:** NEVER perform full-buffer file replacements on existing files.
* **Modular File Isolation:** Every topic or operator group MUST be written to its own dedicated file using explicit index ordering (e.g., `spec/01-lexical-structure.md`).
* **Patch-Only Editing:** If updating an existing file is mandatory, output ONLY a line-bounded unified diff/patch (`git diff` format). Broad file updates or total replacements are strictly forbidden.

## 4. Mandatory Compiler Specification Schema
Every generated specification file MUST contain all four compiler layers:
1. **Lexical Grammar & Concrete Syntax:** EBNF rules, token regex patterns, operator precedence, associativity.
2. **Type Matrix:** Valid operand combinations, return types, casting, compile-time errors.
3. **AST Node Representations:** Concrete field definitions, child evaluation, mapping to Go structs (`package ast`).
4. **Operational Semantics:** Runtime execution, short-circuiting, side-effect ordering.

## 5. Ground-Truth Retrieval
* **Zero Apology Policy:** Respond strictly with structural verification errors, code, or unified diffs.
* **Mandatory Live Retrieval:** FORBIDDEN from generating syntax from internal memory. MUST fetch `https://sagecode.org/projects/bee/` before writing.
* **Keyword AST Validation:** Cross-reference keywords against retrieved payload. If absent, output `[MISSING_SOURCE_DATA]`.

## 6. Resource Constraints & Micro-Batching
* **Rate Limiting:** Enforce a strict "1 Tool Call per 120s" policy.
* **Throttle Implementation:** Before every tool call, wait 120 seconds. If a tool call fails due to rate limits or API constraints, implement an exponential backoff (`delay = initial_delay * 2^n`) with added jitter before retrying.
* **Output Token Cap:** Max 1,000 tokens per turn.
* **Verification:** Conduct write-verification check (confirm file exists and `line_count > 0`) before advancing.
