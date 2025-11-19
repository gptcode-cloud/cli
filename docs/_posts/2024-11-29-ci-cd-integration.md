---
layout: default
title: "Integrating Chuchu with CI/CD: Automating Code Reviews"
---

# Integrating Chuchu with CI/CD: Automating Code Reviews

*November 29, 2024*

Chuchu isn't just for your local terminal. Because it's a CLI tool, it can be easily integrated into your CI/CD pipelines to provide automated code reviews and sanity checks.

## The `chu review` Command

The `chu review` command (triggered via `chu chat` with "review" intent) is perfect for CI. It analyzes code and outputs markdown.

## GitHub Actions Example

Here is a GitHub Action workflow that runs Chuchu on every Pull Request to review changed files.

```yaml
name: AI Code Review

on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install Chuchu
        run: go install github.com/yourusername/chuchu/cmd/chu@latest

      - name: Run Review
        env:
          GROQ_API_KEY: ${{ secrets.GROQ_API_KEY }}
        run: |
          # Get list of changed files
          CHANGED_FILES=$(git diff --name-only origin/main HEAD | grep .go)
          
          # Run review for each file
          for file in $CHANGED_FILES; do
            echo "Reviewing $file..."
            echo "Review $file" | chu chat > review_output.md
            
            # Post comment to PR (using gh cli)
            gh pr comment ${{ github.event.pull_request.number }} -F review_output.md
          done
```

## Benefits

1.  **Instant Feedback**: Developers get immediate feedback on style and obvious bugs before a human reviewer even looks at the code.
2.  **Consistency**: The AI applies the same standards every time.
3.  **Cost**: Using Groq's Llama 3.1 70B, a full review costs fractions of a cent.

## Best Practices

-   **Filter Files**: Only review source code (exclude vendor, generated files).
-   **Use a "Soft" Prompt**: Configure the Review agent to be helpful, not nitpicky.
-   **Human in the Loop**: AI reviews should complement, not replace, human reviews. Use them to catch low-hanging fruit.
