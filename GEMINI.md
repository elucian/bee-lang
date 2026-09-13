# BEE COMPILER ARCHITECT & GENERATOR SYSTEM - LEXER & PARSER INVARIANTS

## 4. Architectural Constraints
- **Go Standards:** Idiomatic Go only. Prohibit `panic` in compiler error flows; propagate explicit `error` returns.
- **Array Semantics:** Bee source is **1-based**. Lists, arrays, matrices, and string indices begin at `1`. The `$` token is the dynamic **end anchor** (index of the last element), not a zero terminator. Indexing across all compiler passes must use 1-based semantics.
- **IR Conversion (bridge to LLVM):** All collection indexing must be lowered to **0-based** at the IR boundary. The lowering rule is `ir_index = source_index - 1`. The `$` anchor lowers to `ir_index = len(collection) - 1`. This conversion is centralized in `internal/compiler/` and must be the only place where the delta is applied; lexer, parser, typechecker, and evaluator stay in 1-based land.
- **LLVM IR Generation:** Ensure every basic block terminates explicitly (`CreateBr`, `CreateRet`, `CreateCondBr`). Validate operands against `llvm.Type` before emission.
- **Diagnostics:** Route debug logs strictly to `os.Stderr`. Reserve `stdout` exclusively for clean build outputs or raw IR.

## 2. Lexer & Parser Unicode & Rune Handling Instructions
When generating, modifying, or debugging Go code for the Bee programming language lexer and parser:
1. **UTF-8 Handling:** Treat all source inputs strictly as decoded UTF-8 sequences using `[]rune` or `bufio.Reader.ReadRune()`. Never use raw byte indexing (`s[i]`) for tokenization.
2. **Operator Syntax:** Bee operators can consist of single Unicode symbols (e.g., `≠`, `¬`), standard ASCII symbols (e.g., `and`, `or`), mixed combinations of Unicode and ASCII characters, and Unicode superscript/subscript ranges (`²`, `³`, `²√`, `³√`).
3. **Lookahead Matching:** Implement lexing with rune lookahead (`peekRune()`) rather than fixed-size assumptions to correctly parse multi-character operators containing mixed scripts or modifiers.
4. **Token Definitions:** Store operators as `[]rune` slices in token lookup tables or transition tries to support arbitrary multi-rune Unicode operators.
5. **Operator Synonymy (Decision 7):** The ASCII keywords `and`, `or`, `xor`, `not` are canonical synonyms for `∧`, `∨`, `⊕`, `¬`. The lexer MUST emit a single stable TokenType for each pair (`LOGICAL_AND` for both `and` and `∧`, etc.) so the parser has one branch per logical connective.
6. **`is not` is one token (Decision 7):** Maximal-Munch collapsing is mandatory: when `is` is followed by ASCII whitespace and `not`, the lexer emits a single `IS_NOT` token. Both literal `is not` and parenthetical `(a is not b)` patterns MUST bind this way.

## 5. Decisions Backlog — Reference
For any new operator, statement, keyword, or grammar production, first consult `todo/DECISIONS.md`. D1–D7 are ratified and mirrored in `MANIFEST.md`. D8 is deferred. D9, D10, D11 are pending the user's ratification before any implementation begins.

## 3. Test Lifecycle & Freeze Protocol (.bee Test Cases)
* **Spec-Driven Generation:** Generate new test files (`.bee`) strictly under `test/levelX/` derived directly from `/spec/`.
* **Locking Created Tests (AI-Immunity):** Every newly generated `.bee` test file MUST include this header tag on line 1:
  `-- @FROZEN: Generated from /spec/. Immutable ground truth for AI agents.`
* **AI Read-Only Enforcement:** Test files in `test/levelX/*.bee` marked `@FROZEN` are strictly immutable for AI agents. AI agents must NEVER modify `.bee` test inputs, assertions, or expected outputs to force a failing compiler build to pass. (Human users retain full permission to modify or author tests).
* **Disable, Never Delete:** If a test fails persistently across fix attempts, AI agents must NEVER delete the file. Disable it by prepending `-- @DISABLED: <reason>` on line 1 of the `.bee` file.

## AI Agent Protocol & Anti-Loop Rules

1.  **Strict Anti-Loop Protocol:** 
    - NEVER attempt more than one edit pass per user prompt.
    - If a test fails after one fix attempt, halt immediately, write error analysis to `os.Stderr`, and yield back without modifying files.
    - NEVER "guess" fixes. Analyze the implementation against the `/spec` first.
    - Before applying an edit, explain the root cause identified by comparing the implementation against the relevant `/spec`.
    - If you are stuck, stop. Do not loop.

2.  **Implementation Invariants:**
    - Always verify the parser and evaluator against the EBNF grammar in `spec/02-statements.md` for mutation operators.
    - Use absolute or relative paths starting from project root (`bee-lang/`) for all file operations.
    - Check the authoritative keyword list in `internal/token/token.go` before introducing or modifying language keywords.
    - All language documentation and specification changes MUST occur in the `/spec` directory.
    - **TDD Integration:** Every feature or fix must maintain test parity:
        1. Audit `/spec/` and update EBNF.
        2. Create/update a `.bee` test case in `test/levelX/` with a `-- @DESC:` tag.
        3. Run `python test/solo.py <test_name>` to verify and auto-update `test/levelX/README.md`.
        4. Run `sh run.sh smoke` for system-wide health check before yielding.

## Final message
When you finish send this message: "Task Completed in <runtime>"
