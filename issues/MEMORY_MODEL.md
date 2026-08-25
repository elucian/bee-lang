# Issue: Memory Model Implementation
- **Goal:** Track the implementation of Region-Based Allocator and verification of `zap` usage in Hot Zones.
- **Region Policy:** Every `rule` context maintains a memory region; local allocations are reclaimed at `return`.
- **Hot Zone Policy:** `zap` usage is strictly reserved for performance-critical loops; must be accompanied by source documentation.
- **Diagnostic:** Runtime `Panic` or compiler diagnostic for access after `zap`.
