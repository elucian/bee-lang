# Bee Comprehensive Test Plan (Levels 1 to 5)

This document outlines the test plan and tracking checklist for the Bee Programming Language compiler test suite across Levels 1 through 5. Each test case corresponds to a test file under `test/levelX/TXXYY.bee`.

---

## Level 1: Lexical Structure & Basic Declarations (`test/level1/`)
**Focus:** Character encoding, Maximal Munch disambiguation, variable declarations (`new`), assignments (`let`), basic arithmetic, expectations (`expect`), and code comments (including expression comments `(: ... :)`).

- [x] **T0101**: Basic variable assignment (`new a ∈ Z; let a := 42;`)
- [x] **T0102**: Variable mutation and expectation check (`let a := 42; expect a = 42; let a := 100; expect a = 100;`)
- [x] **T0103**: Arithmetic addition and multiplication expressions
- [x] **T0104**: String literals and escape sequence handling
- [x] **T0105**: Range operators (`..`, `.!`, `!.`, `!!`) and types (`Z`, `N`, `R`)
- [x] **T0107**: Division error checking (`test/level1/T0107.bee`)
- [x] **T0108**: Block comments (`+- ... -+`) and expression comments (`(: ... :)`) (`test/level1/T0108.bee`)

---

## Level 2: Statements, Control Flow & I/O (`test/level2/`)
**Focus:** Input/Output (`print`, `write`), conditionals (`if` / `else` / `done`), loops (`cycle`, `while`, `for`), and block indentation.

- [x] **T0201**: Basic print statement output
- [x] **T0202**: Conditional `if` / `else` block execution
- [x] **T0203**: While loop execution and counter increment
- [x] **T0204**: For loop iteration over collection ranges
- [x] **T0205**: Match control flow (`match` / `when` / `other` / `done`)

---

## Level 3: Rules, Subroutines & Contracts (`test/level3/`)
**Focus:** Rule definitions, parameters, variadic parameters (`*`), multiple results (`=> (res1, res2)`), contracts (`require` / `ensure`), and recursion.

- [x] **T0301**: Rule definition with single parameter and return result
- [ ] **T0302**: Rule with multiple return results and named parameters
- [ ] **T0303**: Rule precondition contracts (`require`)
- [ ] **T0304**: Rule postcondition contracts (`ensure`)
- [ ] **T0305**: Recursive rule execution (factorial / Fibonacci)

---

## Level 4: Collections & Data Processing (`test/level4/`)
**Focus:** Lists `()`, Arrays `[]`, Matrices `[](r, c)`, Sets `{}` (with set algebra `∩`, `∪`, `\`), Maps `{:}` (with keys), 1-based indexing (`$`), boxing `[x]`, and pipelines (`>>`).

- [x] **T0401**: List and Array indexing (1-based)
- [ ] **T0402**: Set algebra operations (`∩`, `∪`, `\`)
- [ ] **T0403**: Hash map insertion and key lookup
- [ ] **T0404**: Primitive boxing `[x]` and unboxing
- [ ] **T0405**: Pipeline transformation (`>>`) and quantifiers (`∀`, `∃`)

---

## Level 5: Concurrency, Objects & Error Handling (`test/level5/`)
**Focus:** Multithreading (`begin` / `wait`), reduction channels (`+>`), coroutines (`yield`), object entities (`type` with properties and methods), and error trials (`trial` / `try` / `miss` / `final`).

- [x] **T0501**: Trial error handling block (`trial` / `try` / `final`)
- [ ] **T0502**: Multithreaded execution with `begin` and `wait` barrier
- [ ] **T0503**: Thread-safe reduction channels (`+>`)
- [ ] **T0504**: Object entity creation and method dispatch
- [ ] **T0505**: Coroutine pipeline extraction (`yield ... << ...`)

---

## Test Folders & Architecture (`test/`)
- **`test/level1/` through `test/level5/`**: Organized test vector suites spanning lexical structures, control flow, functions, collections, and concurrency.
- **`test/bench/`**: Micro-benchmark suites for measuring performance and throughput.
- **`test/output/`**: Spooled failure reports (`.fail`) generated for failing test runs.
- **`test/status/`**: Execution status records saved as timestamped JSON files (`status_YYYYMMDD_HHMMSS.json`) detailing per-level pass/fail metrics.
- **`test/watch.py`**: Continuous developer watcher script that rebuilds the compiler and executes a specified test file in debug mode.
- **`test/bench.py`**: Performance benchmark execution and telemetry generation script.
- **`test/dryrun.py`**: CLI flag verification script.
- **`test.py`**: Root test orchestrator that executes all level suites and records level-by-level status telemetry.
