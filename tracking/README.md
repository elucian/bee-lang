# Project-State Hub (`tracking/`)

The single home for Bee's live project state. Every design problem, its
proposed resolution, and its open technical debt lives under this one folder so
agents and humans share one place to look — instead of scattered `issues/`,
`solution/`, and `todo/` trees.

| Subfolder | Contents |
| :--- | :--- |
| `issues/` | Logged problems (architecture gaps, spec ambiguities, unimplemented features). Named `NN-slug.md`. |
| `solutions/` | Proposed / ratified resolution strategies. Named `NN-slug.md`. |
| `todo/` | Open work and technical debt (`TECH_DEBT.md`), plus backlogs (`TUTORIAL_TODO.md`). |

## Workflow (spec-driven TDD)

1. **Architectural gap →** write the problem in `solutions/` and log it in
   `todo/TECH_DEBT.md`.
2. **Specification →** formalize in `/spec/` (single source of truth).
3. **Test-first →** add a failing `.bee` test under `test/levelX/`.
4. **Implement →** make the compiler satisfy the spec and pass the test.
5. **Freeze →** mark new tests `-- @FROZEN: ...`.

Issue/solution pairs share a number (`issues/16-rule-tuple-destructure.md` ⇄
`solutions/16-rule-tuple-destructure.md`). Keep the number aligned when
renumbering.

> Note: `registry/` is deliberately **not** under `tracking/` — it is a runtime
> data source for code generation (`scripts/gen_diagnostics_page.py`), not
> project-state documentation. See `registry/README.md`.
