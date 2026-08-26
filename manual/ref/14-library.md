# Bee Standard Library

Bee is a compiled language; you must include only what you use. The standard library provides a minimal footprint for any executable.

## 1. Introspection Rules
Rules for runtime inspection.

| Rule | Purpose |
| :--- | :--- |
| `type` | Return type name |
| `size` | Return type memory size |

## 2. Core Collections
Primitive and composite types are included automatically in the standard library.

| Rule | Purpose |
| :--- | :--- |
| `length` | Collection length |
| `capacity` | Collection capacity |
| `min` / `max` | Limit definitions |

## 3. String & List Utilities
| Rule | Purpose |
| :--- | :--- |
| `split` / `join` | String to list / List to string |
| `find` / `replace` | String search & modification |
| `trim` | Remove whitespace |
| `left`/`right`/`center` | Alignment |

## 4. System I/O
The `system.io` module manages file handles (`F`) and folder structures.

| Rule | Purpose |
| :--- | :--- |
| `open` / `close` | File I/O lifecycle |
| `exist` | File/Folder existence check |
| `list` | Directory listing |
| `delete` | File/Folder deletion |

```bee
-- File I/O Usage
new file := File.open('data.txt', 'w');
```
