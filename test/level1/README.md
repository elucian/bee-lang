# Test Level 1: Lexical and Type Foundation

This level covers the fundamental lexical structure, tokenization, and type inference as defined in:
- `spec/01-lexical-structure.md`
- `spec/05-types.md`

## Test Coverage

| CASE     | DESCRIPTION                      | STATUS     |
| -------- | -------------------------------- | ---------- |
| T0138    | T0138 Decision 13: postfix-st... | PASS       |
| T0137    | T0137 Decision 13: postfix-st... | PASS       |
| T0136    | T0136 Decision 13: postfix-st... | PASS       |
| T0134    | T0134 Decision 13: postfix-st... | PASS       |
| T0135    | T0135 Decision 13: fully-excl... | PASS       |
| T0133    | T0133 Decision 13: fully excl... | PASS       |
| T0132    | T0132 Decision 13: left-exclu... | PASS       |
| T0131    | T0131 Decision 13: right-excl... | PASS       |
| T0101    | T0101 Essential arithmetic op... | PASS       |
| T0102    | T0102 Relational and comparis... | FAIL       |
| T0103    | TBD                              | FAIL       |
| T0104    | T0104 Exponentiation and powe... | PASS       |
| T0105    | T0105 Range membership with i... | PASS       |
| T0106    | T0106 Test assignment, equali... | PASS       |
| T0107    | T0107 Square root, cube root,... | FAIL       |
| T0108    | T0108 Compound arithmetic ass... | PASS       |
| T0115    | T0115 Comprehensive print and... | PASS       |
| T0116    | T0116 Line and block comments... | PASS       |
| T0117    | T0117 Constant declaration wi... | PASS       |
| T0118    | T0118 Variable declaration wi... | PASS       |
| T0119    | file: ending_early.bee           | FAIL       |
| T0120    | file: hello_world.bee            | PASS       |
| T0121    | file: initial_values.bee         | FAIL       |
| T0122    | file: parallel_assignment.bee    | FAIL       |
| T0123    | file: print_write.bee            | PASS       |
| T0124    | T0124 In-place addition mutat... | PASS       |
| T0125    | T0125 Chained arithmetic muta... | PASS       |
| T0128    | @FROZEN: Generated from /spec... | PASS       |
| T0129    | @FROZEN: Generated from /spec... | PASS       |
| T0130    | T0130 Equality and identity r... | PASS       |
