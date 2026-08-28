# Bee Test Infrastructure

This document outlines the test organization for the Bee Programming Language. Tests are organized by complexity levels in `test/levelX/`.

---

## 1. Test Suite Levels

- **Level 1 (Lexical/Declarations)**: Character encoding, operators, declarations, assignments.
- **Level 2 (Statements/Control Flow)**: I/O, `if/else`, loops (`cycle`, `while`, `for`), `match`.
- **Level 3 (Rules/Contracts)**: Rules, parameters, recursion, contracts.
- **Level 4 (Collections/Data)**: Lists, arrays, sets, maps, pipelines, boxing.
- **Level 5 (Concurrency/Objects)**: Threads, objects, error handling (`trial`).

---

## 2. Test Automation Tools

Located in `test/` or `scripts/`:

- **`test/reset.py`**: Clean build artifacts and test report directories.
- **`scripts/check.py`**: Run syntax checks on all test files.
- **`test/test.py`**: Root test orchestrator for full suite execution.
- **`test/bench.py`**: Performance benchmark runner.
- **`test/solo.py`**: Debug mode test runner (verbose code echo and error location).
- **`test/health/health.go`**: Intelligent compiler self-health check and smoke test.

---

## 3. Reporting Infrastructure

- **`test/status/`**: Stores detailed syntax check reports and execution status summaries.

---
