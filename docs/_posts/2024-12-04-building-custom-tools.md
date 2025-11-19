---
layout: default
title: "Building a Custom Tool for Chuchu: Extending Capabilities"
---

# Building a Custom Tool for Chuchu: Extending Capabilities

*December 4, 2024*

Chuchu is designed to be extensible. While it comes with powerful tools like `read_file` and `run_command`, sometimes you need something specific to your domain. This guide shows you how to build a custom tool for Chuchu.

## The `Tool` Interface

In Chuchu, a tool is simply a Go function that implements the `Tool` interface (conceptually). It needs a definition (JSON schema) and an execution function.

### Step 1: Define the Tool

In `internal/tools/tools.go`, add your tool definition:

```go
var MyCustomTool = ToolDefinition{
    Name: "check_database_schema",
    Description: "Checks the current database schema against the migration files",
    Parameters: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "environment": map[string]interface{}{
                "type": "string",
                "description": "The environment to check (dev, staging, prod)",
            },
        },
        "required": []string{"environment"},
    },
}
```

### Step 2: Implement the Logic

Create a new file `internal/tools/database.go`:

```go
package tools

func CheckDatabaseSchema(call ToolCall, workdir string) ToolResult {
    env := call.Arguments["environment"].(string)
    
    // Your custom logic here...
    // e.g., run a CLI command, query a DB, parse a file
    
    return ToolResult{
        Tool: "check_database_schema",
        Result: "Schema is up to date with 14 migrations applied.",
    }
}
```

### Step 3: Register the Tool

Add it to the `GetAvailableTools` and `ExecuteTool` functions in `internal/tools/tools.go`.

## Use Cases

-   **Internal APIs**: Create a tool to fetch feature flags or user data from your internal admin panel.
-   **Linting**: Wrap your company's specific linter (e.g., `golangci-lint` with custom rules) as a tool.
-   **Deployment**: Give Chuchu the ability to trigger a deployment to staging: `deploy_to_staging(branch="feature-x")`.

By building custom tools, you transform Chuchu from a generic coding assistant into a specialized domain expert for your specific codebase.
