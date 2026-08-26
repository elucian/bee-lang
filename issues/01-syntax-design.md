# Language Specification Issues

- [x] **Ambiguous Block Terminations:** Resolved in `spec/02-statements.md`. Block terminators are strictly defined: `done` for `if`/`with`/`match`/`trial`, `repeat` for `cycle`/`while`/`for`, and `return` for `rule`/routines.
- [x] **`match` keyword vs `if-else` ladder:** Formalized in `spec/02-statements.md`. `if-else` is used for binary/boolean branches; `match` supports `first`, `every`, and `total` matching modes.
- [x] **Formal Grammar Definition:** EBNF grammar fully defined in `spec/01-lexical-structure.md` and `spec/02-statements.md`.
