# Test Execution Report: `test\level0\smoke.bee`

--- STDOUT ---
```
Smoke test

```

--- STDERR ---
```
W0901 SignatureGrammarPartial: line=2 rule="main" (signature decoder is heuristic; full type binding deferred to Phase 7.2)

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
