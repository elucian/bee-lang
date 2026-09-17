# Developer Manual — Index

The **developer manual** is the human-facing operational reference for the Bee
compiler repository. It complements the machine-facing agent contract in
`config/AGENTS.md`: the manual explains *why* and *how* to work here, while
`config/AGENTS.md` is the normative rule set every agent must obey.

> Read `config/AGENTS.md` first if you are an AI agent. Read this manual for
> project-wide orientation, history, and procedures.

## Contents

| Document | Purpose |
| :--- | :--- |
| `DEVELOPER.md` | The developer instruction manual — orientation, workflow, and procedures. |
| `MANIFEST.md` | Phase/roadmap state, user-locked decisions, and changelog deltas. |
| `DECISIONS.md` | Normative decision backlog (D1–D16) — the design history. |
| `CONTEXT.md` | The prompt-cache signature spec and the current signature value. |

## Related top-level directories

- `config/` — all agent configuration files and adapters (including this
  manual's sibling `AGENTS.md` rules).
- `.agents/skills/bee/` — the Zed skill implementing repo rules.
- `spec/` — single source of truth for language facts.
- `tutorial/` — derived HTML presentation of the spec (rendered site pages).
- `tracking/` — unified project-state hub (`issues/`, `solutions/`, `todo/`).
- `registry/` — diagnostic-code registry (`diagnostics.json`).
