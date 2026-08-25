# Bee Language & Compiler

Bee is a systems programming language developed locally. This repository holds the official language specification and the reference compiler implementation in Go.

## Specifications

The core language specifications are organized in the [spec](spec/readme.md) directory.

## Reference Compiler

The reference compiler is being implemented in **Go**.

## Project Structure

- `cmd/bee/`: Main entry point for the Bee compiler CLI.
- `internal/`: Core compiler components.
- `spec/`: Formal language specifications.
- `doc/`: Human-readable user documentation and design principles.
- `plan/`: Strategic roadmap and development milestones.
- `solution/`: Reference implementation patterns and solutions.
- `todo/`: Current task tracker and issue backlog.
- `tests/`: Integration test cases (`.bee` files).
- `scripts/`: Utility scripts for development and testing.

## Getting Started

1. **Build:**
   ```bash
   go build -o bee ./cmd/bee
   ```

2. **Run Tests:**
   The project uses a Python test runner to automatically validate `.bee` test cases.
   ```bash
   python3 scripts/test_runner.py
   ```
