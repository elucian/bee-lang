# Solution: Hybrid Execution Model (VM + Compiler)

## 1. Execution Strategy
- **VM/Interpreter:** Provides immediate, in-memory execution of Bee source files using AST evaluation. Used for fast prototyping, testing, and debugging.
- **LLVM Backend:** Generates standalone, machine-native executables for production.

## 2. AST Evaluator (VM Layer)
- **Evaluation:** The parser generates an Abstract Syntax Tree (AST). The Evaluator traverses this tree, resolving identifiers in the `rule` region and outputting results to `stdout`.
- **State management:** Variables are bound to the Region-Based Memory Model during the evaluation phase, ensuring consistent memory behavior between the VM and the eventual LLVM execution.
- **Goal:** Enable "compile-and-run" cycles without needing an external linker during development.
