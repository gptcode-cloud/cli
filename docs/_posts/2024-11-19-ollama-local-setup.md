---
layout: post
title: "Running Chuchu Completely Offline with Ollama"
date: 2024-11-19
author: Jader Corrêa
tags: [configuration, ollama, local, privacy, offline]
---

# Running Chuchu Completely Offline with Ollama

Want to use Chuchu without sending your code to the cloud? Ollama lets you run powerful LLMs locally on your machine, completely free and private.

## Why Run Local Models?

- **Privacy**: Your code never leaves your machine
- **Cost**: $0 per token - run unlimited queries
- **Speed**: No network latency for small models
- **Offline**: Work anywhere, even without internet
- **Control**: Full control over model versions and updates

## Prerequisites

1. Install Ollama: https://ollama.com/download
2. Have at least 8GB RAM (16GB+ recommended for larger models)
3. SSD storage (models can be 4-30GB each)

## Recommended Model Configurations

### Balanced Setup (16GB RAM)

Best all-around configuration for most machines:

```yaml
backend:
  ollama:
    type: ollama
    base_url: http://localhost:11434
    default_model: llama3.1:8b
    agent_models:
      router: llama3.1:8b         # Fast routing
      query: gpt-oss:latest       # 20B for comprehension
      editor: qwen3-coder:latest  # Specialized for code
      research: gpt-oss:latest    # Good at synthesis
```

**Required models:**
```bash
ollama pull llama3.1:8b        # ~4.7GB
ollama pull gpt-oss:latest     # ~13GB
ollama pull qwen3-coder:latest # ~18GB
```

**Total storage**: ~36GB

### Performance Setup (32GB+ RAM)

For powerful machines that can run larger models:

```yaml
backend:
  ollama:
    type: ollama
    base_url: http://localhost:11434
    default_model: deepseek-r1:32b
    agent_models:
      router: llama3.1:8b         # Fast routing
      query: deepseek-r1:32b      # Strong reasoning
      editor: qwen3-coder:latest  # Code specialist
      research: deepseek-r1:32b   # Excellent research
```

**Required models:**
```bash
ollama pull llama3.1:8b
ollama pull deepseek-r1:32b    # ~20GB
ollama pull qwen3-coder:latest
```

**Total storage**: ~43GB

### Minimal Setup (8GB RAM)

For resource-constrained machines:

```yaml
backend:
  ollama:
    type: ollama
    base_url: http://localhost:11434
    default_model: llama3.1:8b
    agent_models:
      router: llama3.1:8b
      query: llama3.1:8b
      editor: llama3.1:8b
      research: llama3.1:8b
```

**Required models:**
```bash
ollama pull llama3.1:8b
```

**Total storage**: ~4.7GB

## Model Recommendations by Task

### Router (Intent Classification)
- **Best**: `llama3.1:8b` - Fast and accurate enough
- **Alternative**: `phi3:mini` - Even smaller/faster

### Query (Code Analysis)
- **Best**: `deepseek-r1:32b` - Excellent reasoning
- **Good**: `gpt-oss:latest` - Strong comprehension
- **Budget**: `llama3.1:8b` - Decent understanding

### Editor (Code Generation)
- **Best**: `qwen3-coder:latest` - Specialized for code (30B MoE)
- **Good**: `deepseek-coder-v2:latest` - Strong coding model
- **Budget**: `codellama:13b` - Decent code generation

### Research (Information Synthesis)
- **Best**: `deepseek-r1:32b` - Excellent at reasoning
- **Good**: `gpt-oss:latest` - Good synthesis
- **Budget**: `llama3.1:8b` - Basic research

## Popular Ollama Models for Development

| Model | Size | RAM | Best For | Quantization |
|-------|------|-----|----------|--------------|
| llama3.1:8b | 4.7GB | 8GB | Fast, general use | Q4_K_M |
| phi3:mini | 2.3GB | 4GB | Extremely fast routing | Q4_K_M |
| gpt-oss:latest | 13GB | 16GB | Comprehension, analysis | MXFP4 |
| qwen3-coder:latest | 18GB | 20GB | Code generation | Q4_K_M |
| deepseek-r1:32b | 20GB | 24GB | Reasoning, research | Q4_K_M |
| deepseek-coder-v2 | 16GB | 18GB | Code-focused tasks | Q4_K_M |
| codellama:13b | 7.4GB | 12GB | Budget code generation | Q4_K_M |

## Setting Up

1. **Pull your chosen models:**
```bash
ollama pull llama3.1:8b
ollama pull gpt-oss:latest
ollama pull qwen3-coder:latest
```

2. **Update Chuchu's model catalog:**
```bash
chu models update
```

This will automatically detect installed Ollama models.

3. **Configure in Neovim:**
```
Ctrl+X (in chat buffer)
Select ollama backend
Configure agent models
```

4. **Or edit `~/.chuchu/setup.yaml` directly**

## Performance Tips

### Speed Up Inference

1. **Use quantized models**: Ollama defaults to Q4_K_M (good balance)
2. **Keep models in RAM**: First run loads model, subsequent runs are fast
3. **Use smaller models for router**: Router is called most frequently
4. **SSD storage**: Models load much faster from SSD

### Memory Management

```bash
# Check running models
ollama ps

# Stop a model to free memory
ollama stop llama3.1:8b

# Models auto-unload after 5 minutes of inactivity
```

### Parallel Model Loading

Ollama can run multiple models simultaneously if you have enough RAM:

```bash
# In separate terminals
ollama run llama3.1:8b
ollama run qwen3-coder:latest
```

## Hybrid Setup: Local + Cloud

Mix local and cloud for best of both worlds:

```yaml
backend:
  ollama:
    type: ollama
    base_url: http://localhost:11434
    default_model: llama3.1:8b
    agent_models:
      router: llama3.1:8b         # Local - fast
      query: llama3.1:8b           # Local - private
      editor: llama3.1:8b          # Local - private
      research: llama3.1:8b        # Local - no API cost
  
  groq:
    type: openai
    base_url: https://api.groq.com/openai/v1
    default_model: llama-33-70b-versatile-128k
    agent_models:
      router: llama-31-8b-instant-128k
      query: gpt-oss-20b-128k
      editor: llama-33-70b-versatile-128k
      research: groq-compound
```

Switch between them in Neovim (`Ctrl+X`) or CLI:
```bash
chu chat --backend ollama    # Private, local
chu chat --backend groq      # Fast, cloud
```

## Troubleshooting

### Model Loading Slowly
- Check if you have enough RAM: `ollama ps`
- Ensure model is on SSD, not HDD
- Close other applications to free memory

### Out of Memory
- Use smaller models or quantizations
- Run one model at a time
- Increase swap space (not recommended for performance)

### Poor Quality Responses
- Try larger models (requires more RAM)
- Use specialized models (e.g., qwen3-coder for code)
- Check model quantization (Q4_K_M is good balance)

## Comparing Local vs Cloud

| Aspect | Ollama (Local) | Groq (Cloud) |
|--------|----------------|--------------|
| Privacy | ✅ Complete | ❌ Code sent to API |
| Cost | ✅ Free | 💰 Pay per token |
| Speed (first run) | ⚠️ Model load time | ✅ Instant |
| Speed (loaded) | ✅ No network | ✅ Very fast |
| Model quality | ⚠️ Limited by RAM | ✅ Largest models |
| Offline | ✅ Works offline | ❌ Requires internet |
| Setup | ⚠️ Download models | ✅ Just API key |

## Model Discovery

List all available Ollama models:
```bash
chu ollama models
```

*(Feature coming soon - will scrape ollama.com/library)*

Install models directly:
```bash
chu ollama install llama3.1:8b
```

*(Feature coming soon)*

## Community Recommendations

Share your Ollama configuration on [GitHub Discussions](https://github.com/yourusername/chuchu/discussions) and help others find the best setup for their hardware!

---

*Running into issues? Check the [troubleshooting guide](/docs/troubleshooting) or ask in [Discord](https://discord.gg/chuchu)*
