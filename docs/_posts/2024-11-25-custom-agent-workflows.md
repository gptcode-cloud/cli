---
layout: default
title: "Custom Agent Workflows: Build Specialized Pipelines"
---

# Custom Agent Workflows: Build Specialized Pipelines

*November 25, 2024*

Chuchu comes with a standard set of agents (Router, Editor, Query, Research, Review), but every developer's workflow is unique. Chuchu allows you to define **Custom Agent Workflows** to automate repetitive tasks specific to your project.

## Defining a Workflow

Workflows are defined in `.chuchu/workflows.yaml`. A workflow is a sequence of steps, where each step can use a specific agent and prompt template.

### Example: Automated PR Description

Imagine you want to generate a PR description based on the changes in your current branch relative to `main`.

```yaml
workflows:
  pr_description:
    steps:
      - name: "Get Diff"
        tool: run_command
        args: "git diff main..."
        save_output: "git_diff"
      
      - name: "Generate Description"
        agent: query
        prompt: |
          Write a Pull Request description for the following changes.
          Follow the template:
          ## Summary
          ## Changes
          ## Verification
          
          Diff:
          {{git_diff}}
```

## Running a Workflow

To run this workflow, simply use the `run` command:

```bash
chu run pr_description
```

Chuchu will execute the steps in order. It will first run the git command, capture the output, and then pass it to the Query agent to generate the text.

## Advanced: Chained Agents

You can also chain agents together. For example, a "Refactor and Verify" workflow:

1.  **Query Agent**: Analyze file for code smells.
2.  **Editor Agent**: Apply refactoring suggestions.
3.  **Test Agent**: Run `go test ./...`.
4.  **Review Agent**: Review the changes if tests pass.

This level of automation turns Chuchu from a passive assistant into an active member of your team, handling the grunt work while you focus on architecture and logic.
