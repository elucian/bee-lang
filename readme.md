# Bee Compiler

Welcome to the **Bee Programming Language** repository. This project contains the compiler source and test infrastructure.

---

## 1. Environment Setup & Prerequisites

To build and develop the Bee compiler from source, ensure you have:
- **Go 1.26+**: [Download & Install Go](https://go.dev/dl/)
- **LLVM 18+**: Installed and added to system `PATH` (verify via `llc --version`)
- **Git Bash / Terminal**: Recommended on Windows.

---

## 2. Project Build & Execution

### Building the Compiler
```bash
python build.py
```

### Installation & PATH Setup
```bash
sh setup.sh
```
*(Run once to add `bee` to your PATH).*

### Compiler Modes & Flags
- `-d` / `--debug`: Output token stream to `stderr`.
- `-e` / `--execute`: Enable in-memory AST evaluator (VM) to execute code.
- `-c` / `--compile`: Parse and validate syntax only.

---

## 3. Workflow Automation & Developer Notes

### Automation Scripts
- **Reset Environment**: `python test/reset.py`
  *(Cleans build artifacts and test outputs/statuses).*
- **Run Syntax Check**: `python scripts/check.py`
  *(Runs syntax checks on all test files. Generates reports in `test/output/` and statuses in `test/status/`).*
- **Master Workflow (`run.sh`)**:
  - Build compiler: `sh run.sh build`
  - Run test pipeline: `sh run.sh test`
  - Iterative fix loop: `sh run.sh fix level1` (up to 10 iterations)

---

## 4. Repository Structure

- **[/spec](spec/)**: Official language specification and documentation.
- **[/test](test/)**: Test suites, test runners, and infrastructure.
- **[/issues](issues/)**: Issue tracker, logs, and known bug documentation.
- **[/solution](solution/)**: Implementation strategies and architectural notes.

---
