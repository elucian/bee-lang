# BEE COMPILER ARCHITECT & GENERATOR SYSTEM - LEXER & PARSER INVARIANTS

## 1. Engineering Invariants
* Pure Go standard library. Zero external dependencies.
* Explicit `error` return values only. Never call `panic` or `recover` for compiler errors.
* Enforce zero-based array indexing across parser, AST, type-checker, and codegen passes.
* Output debug traces strictly to `os.Stderr`. Keep `stdout` clear for generated code or execution output.

## 2. Lexer & Parser Unicode & Rune Handling Instructions
When generating, modifying, or debugging Go code for the Bee programming language lexer and parser:
1. **UTF-8 Handling:** Treat all source inputs strictly as decoded UTF-8 sequences using `[]rune` or `bufio.Reader.ReadRune()`. Never use raw byte indexing (`s[i]`) for tokenization.
2. **Operator Syntax:** Bee operators can consist of single Unicode symbols (e.g., `≠`), standard ASCII symbols, mixed combinations of Unicode and ASCII characters, and Unicode superscript/subscript ranges.
3. **Lookahead Matching:** Implement lexing with rune lookahead (`peekRune()`) rather than fixed-size assumptions to correctly parse multi-character operators containing mixed scripts or modifiers.
4. **Token Definitions:** Store operators as `[]rune` slices in token lookup tables or transition tries to support arbitrary multi-rune Unicode operators.

## 3. Test Lifecycle & Freeze Protocol (.bee Test Cases)
* **Spec-Driven Generation:** Generate new test files (`.bee`) strictly under `test/levelX/` derived directly from `/spec/`.
* **Locking Created Tests (AI-Immunity):** Every newly generated `.bee` test file MUST include this header tag on line 1:
  `// @FROZEN: Generated from /spec/. Immutable ground truth for AI agents.`
* **AI Read-Only Enforcement:** Test files in `test/levelX/*.bee` marked `@FROZEN` are strictly immutable for AI agents. AI agents must NEVER modify `.bee` test inputs, assertions, or expected outputs to force a failing compiler build to pass. (Human users retain full permission to modify or author tests).
* **Disable, Never Delete:** If a test fails persistently across fix attempts, AI agents must NEVER delete the file. Disable it by prepending `// @DISABLED: <reason>` on line 1 of the `.bee` file.

## AI Agent Protocol & Anti-Loop Rules

1.  **Strict Anti-Loop Protocol:** 
    - Never attempt more than one edit pass per user prompt.
    - If a test fails after one fix attempt, halt immediately, write error analysis to `os.Stderr`, and yield back without modifying `.bee` files.
    - Never use loops in reasoning or tool execution that rely on "guessing" fixes. Analyze the implementation against the `/spec` first.
    - Before applying an edit, explain the root cause identified by comparing the implementation against the relevant `/spec`.

2.  **Implementation Invariants:**
    - Always verify the parser and evaluator against the EBNF grammar in `spec/02-statements.md` for mutation operators.
    - If an operator exists in the grammar but fails, check if the `parser.go` lookup logic or the `evaluator.go` switch statement handles all valid `assign_op` variants (`:=`, `::`, `+=`, `-=`, `*=`, `/=`, `%=`, `^=`, `√=`).
    - Use absolute or relative paths starting from project root (`bee-lang/`) for all file operations.

## Final message
When you finish send this message: "Task Completed in <runtime>"
