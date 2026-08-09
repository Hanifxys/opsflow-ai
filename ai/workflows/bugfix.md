# AI Bugfix Workflow

```text
Reproduce
 ↓
Identify expected behaviour
 ↓
Trace request/data flow
 ↓
Find root cause
 ↓
Add regression test
 ↓
Fix
 ↓
Run relevant tests
 ↓
Run broader tests
 ↓
Review
```

Do not change architecture unless the bug proves the architecture itself is incorrect.

Every fixed bug should have a regression test where practical.
