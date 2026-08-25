# Issue: Documentation Inconsistency
- **Status:** Open
- **Description:** Discrepancy between documentation regarding 0-based vs 1-based indexing.
- **Resolution:** All indexed structures (Arrays, Matrices, Lists) will be strictly 1-based in Bee syntax. The compiler will perform offset normalization (`index - 1`) during the AST lowering to LLVM IR (0-based).
- **Documentation Update:** Documentation needs a final pass to ensure all examples (especially `11-processing.md`) reflect 1-based indexing.
