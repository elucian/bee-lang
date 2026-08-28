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

## 3. Workflow Automation & Developer Notes

### Automation Scripts
- **Reset Environment**: `python test/reset.py`
  *(Cleans build artifacts and test outputs/statuses).*
- **Run Syntax Check**: `python scripts/check.py`
  *(Runs syntax checks on all test files).*
- **Master Workflow (`run.sh`)**:
  - Build compiler: `sh run.sh build`
  - Run test pipeline: `sh run.sh test`
  - Run self-health check: `sh run.sh smoke`

### Development Methodology (TDD & Specification-Driven)
Bee development follows a **Strict Test-Driven Development (TDD)** lifecycle:
- [x] Operator Taxonomy (Binding `:` vs. Mutation `:=` vs. Clone `::`)
- [x] Operator Taxonomy (Equality `=` vs. Reference `==` vs. Equivalence `≡`)
1. **Architectural Gap**: Any design change must first be documented in `/issues/` and `/solution/`.
2. **Specification Update**: Update relevant `/spec/` EBNF and rules.
3. **Test-First**: Create a failing `.bee` test case under `/test/levelX/`.
4. **Implementation**: Modify the compiler to satisfy the spec and pass the test.
5. **Freeze**: Mark new tests as `@FROZEN` in `/test/levelX/`.

---

## 4. Repository Structure

- **[/spec](spec/)**: Official language specification and documentation.
- **[/test](test/)**: Test suites, test runners, and infrastructure.
- **[/issues](issues/)**: Issue tracker, logs, and known bug documentation.
- **[/solution](solution/)**: Implementation strategies and architectural notes.

---
