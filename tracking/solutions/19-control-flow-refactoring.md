# Solution: Control Flow Syntax Refactoring (D14)

## Design

### 1. Uniform Block Termination

Replace `repeat [label];` with `done [label];` as the block terminator for `cycle` and `for`. This unifies all control blocks under a single terminator rule:

```
"if"     ... "done" [label] ";"
"match"  ... "done" [label] ";"
"start"  ... "done" [label] ";"
"with"   ... "done" ";"
"trial"  ... "done" [label] ";"
"cycle"  ... "done" [label] ";"
"for"    ... "done" [label] ";"
```

Parser benefit: `parse_block()` checks only for `TOKEN_DONE` (and its variants) to close any control scope.

### 2. Anonymous Scope Prologue (`cycle:`)

Decouple scope allocation from label binding — **optional label, optional scope**:

| Form | Label? | Prologue? | Semantics |
| :--- | :--- | :--- | :--- |
| `cycle name:` decls `do` body `done name;` | ✅ | ✅ | Labeled with stable prologue |
| `cycle:` decls `do` body `done;` | ❌ | ✅ | Anonymous with stable prologue |
| `cycle name` `do` body `done name;` | ✅ | ❌ | Labeled volatile body (label is jump target only) |
| `cycle do` body `done;` | ❌ | ❌ | Anonymous volatile body |
| `cycle while cond do` body `done;` | ❌ | ❌ | While-loop volatile body |

The colon (`:`) becomes the **prologue marker**, independent of label presence. A label only serves as a jump target for `repeat label;` / `stop label;` / `redo label;`.

### 3. `repeat` Repurposed as Inline Jump

`repeat` is demoted from block terminator to inline transfer statement:

```ebnf
jump_stmt ::= ( "repeat" | "stop" | "redo" ) [ label ] [ "if" condition ] ";" ;
```

- **`repeat`** — skip remainder of current iteration body; jump to loop-header re-evaluation. In a `for` loop, advance to the next element.
- **`stop`** — terminate the loop immediately; transfer execution past `done [label];`.
- **`redo`** — restart current iteration without advancing the iterator (non-`for` cycles only).

### 4. `next` Retirement

The `next` keyword is retired; its semantic (advance to next iteration) is absorbed by `repeat`. The distinction was:
- `next` — advance iterator and continue (like `continue`).
- `redo` — restart current iteration without advancing.

With `repeat` as the universal "continue" keyword, `next` becomes redundant. In a `for` loop, `repeat` implies "advance to next element." In a non-`for` cycle, `repeat` implies "re-evaluate the while condition."

### 5. `then` Post-Loop Epilogue

The optional `then` block is repositioned as a post-loop epilogue:

```
cycle [label] [:] [prologue]
  ( "do" | "while" cond "do" | "for" ... "do" )
  body_block
[ "then" epilogue_block ]
"done" [label] ";"
```

The `then` block executes exactly once after loop exit (via `stop` or condition exhaustion), before the prologue scope is popped. It does NOT execute if the loop is exited via `repeat`/`redo`/`next` (those continue iteration, not exit).

## Grammar Changes (EBNF)

### New `cycle_stmt` EBNF

```ebnf
cycle_stmt  ::= "cycle" [ label ] [ ":" decl_block ]
                ( "do" | "while" expression "do"
                | "for" [ "∀" ] identifier ( "∈" | "in" ) expression "do" )
                block
                [ "then" block ]
                "done" [ label ] ";"
              | "for" [ "∀" ] identifier ( "∈" | "in" ) expression "do"
                block "done" ";" ;
decl_block  ::= { declaration_stmt } ;
```

### New `jump_stmt` EBNF

```ebnf
jump_stmt ::= ( "repeat" | "stop" | "redo" ) [ label ] [ "if" condition ] ";" ;
```

### `transfer_stmt` Update

Remove `"next"` from the transfer statement alternatives. Add `"repeat"` as a jump statement. The `jump_stmt` becomes a sub-production of `transfer_stmt`.

## Worked Examples

### Infinite loop with stop
```bee
cycle:
  new count ∈ N := 0;
do
  let count += 1;
  repeat if count % 2 = 0;  -- skip even numbers
  write count;
  stop if count ≥ 100;      -- exit condition
done;
```

### While loop with prologue
```bee
cycle:
  new total ∈ Z := 0;
  new n ∈ Z := 0;
while n < 10 do
  let n += 1;
  let total += n;
then
  print ("Sum:", total);
done;
```

### Nested labeled loops
```bee
cycle outer:
  new x ∈ N := 0;
do
  cycle inner:
  for ∀ y ∈ (1..10) do
    stop outer if check_critical(x, y);
    repeat outer if check_retry(x, y);
  done inner;
done outer;
```

### For with anonymous prologue
```bee
cycle:
  new results ∈ List(Z);
for ∀ i ∈ (1..100) do
  repeat if i % 5 = 0;  -- skip multiples of 5
  let results append process(i);
done;
```

## Compiler Impact (for next implementation pass)

1. **Lexer:** No new tokens needed. `repeat` remains a keyword but its parse context changes.
2. **Parser:** `parse_block()` closes on `TOKEN_DONE` for all control blocks. `repeat` in statement position → `TransferStatement`.
3. **AST:** `CycleStatement.Terminator` changes from `repeat` to `done`. `RepeatStatement` (or `TransferStatement`) added for inline `repeat`.
4. **Evaluator:** `done` triggers prologue scope cleanup. `repeat` sets a `continuing` flag that jumps to loop-header. `stop` sets an `exiting` flag that breaks out past `done`. `then` block runs on `exiting` before scope cleanup.

## Test Strategy

New `@FROZEN` tests to be authored after compiler implementation:
- `T0210.bee` — anonymous `cycle:` prologue, `repeat if`, `stop if`, `done;`
- `T0211.bee` — labeled `cycle outer:` with nested `done outer;` / `done inner;`
- `T0212.bee` — `for` with `repeat` jump (skip evens)
- `T0213.bee` — `then` epilogue after `stop`
- `T0214.bee` — `redo` restart without advance

Existing tests using legacy `repeat;` terminators will be flagged `@DISABLED` (per freeze protocol) and eventually superseded by the new-syntax tests.
