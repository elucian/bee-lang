# Bee Specification: Structure (04-structure.md)

## 1. Module Model
- **Main Module:** The entry point of an application. Must contain exactly one `rule main`. Prohibited from being loaded/imported into other modules.
- **Secondary Module:** Located in `src/`. Reusable. Public members are accessible via dot-notation. Forbidden from containing `rule main`.
- **Library Module:** Located in `lib/`. Globally reusable across projects. Loaded once per execution context. Forbidden from containing `rule main`.

## 2. Name Space & Scoping
- **Public Members:** Identified by a `.` prefix (e.g., `rule .my_rule`).
- **Private Members:** No prefix. Visible only within the module scope.
- **Qualifier Suppression:** `with identifier do block done` suppresses module qualifiers for all identifiers within the block.
- **Aliasing:** `alias new_name: qualifier.member_name;` exports/renames symbols into the local scope.

## 3. Execution Primitives
- **Synchronous:** `apply identifier(...)` executes a rule and discards results.
- **Asynchronous:** `begin identifier(...)` spawns a new thread with an isolated memory region.
- **Synchronization:** `wait` block acts as a barrier, joining all spawned threads in the current scope.
