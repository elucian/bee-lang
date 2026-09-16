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
---

## 4. Authoring Assert-Driven Test Cases

Every `.bee` case in `test/levelX/` is a **self-verifying program**: it must
*pass or fail on its own* after execution, never require a human (or agent) to
eyeball printed output. Verify program truth with **`expect <cond>;`**
assertions inside `rule main`; the evaluator fails (non-zero exit) the moment
an `expect` is false.

### Verification rules

- **Assert, don't print.** Replace `print value;` with `expect value = <expected>;`
  (or a bare boolean `expect 2 ∈ coll;`). `print` is for *debugging only* and
  never constitutes a pass.
- **A test "PASSES" iff `bee -e <file>` exits 0.** The harness
  (`test/test.py`) runs each non-disabled case and records that exit status.
  Unimplemented features must therefore **fail** the test (a genuine signal),
  not silently print `0`.
- **Membership** uses `expect x ∈ coll;` (arrays, lists, sets). Positive
  membership is supported; there is no `!∈` spelling — assert positives only.
- **Boolean literals** are `1`/`0` (integers), so `expect sub = 1;` checks a
  true-valued result. Comparisons (`=`, `<`, `≤`, `¬`) evaluate to `1`/`0`.

### Header tags (line 1, in order)

| Tag | Meaning |
| :--- | :--- |
| `-- @DESC:<name> <purpose>` | Required. Human + harness description of the test's purpose (shown in the level `README.md` table). |
| `-- @NEGATIVE` | The test supplies a *negative* program: it **passes** on a non-zero exit and **fails** on exit 0 (e.g. a rejected `a[-1]` index). |
| `-- @DISABLED: <reason>` | Skipped by the harness. Inserted automatically when a case fails; the reason is the first stderr diagnostic. Powerful, honest positive signal: disabled == feature not yet implemented. |
| `-- @FROZEN` / `@ENABLED` | Legacy lifecycle markers (spec-protocol). `@FROZEN` marks immutable ground truth; `@ENABLED` manual re-activation. |

### Lifecycle & automatic testing

1. Author the case with `-- @DESC:` and `expect` assertions for the intended
   (spec) behaviour.
2. Run `python test/solo.py <name>` to execute it and refresh the level
   `README.md` status row.
3. Run `python test/test.py <level>` for the full orchestrated sweep: it
   auto-disables every failing case with a fresh `@DISABLED` reason and updates
   the status table. A case that fails on an unimplemented feature is
   *honestly* recorded as `FAIL`/`@DISABLED` until that feature lands — never
   paper over it with `print`.
