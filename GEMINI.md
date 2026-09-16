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
For any new operator, statement, keyword, or grammar production, first consult `todo/DECISIONS.md`. D1–D7 and D9–D16 are ratified and mirrored in `MANIFEST.md`. D8 (rule-call result destructuring) is deferred. All ratified decisions (including D10 radical precedence and D11 parallel colon-initialisation, both ratified 2026-09-14) are ready for implementation.

## 3. Test Lifecycle & Freeze Protocol (.bee Test Cases)
* **Spec-Driven Generation:** Generate new test files (`.bee`) strictly under `test/levelX/` derived directly from `/spec/`.
* **Locking Created Tests (AI-Immunity):** Every newly generated `.bee` test file MUST include this header tag on line 1:
  `-- @FROZEN: Generated from /spec/. Immutable ground truth for AI agents.`
* **AI Read-Only Enforcement:** Test files in `test/levelX/*.bee` marked `@FROZEN` are strictly immutable for AI agents. AI agents must NEVER modify `.bee` test inputs, assertions, or expected outputs to force a failing compiler build to pass. (Human users retain full permission to modify or author tests).
* **Disable, Never Delete:** If a test fails persistently across fix attempts, AI agents must NEVER delete the file. Disable it by prepending `-- @DISABLED: <reason>` on line 1 of the `.bee` file.

## 6. Tutorial Maintenance & Synchronization (Decision 5)

### Source-of-Truth Model
- `/spec/` is the single source of truth for **language facts** (grammar,
  EBNF, operators, keywords, diagnostics, semantics).
- `/tutorial/` is the single source of truth for the **rendered HTML pages**
  served on the public SCL site.
- The tutorial is a *derived presentation* of the spec, never the reverse.
  A language fact is decided in `/spec/` first and then expressed in
  `/tutorial/`; never author it in a tutorial page and treat the spec as
  "behind".

### Where to Work (local copy + one-way `sync`)
- **Edit tutorial pages only under `bee-lang/tutorial/`.** This is a
  *versioned, local* directory in this repository — it is the source of
  truth for the rendered HTML, and it is **not** a symlink/junction.
- The external SCL repository directory (`C:\Users\eluci\sage-code\scl\projects\bee`)
  is the *live site content* target. It is kept in sync with this repo via
  a one-way mirror (`scripts/sync_tutorial.py`), not via a filesystem link.
- `/web/` has been removed; there is no legacy local copy.
- **External git is user-owned:** committing and pushing the external SCL
  repository is a separate step owned by the user. The sync script only
  mirrors files; it never performs git operations against the SCL repo.

### Running the Sync
Mirror files between `bee-lang/tutorial/` and the SCL site directory:

| Command | Effect |
| :--- | :--- |
| `sh run.sh sync` | Push local → SCL (publish edits). Default direction. |
| `sh run.sh sync pull` | Pull SCL → local (adopt SCL-side edits). |
| `sh run.sh sync --dry-run` | Preview the push without writing. |
| `sh run.sh sync --delete` | Push *and* remove SCL files absent locally (true mirror). |
| `sh run.sh sync --scl <PATH>` | Override the SCL directory path. |

Underlying helper: `scripts/sync_tutorial.py` (plain-Python mirror; the
environment has no `rsync`). It copies in one direction at a time, is
default-safe (never deletes unless `--delete` is passed), and skips files
whose size/mtime already match. Push or pull are the two supported
operations — the sync is **not** a bidirectional merge.

### Spec ↔ Tutorial Mapping
When a spec module changes, update its mapped page(s) in the same change
set:

| Spec module | Tutorial page(s) |
| :--- | :--- |
| `01-lexical-structure.md` | `syntax.html`, `operators.html`, `js/bee.js` (highlighter) |
| `02-statements.md` | `control.html`, `structure.html` |
| `03-rules.md` | `rules.html` |
| `04-structure.md` | `structure.html` |
| `05-types.md` | `types.html` |
| `06-objects.md` | `objects.html` |
| `07-functions.md` | `functions.html` |
| `10-collections.md` | `collections.html` |
| `11-processing.md` | `processing.html` |
| `12-concurrency.md` | `concurrency.html` |
| `13-graphics.md` | `graphics.html` |
| `14-library.md` | `library.html` |
| `00-memory-model.md` | `features.html` (overview notes; no dedicated page) |

Meta/navigation pages (`index.html`, `template.html`, `features.html`) are
not tied to a single spec module and change only for structural reasons.

### Synchronization Invariant
Every `/spec/` change MUST be accompanied by the corresponding `/tutorial/`
update in the **same change set**. A spec/tutorial pair that disagrees about
the same operator, keyword, or grammar production is a defect. Before any
language edit, ask: *"does the tutorial still agree with `/spec/` on this
fact?"* and update both sides together.

### Edit Authorization
- **Pre-Authorized Target:** `/tutorial/` is a versioned local directory in
  this repository. It is mirrored to the user's external live repository
  (`C:\Users\eluci\sage-code\scl\projects\bee`) via `sh run.sh sync`.
- **No Confirmation Prompts:** When the user asks to modify the tutorial, AI
  agents MUST edit files under `bee-lang/tutorial/` directly. NEVER ask for
  confirmation, warn that the target is external, or treat the sync boundary
  as a reason to halt.
- **Scope:** This authorization covers all files under `/tutorial/` (HTML
  pages, `js/`, `data/`, `symbols/`, `img/`). It does NOT authorize pushing,
  committing, or any git operations against the external SCL repository —
  file edits and `sh run.sh sync` mirroring only, unless the user explicitly
  asks.

## AI Agent Protocol & Anti-Loop Rules

1.  **Edit Discipline (diff-first, chunked):**
    - **Existing files → diffs.** Modify with targeted, minimal edits (`edit_file`).
      Never rewrite a whole file for a small change; never re-read a file after
      your own edit (the tool confirms success).
    - **New files → chunked appends.** Create small, self-contained chunks
      (a section, a function, a table) and append them incrementally rather than
      emitting one enormous file. Every chunk should be independently reviewable.
    - **Spec ↔ tutorial pairs** must ship together in the same change set (§6
      synchronization invariant).
    - Keep changes surgical: no unrelated refactors, no drifted formatting, no
      logic-touching edits beyond the task.

2.  **Stop-Loop Protocol:**
    - **One fix attempt per prompt.** If a test fails after one targeted fix,
      halt immediately: explain the failure to `os.Stderr`, yield back, and do
      **not** modify more files.
    - **Never guess.** Diagnose the root cause by comparing the implementation
      against the relevant `/spec`/EBNF before editing; state that root cause.
    - **Ask when blocked or in doubt.** If you cannot identify the root cause,
      the fix risks a collision, or you are unsure of intent, **stop and ask the
      user** with the specific question and options — do not re-attempt the same
      error in a loop.

3.  **Efficient Communication:**
    - Lead each work unit with a one-to-two sentence preamble; skip narration
      for trivial reads.
    - Report decisions/state in concise bullets. Reference files by
      project-relative path.
    - When you finish, summarize: what changed, files touched, and exactly what
      validation you ran (commands + pass/fail).

4.  **Implementation Invariants:**
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
