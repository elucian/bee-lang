# Bee Compiler: Developer Manual

## 1. Environment Setup
To build the Bee compiler from source, the following dependencies are required:

### Dependencies
- **Go 1.26+:** [Download & Install](https://go.dev/dl/)
- **LLVM 18+:**
  - **Windows:** Download the [LLVM Windows Installer](https://github.com/llvm/llvm-project/releases). Ensure "Add LLVM to PATH" is selected during installation.
  - **Verification:** Run `llc --version` and `go version` in your terminal.

## 2. Project Build
Bee uses standard Go modules. Build the compiler from the root directory:
```bash
go build -o bee ./cmd/bee
```

## 3. Contributing Guidelines
- **Commit Strategy:** Use atomic, logical commits. Every milestone must be marked in `MANIFEST.md`.
- **Specification First:** Any language change must be reflected in `spec/` before implementation in `internal/`.
- **Memory Model:** Adhere to the hybrid memory strategy. Use `zap` only in documented "Hot Zones."
- **Testing:** Add new test cases in `test/levelX/`. Ensure all tests pass via `test/test_runner.py` before submitting.

## 5. Execution Model
- **Hybrid VM/Compiler:** Bee features an integrated in-memory AST evaluator (VM) for immediate execution and testing, alongside the LLVM-based machine-code compiler.
- **Workflow:** Developers can run source files directly via the VM to verify logic before performing full native compilation.
- **Memory Consistency:** Both the VM and the LLVM backend strictly enforce the region-based memory model and 1-based indexing rules.
