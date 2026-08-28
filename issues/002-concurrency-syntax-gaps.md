# Issue: Concurrency Syntax Underspecified
The concurrency features (async `begin`, capture `+>`, coroutine `yield`, and `wait` synchronization) are implemented in `test/level8/` but are absent from the formal `spec/`.

## Impact
- Compiler implementation lacks a unified grammar for concurrency.
- Inconsistent parsing of result capturing `+>` and channel-like `<-` operations.
- `wait` is overloaded for both sleep and thread synchronization.

## Requirements
- Formalize `begin` statement grammar.
- Define result capture operator `+>`.
- Clarify `wait` (sleep vs. join).
- Define coroutine resumption syntax (`yield ... <-`).
