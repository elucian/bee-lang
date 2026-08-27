# Bee Compiler: Developer Manual

## 1. Environment Setup
To build the Bee compiler from source, the following dependencies are required:

### Editor

We use Zed editor for performance and Gemini 3.5 Flash-Light. Following instructions depend on it.

### On windows

Use github terminal instead of powershell. To do this configure Zed terminal using this json entry in settings.json

  "terminal": {
    "shell": {
      "program": "C:\\Program Files\\Git\\bin\\bash.exe"
    }
  },

### Dependencies
- **Go 1.26+:** [Download & Install](https://go.dev/dl/)
- **LLVM 18+:**
  - **Windows:** Download the [LLVM Windows Installer](https://github.com/llvm/llvm-project/releases). Ensure "Add LLVM to PATH" is selected during installation.
  - **Verification:** Run `llc --version` and `go version` in your terminal.

## 2. Project Build & Execution
Build the compiler:
```bash
python build.py
```

Install to PATH:
```bash
sh setup.sh
```
*(Note: Run this once as Administrator to permanently add `bee` to your User PATH. Restart your terminal after execution.)*

### Compiler Modes
- **`-d` / `--debug`**: Debug mode. Outputs the token stream to `stderr`.
- **`-e` / `--execute`**: Enables the in-memory AST evaluator (VM) to execute the code.
- **`-c` / `--compile`**: Parse-only mode. Validates syntax and outputs "Syntax OK" if successful.

### Usage Examples
```bash
# Debugging lexer tokens
./bin/bee.exe -d test/level2/T0202.bee

# Execute logic in the VM
./bin/bee.exe -e test/level2/T0202.bee

# Syntax check only
./bin/bee.exe -c test/level1/T0101.bee
```

## 7. Master Workflow Automation (`run.sh`)
- **Build:** `sh run.sh build` builds the Bee compiler.
- **Complete Test Pipeline:** `sh run.sh test` runs `test.py` (Dryrun verification $\rightarrow$ Benchmark Suite $\rightarrow$ All Level Test Suites with timestamped JSON status reporting in `test/status/`).
- **Iterative Fix Loop:** `sh run.sh fix level1` attempts automated iterative test fixes for a target level up to a maximum of 10 iterations, rebuilding the compiler on failure and stopping for user input.

## Author Note (elucian)

I preffer to use $run command directly in terminal instead of ./run.sh that is too much to type. So I have learned a little trick: To make run command permanent easy to access run this:

```
$echo 'alias run="./run.sh"' >> ~/.bashrc
$source ~/.bashrc
```
