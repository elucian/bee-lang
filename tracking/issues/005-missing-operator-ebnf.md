# Issue: Incomplete Operator EBNF in Specification
The current `spec/01-lexical-structure.md` is missing formal definitions for several operators present in the compiler implementation (`evaluator`, `beautifier`) and test suites.

## Impact
- Compiler implementation relies on undocumented syntax.
- Static analysis and type checking cannot reliably validate operators not present in the EBNF.
- Discrepancy between implemented features (e.g., `+=`, `√=`, `+>`) and official language spec.

## Requirements
- Identify and formalize all missing arithmetic, modifier, and collection operators.
- Update EBNF grammar to include these operators.
- Define precedence for the newly added operators.
