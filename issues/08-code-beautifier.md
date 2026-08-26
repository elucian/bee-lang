# Issue 08: Compiler-Native Code Beautifier (-b / --beautify)

- **Status:** Open
- **Impact:** High (Eliminates AI token consumption and human time spent formatting Bee source code).
- **Description:** AI agents and human developers spend significant effort manually formatting and aligning Bee source code. Large language models waste context tokens re-printing full files to fix indentation or spacing.
- **Goal:** Integrate a deterministic, high-performance source code beautifier into the `bee` compiler toolchain accessible via the `-b` or `--beautify` CLI flag.

## Key Functional Requirements
1. **In-Place File Modification:** Running `bee -b <file.bee>` parses and formats the file, writing back to disk if changes are detected.
2. **2-Space Indentation Enforcement:** Strictly enforces 2-space offset per block nesting level, aligning block terminators (`done`, `repeat`, `return`) to 0 relative indentation.
3. **Trailing Comment Alignment:** Aligns end-of-line comments (`-- ...`) at consistent column offsets per block.
4. **Automatic Syntax Auto-Fixes (Deterministic Corrections):**
   - **Implicit Multiplication:** Automatically converts implicit multiplication syntax like `2(a+b)` or `n(x+y)` to explicit `2 * (a + b)` or `n * (x + y)`.
   - **Operator Spacing:** Normalizes spacing around assignment (`:=`, `::`), arithmetic (`+`, `-`, `*`, `/`), relational (`=`, `≠`, `∈`), and range operators (`..`).
   - **Block Header Formatting:** Normalizes colons `:` and keyword alignment for `rule`, `if`, `while`, `for`, `cycle`, `match`, and `trial` (including `try`, `miss`, `final` section alignment).
5. **Specification Reference:** `solution/12-code-beautifier.md`
