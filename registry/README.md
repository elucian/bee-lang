# Bee Diagnostic Code Registry

`registry/diagnostics.json` is the **single source of truth** for every Bee
diagnostic code (`E` hard error / `W` soft warning) referenced anywhere in
`/spec` and `/tutorial`.

## When to register

- **Register first, then use.** Never inject a diagnostic table into a `/spec`
  module or a `/tutorial` page before the codes are present here.
- **Collision policy.** If a proposed code already exists in this registry, the
  newer / less-canonical use must be **bumped** to the next free code in its
  module block and re-registered. Never silently reuse an occupied code with a
  different meaning.

## Block allocation

| Module block | Owner |
| :--- | :--- |
| `E00xx` | Global / deprecation codes (`E0009`, `E0010`, `E0011`), cross-module |
| `E01xx` | `spec/01-lexical-structure.md` |
| `E02xx` | `spec/02-statements.md` |
| `E03xx` | `spec/03-rules.md` |
| `E04xx` | `spec/04-structure.md` |
| `E05xx` | `spec/05-types.md` |
| `E06xx` | `spec/06-objects.md` |
| `E07xx` | `spec/07-functions.md` |
| `E08xx` | `spec/00-memory-model.md` *(bumped from E04xx on 2026-09-14)* |
| `E10xx` | `spec/10-collections.md` |
| `E11xx` | `spec/11-processing.md` |
| `E12xx` | `spec/12-concurrency.md` |
| `E13xx` | `spec/13-graphics.md` |
| `E14xx` | `spec/14-library.md` |

Blocks `E09xx` remain free for future modules.

## Known semantic twins (not collisions)

- `E0406` (`spec/04-structure.md`) and `E1201` (`spec/12-concurrency.md`) both
  describe *scope exited with pending `begin` threads prior to `wait`*. Distinct
  codes, so not a registry collision; the concurrency `E1201` is the canonical
  working code, `E0406` is retained for module-level structure completeness.

## Reconciliation history

- **2026-09-14** — Created registry. Reassigned `spec/00-memory-model.md` from
  `E0401`–`E0404` (colliding with `spec/04-structure.md`) to `E0801`–`E0804`.
  Updated `spec/00` and `tutorial/memory.html`.
