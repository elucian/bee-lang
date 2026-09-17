# Zed Coding Agent — Instructions (adapter)

Tool-specific adapter for the **Zed** editor's built-in coding agent. It is a
thin, discovery-only adapter: the canonical, single source of truth for
repository rules is `config/AGENTS.md`. This file adds nothing and must never
fork rules — keep it in sync (empty here by design).

## Discovery chain

- Zed auto-attaches `AGENTS.md` at the repository root. That root file is a
  stub that resolves here: **read `config/AGENTS.md`** before touching the
  codebase.
- The repo skill (`.agents/skills/bee/SKILL.md`) is defined in Zed's Skill
  system; it references `config/AGENTS.md`. If config/AGENTS.md and the skill
  disagree, `config/AGENTS.md` wins (§8 source-of-truth reflex).

## Zed-specific habits (Zed only; not repo canon)

1. **Raw output by default.** For code changes emit target code, unified diff
   patches, or the exact command — directly, no wrapper, no preamble, no
   filler. Zed's panel narrows the readable budget; lean output is a feature.
2. **One prose block per work unit** — a closing summary of ≤3 lines: what
   changed, files touched, validation run and pass/fail. Then finish with the
   canonical sign-off from `config/AGENTS.md` §9.
3. **Diff-first, chunked edits.** Prefer targeted diffs over whole-file
   rewrites; append small reviewable chunks for new files. Re-read nothing
   after your own edit — the tool confirms success.
4. **Use Zed's native tools** (grep, search, terminal) and the repo's `bee-ed`
   via `sh run.sh ed ...` when modifying Markdown/HTML/source files. Follow the
   `bee-ed` pre/post `sh run.sh ed balance` validation.
5. **Never bypass bee-ed for repo file changes** in HTML/Markdown where the
   SKILL mandates it; manual edits are the single largest source of markup
   defects.

## Prompt-cache signature

Prepend the current signature (from `manual/CONTEXT.md`;
`python scripts/context_signature.py`) to the task prompt before any project
context whenever a task-scoped session starts.
