# Solution: 1-Based Indexing Normalization
- **Compiler Normalization:** The Parser will consume Bee syntax `[1..$]`. 
- **Internal Mapping:** The Lowering phase maps `n` to `n-1` for LLVM memory offsets.
- **Matrix Notation:** `$` represents the last dimension. Slice notation `[$-2..$]` is enforced for relative range access.
- **Error Handling:** Raw negative indexing (e.g., `[-3]`) is a hard compile-time error.
