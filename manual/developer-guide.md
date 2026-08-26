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

Run tests (compile-only mode):
```bash
python test.py
```

## 5. Performance Flags
- **`-e` / `--execute`**: Enables in-memory AST evaluation (VM).
- **`-c` / `--compile`**: Compiler/parse-only mode.
- **`-m <threads>`**: Activates multi-threaded Lexer/Parser pipeline using Producer-Consumer model.

```bash
# Execute a script
./bin/bee.exe -e test/level2/T0201.bee
```

## 5. Execution Model
- **Hybrid VM/Compiler:** Bee features an integrated in-memory AST evaluator (VM) for immediate execution and testing, alongside the LLVM-based machine-code compiler.
- **Workflow:** Developers can run source files directly via the VM to verify logic before performing full native compilation.
- **Memory Consistency:** Both the VM and the LLVM backend strictly enforce the region-based memory model and 1-based indexing rules.
