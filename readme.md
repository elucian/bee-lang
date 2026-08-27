# Bee Compiler & Language Manual

Welcome to the **Bee Programming Language** official manual and documentation repository.

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

### Usage Examples
```bash
# Debugging lexer tokens
./bin/bee.exe -d test/level2/T0202.bee

# Execute logic in VM
./bin/bee.exe -e test/level2/T0202.bee

# Syntax check only
./bin/bee.exe -c test/level1/T0101.bee
```

---

## 3. Workflow Automation & Developer Notes

### Master Workflow (`run.sh`)
- Build compiler: `sh run.sh build`
- Run test pipeline: `sh run.sh test`
- Iterative fix loop: `sh run.sh fix level1` (up to 10 iterations)

*(Tip: You can alias `run` in your shell configuration: `echo 'alias run="./run.sh"' >> ~/.bashrc && source ~/.bashrc`)*

### Individual Test Execution
- Syntax Check: `./bin/bee.exe -c test/level1/T0101.bee`
- Debug Tokens: `./bin/bee.exe -d test/level1/T0101.bee`
- Evaluator: `./bin/bee.exe -e test/level2/T0202.bee`

---

## 4. Repository Structure & Links

- **[/manual](manual/)**: Complete language manual and specification guide.
- **[/test](test/)**: Test suites, test runners (`test.py`), benchmark tools (`test/bench.py`), and test vectors.
- **[/issues](issues/)**: Issue tracker, logs, and known bug documentation.
- **[/solution](solution/)**: Implementation strategies, architectural notes, and solutions.

---

## 5. Language Reference

| # | Topic | Description |
|---|-------|-------------|
| 01 | [Features](manual/01-features.md) | Design principles |
| 02 | [Syntax](manual/02-syntax.md) | Language grammar |
| 03 | [Operators](manual/03-operators.md) | Symbol definitions |
| 04 | [Structure](manual/04-structure.md) | Modular architecture |
| 05 | [Types](manual/05-types.md) | Type system |
| 06 | [Control Flow](manual/06-control.md) | Conditional logic |
| 07 | [Rules](manual/07-rules.md) | Rule-oriented design |
| 08 | [Functions](manual/08-functions.md) | Lambda expressions |
| 09 | [Objects](manual/09-objects.md) | OOP in Bee |
| 10 | [Collections](manual/10-collections.md) | Data structures |
| 11 | [Processing](manual/11-processing.md) | Operators & manipulation |
| 12 | [Concurrency](manual/12-concurrency.md) | Multithreading & coroutines |
| 13 | [Graphics](manual/13-graphics.md) | 2D Geometry |
| 14 | [Library](manual/14-library.md) | Standard API |
