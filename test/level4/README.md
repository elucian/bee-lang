# Test Level 4: Collections and Pipelines

This level covers collections and pipelines as defined in:
- `spec/10-collections.md`
- `spec/11-processing.md`

## Test Coverage
| CASE     | DESCRIPTION                      | AI    | STATUS     |
| -------- | -------------------------------- | ----- | ---------- |
| T0401    | Complex Data Processing test ... | No     | PASS       |
| T0402    | file: contract_argument_mutat... | No     | PASS       |
| T0403    | contract deposit with expecta... | No     | PASS       |
| T0404    | contract grow on a collection    | No     | PASS       |
| T0405    | contract write-through argument  | No     | PASS       |
| T0406    | contract clamp with named res... | No     | PASS       |
| T0407    | contract postcondition with r... | No     | PASS       |
| T0408    | contract precondition with pa... | No     | PASS       |
| T0409    | contract store item in collec... | No     | PASS       |
| T0410    | file: arrays.bee                 | No     | PASS       |
| T0411    | array traversal with done ter... | No     | FAIL       |
| T0412    | bubble sort with done termina... | No     | PASS       |
| T0413    | file: builders.bee               | Yes    | SKIP       |
| T0414    | file: concatenation.bee          | No     | PASS       |
| T0415    | growing a list with <+ and +>    | No     | PASS       |
| T0416    | file: lists.bee                  | No     | PASS       |
| T0417    | file: maps.bee                   | Yes    | SKIP       |
| T0418    | file: map_builder.bee            | Yes    | SKIP       |
| T0419    | file: quantifiers.bee            | No     | PASS       |
| T0420    | file: sets_sorted.bee            | No     | PASS       |
| T0421    | file: set_add_remove.bee         | No     | PASS       |
| T0422    | file: set_algebra.bee            | No     | PASS       |
| T0423    | file: set_algebra_unicode.bee    | Yes    | SKIP       |
| T0424    | file: sharing_copying.bee        | No     | PASS       |
| T0425    | file: sized_array.bee            | No     | FAIL       |
| T0426    | slice extraction with $ anchor   | No     | PASS       |
| T0427    | walking a map with done termi... | Yes    | SKIP       |
| T0430    | collection assertions for arr... | No     | PASS       |
| T0431    | dollar-anchor arithmetic and ... | No     | PASS       |
| T0432    | raw negative index a[-1] is r... | No     | PASS       |
| T0433    | array slicing with $ anchor a... | No     | PASS       |
| T0434    | range endpoint variants with ... | No     | PASS       |
| T0435    | set builder with filter over ... | No     | PASS       |
| T0436    | array and hash-map builders o... | No     | PASS       |
| T0437    | collection iteration with $ a... | No     | PASS       |
| T0438    | list concatenation and append... | No     | FAIL       |
| T0439    | array decomposition, spreadin... | No     | PASS       |
| T0440    | matrix row/column slice mutat... | No     | PASS       |
| T0441    | set algebra intersection, uni... | No     | PASS       |
| T0442    | quantifiers forall/exists ove... | No     | PASS       |
| T0443    | collection casting between ar... | No     | PASS       |
| T0444    | list as queue FIFO using appe... | No     | PASS       |
