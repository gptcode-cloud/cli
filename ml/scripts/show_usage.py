#!/usr/bin/env python3
import json
from pathlib import Path
from datetime import datetime, timedelta

def main():
    home = Path.home()
    usage_path = home / ".gptcode" / "usage.json"
    
    if not usage_path.exists():
        print("No usage data yet.")
        return
    
    with open(usage_path) as f:
        usage = json.load(f)
    
    today = datetime.now().strftime("%Y-%m-%d")
    yesterday = (datetime.now() - timedelta(days=1)).strftime("%Y-%m-%d")
    
    print("📊 Model Usage Statistics\n")
    
    for date in sorted(usage.keys(), reverse=True)[:7]:
        models_data = usage[date]
        print(f"{'🔥' if date == today else '📅'} {date}")
        
        for model_key in sorted(models_data.keys(), key=lambda k: models_data[k]["requests"], reverse=True):
            model_usage = models_data[model_key]
            requests = model_usage["requests"]
            last_error = model_usage.get("last_error")
            
            status = "❌" if last_error else "✓"
            print(f"  {status} {model_key}: {requests} requests")
            if last_error:
                print(f"     └─ Last error: {last_error[:60]}...")
        print()

if __name__ == "__main__":
    main()
