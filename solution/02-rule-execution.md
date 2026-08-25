# Solution: Rule Execution & Scheduling Strategy

## Overview
Bee handles rule execution and workflow through two distinct modes: Synchronous (single-threaded block) and Asynchronous (multi-threaded dispatch). This strategy ensures auditability while enabling high-performance parallel execution.

## 1. Synchronous Execution (`apply`)
- **Primary Use:** Deterministic, sequential processing.
- **Mechanism:** Direct execution in the caller's stack frame.
- **Safety:** Inherits the caller's scope and RC context.

## 2. Asynchronous Execution (`begin` / `wait`)
- **Primary Use:** Task parallelism (e.g., image processing, batch computations).
- **Mechanism:** Dispatches the rule to a worker thread pool managed by the runtime.
- **Synchronization:** The `wait` statement acts as a thread-barrier, collecting results from all outstanding threads spawned within the current block scope.

## 3. Scope & Naming
- **Local Scope Enforcement:** Every `begin` block creates a distinct local namespace to prevent data race conditions between threads.
- **Closure States:** Mutable states within closures (captured by `[]`) are managed via atomic reference counting to ensure consistency across thread boundaries.

## 4. Auditability & Scheduling
- **Rule Orchestration:** The `main` rule acts as the primary orchestrator. Asynchronous threads must be joined via `wait` before the `main` rule reaches the `return` statement.
- **Error Propagation:** Errors within `begin` spawned threads are caught by the runtime and reported in the `$trial` object during the subsequent `wait` synchronization, maintaining a flat, predictable error log.

---
**Status:** Implemented/Resolved in design brainstorming.
**Linked Issues:** `bee/todo/FLOW_CONTROL.md`.
