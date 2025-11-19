---
layout: default
title: "Advanced Context Management: Handling Large Repositories"
---

# Advanced Context Management: Handling Large Repositories

*December 2, 2024*

One of the biggest challenges in AI coding is the **Context Window**. Even with 128k token windows, dumping your entire repository into the prompt is slow, expensive, and often confuses the model. Chuchu uses **Smart Context Management** to solve this.

## How Chuchu Manages Context

Chuchu doesn't just read files; it understands them.

1.  **Project Map**: When you start a session, Chuchu generates a tree-like map of your project structure. This fits in ~500 tokens and gives the model a "mental map" of where things are.
2.  **Relevance Search**: When you ask a question, the `Query` agent uses `search_code` (grep) and `list_files` to find *only* the relevant files.
3.  **Summarization**: For large files, Chuchu can read just the outline (function signatures) instead of the full implementation details.

## Tips for Large Repos

If you are working in a massive monorepo, here are some tips to help Chuchu stay focused:

### 1. Use `.chuchuignore`
Create a `.chuchuignore` file (works just like `.gitignore`) to exclude directories that the AI should never see.
```text
# .chuchuignore
vendor/
node_modules/
legacy_code/
docs/images/
```

### 2. Be Specific in Prompts
Instead of "Fix the bug in the auth system", try:
> "Fix the nil pointer in `auth/login.go` when the user ID is empty. Check `auth/types.go` for the struct definition."

This guides the agent to read exactly what it needs, saving tokens and improving accuracy.

### 3. Reset Context
If a conversation gets too long, the context can get "polluted" with old information. Use the `reset` command (or just restart `chu chat`) to clear the history and start fresh with a focused goal.

## The Future: RAG
We are actively working on a local RAG (Retrieval-Augmented Generation) system for Chuchu. This will index your code into a vector database, allowing semantic search ("Find code that handles user logout") to instantly retrieve relevant snippets from millions of lines of code. Stay tuned!
