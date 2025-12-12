---
title: Dependency Graph & Context Optimization
description: Intelligent context selection using PageRank and multi-language dependency analysis
---

# Dependency Graph & Context Optimization

GPTCode builds a dependency graph of your codebase to intelligently select the most relevant files for each query.

---

## Overview

The dependency graph feature:

1. **Parses imports** across multiple languages (Go, Python, JS/TS, Ruby, Rust)
2. **Builds a graph** where nodes are files and edges are dependencies
3. **Ranks files** using PageRank algorithm (like Google Search)
4. **Optimizes context** by selecting the most relevant files for your query
5. **Caches results** for fast subsequent queries

**Benefits:**
- 5x token reduction (100k → 20k typical)
- Better LLM responses (focused context)
- Cost savings (fewer tokens = lower API costs)
- Faster responses (less to process)

---

## How It Works

### 1. Import Detection

The graph builder scans your codebase and extracts import relationships:

**Go:**
```go
import "github.com/user/repo/internal/auth"  // External
import "mymodule/internal/config"            // Internal (uses go.mod)
```

**Python:**
```python
from auth import User          # Relative
from project.db import connect # Absolute
```

**JavaScript/TypeScript:**
```javascript
import { API } from './api'              // Relative
import { auth } from '@/lib/auth'        // Alias
```

**Ruby:**
```ruby
require 'rails/all'           # Gem
require_relative '../config'  # Relative
```

**Rust:**
```rust
use crate::models::User;      // Internal
use std::collections::HashMap; // Standard
```

### 2. Graph Construction

Files become nodes, imports become edges:

```
main.go → auth.go → user.go
       → api.go  → user.go
       → config.go
```

### 3. PageRank Scoring

Files are scored by importance using PageRank algorithm:

- Files imported by many others = high score
- Files importing many others = lower score
- Scores sum to ~1.0 across all files

**Example scores:**
```
user.go:    0.187  (imported by auth, api, tests)
auth.go:    0.142  (imported by main, api)
config.go:  0.095  (imported by main)
main.go:    0.063  (entry point, imports many)
```

### 4. Context Optimization

When you ask a query like "how does authentication work?":

1. **Keyword matching**: Find files containing "auth", "login", "user"
2. **Neighbor expansion**: Include files that import/are imported by matches
3. **PageRank weighting**: Sort by importance score
4. **Top N selection**: Select top 5 (configurable) most relevant files

### 5. Smart Truncation

Selected files are truncated to ~3000 chars each:

- **Head** (first 30 lines): Imports, package declaration, type definitions
- **Tail** (last 20 lines): Recent code, likely most relevant

This keeps essential context while reducing token usage.

---

## Configuration

### Max Files

Control how many files are added to context:

```bash
# View current setting (default: 5)
chu config get defaults.graph_max_files

# Increase for more context (1-20)
chu config set defaults.graph_max_files 10

# Decrease for fewer tokens
chu config set defaults.graph_max_files 3
```

**Recommendations:**
- **Small projects** (<50 files): 3-5 files
- **Medium projects** (50-500 files): 5-8 files
- **Large projects** (500+ files): 8-12 files

### Environment Variables

**Enable debug mode:**
```bash
export CHUCHU_DEBUG=1
chu chat "your query"  # Shows graph stats
```

Debug output example:
```
[Graph] Built: 143 nodes, 287 edges (from cache)
[Graph] Selected 5 files for query "authentication":
  - internal/auth/handler.go (PR: 0.187)
  - internal/auth/middleware.go (PR: 0.142)
  - internal/models/user.go (PR: 0.095)
```

---

## CLI Commands

### Build Graph

Force rebuild ignoring cache:

```bash
chu graph build
```

Output:
```
🏗️  Building dependency graph...
   Nodes: 143
   Edges: 287
📊 Calculating PageRank...
✅ Done in 234ms
```

**When to use:**
- After major refactoring
- After adding/removing many files
- If cache seems stale

### Query Graph

Find relevant files for a query:

```bash
chu graph query "authentication"
chu graph query "database connection"
chu graph query "api routes"
```

Output:
```
🔍 Query: "authentication"
📂 Relevant Context:
   - internal/auth/handler.go (PR: 0.187)
   - internal/auth/middleware.go (PR: 0.142)
   - internal/models/user.go (PR: 0.095)
   - cmd/server/main.go (PR: 0.063)
   - internal/config/auth.go (PR: 0.051)
```

---

## Auto-Integration in Chat Mode

The graph automatically enhances `chu chat`:

```bash
chu chat "how does authentication work?"
```

**What happens:**
1. Query is analyzed for keywords ("authentication", "auth")
2. Graph finds top 5 relevant files
3. Files are truncated to ~3000 chars each
4. Context is appended to your message:
```
[Context from Dependency Graph]

File: internal/auth/handler.go (lines 1-30, 180-200)
[truncated content...]

File: internal/auth/middleware.go (lines 1-30, 95-115)
[truncated content...]
```
5. LLM receives enhanced context for better answers

**Comparison:**

| Without Graph | With Graph |
|--------------|------------|
| All files (100k tokens) | Top 5 files (20k tokens) |
| $0.50/query | $0.10/query |
| Slower response | Faster response |
| Generic answers | Focused answers |

---

## Cache System

### How Caching Works

The graph is expensive to build (300ms for 500 files), so results are cached:

**Cache key:** MD5 of all file modification times
**Cache location:** `~/.gptcode/cache/graph_<md5>.json`
**Staleness:** 24 hours

### Cache Lifecycle

1. **First query:** Build graph, cache result
2. **Subsequent queries:** Load from cache (instant)
3. **File changes:** Detects via mtime hash, rebuilds
4. **24h expiry:** Rebuilds even if no changes detected

### Manual Cache Control

```bash
# Force rebuild (clears cache)
chu graph build

# Clear all caches manually
rm -rf ~/.gptcode/cache/graph_*.json
```

---

## Supported Languages

| Language | Extensions | Import Detection |
|----------|-----------|------------------|
| Go | `.go` | `import`, uses `go.mod` for module resolution |
| Python | `.py` | `import`, `from...import` |
| JavaScript | `.js`, `.jsx` | `import`, `require()` |
| TypeScript | `.ts`, `.tsx` | `import`, handles aliases |
| Ruby | `.rb` | `require`, `require_relative` |
| Rust | `.rs` | `use`, `mod` declarations |

### Go Module Resolution

For Go projects, the builder reads `go.mod` to resolve internal imports:

```go
// go.mod
module github.com/user/myproject

// main.go
import "github.com/user/myproject/internal/auth"  // Resolved as internal
```

This ensures internal package imports are correctly linked in the graph.

---

## Performance

### Benchmarks

**Medium Go project** (150 files, 15k LOC):
- Build time: 234ms (first run)
- Cache load: 12ms (subsequent)
- PageRank: 18ms
- Context optimization: 3ms

**Large TypeScript project** (500 files, 80k LOC):
- Build time: 1.2s (first run)
- Cache load: 45ms (subsequent)
- PageRank: 67ms
- Context optimization: 8ms

### Memory Usage

- Graph structure: ~500 bytes per file
- 500 files = ~250KB in memory
- Cache on disk: ~500KB for 500 files

---

## Algorithm Details

### PageRank Implementation

Classic PageRank with damping:

```
PR(A) = (1-d)/N + d * Σ(PR(Ti) / C(Ti))

Where:
- PR(A) = PageRank score of file A
- d = damping factor (0.85)
- N = total number of files
- Ti = files that link to A
- C(Ti) = number of outgoing links from Ti
```

**Iteration:**
- Runs for 20 iterations or until convergence
- Convergence threshold: 0.0001 delta
- Typical convergence: 8-12 iterations

### Context Optimizer Algorithm

```python
def optimize_context(query, max_files):
    # 1. Keyword matching
    candidates = []
    for file in graph.nodes:
        if query_keywords_in(file.path):
            candidates.append(file)
    
    # 2. Neighbor expansion
    for candidate in candidates:
        candidates += graph.neighbors(candidate)
    
    # 3. Deduplicate
    candidates = unique(candidates)
    
    # 4. Sort by PageRank
    candidates.sort(key=lambda f: f.score, reverse=True)
    
    # 5. Select top N
    return candidates[:max_files]
```

**Keyword matching:**
- Splits query into tokens
- Matches against file paths (case-insensitive)
- Supports partial matching ("auth" matches "authentication.go")

---

## Examples

### Example 1: Authentication Query

```bash
chu chat "explain the authentication flow"
```

**Graph selects:**
1. `internal/auth/handler.go` - Contains "auth", high PageRank
2. `internal/auth/middleware.go` - Neighbor of handler
3. `internal/models/user.go` - Imported by auth files
4. `cmd/server/main.go` - Entry point, imports auth
5. `internal/config/auth.go` - Contains "auth" keyword

**Result:** LLM gets focused auth-related files, gives accurate answer

### Example 2: Database Query

```bash
chu chat "how do we connect to the database?"
```

**Graph selects:**
1. `internal/db/connection.go` - Contains "db", "connection"
2. `internal/db/migrations.go` - Neighbor of connection
3. `internal/models/base.go` - Imports db package
4. `internal/config/database.go` - Contains "database"
5. `cmd/migrate/main.go` - Uses db connection

**Result:** Comprehensive database context without unrelated files

### Example 3: API Routes Query

```bash
chu chat "list all api endpoints"
```

**Graph selects:**
1. `internal/api/routes.go` - Contains "routes", "api"
2. `internal/api/handlers.go` - Neighbor of routes
3. `internal/api/middleware.go` - Imported by routes
4. `cmd/server/main.go` - Registers routes
5. `internal/auth/api.go` - Contains "api"

**Result:** Complete API overview from relevant route files

---

## Troubleshooting

### Graph not building

**Issue:** No graph data generated

**Solutions:**
```bash
# Check if project has supported files
ls **/*.go **/*.py **/*.ts

# Enable debug mode
export CHUCHU_DEBUG=1
chu graph build

# Check for errors in output
```

### Incorrect file selection

**Issue:** Graph selects wrong files

**Solutions:**
```bash
# Test query matching
chu graph query "your search term"

# Adjust max_files if too few/many
chu config set defaults.graph_max_files 8

# Rebuild graph if stale
chu graph build
```

### Cache not updating

**Issue:** Old files still in context

**Solutions:**
```bash
# Force rebuild
chu graph build

# Clear cache manually
rm ~/.gptcode/cache/graph_*.json

# Check file mtimes
ls -la <file>
```

### Performance issues

**Issue:** Graph build is slow

**Solutions:**
- Exclude large directories in `.gitignore` (already respected)
- Reduce `graph_max_files` for faster queries
- Use cache (automatic after first build)
- Check for huge files (>100k LOC)

---

## Limitations

### Current Limitations

1. **Static analysis only**
   - No runtime dependency tracking
   - Dynamic imports not detected
   
2. **Path-based matching**
   - Keyword search is simple substring matching
   - No semantic understanding of code
   
3. **Truncation trade-off**
   - ~3000 chars per file may miss some context
   - Configurable via code, not CLI (yet)

4. **Language support**
   - Only 6 languages supported
   - No Java, C#, PHP, etc.

### Future Enhancements

- [ ] Semantic code search (embeddings)
- [ ] Function-level granularity
- [ ] More languages (Java, C#, PHP)
- [ ] Configurable truncation
- [ ] Graph visualization
- [ ] Export to GraphML/DOT

---

## Next Steps

- Explore [ML Features](./ml-features.html) for intent classification
- See [Commands Reference](./commands.html) for full CLI
- Read implementation in `internal/graph/`
- Check tests in `internal/graph/builder_test.go`
