# Self-Testing & Performance Profiling Plan (`-t` / `--test`)

## 1. Objectives & Scope
- **Self-Testing Mode (`-t` / `--test`):** Embed an internal self-test runner directly inside the compiler binary (`bin/bee.exe`) so it can validate its own subsystems (lexer, parser, evaluator, standard library) programmatically without relying solely on external test runner scripts.
- **Performance Profiling:** Measure micro-benchmarks and end-to-end execution latency (parsing speed, symbol resolution, memory allocation tracking, evaluation time) and output structured reports.
- **Automated Integration:** Introduce a dedicated `test/bench/` folder containing performance use cases and micro-benchmarks.

## 2. CLI Flag Integration
- `--test` or `-t`: Executes the embedded compiler benchmark and self-test suite.
- JSON performance report generated automatically in `test/output/perf_report.json`.

## 3. Use Cases for Self-Tests & Benchmarks
1. **Lexical Munch Speed Test:** Benchmarking UTF-8 tokenization throughput on large source files (`test/bench/large_source.bee`).
2. **AST Parsing Complexity Test:** Measuring recursive descent parsing depth and tree construction latency.
3. **Array & Collection Indexing Benchmark:** Evaluating 1-based indexing (`lista[$]`, `lista[x]`) and collection iteration performance.
4. **Concurrency Barrier Simulation:** Testing thread barrier spawn/wait correctness (`begin`/`wait`).
