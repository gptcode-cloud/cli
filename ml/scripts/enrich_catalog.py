#!/usr/bin/env python3
import json
import sys
from pathlib import Path

ENRICHMENT = {
    "openrouter": {
        "google/gemini-2.0-flash-exp:free": {
            "cost_per_1m": 0,
            "rate_limit_daily": 1000,
            "context_window": 1000000,
            "tokens_per_sec": 150
        },
        "meta-llama/llama-3.2-3b-instruct:free": {
            "cost_per_1m": 0,
            "rate_limit_daily": 1000,
            "context_window": 128000,
            "tokens_per_sec": 200
        },
        "meta-llama/llama-3.3-70b-instruct:free": {
            "cost_per_1m": 0,
            "rate_limit_daily": 1000,
            "context_window": 128000,
            "tokens_per_sec": 120
        },
        "moonshotai/kimi-k2:free": {
            "cost_per_1m": 0,
            "rate_limit_daily": 1000,
            "context_window": 262144,
            "tokens_per_sec": 100
        },
        "qwen/qwen3-coder:free": {
            "cost_per_1m": 0,
            "rate_limit_daily": 1000,
            "context_window": 32768,
            "tokens_per_sec": 140
        },
        "qwen/qwen-2.5-coder-32b-instruct": {
            "cost_per_1m": 0.3,
            "rate_limit_daily": 10000,
            "context_window": 32768,
            "tokens_per_sec": 120
        },
        "meta-llama/llama-3.1-8b-instruct": {
            "cost_per_1m": 0.02,
            "rate_limit_daily": 10000,
            "context_window": 128000,
            "tokens_per_sec": 180
        }
    },
    "groq": {
        "llama-3.3-70b-versatile": {
            "cost_per_1m": 0.59,
            "rate_limit_daily": 14400,
            "context_window": 131072,
            "tokens_per_sec": 300
        },
        "llama-3.1-8b-instant": {
            "cost_per_1m": 0.05,
            "rate_limit_daily": 14400,
            "context_window": 131072,
            "tokens_per_sec": 500
        },
        "groq/compound": {
            "cost_per_1m": 0,
            "rate_limit_daily": 14400,
            "context_window": 131072,
            "tokens_per_sec": 250
        }
    },
    "ollama": {
        "qwen2.5-coder:32b": {
            "cost_per_1m": 0,
            "rate_limit_daily": 999999,
            "context_window": 32768,
            "tokens_per_sec": 50
        },
        "llama3.2:3b": {
            "cost_per_1m": 0,
            "rate_limit_daily": 999999,
            "context_window": 128000,
            "tokens_per_sec": 80
        }
    }
}

def main():
    home = Path.home()
    catalog_path = home / ".gptcode" / "models_catalog.json"
    
    with open(catalog_path) as f:
        catalog = json.load(f)
    
    enriched = 0
    for backend, models_data in ENRICHMENT.items():
        if backend not in catalog:
            continue
            
        for model_dict in catalog[backend]["models"]:
            model_id = model_dict.get("id")
            if model_id in models_data:
                enrichment = models_data[model_id]
                model_dict.update(enrichment)
                enriched += 1
                print(f"✓ Enriched {backend}/{model_id}")
    
    with open(catalog_path, "w") as f:
        json.dump(catalog, f, indent=2)
    
    print(f"\n✓ Enriched {enriched} models")

if __name__ == "__main__":
    main()
