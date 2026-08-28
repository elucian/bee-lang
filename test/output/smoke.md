# Test Execution Report: `test\level0\smoke.bee`

--- STDOUT ---
```
Smoke test

```

--- STDERR ---
```
DEBUG: parsing args...
DEBUG: Parsing expr, token: "Smoke test" type: STRING
DEBUG: In binary loop, peekTok type: ;, literal: ";"
DEBUG: Added expr, len: 1
DEBUG: peeked: type=;, lit=";"
DEBUG: Done parsing args, count: 1
DEBUG: Done parsing args, count: 1

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
