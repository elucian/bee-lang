# GLM — Agent Instructions (adapter)

This is the GLM-specific adapter in `config/`. Read `config/AGENTS.md` — it
is the canonical, single source of truth for AI-agent behavior in this
repository. This adapter adds no rules and must never fork rules; the GLM
agent (the Zed coding agent powered by GLM) resolves here for its onboarding
fast path only. Everything normative lives in `config/AGENTS.md`,
`.agents/skills/bee/SKILL.md`, and `/spec/`.

> Prompt-cache signature: prepend the current value from `manual/CONTEXT.md`
> (`python scripts/context_signature.py`) before any task prompt.

## Onboarding fast path (navigation only — not rules)

Read in this order at session start. For *process*, `config/AGENTS.md` >
`.agents/skills/bee/SKILL.md` > ad-hoc notes; for *language facts*, `/spec/`
wins.

1. `config/AGENTS.md` — canonical contract: output discipline, architectural
   invariants, Unicode/rune lexing rules, test lifecycle, anti-loop protocol.
2. `.agents/skills/bee/SKILL.md` — Zed skill: lazy-loading context budget,
   `bee-ed` mandate, context-signature emission.
3. `manual/MANIFEST.md` — roadmap phases, user-locked decisions, audit log.
4. `manual/DECISIONS.md` — D0–D16 backlog (D8 deferred; D14 implementation
   deferred).
5. `manual/CONTEXT.md` — prompt-cache signature spec and procedure.
6. `manual/DEVELOPER.md` — human-facing procedures: build, test pipeline,
   tutorial sync, `bee-ed` reference.
7. `spec/` — single source of truth for grammar, EBNF, operators, keywords,
   diagnostics, semantics.

## GLM session-start habits (GLM only; not repo canon)

1. Compute the signature fresh (`python scripts/context_signature.py`) and
   prepend it to the task prompt — never hand-copy the recorded value.
2. Skim `manual/MANIFEST.md` before touching `internal/` — the Anti-Loop
   Gate blocks edits until the governing `/spec` sections are harmonized.
3. Route every Markdown/HTML edit through `sh run.sh ed ...`: `balance`
   before and after, unique anchors for `edit`, multi-line chunks spooled
   via `.temp/`, `--dry-run` before `sed`.
4. One targeted fix per prompt, then yield; single-test runs only
   (`sh run.sh solo <test>`) — never bulk passes without explicit approval.
5. Close every work unit with the ≤3-line summary and the
   `config/AGENTS.md` §9 sign-off.
