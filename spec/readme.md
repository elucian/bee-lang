# Bee Language Specification Index

This repository contains the formal specification of the Bee programming language. The specification is structured into modular, numbered sections for human readability and machine-parsing efficiency.

## Specification Index

| ID | Topic | Description |
|----|-------|-------------|
| 00 | [Memory Model](00-memory-model.md) | Hybrid memory management (RC, MMM, GC) and the `zap` directive. |
| 01 | [Lexical Structure](01-lexical-structure.md) | UTF-8 encoding, tokenization, and Unicode operator definitions. |
| 02 | [Syntax Statements](02-syntax-statements.md) | Grammar for statements, blocks, and declarations. |
| 03 | [Rules & Lambdas](03-rules-lambdas.md) | Rule definitions, parameter binding, and lambda expressions. |
| 04 | [Concurrency & Types](04-concurrency-types.md) | Threading, coroutines, and primitive/composite types. |

*Note: For the implementation, refer to the `bee/solution/` folder for architectural decisions.*
