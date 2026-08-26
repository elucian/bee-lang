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

## 4. Architecture Overview
- **Frontend:** Hand-written Lexer and recursive descent Parser.
- **Backend:** LLVM IR generation via `llir/llvm`.
- **Memory:** Region-based allocation bound to `rule` lifecycles.
