# Bee-Lang Context Signature

The **context signature** is a unique, deterministic identifier for the
current Bee design context. It is sent to the LLM **before a task is created**
so the prompt cache can recognize an unchanged context and reuse it — raising
the cache-hit rate and lowering the cost (and latency) of rebuilding context.

## Why

An LLM's prompt cache keys on the prompt **prefix**. When the same large,
stable context (specs, decisions, manifest) is prepended to many tasks, the
cache serves the identical prefix from cache instead of recomputing it. But
the cache can only do this if the agent can *prove* the context is unchanged.
The signature is that proof: a fingerprint that changes if and only if a
normative fact changes.

## Format

```
BEE-LANG-V<phase>-D<epoch>-<hash8>
```

| Part         | Meaning                                                        | Source |
| :---         | :---                                                           | :--- |
| `BEE-LANG`   | Fixed project id                                               | — |
| `V<phase>`   | Current Manifest Phase/version (e.g. `8.7`)                    | `manual/MANIFEST.md` |
| `D<epoch>`   | Highest authored decision id (e.g. `15`)                       | `manual/DECISIONS.md` |
| `<hash8>`    | First 8 hex of SHA-256 over the canonical context files        | see below |

**Current value:** `BEE-LANG-V8.8-D16-4897df02`

> The `<hash8>` changes whenever the content of the context bundle changes,
> even if the human-readable prefix (`V8.7-D15`) is unchanged. The prefix is
> for humans; the hash is the authoritative cache key.

## Canonical context files (hashed, in fixed order)

1. `config/AGENTS.md` (the canonical agent-rule contract for all agents/users)
2. `manual/MANIFEST.md`
3. `manual/DECISIONS.md`
4. All `spec/*.md` (sorted, excluding `readme.md`)

These are the *normative facts* that define the context an agent must hold.
Per-agent adapters (`config/GEMINI.md`, `config/CLAUDE.md`,
`config/copilot-instructions.md`, `config/ZED.md`) and the root discovery
stubs are excluded by design: they are thin imports of `config/AGENTS.md` and
hold no normative facts, so editing the adapter text must not invalidate the
cache.
The tutorial is derived output and is **not** part of the signature (editing
a tutorial page does not change the design facts; a cache should remain valid).

## Procedure

1. **Compute** the signature (never hand-copy it):
   ```sh
   python scripts/context_signature.py
   ```
2. **Emit it first.** Prepend the exact returned string to the task prompt,
   on its own line, before any other project context is sent. Example:
   ```
   BEE-LANG-V8.7-D15-c58f3dc8

   <task description>
   ```
3. **Recompute after any normative change.** Re-run the script whenever a
   ratified decision is added/reversed, the manifest phase/version advances,
   or any spec module is edited. A new hash correctly signals a cache miss.
4. **Do not mix sync state into the signature.** Tutorial sync state
   (`SYNC` vs `LOCAL`) is orthogonal to the design facts and is tracked
   separately (it must never invalidate the design-context cache).

## Integration

- `scripts/context_signature.py` — deterministic generator (single source of
  truth for the hash).
- `.agents/skills/bee/SKILL.md` §5 — instructs the agent to compute and
  prepend the signature at session start; `config/AGENTS.md` §How-to-Use reinforces
the same.
- This file (`manual/CONTEXT.md`) — the human-readable specification and the
  record of the current value.

## Notes / assumptions

- Determinism depends on the canonical file list and order staying stable. If
  a new normative file is introduced (e.g. a new decisions source), add it to
  `FILES` in `scripts/context_signature.py` and note the change here.
- The signature relies on `manual/DECISIONS.md` being internally consistent.
  A stale/contradictory decision index (e.g. D10/D11 status drift, missing
  D15) weakens the human-readable `D<epoch>` but not the hash — fix the index
  separately (see `tracking/todo/TUTORIAL_TODO.md` §3.1).
- The signature depends on the canonical agent-rule file being
  `config/AGENTS.md`. If a new normative file is ever promoted (or the
  canonical rule file is renamed), it must be reflected in `FILES` of
  `scripts/context_signature.py` and in the list above, together in the same
  change set.
