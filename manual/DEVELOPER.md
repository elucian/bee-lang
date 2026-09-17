# Bee Compiler — Developer Instruction Manual

This manual is the **human-facing operational reference** for working on the
Bee programming-language compiler. It sits side-by-side with the
machine-facing agent contract (`config/AGENTS.md`). If you read this and you
are an AI agent, the normative rules in `config/AGENTS.md` win on any
conflict; this manual is where the *context and procedure* live.

> Prompt-cache signature: prepend the current value from `manual/CONTEXT.md`
> (`python scripts/context_signature.py`) before any task prompt.

---

## 1. Repository map

| Path | Role |
| :--- | :--- |
| `config/` | Agent configuration files and adapters (`AGENTS.md`, `GEMINI.md`, `CLAUDE.md`, `copilot-instructions.md`, `ZED.md`). |
| `manual/` | Developer manual and normative design docs (`MANIFEST.md`, `DECISIONS.md`, `CONTEXT.md`). |
| `spec/` | **Single source of truth** for language facts: grammar, EBNF, operators, keywords, diagnostics, semantics. |
| `tutorial/` | **Derived** HTML presentation of the spec — mirrored one-way to the live SCL site. |
| `internal/` | Go compiler source (`lexer`, `parser`, `evaluator`, `typechecker`, `compiler`). |
| `test/` | Test suites (`level0`–`level8`, `debt/`), runners, and infrastructure. |
| `scripts/` | Python automation (`context_signature.py`, `sync_tutorial.py`, etc.). |
| `tracking/` | Unified project-state hub: `issues/`, `solutions/`, `todo/` (tech debt, backlog, tutorial tracking). |
| `registry/` | Diagnostic-code registry (`diagnostics.json`, README) — the runtime source of truth for every `E`/`W` code. |
| `run.sh` | Master CLI helper (`build`, `test`, `solo`, `smoke`, `sync`, `ed`, `commit`). |

---

## 2. Environment

- **Go 1.26+** and **LLVM 18+** on `PATH` (`llc --version`).
- Build: `python build.py`; install/PATH once: `sh setup.sh`.
- Git Bash / terminal recommended on Windows.

### Compiler flags
- `-d` / `--debug` — token stream to `stderr`.
- `-e` / `--execute` — in-memory AST evaluator (VM).
- `-c` / `--compile` — parse and validate syntax only.

---

## 3. Core invariants (never violate)

These are enforced by `config/AGENTS.md` §2 and re-stated here for humans.

- **Go standards:** idiomatic Go; **no `panic`** in compiler error flows —
  propagate explicit `error` returns.
- **1-based semantics at source:** Bee source is 1-based. Lists, arrays,
  matrices, and string indices begin at `1`; `$` is the dynamic *end anchor*
  (index of last element), not a zero terminator.
- **0-based only at the IR boundary:** `ir_index = source_index - 1`;
  `$` → `len(collection) - 1`. Centralized in `internal/compiler/` — the only
  place the delta is applied.
- **LLVM IR generation:** every basic block terminates explicitly
  (`CreateBr`, `CreateRet`, `CreateCondBr`); validate operands against
  `llvm.Type` before emission.
- **Diagnostic routing:** debug logs strictly to `os.Stderr`; `os.Stdout` is
  reserved for clean build outputs or raw IR only.
- **Enabled operator taxonomy (Decision 7):** `and`/`∧`, `or`/`∨`, `xor`/`⊕`,
  `not`/`¬` are synonyms; the lexer emits one stable `TokenType` per pair.
- **`is not` is one token:** maximal-munch collapses whitespace-separated
  `is` + `not` into a single `IS_NOT` token.

---

## 4. Development lifecycle (spec-driven TDD)

1. **Architectural gap** → document in `tracking/solutions/` (and log in `tracking/todo/`).
2. **Specification** → update `/spec/` EBNF and rules.
3. **Test-first** → create a failing `.bee` test under `test/levelX/` with a
   `-- @DESC:` tag.
4. **Implementation** → modify the compiler to satisfy spec and pass the test.
5. **Freeze** → mark new tests `-- @FROZEN: ...` (immutable ground truth).

### Test rules (AI-immunity)
- Generated `.bee` files carry `@FROZEN` on line 1 and are **immutable for AI
  agents** (humans may edit).
- Disable, never delete: prepend `-- @DISABLED: <reason>` on line 1.

### Manual loop discipline
- **One fix attempt per prompt.** If a test fails after one targeted fix, stop:
  diagnose the root cause against `/spec`, report it, and yield — do not loop.
- Never guess; ask when blocked or the fix would collide.

---

## 5. Tutorial maintenance & sync

- `/spec/` decides a language fact **first**; `/tutorial/` expresses it
  (never the reverse).
- A spec change MUST ship with the corresponding tutorial update in the
  **same change set** — a spec/tutorial pair that disagrees is a defect.
- Edit only under `bee-lang/tutorial/` (versioned local copy). The external SCL
  site (`C:\Users\eluci\sage-code\scl\projects\bee`) is a one-way mirror target.

| Command | Effect |
| :--- | :--- |
| `sh run.sh sync` | Push local → SCL (publish edits). Default. |
| `sh run.sh sync pull` | Pull SCL → local. |
| `sh run.sh sync --dry-run` | Preview the push without writing. |
| `sh run.sh sync --delete` | Push and remove SCL files absent locally (true mirror). |
| `sh run.sh sync --scl <PATH>` | Override the SCL directory path. |

Underlying helper: `scripts/sync_tutorial.py` (default-safe, never deletes
unless `--delete`, skips already-matched files).

---

## 6. Workflow automation (`run.sh`)

| Command | Action |
| :--- | :--- |
| `sh run.sh build` | Clean + build. |
| `sh run.sh test [level]` | Run the full test pipeline (or a level). |
| `sh run.sh check [level]` | Syntax-check test files. |
| `sh run.sh solo TNNNN` | Run one named test and auto-update its README. |
| `sh run.sh reset [level]` | Re-enable disabled tests. |
| `sh run.sh smoke` | System-wide self-health check. |
| `sh run.sh ed <cmd...>` | Repo file maintenance via `bee-ed`. |
| `sh run.sh sync [...]` | Tutorial mirror (see §5). |
| `sh run.sh commit [msg]` | Stage, synthesize/use message, commit, push. |

### `bee-ed` (mandatory for Markdown/HTML edits)
- `sh run.sh ed apply <patch> <file>` — unified-diff apply (atomic; refuses on
  mismatch).
- `sh run.sh ed edit <file> <old> <new>` — replace a unique substring.
- `sh run.sh ed append <file> [chunk|@chunk.txt]` — append a reviewable chunk.
- `sh run.sh ed balance <file>` — validate tag structure; run **before and
  after** every edit to any HTML/Markdown file.
- `sh run.sh ed sed <pattern> <repl> <glob...> [--dry-run]` — parallel RE2
  regex find-and-replace across a tree; always `--dry-run` first.

---

## 7. Context signature (prompt cache)

The **context signature** fingerprints the normative context bundle so an LLM
prompt cache can reuse a stable prefix across tasks. Spec: `manual/CONTEXT.md`.

- Compute — never hand-copy:
  ```sh
  python scripts/context_signature.py
  ```
- Emit it **first**: prepend the exact returned string to the task prompt, on
  its own line, before any other project context.
- Recompute after any normative change (a ratified decision, a manifest phase
  bump, or any spec edit) — a new hash correctly signals a cache miss.

Canonical hashed files (fixed order): `config/AGENTS.md`,
`manual/MANIFEST.md`, `manual/DECISIONS.md`, then all `spec/*.md` (sorted,
excluding `readme.md`).

---

## 8. Multi-user / multi-agent coordination

- **One contract.** `config/AGENTS.md` is canonical for *every* agent and user.
  Tool-specific names (`GEMINI.md`, `CLAUDE.md`, `copilot-instructions.md`,
  `ZED.md`, and the root `AGENTS.md` stub) are thin discovery adapters that
  point to it — never a second set of rules.
- **One change set = one focused task.** No unrelated refactors, no drifted
  formatting.
- **Own your target.** Before editing, confirm the file is not another agent's
  in-flight work; prefer disjoint write scopes when parallelizing.
- **Accountability (optional).** Append `Attributed-to: <agent-id or user>` to
  commits so multi-agent history stays debuggable; never rewrite another
  author's commit.
- **Source-of-truth reflex.** For *process*: `config/AGENTS.md` >
  `.agents/skills/bee/SKILL.md` > ad-hoc notes. For *language facts*:
  `/spec/` wins.
