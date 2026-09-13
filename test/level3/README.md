# Test Level 3: Rules and Functions

This level covers rules, contracts, and function definitions as defined in:
- `spec/03-rules.md`
- `spec/07-functions.md`

## Test Coverage

| CASE     | DESCRIPTION                      | STATUS     |
| -------- | -------------------------------- | ---------- |
| T0301    | Rule anatomy no params — mini... | PASS       |
| T0302    | Rule anatomy with params — ru... | FAIL       |
| T0303    | Assert precondition pass — as... | PASS       |
| T0304    | Assert precondition fail — as... | PASS       |
| T0305    | Expect postcondition pass — e... | PASS       |
| T0306    | Expect postcondition fail — e... | PASS       |
| T0307    | Rule docstrings — triple-dash... | PASS       |
| T0308    | Scoping parameter shadowing —... | PASS       |
| T0309    | Scoping variable locality — r... | PASS       |
| T0310    | Closures state capture — rule... | FAIL       |
| T0311    | Early return with exit — exit... | PASS       |
| T0312    | Nested rule calls — one rule ... | FAIL       |
| T0313    | Recursion — factorial via rec... | FAIL       |
| T0314    | Multi-result deconstruction —... | FAIL       |
| T0315    | Lambda expression — explicit ... | FAIL       |
| T0316    | Lambda in pipeline — inline s... | FAIL       |
| T0317    | Compose operator — f ∘ g comp... | FAIL       |
| T0318    | Pipe operator — value |> f |>... | FAIL       |
| T0319    | Partial application — ? place... | FAIL       |
| T0320    | Rule contract verification wi... | PASS       |
| T0321    | Rule contract assert failure ... | PASS       |
