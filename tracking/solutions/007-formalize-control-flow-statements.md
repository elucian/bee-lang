# Solution: Update Control Flow Specification
Formalize the language grammar to include `start` and `pass` control flow primitives.

## Design
1. **`start` block**: Define as a non-repetitive local scope initiator.
2. **`pass` statement**: Add to `transfer_stmt` EBNF as a skip/continue instruction.
3. **Grammar Update**: Update `spec/02-statements.md` to reflect these primitives in the formal grammar.
