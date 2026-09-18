# Level 0: Compiler Self-Bootstrap & Smoke

Boot-level tests that validate the Bee compiler can lex, parse, evaluate, and print outputs without crashing. These tests intentionally cover the minimum surface required to exercise the print/io grammar path and entry-rule header (`rule main: ... return;`).

| CASE  | AI  | STATUS | DESCRIPTION               |
| ----- | --- | ------ | ------------------------- |
| T0001 | Yes | PASS   | Hello World Test          |
| T0002 | Yes | PASS   | Large list of numbers     |
| T0003 | Yes | PASS   | Multiple arguments        |
| T0004 | Yes | PASS   | Test print with custom... |
| T0005 | No  | PASS   | Negative test — compil... |
| smoke | Yes | PASS   | Smoke test validation     |
