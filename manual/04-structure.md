# Code Structure

Bee has a modular architecture. A source file represents a module. A project can have many modules and can contain other files like configuration files, images and data files. Each module has extension *.bee and has a distinct role depending on its declaration and location. 

## Projects
A Bee project is a folder with a specific structure containing one or more applications. 

**Recommended Project Tree:**
```text
$pro_home/
  bin/           # Compiled executables
  src/           # Application modules
    module1.bee
    module2.bee
  lib/           # Library modules
    library1.bee
  doc/           # Documentation
  client.bee     # Main entry point
  server.bee     # Main entry point
```

## System Variables
System variables use the "$" prefix to locate project files or configuration.

| Variable | Description |
| :--- | :--- |
| `$bee_home` | Bee home installation folder |
| `$bee_lib`  | Bee library base directory |
| `$pro_home` | Project root folder |
| `$pro_lib`  | Project-specific libraries |
| `$pro_mod`  | Project module search path |
| `$pro_log`  | Log output directory |

## Compiler Directives
Directives control the compilation process and are set in configuration files or the main module.

| Constant | Default | Description |
| :--- | :--- | :--- |
| `$max_precision` | 0.00001 | Rational number precision |
| `$max_recursion` | 10000 | Recursion depth limit |
| `$log_debug` | "Off" | Debug info toggle |
| `$platform` | "Windows" | Target: Windows/Linux/Mac |

## Modules
Bee applications consist of one `main` module and multiple secondary/library modules.

- **Main Module:** The entry point for an application. MUST contain a `rule main`. Cannot be imported/loaded into other modules.
- **Secondary Modules:** Located in `src/`. Contain reusable `rule` sets but **do not** contain `rule main`.
- **Library Modules:** Located in `lib/`. Globally reusable, loaded once per module. **Do not** contain `rule main`.

### Main rule
The `main` rule is the orchestration entry point.
```bee
-- *params accepts variable-length list of strings
rule main(*params ∈ S):
  new c := params.count;
  panic if (c = 0);
  
  new i := 0 ∈ Z;
  while (i < c) do
    write params[i];
    let i += 1;
    write "," if (i < c);
  repeat;
  print;
  return;
```

## Name space
Modules define their own scope (namespace).
- Public members: start with `.`
- Private members: no prefix.

```bee
-- demo module:
set .pi: 3.14; -- Public constant

rule .bar(x, y ∈ N) => (r ∈ N):
  new str := "test";
  let r := x + y;
return;
```

## Execution
- **Synchronous (`apply`):** Sequential execution in caller context.
- **Asynchronous (`begin` / `wait`):** Spawns threads and synchronizes at `wait`.

* * *

**Read next:** [Data Types](/projects/bee/types/)
