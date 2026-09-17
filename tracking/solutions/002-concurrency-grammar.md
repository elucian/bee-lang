# Solution: Concurrency Grammar Integration
Formalize concurrency based on `test/level8/` implementations and `web/concurrency.html`.

## Design
1. **Async Execution**: `begin rule_call([args])`
2. **Result Capture**: `begin rule_call([args]) +> [collection]`
3. **Synchronization**:
   - `wait;` (no args): Synchronize (join) all background threads.
   - `wait [expr];`: Sleep for `[expr]` seconds.
4. **Coroutine Resumption**:
   - `yield [var] <- [routine_name]`
5. **Specification**: Add new grammar rules to `spec/12-concurrency.md`.
