# Implementation & Compiler Issues

- **Unicode Operator Inconsistency:** The spec mixes ASCII alternatives with Unicode operators (e.g., `!` vs `¬`). The compiler needs a clear specification on whether source files must be encoded in UTF-8 and if ASCII/Unicode equivalents are strictly interchangeable.
- **Hoisting Absence:** The explicit "no hoisting" rule (defined in `07-rules.md`) combined with the directive that `main()` must be at the bottom of the module creates a restrictive ordering requirement that might complicate large-scale code generation.
- **`$trial` Object Management:** The `$trial.messages` hash-map is managed automatically, but the life-cycle of this object and its interaction with multiple concurrent threads in a multi-threaded application is not explicitly defined in the memory model.
