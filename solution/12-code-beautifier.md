# Solution 12: Compiler-Native Code Beautifier Engine

## Architectural Overview

The Bee compiler incorporates a native code beautifier subsystem into the frontend pipeline. Activated via `bee -b <file.bee>` or `bee --beautify <file.bee>`, this engine formats, aligns, and auto-corrects Bee source code deterministically with zero runtime or LLM overhead.

```text
Raw Source (.bee) ──► Lexer & Preserving Tokenizer ──► AST & CST Builder
                                                           │
Formatted Source (.bee) ◄── Code Printer Engine ◄── AST Transformer / Auto-Fixer
```

---

## 1. CLI Integration & Execution Policy

- **Command Flags:** `bee -b <file.bee>` or `bee --beautify <file.bee>`.
- **Dry-Run Mode:** `bee -b --dry-run <file.bee>` previews beautification diffs in terminal stdout without modifying files.
- **In-Place Mutation:** Modifies the target `.bee` file on disk only when structural or formatting differences are detected.

---

## 2. Formatting & Alignment Rules

### 2.1 Indentation & Block Alignment
- **Strict 2-Space Rule:** Every nested statement block body is indented by exactly +2 spaces per nesting level.
- **Block Terminators:** Block closing keywords (`done`, `repeat`, `return`) are aligned to 0 relative indentation relative to their block header (`rule`, `if`, `while`, `for`, `cycle`, `match`, `trial`).

### 2.2 End-of-Line Comment Alignment
- EOL comments (`-- comment`) within the same code block are right-aligned to a unified column boundary (default: Column 40 or +2 spaces beyond the longest statement line in the block).

### 2.3 Operator & Delimiter Spacing
- Spaces surrounding assignment (`:=`, `::`, `+=`, `-=`), comparison (`=`, `≠`, `∈`), logic (`∧`, `∨`), and arithmetic operators (`+`, `-`, `*`, `/`).
- Space after colons `:` in type declarations and function parameters (`x: 0 ∈ Z`).
- Compact range operator representation (`1..10`, `1.!10`).

---

## 3. Automatic Deterministic Syntax Auto-Fixes

The beautifier AST transformer safely corrects common shorthand or omitted syntax:

1. **Implicit Multiplication Insertion:**
   - Input: `2(a + b)` $\rightarrow$ Output: `2 * (a + b)`
   - Input: `n(x + y)` $\rightarrow$ Output: `n * (x + y)`
   - Input: `(a + b)(c + d)` $\rightarrow$ Output: `(a + b) * (c + d)`

2. **Operator Normalization:**
   - Normalizes ASCII aliases to Unicode standard operators where configured (e.g. `in` $\rightarrow$ `∈`, `forall` $\rightarrow$ `∀`, `exists` $\rightarrow$ `∃`).

3. **Block Header Colon Normalization:**
   - Ensures missing colons after block headers (`rule main:`, `if cond then:`, `else:`, `do:`) are appended automatically.

---

## 4. Compiler Subsystem Implementation Design (Go)

The beautifier is implemented in `internal/beautifier` using a **Concrete Syntax Tree (CST)** that preserves white spaces and comments alongside AST nodes:

```go
package beautifier

import (
	"bee/internal/lexer"
	"bee/internal/parser"
	"io"
)

type Beautifier struct {
	indentWidth int
	alignComments bool
	autoFix     bool
}

func New() *Beautifier {
	return &Beautifier{
		indentWidth:   2,
		alignComments: true,
		autoFix:       true,
	}
}

func (b *Beautifier) FormatSource(input string) (string, error) {
	// 1. Lex and Parse into Concrete Syntax Tree
	// 2. Apply AST auto-fix transformations (e.g. implicit multiplication)
	// 3. Pretty-print CST with strict 2-space indentation and comment alignment
	return formattedSource, nil
}
```

---

## 5. Summary of Benefits

- **Token Cost Elimination:** AI coding agents invoke `bee -b` instead of spending thousands of prompt tokens re-formatting code.
- **Readability & Auditability:** Ensures all Bee code across projects adheres to unified 2-space indentation and mathematical layout standards.
