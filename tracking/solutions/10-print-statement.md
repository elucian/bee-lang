# Solution: Print Statement Formatting

## 1. Syntax Design
Extend the `io_stmt` in `spec/02-statements.md`:
```ebnf
io_stmt           ::= "print" "(" expression ( "," expression )* ")" [ "using" ":" expression ]
```

## 2. Implementation
- **Parser**: Modify `parsePrintStatement` to check for an optional `using` clause.
- **Evaluator**: Update `evalStatement` for `*parser.PrintStatement` to read the separator expression and use it during concatenation of expressions.

## 3. Compatibility
- Standard `print a, b;` continues to use space separation.
- Advanced `print a, b using: " | ";` uses the specified separator.
