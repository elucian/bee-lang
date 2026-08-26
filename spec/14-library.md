# Bee Specification: System Standard Library API (14-library.md)

## 1. Executive Library Architecture

The Bee Standard Library is designed for zero-overhead performance, modular tree-shaking static linkage, and OS-agnostic system access:

1. **Pay-For-What-You-Use Linkage:** The compiler statically links only the specific library routines imported or referenced by source code, eliminating binary bloat.
2. **System Namespaces (`$bee.sys`):** Core API modules are organized under the `$bee.sys` namespace hierarchy (`$bee.sys.io`, `$bee.sys.math`, `$bee.sys.string`, `$bee.sys.time`, `$bee.sys.env`).
3. **ARC File Handles (`F`):** Operating system file and directory descriptors are encapsulated in Atomic Reference Counted (ARC) handle objects ensuring immediate, deterministic cleanup upon scope exit or `zap`.

---

## 2. Introspection & Memory Utilities

| Utility Method / Routine | Signature | Description |
| :--- | :--- | :--- |
| **`entity.type()`** | `entity.type() => (t ∈ S)` | Returns type name string for any entity |
| **`size(entity)`** | `size(entity ∈ Object) => (b ∈ N)` | Returns exact memory footprint in bytes |
| **`length(coll)`** | `length(coll ∈ Collection) => (count ∈ N)` | Returns current element count |
| **`capacity(coll)`** | `capacity(coll ∈ Array) => (cap ∈ N)` | Returns allocated memory capacity |

---

## 3. String & Text Manipulation Utilities (`$bee.sys.string`)

```bee
use $bee.sys.string as StringUtil;

-- String utility methods
new str := "  hello, world!  ";
new trimmed := StringUtil.trim(str);                     -- "hello, world!"
new parts   := StringUtil.split(trimmed, ", ");          -- ("hello", "world!")
new joined  := StringUtil.join(parts, " - ");            -- "hello - world!"
new idx     := StringUtil.find(joined, "world");         -- 9 (1-based index)
new updated := StringUtil.replace(joined, "world", "Bee"); -- "hello - Bee!"
```

---

## 4. System I/O & File Operations (`$bee.sys.io`)

File operations use opaque file handles typed as **`F`** (File Handle):

### 4.1 File & Folder Lifecycle API
```bee
use $bee.sys.io as IO;

-- Check existence
if IO.Folder.exist("data") = false then:
  apply IO.Folder.create("data");
done;

-- Open file handle in write mode ("w", "r", "a", "rw")
new file_handle := IO.File.open("data/output.txt", "w");

-- Write data
apply IO.File.write(file_handle, "Hello, Bee Compiler!");

-- Flush and close handle
apply IO.File.close(file_handle);
```

### 4.2 Directory Processing
```bee
-- List directory contents into a List of string paths
new file_list := IO.Folder.list("data/");

for ∀ file_name ∈ file_list do:
  print "Found file: ", file_name;
repeat;
```

---

## 5. System Exception System (`$error`)

Exceptions and system error reports populate the global system variable **`$error`**:

```bee
type SystemError: {code ∈ Z, message ∈ S} <: Object;
```

### Error Code Allocation Standard:
- **`1 .. 199`:** System Reserved Standard Error Codes (I/O failures, network errors, allocation errors).
- **`200 .. 9999`:** User / Application Defined Business Domain Error Codes.
- **`≤ -1`:** Unrecoverable Hard Hardware or Memory Security Panics.

```bee
trial:
  try()
  new handle := IO.File.open("non_existent.txt", "r");
case $error.code = 404 do:
  print "File not found: ", $error.message;
  resume;
miss:
  print "Unhandled system error code: ", $error.code;
done;
```

---

## 6. Formal EBNF Grammar

```ebnf
(* Library Invocations & Directives *)
library_import    ::= "use" "$bee." sys_submodule [ "as" identifier ] ";" ;
sys_submodule     ::= "sys." ( "io" | "math" | "string" | "time" | "env" ) ;

(* File & Folder Operations *)
file_open         ::= "IO.File.open" "(" path_expr "," mode_expr ")" ;
file_close        ::= "IO.File.close" "(" handle_expr ")" ";" ;
folder_list       ::= "IO.Folder.list" "(" path_expr ")" ;

(* System Variables *)
system_var        ::= "$" ( "error" | "trial" | "precision" | "platform" ) ;
```

---

## 7. Diagnostic Error Codes

| Error Code | Error Condition | Description |
| :--- | :--- | :--- |
| `E1401` | `FileNotFound` | Target file path missing during open operation |
| `E1402` | `FileAccessDenied` | Insufficient OS permissions to open or write file handle |
| `E1403` | `UnclosedFileHandle` | Scope exited with unclosed file handle `F` |
| `E1404` | `LibraryNotFound` | Requested `$bee.sys` submodule not available in runtime build |
| `E1405` | `InvalidFileMode` | File open mode parameter is not `"r"`, `"w"`, `"a"`, or `"rw"` |

---

## 8. Alignment Status

- **Issues Addressed:** Formalized Standard Library tree-shaking static linkage, `$bee.sys` namespaces, universal `entity.type()`, `$error` codes (`1..199` vs `200+`), File handles (`F`), directory processing, and diagnostic codes.
- **Manifest Tracking:** Updated `MANIFEST.md` to mark all specifications in `/spec` fully formalized.
