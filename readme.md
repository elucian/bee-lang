# Bee Compiler: Documentation & Manual

## Chapter 1: Introduction & Architecture
- [Architecture Overview](manual/developer-guide.md)

## Chapter 2: Build & Testing Suite
- **Environment Setup:** `sh setup.sh`
- **Master Automation (`run.sh`):**
  - Build: `sh run.sh build`
  - Test: `sh run.sh test`, `sh run.sh test level1`, or `sh run.sh test T0104`
  - Iterative Fix: `sh run.sh fix level1` (max 3 iterations) or `sh run.sh fix T0104` (max 5 iterations)
- **Build Pipeline:** `python build.py`
- **Feature Testing (`test/level.py`):** Runs compliance and feature test vectors across levels 1–5 (`python test/level.py` or `python test/level.py --level <N>`).
- **Full Test Suite (`test.py`):** Root orchestrator script that runs all level tests (`python test.py` or `python test.py level1`).
- **Benchmark Performance Testing (`test/bench.py`):** Executes performance benchmarks and outputs JSON telemetry to `test/output/perf_report.json` (`python test/bench.py`).
- **Dryrun Flag Verification (`test/dryrun.py`):** Verifies all compiler flags (`-c`, `--compile`, `-e`, `--execute`, `-b`, `--beautify`) work properly (`python test/dryrun.py`).

## Chapter 3: Language Reference
- [Features](manual/ref/01-features.md)
- [Syntax](manual/ref/02-syntax.md)
- [Operators](manual/ref/03-operators.md)
- [Structure](manual/ref/04-structure.md)
- [Types](manual/ref/05-types.md)
- [Control Flow](manual/ref/06-control.md)
- [Rules](manual/ref/07-rules.md)
- [Functions](manual/ref/08-functions.md)
- [Objects](manual/ref/09-objects.md)
- [Collections](manual/ref/10-collections.md)
- [Processing](manual/ref/11-processing.md)
- [Concurrency](manual/ref/12-concurrency.md)
- [Graphics](manual/ref/13-graphics.md)
- [Library](manual/ref/14-library.md)

## Chapter 6: Self-Testing & Performance Benchmarks
- **Performance Suite (`test/bench/`):** Dedicated benchmark use cases for collection indexing, memory handling, and evaluation throughput.
- **Performance Reporting:** Run `python test/bench.py` to execute benchmarks and generate JSON execution telemetry in `test/output/perf_report.json`.
