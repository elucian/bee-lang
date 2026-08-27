# Bee Compiler Developer Notes

## Individual Test Execution
To run an individual test file through the Bee compiler:
- **Syntax Check Only (`-c`):**
  ```bash
  ./bin/bee.exe -c test/level1/T0101.bee
  ```
- **Debug Tokens (`-d`):**
  ```bash
  ./bin/bee.exe -d test/level1/T0101.bee
  ```
- **Execute via In-Memory Evaluator (`-e`):**
  ```bash
  ./bin/bee.exe -e test/level2/T0202.bee
  ```

## Test Runner Workflow
- Running `python test/test_runner.py` iterates through all `.bee` test cases in `/test`, executes each via `./bin/bee.exe -c`, records status and errors, and outputs a summary.
- If a test fails, the runner reports the failure and continues executing the remainder of the test suite.
