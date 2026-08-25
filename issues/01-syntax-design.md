# Language Specification Issues

- **Ambiguous Block Terminations:** While `done` is documented as the standard terminator, some `cycle` structures mention `repeat` or `cycle` as terminators. The spec needs to explicitly define which terminator is required for which construct.
- **`match` keyword vs `if-else` ladder:** The distinction between when to use a `match` statement versus a `ladder` is not clearly defined in terms of performance or design intent.
- **Missing Grammar Definition:** The current documentation describes features through examples but lacks a formal BNF/EBNF grammar definition, leading to potential implementation divergence.
