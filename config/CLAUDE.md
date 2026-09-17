# Claude — Agent Instructions (adapter)

This is the Claude-specific adapter in `config/`. Read `config/AGENTS.md` — it
is the canonical, single source of truth for AI-agent behavior in this
repository. The repository-root `CLAUDE.md` is a discovery stub that hands off
to this file and to `config/AGENTS.md`. This adapter adds nothing and must
never fork rules — keep it in sync (empty here by design).

> Prompt-cache signature: prepend the current value from `manual/CONTEXT.md`
> (`python scripts/context_signature.py`) before any task prompt.
