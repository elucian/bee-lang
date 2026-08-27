# Bee Compiler: Documentation & Manual

## Chapter 1: Introduction & Architecture
- [Architecture Overview](manual/developer-guide.md)

## Chapter 2: Build & Test
- [Build Pipeline](manual/developer-guide.md#2-project-build--execution)
- [Testing Suite](test/readme.md)

## Chapter 3: Language Reference
- [Features](manual/ref/01-features.md)
- [Syntax](manual/ref/02-syntax.md)
- [Operators](manual/ref/03-operators.md)
- [Structure](manual/ref/04-structure.md)
- [Types](manual/ref/05-types.md)
- [Control Flow](manual/ref/06-control.md)
- [Rules](manual/ref/07-rules.md)
- [Functions](manual/ref/08-functions.md)
- [Objects](manual/ref/09-objects.md)
- [Collections](manual/ref/10-collections.md)
- [Processing](manual/ref/11-processing.md)
- [Concurrency](manual/ref/12-concurrency.md)
- [Graphics](manual/ref/13-graphics.md)
- [Library](manual/ref/14-library.md)

## Chapter 5: AST Evaluator & Execution Engine (Phase 4)
- Implemented `internal/evaluator/evaluator.go` to evaluate AST programs, variables, binary operations, assertions (`expect`), and print statements.
