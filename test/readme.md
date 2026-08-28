# Bee Test Infrastructure

This document outlines the test organization for the Bee Programming Language. Tests are organized by complexity levels in `test/levelX/`.

---

## 1. Test Suite Levels

| Level | Description | Spec Reference |
| :--- | :--- | :--- |
| [Level 1](level1/README.md) | Lexical/Declarations/Types | [Spec 01/05](spec/01-lexical-structure.md) |
| [Level 2](level2/README.md) | Statements/Control Flow | [Spec 02/04](spec/02-statements.md) |
| [Level 3](level3/README.md) | Rules/Contracts/Functions | [Spec 03/07](spec/03-rules.md) |
| [Level 4](level4/README.md) | Collections/Pipelines | [Spec 10/11](spec/10-collections.md) |
| [Level 5](level5/README.md) | Objects/Concurrency | [Spec 06/12](spec/06-objects.md) |
| [Level 6](level6/README.md) | Advanced Features | N/A |
| [Level 7](level7/README.md) | System Integration | [Spec 14](spec/14-library.md) |
| [Level 8](level8/README.md) | Experimental | N/A |

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
- **`test/output/`**: Detailed Markdown execution reports for each individual test run.
