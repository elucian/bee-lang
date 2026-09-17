# Issue: Tutorial Code Examples Still Use Deprecated `==` Equality

Several `tutorial/` HTML pages show `==` as the value-equality operator in code
examples. This contradicts Decision 12 (2026-09-13), which made `=` the
canonical **value-equality** operator and demoted `==` to a deprecated form
that emits `E0010 deprecated-symbol: '==' — use '='`.

`spec/01-lexical-structure.md` §3.3 states the canonical equality operator is
`=`, and §3.3 deprecation note lists `==` as deprecated.

## Impact

- The public tutorial teaches authors to write `==`, the deprecated form.
- Spec and tutorial disagree on the same operator, violating the
  synchronization invariant in `config/AGENTS.md` §6.
- New language adopters copy examples into Bee source and immediately trigger
  `E0010` deprecation diagnostics.

## Requirements

- Replace `==` with canonical `=` in all tutorial code examples.
- Locate every occurrence (searched 2026-09-14) and update:

| File | Line | Location |
| :--- | :--- | :--- |
| `tutorial/control.html` | 110 | `expect x == 3;` |
| `tutorial/control.html` | 629 | `case $error.code == 200 do` |
| `tutorial/processing.html` | 94 | `print n == k;` |
| `tutorial/rules.html` | 184 | `if (c == 0) do` |
| `tutorial/types.html` | 521–523 | `a == 10`, `a == 11`, `a == 10` |
| `tutorial/types.html` | 682–683 | `1 == 1`, `"s" == "s"` |

- Confirm no `==` remains in `tutorial/*.html` code samples thereafter.
- Re-run `sh run.sh sync --dry-run` to confirm local/SCL parity after the edit.
