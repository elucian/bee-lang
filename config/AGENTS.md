# BEE COMPILER — AGENT INSTRUCTIONS (CANONICAL)

Status: canonical, single source of truth for AI-agent behavior in this
repository. Applies to **all** AI agents (Claude, ChatGPT/Copilot, Gemini,
DeepSeek, Zed, …) and all human contributors on every machine and in every
session.

## How to use this file

- **This is the contract.** Every agent resolves `AGENTS.md` to this file
  (`config/AGENTS.md`) before editing. A root `AGENTS.md` stub and the root
  `GEMINI.md`/`CLAUDE.md` and `.github/copilot-instructions.md` stubs exist
  only so tools auto-discover a well-known name; each stub hands off here. Do
  not rely on a differently-named file holding the real rules.
- **Adapters are thin and may not fork rules.** `config/GEMINI.md`,
  `config/CLAUDE.md`, and `config/copilot-instructions.md` exist only because
  some tools look for their own name first. Each is a one-line import pointing
  here. Never copy a rule into an adapter — drift between an adapter and this
  file is a defect.
- **To change agent behavior, edit this file and** keep the adapters'
  import text in sync in the **same change set** (mirrors the §6
  spec↔tutorial invariant).
- **Prompt-cache signature.** Prepend the current signature (see
  `manual/CONTEXT.md`) to the task prompt before any other project context
  whenever a task-scoped session starts. `config/AGENTS.md` is one of the
  canonical hashed files.

---

## 1. Output & verbosity discipline (maximize the usable output budget)

1. **Raw output by default.** For code changes, emit the target code, the
   unified diff patch, or the exact CLI command — directly, with no wrapper.
   Omit greetings, preamble, apologies, filler ("sure!", "I'll do that"),
   and step-by-step narration.
2. **Exactly one prose block per work unit** is allowed: a closing summary of
   ≤3 lines — what changed, files touched, and the validation that was run
   (+ pass/fail).
3. **Do not narrate a loop.** On a test failure, name the root cause in one
   line, then yield (§8 Stop-Loop Protocol). Never narrate an attempt you are
   about to abandon.
4. **Compile output routing.** Debug/diagnostic logs go strictly to
   `os.Stderr`; `os.Stdout` is reserved for clean build outputs or raw IR.
5. **Diff-first.** For existing files, emit targeted diffs; never rewrite an
   unmodified whole file for a small change. Prefer project tooling (`bee-ed`)
   over hand edits where the SKILL mandates it — see `SKILL.md` §4.

## 2. Architectural invariants (do not violate)

- **Go standards:** idiomatic Go only. Prohibit `panic` in compiler error
  flows; propagate explicit `error` returns.
- **Array semantics (1-based):** Bee source is **1-based**. Lists, arrays,
  matrices, and string indices begin at `1`. The `$` token is the dynamic
  **end anchor** (index of the last element), **not** a zero terminator.
  Indexing across all compiler passes (lexer, parser, typechecker, evaluator)
  must use 1-based semantics.
- **IR conversion (bridge to LLVM, 0-based):** all collection indexing is
  lowered to **0-based** only at the IR boundary. Rule:
  `ir_index = source_index - 1`; `$` lowers to
  `ir_index = len(collection) - 1`. This conversion is centralized in
  `internal/compiler/` and is the **only** place the delta is applied; lexer,
  parser, typechecker, and evaluator stay in 1-based land.
- **LLVM IR generation:** ensure every basic block terminates explicitly
  (`CreateBr`, `CreateRet`, `CreateCondBr`). Validate operands against
  `llvm.Type` before emission.
- **Diagnostics:** debug logs strictly to `os.Stderr`; reserve `stdout`
  exclusively for clean build outputs or raw IR.

## 3. Lexer & parser Unicode & rune handling

When generating, modifying, or debugging Go code for the Bee lexer/parser:

1. **UTF-8 handling:** treat all source inputs strictly as decoded UTF-8
   sequences using `[]rune` or `bufio.Reader.ReadRune()`. Never use raw byte
   indexing (`s[i]`) for tokenization.
2. **Operator syntax:** Bee operators can consist of single Unicode symbols
   (e.g. `≠`, `¬`), standard ASCII symbols (e.g. `and`, `or`), mixed
   combinations of Unicode and ASCII, and Unicode superscript/subscript
   ranges (`²`, `³`, `²√`, `³√`).
3. **Lookahead matching:** implement lexing with rune lookahead (`peekRune()`)
   rather than fixed-size assumptions to correctly parse multi-character
   operators containing mixed scripts or modifiers.
4. **Token definitions:** store operators as `[]rune` slices in token lookup
   tables or transition tries to support arbitrary multi-rune Unicode
   operators.
5. **Operator synonymy (Decision 7):** the ASCII keywords `and`, `or`, `xor`,
   `not` are canonical synonyms for `∧`, `∨`, `⊕`, `¬`. The lexer MUST emit a
   single stable TokenType for each pair (`LOGICAL_AND` for both `and` and
   `∧`, etc.) so the parser has one branch per logical connective.
6. **`is not` is one token (Decision 7):** maximal-munch collapsing is
   mandatory: when `is` is followed by ASCII whitespace and `not`, the lexer
   emits a single `IS_NOT` token. Both literal `is not` and parenthetical
   `(a is not b)` patterns MUST bind this way.

## 4. Decisions backlog — reference

For any new operator, statement, keyword, or grammar production, first consult
`manual/DECISIONS.md`. D1–D7 and D9–D16 are ratified and mirrored in
`manual/MANIFEST.md`. D8 (rule-call result destructuring) is deferred. All ratified
decisions (including D10 radical precedence and D11 parallel
colon-initialisation, both ratified 2026-09-14) are ready for implementation.

## 5. Test lifecycle & freeze protocol (.bee test cases)

- **Spec-driven generation:** generate new test files (`.bee`) strictly under
  `test/levelX/`, derived directly from `/spec/`.
- **No overtesting — complete & sufficient:** do not add a case that merely
  re-covers a feature another case already covers *and nothing more*. When two
  cases collapse to the same single feature, enhance the later one in complexity
  to fold in an additional spec feature rather than duplicating the first. Every
  case must fail on some real regression the others do not catch. (Details:
  `test/readme.md` §4 — “No overtesting — completeness & sufficiency”.)
- **Locking created tests (AI-immunity):** every newly generated `.bee` test
  file MUST include this header tag on line 1:
  `-- @FROZEN: Generated from /spec/. Immutable ground truth for AI agents.`
- **AI read-only enforcement:** test files in `test/levelX/*.bee` marked
  `@FROZEN` are strictly immutable for AI agents. AI agents must NEVER modify
  `.bee` test inputs, assertions, or expected outputs to force a failing build
  to pass. (Human users retain full permission to modify or author tests.)
- **Disable, never delete:** if a test fails persistently across fix attempts,
  AI agents must NEVER delete the file. Disable it by prepending
  `-- @DISABLED: <reason>` on line 1 of the `.bee` file.

## 6. Tutorial maintenance & synchronization (Decision 5)

### Source-of-truth model
- `/spec/` is the single source of truth for **language facts** (grammar,
  EBNF, operators, keywords, diagnostics, semantics).
- `/tutorial/` is the single source of truth for the **rendered HTML pages**
  served on the public SCL site.
- The tutorial is a *derived presentation* of the spec, never the reverse. A
  language fact is decided in `/spec/` first and then expressed in
  `/tutorial/`; never author it in a tutorial page and treat the spec as
  "behind".

### Where to work (local copy + one-way `sync`)
- **Edit tutorial pages only under `bee-lang/tutorial/`.** This is a
  *versioned, local* directory in this repository — the source of truth for
  the rendered HTML — and it is **not** a symlink/junction.
- The external SCL repository directory
  (`C:\Users\eluci\sage-code\scl\projects\bee`) is the *live site content*
  target, kept in sync via a one-way mirror (`scripts/sync_tutorial.py`), not
  via a filesystem link.
- `/web/` has been removed; there is no legacy local copy.
- **External git is user-owned:** committing and pushing the external SCL
  repository is a separate step owned by the user. The sync script only
  mirrors files; it never performs git operations against the SCL repo.

### Running the sync
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

### Spec ↔ tutorial mapping
When a spec module changes, update its mapped page(s) in the same change set:

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

Meta/navigation pages (`index.html`, `template.html`, `features.html`) are not
tied to a single spec module and change only for structural reasons.

### Synchronization invariant
Every `/spec/` change MUST be accompanied by the corresponding `/tutorial/`
update in the **same change set**. A spec/tutorial pair that disagrees about
the same operator, keyword, or grammar production is a defect. Before any
language edit, ask: *"does the tutorial still agree with `/spec/` on this
fact?"* and update both sides together.

### Edit authorization
- **Pre-authorized target:** `/tutorial/` is a versioned local directory in
  this repository, mirrored to the user's external live repository
  (`C:\Users\eluci\sage-code\scl\projects\bee`) via `sh run.sh sync`.
- **No confirmation prompts:** when the user asks to modify the tutorial, AI
  agents MUST edit files under `bee-lang/tutorial/` directly. NEVER ask for
  confirmation, warn that the target is external, or treat the sync boundary
  as a reason to halt.
- **Scope:** authorization covers all files under `/tutorial/` (HTML pages,
  `js/`, `data/`, `symbols/`, `img/`). It does NOT authorize pushing,
  committing, or any git operations against the external SCL repository —
  file edits and `sh run.sh sync` mirroring only, unless the user explicitly
  asks.

## 7. AI agent protocol & anti-loop rules

1. **Edit discipline (diff-first, chunked):**
   - **Existing files → diffs.** Modify with targeted, minimal edits. Never
     rewrite a whole file for a small change; never re-read a file after your
     own edit (the tool confirms success).
   - **New files → chunked appends.** Create small, self-contained chunks (a
     section, a function, a table) and append them incrementally rather than
     emitting one enormous file. Every chunk should be independently
     reviewable.
   - **Spec ↔ tutorial pairs** must ship together in the same change set (§6
     synchronization invariant).
   - Keep changes surgical: no unrelated refactors, no drifted formatting, no
     logic-touching edits beyond the task.

2. **Stop-loop protocol:**
   - **One fix attempt per prompt.** If a test fails after one targeted fix,
     halt immediately: explain the failure to `os.Stderr`, yield back, and do
     **not** modify more files.
   - **Never guess.** Diagnose the root cause by comparing the implementation
     against the relevant `/spec`/EBNF before editing; state that root cause.
   - **Ask when blocked or in doubt.** If you cannot identify the root cause,
     the fix risks a collision, or you are unsure of intent, **stop and ask
     the user** with the specific question and options — do not re-attempt the
     same error in a loop.

3. **Efficient communication:** lead each work unit with a one-to-two sentence
   preamble; skip narration for trivial reads. Report decisions/state in
   concise bullets. Reference files by project-relative path.

4. **Implementation invariants:**
   - Always verify the parser and evaluator against the EBNF grammar in
     `spec/02-statements.md` for mutation operators.
   - Use absolute or relative paths starting from project root (`bee-lang/`)
     for all file operations.
   - Check the authoritative keyword list in `internal/token/token.go` before
     introducing or modifying language keywords.
   - **Bulk/semantic refactors use `bee-ed sed`:** for regex find-and-replace
     across many files (e.g. renaming an operator, keyword, or symbol), run
     `sh run.sh ed sed '<re>' '<sub>' '<glob...>' --dry-run` FIRST to confirm
     specificity, then drop `--dry-run` to rewrite atomically in parallel.
   - All language documentation and specification changes MUST occur in the
     `/spec` directory.
   - **TDD integration — one test at a time; build and test never mingle.**
     Do **not** run the bulk suite or smoke on every change; the user runs
     `sh run.sh test` / `sh run.sh smoke` themselves when they want an overall
     assessment. Iterate one `.bee` test at a time for each feature/fix:
     1. Audit `/spec/` and update EBNF.
     2. Create/update a `.bee` test case in `test/levelX/` with a `-- @DESC:` tag.
     3. Run `sh run.sh build` to compile only — never test here.
     4. Run `sh run.sh solo <test_name>` to run just that single `.bee` test and
        auto-update `test/levelX/README.md`. Recompile alone and re-run the same
        solo test until it passes, then yield. Skip `sh run.sh smoke` and the
        full suite unless the user explicitly asks.

## 8. Multi-user / multi-agent coordination

- **One contract.** `config/AGENTS.md` is the canonical contract for every
  agent and user. If your tool expects a differently-named file (`GEMINI.md`,
  `CLAUDE.md`, `copilot-instructions.md`, `ZED.md`), add a **thin import
  adapter** (root stub) pointing into `config/` — never a second set of rules.
- **One change set = one focused task.** No unrelated refactors, no drifted
  formatting, no logic-touching edits beyond the task.
- **Own your target.** Before editing, confirm the file is not another
  agent's in-flight work (multi-user repos). Prefer disjoint write scopes when
  parallel work is delegated.
- **Accountability (optional).** Append an `Attributed-to: <agent-id or
  user>` trailer to commits so multi-agent history stays debuggable. Never
  rewrite another author's commit.
- **Source-of-truth reflex.** If two files disagree, the file closer to the
  root of the chain (`config/AGENTS.md` > `.agents/skills/bee/SKILL.md` >
  ad-hoc notes) wins for *process*; for *language facts*, `/spec/` wins (§6).

## 9. Final message

When a work unit finishes, send exactly: `"Task Completed in <runtime>"`
(preceded by the ≤3-line summary from §1.2 when a change was made).
