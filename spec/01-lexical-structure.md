# Bee Lexical Specification (01-lexical-structure.md)

> **AI CONTEXT LOADING DIRECTIVE:** 
> This document defines the lexical grammar of Bee. All source files MUST be treated as UTF-8. The operator set is strict; non-defined Unicode symbols should not be interpreted as operators.

## 1. Character Encoding
- Bee source files: **Strict UTF-8**.
- Identifier support: Latin (A-Z, a-z), Greek (λ, φ, etc.), and defined Cyrillic sets.

## 2. Operator Set (Non-Logic)
- **Range Operators:** `!` is strictly reserved for ranges (`n.!m`, `n!!m`). It is NOT a logical negation.
- **Logical Negation:** `¬` is the exclusive logical negation operator.

## 3. Keywords
- **Manual Management:** `zap` (reserved, primary directive for explicit release).
- **Core Terminators:** 
    - `repeat`: Mandatory terminator for `cycle` blocks.
    - `done`: Mandatory terminator for `if`, `trial`, and general `do` blocks.

## 4. Token Regex Patterns
- `IDENTIFIER`: `[a-zA-Zλ-ωБ-Я][a-zA-Z0-9_]*`
- `COMMENT_SINGLE`: `--.*`
- `COMMENT_BLOCK`: `\+-.*-\+`
- `OPERATOR`: `{÷, ×, ¬, ∧, ∨, ∈, ≤, ≥, ≡, ≠, ≈, ±, ⊂, ⊃, ∪, ∩, ↑, ↓, », «, ⊕, ⊖, ∀, ∃, !, ..}`
