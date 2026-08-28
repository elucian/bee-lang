# Issue: Print Statement Ambiguity and Advanced Formatting

## Description
The current `print` statement implementation is limited to basic comma-separated printing. The specification `spec/02-statements.md` defines it as `print expression ( "," expression )*`. However, there is a need to support advanced string formatting (e.g., custom separators, alignment, padding) which are currently outside the scope of `level0`.

## Solution
1. Define a clear syntax for advanced formatting in the specification.
2. Ensure the base `print` statement remains compatible with the comma-separated list syntax.
3. Introduce a structured way to handle formatting without polluting the core `print` grammar for simple cases.

## Proposed Syntax Extension
- **Separator specification**: Allow a format-like string or directive within the `print` statement.
- **Example**: `print "Value: ", a, b with separator: " | ";`

## Tracking
- [ ] Create issue/solution files for `print` extensions.
- [ ] Update `spec/02-statements.md` to define formatting syntax.
- [ ] Update `internal/parser` and `internal/evaluator` to handle new print grammar.
- [ ] Create new test cases in `test/level4/` (for advanced collection/processing interaction with print).
