# Bee Memory Model Specification (00-memory-model.md)

## 1. Overview
Bee employs a three-tier hybrid memory management architecture:
- **Reference Counting (RC):** Default for mutable objects.
- **Manual Management (MMM):** Explicit `new`/`zap` for Hot Zones.
- **Garbage Collection (GC):** Compacts immutable strings.
