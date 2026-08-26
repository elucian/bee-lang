# Bee Specification: Concurrency (12-concurrency.md)

## 1. Threading Grammar
```ebnf
thread_stmt  ::= "begin" rule_call ";" ;
sync_stmt    ::= "wait" ";" ;
yield_stmt   ::= "yield" identifier? ";" ;
```

## 2. Parallel Orchestration
- **Isolation:** Each `begin` spawned thread is allocated a private Memory Region (Stack + Heap).
- **Communication:** Shared mutations (e.g., `<+` into a shared List) are managed via a runtime-level, lock-free message-passing queue.
- **Synchronization:** The `wait` barrier forces the parent thread to block until all child regions have synchronized their message-queues into the parent region.
- **Error Propagation:** Worker thread exceptions are captured in the thread-local `$trial` report and raised by the parent `wait` call.
