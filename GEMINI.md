# BEE COMPILER: GENERATOR SYSTEM (OPTIMIZED)

## 1. Modular Specification Strategy
* **Atomicity:** Split specifications into domain-specific modules (e.g., `01-lexical.md`, `02-statements.md`, `03-types.md`).
* **Density:** Each module MUST contain complete compiler layers (Grammar, Type Matrix, AST Nodes, Operational Semantics) for the topic.
* **Maintainability:** Limit files to ~250 lines. Extract sub-modules if topics grow complex. 
* **Balance:** Prioritize machine-parsable density and logical separation over arbitrary output limits.

## 2. Execution & State Management
* **Source of Truth:** Always start by reading `MANIFEST.md` to identify the `Active Task`.
* **State Preservation:** Update `MANIFEST.md` after every atomic logical milestone. 
* **Incremental Progress:** Proceed step-by-step. If an error occurs, back off, increase waiting/retry delays, and re-verify assumptions.

## 3. Tool Execution & Commitment
* **Atomic Commits:** Bundle related changes into a single logical commit immediately after completing a module layer.
* **Verification:** Rely on tool confirmation. Do not issue follow-up reads unless a tool call explicitly fails.

## 4. Machine-Readable Gate Protocol
* Omit conversational questions and confirmations. Append status in JSON:
  {"status": "TASK_COMPLETE", "completed_task": "<TASK_NAME>", "next_task": "<NEXT_TASK>"}
