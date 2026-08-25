# Solution: Region-Based Memory Management
- **Primary Mechanism:** Each `rule` invocation defines a memory region.
- **Lifecycle:** All objects (unless RC-managed) allocated within a rule are automatically reclaimed upon the `return` statement.
- **Safety:** Because rules update pre-allocated result containers (passed by the caller), no "data escape" analysis is required. The lifetime is explicitly owned by the caller.
- **Explicit Zap:** In Hot Zones, `zap` is used to trigger early reclaim within a region before the `return` barrier is reached.
