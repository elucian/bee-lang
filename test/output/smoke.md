# Test Execution Report: `test\level0\smoke.bee`

--- STDOUT ---
```
Smoke test

```

--- STDERR ---
```
DEBUG: Parser peek char: "\""
DEBUG: peek after first expr: "\""
DEBUG: peek after first: "\n"
DEBUG: peek after loop: "\n"
DEBUG: S.Separator is nil? true

```

--- Source Code ---
```bee
-- @DESC:Smoke test validation
rule main:
  print "Smoke test";
return;
```

--- Conclusion ---
Status: PASS
