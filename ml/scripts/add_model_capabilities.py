#!/usr/bin/env python3
"""
Add capabilities to model catalog based on known model characteristics.
"""
import json
from pathlib import Path

# Known capabilities for specific models/families
MODEL_CAPABILITIES = {
    # Groq models
    "llama-3.3-70b-versatile": {
        "supports_tools": True,
        "supports_file_operations": True,
        "supports_code_execution": False,
        "notes": "Full tool calling support for file editing"
    },
    "llama-3.1-8b-instant": {
        "supports_tools": True,
        "supports_file_operations": True,
        "supports_code_execution": False,
        "notes": "Fast model with basic tool support"
    },
    "moonshotai/kimi-k2-instruct": {
        "supports_tools": False,
        "supports_file_operations": False,
        "supports_code_execution": False,
        "notes": "No tool calling support via Groq"
    },
    "groq/compound": {
        "supports_tools": True,
        "supports_file_operations": False,
        "supports_code_execution": True,
        "notes": "Has web_search, wolfram, code_interpreter but not file tools"
    },
    
    # OpenRouter models  
    "google/gemini-2.0-flash-exp:free": {
        "supports_tools": True,
        "supports_file_operations": True,
        "supports_code_execution": False,
        "notes": "Supports tool calling when available"
    },
    "moonshotai/kimi-k2:free": {
        "supports_tools": True,
        "supports_file_operations": True,
        "supports_code_execution": False,
        "notes": "Full tool support via OpenRouter"
    },
    
    # Ollama models
    "qwen3-coder:latest": {
        "supports_tools": True,
        "supports_file_operations": True,
        "supports_code_execution": False,
        "notes": "Code-specialized model with tool support"
    },
    "gpt-oss:latest": {
        "supports_tools": True,
        "supports_file_operations": True,
        "supports_code_execution": False,
        "notes": "Large 120B model, good for complex tasks"
    },
    "deepseek-r1:latest": {
        "supports_tools": True,
        "supports_file_operations": True,
        "supports_code_execution": False,
        "notes": "Reasoning model with tool support"
    }
}

def add_capabilities_to_catalog(catalog_path: Path):
    """Add capabilities field to all models in catalog."""
    print(f"Loading catalog from {catalog_path}")
    with open(catalog_path) as f:
        catalog = json.load(f)
    
    updated_count = 0
    total_count = 0
    
    # Iterate through each backend
    for backend_name, backend_data in catalog.items():
        if "models" not in backend_data:
            continue
            
        models = backend_data["models"]
        print(f"\nProcessing {backend_name}: {len(models)} models")
        
        for model in models:
            total_count += 1
            model_id = model.get("id", "")
            model_name = model.get("name", "")
            
            # Check if we have known capabilities for this model
            capabilities = None
            
            # Try exact match first
            if model_id in MODEL_CAPABILITIES:
                capabilities = MODEL_CAPABILITIES[model_id]
            elif model_name in MODEL_CAPABILITIES:
                capabilities = MODEL_CAPABILITIES[model_name]
            else:
                # Try partial matches for model families
                for pattern, caps in MODEL_CAPABILITIES.items():
                    if pattern in model_id or pattern in model_name:
                        capabilities = caps
                        break
            
            # Add capabilities if found
            if capabilities:
                model["capabilities"] = capabilities
                updated_count += 1
                print(f"  ✓ {model_name}: {capabilities.get('notes', '')}")
            else:
                # Set unknown capabilities
                model["capabilities"] = {
                    "supports_tools": None,
                    "supports_file_operations": None,
                    "supports_code_execution": None,
                    "notes": "Capabilities unknown - needs testing"
                }
    
    # Save updated catalog
    print(f"\n📊 Updated {updated_count}/{total_count} models")
    print(f"💾 Saving to {catalog_path}")
    
    with open(catalog_path, 'w') as f:
        json.dump(catalog, f, indent=2)
    
    print("✅ Done!")

if __name__ == "__main__":
    catalog_path = Path.home() / ".gptcode" / "models_catalog.json"
    add_capabilities_to_catalog(catalog_path)
