# Library & Standard Modules

- **Qualified Loading Collision:** The `use module as qualifier` mechanism is documented to avoid name collisions, but it is unclear how `alias` interacts when multiple modules are imported with different qualifiers but shared members.
- **`$bee_lib` Path Resolution:** The dynamic path concatenation using the `.` operator (e.g., `$bee_lib.folder.mod`) is OS-dependent. The specification needs to clarify how this maps to native file systems while maintaining portability.
