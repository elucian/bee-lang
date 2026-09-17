# Bee Compiler

Welcome to the **Bee Programming Language** repository. This project contains the compiler source and test infrastructure.

> **Agent rules** live in [`config/`](config/AGENTS.md) — the canonical
> contract for every AI agent and contributor. The **developer manual** (build
> & workflow procedures, design docs, decision history) lives in
> [`manual/`](manual/DEVELOPER.md).

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

This project is optimized for automation testing. We use TDD method. First we create the test, then we automate the test then we run the test and then we improve the compiler to pass the test suites.

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
- [x] Test Infrastructure Migration (Level 0 - Level 8 Audit & README Sync)
- [x] Print Statement Syntax Extension (Issue #10)
1. **Architectural Gap**: Any design change must first be documented in `/tracking/issues/` and `/tracking/solutions/`.
2. **Specification Update**: Update relevant `/spec/` EBNF and rules.
3. **Test-First**: Create a failing `.bee` test case under `/test/levelX/`.
4. **Implementation**: Modify the compiler to satisfy the spec and pass the test.
5. **Freeze**: Mark new tests as `@FROZEN` in `/test/levelX/`.

---

## 4. Repository Structure

- **[/config](config/)**: Agent configuration and adapters — `AGENTS.md` (canonical rules), `GEMINI.md`, `CLAUDE.md`, `copilot-instructions.md`, `ZED.md`.
- **[/manual](manual/)**: Developer manual and design docs — `DEVELOPER.md`, `MANIFEST.md`, `DECISIONS.md`, `CONTEXT.md`.
- **[/spec](spec/)**: Official language specification and documentation.
- **[/test](test/)**: Test suites, test runners, and infrastructure.
- **[/tracking](tracking/)**: Unified project-state hub — `issues/` (tracker), `solutions/` (strategies), `todo/` (tech debt & backlog).
- **[/registry](registry/)**: Diagnostic-code registry — the runtime source of truth for every `E`/`W` code via `diagnostics.json`.

### Getting started
- **Read the manual**: [manual/DEVELOPER.md](manual/DEVELOPER.md) — orientation, spec-driven TDD lifecycle, tutorial sync, and `run.sh` reference.
- **See the agent contract**: [config/AGENTS.md](config/AGENTS.md) — the normative rules all AI agents must follow.

---
