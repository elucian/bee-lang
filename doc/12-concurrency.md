# Bee Concurrency

Bee supports multi-threading and parallel execution via asynchronous rule dispatching.

## 1. Multi-threading
Rules can be invoked asynchronously using `begin`, creating isolated memory regions (Stack + Heap) per thread. Threads are synchronized using `wait`.

```bee
-- Asynchronous rule dispatch
rule main:
  for ∀ i ∈ (1..4) do
    begin test(i); 
  repeat;
  wait; -- Synchronize thread barrier
  return;
```

## 2. Map-Reduce Pattern
Parallel processing is typically achieved via the map-reduce design pattern, capturing results in a thread-safe list.

```bee
new results ∈ List(N);
-- Parallel dispatch to worker threads
begin sum(1, 25) +> results;
begin sum(26, 50) +> results;
wait; 
```

## 3. Coroutines (`yield`)
Coroutines allow suspension and resumption of stateful execution.

```bee
rule test(n ∈ N) => (result ∈ N):
  cycle:
    for ∀ i ∈ (1..n) do
      let result := i;
      yield; -- Suspend for main thread
    repeat;
  return;
```

## 4. Producer-Consumer
Asynchronous pattern where a producer dispatches tasks to a worker pool (consumers).
- **Producer:** Single-threaded dispatcher.
- **Consumer:** Multi-threaded worker pool.

**Read more:** [Graphics](/projects/bee/graphics/)
