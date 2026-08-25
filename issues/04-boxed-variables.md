# Issue: Scoped Mutability Wrapper Implementation
- **Goal:** Implement the `[]` boxed variable semantics.
- **Requirement:** Distinguish between Heap allocation (for public `.` states) and Region-based allocation (for local `new` variables).
- **Compiler Logic:**
    - `[]` inside `self.` or `.` context -> Promote to Instance Heap.
    - `[]` inside rule block -> Local Region allocation (Auto-reclaim at `return`).
- **Auditability:** Ensure the compiler emits warnings if a boxed variable is used across thread boundaries without proper RC/atomic wrapping.
