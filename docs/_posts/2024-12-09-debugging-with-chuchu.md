---
layout: default
title: "Debugging with Chuchu: From Stack Trace to Fix"
---

# Debugging with Chuchu: From Stack Trace to Fix

*December 9, 2024*

Debugging is often 90% reading and 10% fixing. Chuchu excels at the reading part. Here is a workflow for debugging complex issues using Chuchu.

## The Workflow

1.  **Capture the Error**: Copy the stack trace or error message.
2.  **Start a Session**: `chu chat`
3.  **Provide Context**: Paste the error.
    > "I'm getting this panic in `main.go`: [paste stack trace]"
4.  **Let Chuchu Investigate**:
    -   Chuchu will use `read_file` to examine the lines mentioned in the stack trace.
    -   It will trace the variable values back through the call stack.
    -   It might use `search_code` to find where the variable was initialized.

## Example: The Nil Pointer Hunt

**User**: "Fix this panic: runtime error: invalid memory address or nil pointer dereference at `server.go:45`"

**Chuchu (Thinking)**:
1.  `read_file("server.go", start=40, end=50)` -> Sees `user.Profile.Name` access.
2.  "Ah, `user.Profile` might be nil."
3.  `search_code("Profile")` -> Finds `User` struct definition.
4.  `read_file("auth.go")` -> Checks where `User` is created.
5.  **Conclusion**: "The `GetUser` function returns a User with a nil Profile if the DB join fails. We need to check for nil before accessing Name."

**Chuchu (Action)**:
Uses `apply_patch` to wrap the access in an `if user.Profile != nil` block.

## Pro Tip: The "Explain" Intent

If you don't want Chuchu to fix it, just ask it to explain.
> "Explain why this code might panic."

This is a great way to learn and improve your own debugging skills without just copy-pasting a fix.
