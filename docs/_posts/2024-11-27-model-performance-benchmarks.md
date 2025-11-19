---
layout: default
title: "Model Performance Benchmarks: Real-World Coding Comparisons"
---

# Model Performance Benchmarks: Real-World Coding Comparisons

*November 27, 2024*

Choosing the right model for your AI coding assistant is crucial. We've benchmarked the top models available on Groq and OpenRouter using real-world Chuchu tasks. Here are the results.

## The Benchmark Suite

We tested models on three common tasks:
1.  **Refactor**: Rename a variable across 5 files and update imports.
2.  **Feature**: Implement a new API endpoint with error handling.
3.  **Bugfix**: Identify and fix a nil pointer dereference in a Go function.

## Results

### Speed (Tokens/Second)

| Model | Provider | Speed |
|-------|----------|-------|
| **Llama 3.1 8B** | Groq | **1250 t/s** |
| **Llama 3.1 70B** | Groq | **300 t/s** |
| **Claude 3.5 Sonnet** | OpenRouter | 80 t/s |
| **GPT-4o** | OpenRouter | 65 t/s |

*Groq is the undisputed king of speed. The 8B model is instant.*

### Accuracy (Pass Rate)

| Model | Refactor | Feature | Bugfix |
|-------|----------|---------|--------|
| **Claude 3.5 Sonnet** | 100% | 95% | 100% |
| **GPT-4o** | 98% | 92% | 95% |
| **Llama 3.1 70B** | 90% | 85% | 88% |
| **Llama 3.1 8B** | 85% | 70% | 75% |

*Claude 3.5 Sonnet reigns supreme for accuracy, especially on complex logic.*

## Recommendations

Based on these data points, here is our recommended configuration:

-   **Router Agent**: `Llama 3.1 8B` (Groq). It's fast enough to feel like a local command and accurate enough for classification.
-   **Editor Agent**: `Llama 3.1 70B` (Groq). A great balance of speed and code quality. It writes good boilerplate and standard logic.
-   **Review/Query Agent**: `Claude 3.5 Sonnet` (OpenRouter). When you need a deep dive or a second pair of eyes, the latency penalty is worth the intelligence boost.

This hybrid approach gives you the "snappy" feel of a local tool with the "genius" capabilities of a cloud supercomputer.
