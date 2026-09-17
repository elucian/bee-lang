# Copilot — Agent Instructions (adapter)

This is the Copilot-specific adapter in `config/`. Read `config/AGENTS.md` — it
is the canonical, single source of truth for AI-agent behavior in this
repository. The `.github/copilot-instructions.md` discovery stub (the path
Copilot requires) hands off to this file and to `config/AGENTS.md`. This
adapter adds nothing and must never fork rules — keep it in sync (empty here
by design).

> Prompt-cache signature: prepend the current value from `manual/CONTEXT.md`
> (`python scripts/context_signature.py`) before any task prompt.
