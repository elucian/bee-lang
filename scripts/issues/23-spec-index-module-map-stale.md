# Issue: `spec/readme.md` Module Map Is Stale

The `Module Map` section of `spec/readme.md` does not match the modules that
actually exist in `/spec/`, which misleads the spec ↔ tutorial mapping and any
tooling that reads the index as a source of truth.

## Impact

- Lists a non-existent module (`04-functions.md`) and omits a real one
  (`07-functions.md`), so the index cannot be relied on to derive the
  tutorial mapping.
- `GEMINI.md` §6 already uses the correct names (`04-structure.md`,
  `07-functions.md`); the spec index disagrees with it, which is itself the
  kind of documentation defect the synchronization invariant is meant to
  prevent.
- Confuses the description of "functions/lambda" (`04-functions.md` is
  described as Lambda) with the actual `04-structure.md` module.

## Requirements

- Correct the entry:
  - `04-functions.md` → `04-structure.md` (its real identity).
  - Add the missing `07-functions.md` entry.
- Reconcile `spec/readme.md` with `GEMINI.md` §6 and the actual files in
  `spec/` so all three agree.

Actual `/spec/` modules: `00-memory-model.md`, `01-lexical-structure.md`,
`02-statements.md`, `03-rules.md`, `04-structure.md`, `05-types.md`,
`06-objects.md`, `07-functions.md`, `10-collections.md`, `11-processing.md`,
`12-concurrency.md`, `13-graphics.md`, `14-library.md`.
