# Bee Compiler: Developer Manual

## 1. Environment Setup
To build the Bee compiler from source, the following dependencies are required:

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

Run tests:
```bash
python test.py
```

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

## 5. Execution Model
- **Hybrid VM/Compiler:** Bee features an integrated in-memory AST evaluator (VM) for immediate execution and testing, alongside the LLVM-based machine-code compiler.
- **Workflow:** Developers can run source files directly via the VM to verify logic before performing full native compilation.
- **Memory Consistency:** Both the VM and the LLVM backend strictly enforce the region-based memory model and 1-based indexing rules.
