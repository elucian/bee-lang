# Solution 15: Contract & Invariant Verification System (`assert` & `expect`)

## 1. Context & Motivation
Previously, Bee explored Eiffel-style `require` and `ensure` keywords for Design by Contract. To better align with pragmatic engineering diagnostics and modern test-driven workflows:
1. **Preconditions / Diagnostics (`assert`):** Replaces `require`. When an `assert` expression evaluates to false, it logs a non-fatal warning diagnostic to `stderr` (`os.Stderr`) and allows execution to proceed.
2. **Invariants / Postconditions (`expect`):** Replaces `ensure`. When an `expect` expression evaluates to false, it raises an unrecoverable runtime error (or propagates through `trial` transactional error handlers), terminating the program with a non-zero exit status if unhandled.

---

## 2. Grammar & Syntax Specification

### 2.1 EBNF Statement & Contract Grammar
```ebnf
(* Statements *)
statement         ::= decl_stmt | mutation_stmt | memory_stmt | io_stmt | contract_stmt | control_stmt | trial_stmt | transfer_stmt ";" ;

contract_stmt     ::= "assert" expression
                    | "expect" expression ;

(* Rule Contract Clauses *)
contract_clause   ::= ( "assert" condition ";" )* ( "expect" condition ";" )* ;
```

### 2.2 Rule Contract Layout
```bee
rule calculate_ratio(numerator ∈ R, denominator ∈ R) => (result ∈ R):
  assert denominator ≠ 0;             -- Emits warning to stderr if denominator is 0
  let result := numerator / denominator;
  expect result * denominator ≈ numerator; -- Halts with failure if invariant is violated
return;
```

---

## 3. Runtime Semantics & Diagnostics

### 3.1 `assert condition;`
- **Evaluation:** Evaluates `condition` as a boolean expression.
- **On Success (`true` / `≠ 0`):** No action taken.
- **On Failure (`false` / `0`):** Formats and prints a warning message to `stderr`:
  ```
  [WARNING] Assertion failed at line <line_num>
  ```
- **Execution Flow:** Resumes normal execution without interrupting control flow or modifying program state.

### 3.2 `expect condition;`
- **Evaluation:** Evaluates `condition` as a boolean expression.
- **On Success (`true` / `≠ 0`):** Execution continues normally.
- **On Failure (`false` / `0`):** Dumps the variable context to `stderr` and raises `E0303: ExpectationFailed`.
- **Execution Flow:** Halts program execution with a non-zero exit code unless intercepted by an enclosing `trial` block.

---

## 4. Compiler & Toolchain Implementation Plan
1. **Token Layer (`internal/token`):** Register `ASSERT` (`assert`) and `EXPECT` (`expect`) keywords in `token.go`.
2. **Parser Layer (`internal/parser`):** 
   - Define `AssertStatement` and `ExpectStatement` AST nodes.
   - Parse `assert <expr>;` and `expect <expr>;` within general statement sequences and rule contract headers.
3. **Evaluator Layer (`internal/evaluator`):**
   - Evaluate `*parser.AssertStatement` with warning emission to `os.Stderr`.
   - Evaluate `*parser.ExpectStatement` with error propagation / panic halt.
