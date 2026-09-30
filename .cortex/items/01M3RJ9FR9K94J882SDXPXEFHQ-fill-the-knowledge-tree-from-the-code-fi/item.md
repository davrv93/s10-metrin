---
id: 01M3RJ9FR9K94J882SDXPXEFHQ
type: task
title: Fill the knowledge tree from the code (first setup)
status: todo
author: owner
assignee: "@ai"
tags:
  - bootstrap
fields:
  priority: high
created_at: 2026-09-30T07:08:32.393Z
updated_at: 2026-09-30T07:08:32.393Z
updated_by: owner
---

# Cortex bootstrap task

You are setting up the Cortex knowledge tree for "S10 Conocimiento".
Cortex is this project's single source of truth. Humans review everything you write.

Write all Cortex content in Spanish (es), in plain words. Keep code, paths and identifiers as they are.

1. Call cortex_brief to see the current branches.
2. Explore the codebase (folders, package files, entry points). Do not guess.
3. For the root node ("") write a clear project summary (<= 300 chars) and a body describing
   purpose, main components and how they talk to each other.
4. For each top-level branch in the brief:
   - update its summary to describe THIS project
   - if a branch does not apply, say so in its summary ("Does not apply: ...") and suggest in a question that a human deletes it
   - add child nodes for important subsystems (e.g. backend/auth, backend/payments)
   - link the code files each node describes (links.code)
5. Existing markdown worth importing:
   - FUENTES.md
   - METRIN.md
   - README.md
   Move the durable knowledge into nodes. Note contradictions instead of silently picking one.
   Everything in .cortex/ is committed: never copy secrets (passwords, keys, tokens, addresses with
   credentials) into Cortex, and do not read files git ignores (e.g. *.local.md) for this task.
6. Past decisions you find in docs or commit history: record them with cortex_create_item (type "decision").
   Open questions and contradictions: cortex_create_item (type "question", assignee "@humans").
7. Always include a short "reason" with each write.
8. Claim this task (cortex_claim) when you start and move it to "review" when you finish, with a reply listing what you wrote.
   Work in passes if the project is large: a handoff note on release tells the next session where you stopped.

Knowledge nodes you write become drafts until a human approves them in Cortex.
