#!/usr/bin/env python3
"""
Process user feedback from ~/.gptcode/feedback/*.json into training data.

Converts feedback events into intent classification training examples.
Bad feedback with task info can help improve model accuracy.
"""
import json
import csv
from pathlib import Path
from collections import Counter
from datetime import datetime

SCRIPT_DIR = Path(__file__).parent
PROJECT_ROOT = SCRIPT_DIR.parent
FEEDBACK_DIR = Path.home() / ".gptcode" / "feedback"
OUTPUT_CSV = PROJECT_ROOT / "data" / "training_data_feedback.csv"

def infer_intent_from_task(task):
    """
    Infer intent from task description.
    
    This is a heuristic - ideally user would provide correct intent in feedback.
    """
    task_lower = task.lower()
    
    # Editor patterns
    edit_keywords = [
        'add ', 'create ', 'implement ', 'fix ', 'refactor ', 'update ',
        'modify ', 'change ', 'remove ', 'delete ', 'write '
    ]
    
    # Query patterns  
    query_keywords = [
        'how ', 'what ', 'where ', 'show ', 'list ', 'find ', 'search ',
        'explain ', 'display ', 'get ', 'check ', 'run ', 'execute ',
        'rodar ', 'como ', 'onde '  # Portuguese
    ]
    
    # Research patterns
    research_keywords = [
        'best practices', 'compare ', 'research ', 'investigate ',
        'documentation ', 'tutorial ', 'example '
    ]
    
    # Review patterns
    review_keywords = [
        'review ', 'analyze ', 'audit ', 'check for bugs',
        'security ', 'vulnerability '
    ]
    
    # Check in order of specificity
    for keyword in edit_keywords:
        if keyword in task_lower:
            return 'editor'
    
    for keyword in review_keywords:
        if keyword in task_lower:
            return 'review'
    
    for keyword in research_keywords:
        if keyword in task_lower:
            return 'research'
    
    for keyword in query_keywords:
        if keyword in task_lower:
            return 'query'
    
    # Default for shell/command execution
    return 'query'

def load_feedback_events():
    """Load all feedback events from ~/.gptcode/feedback/*.json"""
    if not FEEDBACK_DIR.exists():
        print(f"No feedback directory found: {FEEDBACK_DIR}")
        return []
    
    events = []
    for json_file in FEEDBACK_DIR.glob("*.json"):
        try:
            with open(json_file, 'r', encoding='utf-8') as f:
                file_events = json.load(f)
                if isinstance(file_events, list):
                    events.extend(file_events)
                elif isinstance(file_events, dict):
                    events.append(file_events)
        except Exception as e:
            print(f"Warning: Failed to load {json_file}: {e}")
            continue
    
    return events

def process_feedback_to_training_data():
    """Convert feedback events to training examples."""
    print(f"📂 Loading feedback from {FEEDBACK_DIR}")
    
    events = load_feedback_events()
    
    if not events:
        print("   No feedback events found")
        return []
    
    print(f"   Loaded {len(events)} feedback events")
    
    # Filter and convert to training examples
    training_examples = []
    
    for event in events:
        task = event.get('task', '') or event.get('context', '')
        
        if not task or len(task) < 5:
            continue
        
        sentiment = event.get('sentiment', '')
        
        # For now, we only use bad feedback to generate training data
        # Good feedback confirms existing behavior
        if sentiment == 'bad' and task:
            # Infer the correct intent
            # In future, we could prompt user for correct intent
            intent = infer_intent_from_task(task)
            
            training_examples.append({
                'message': task,
                'label': intent,
                'source': 'feedback',
                'timestamp': event.get('timestamp', '')
            })
    
    print(f"   Generated {len(training_examples)} training examples from bad feedback")
    
    # Count by intent
    intent_counts = Counter(ex['label'] for ex in training_examples)
    print(f"\n   Intent distribution:")
    for intent, count in sorted(intent_counts.items()):
        print(f"     {intent}: {count}")
    
    return training_examples

def save_training_data(examples):
    """Save training examples to CSV."""
    if not examples:
        print("\n⚠️  No training examples to save")
        return
    
    OUTPUT_CSV.parent.mkdir(parents=True, exist_ok=True)
    
    with open(OUTPUT_CSV, 'w', newline='', encoding='utf-8') as f:
        writer = csv.DictWriter(f, fieldnames=['message', 'label'])
        writer.writeheader()
        for ex in examples:
            writer.writerow({
                'message': ex['message'],
                'label': ex['label']
            })
    
    print(f"\n✅ Saved {len(examples)} examples to {OUTPUT_CSV}")

def main():
    print("🤖 Feedback to Training Data Processor")
    print("=" * 60)
    print("Processes user feedback into intent classification training data")
    print("=" * 60)
    
    try:
        examples = process_feedback_to_training_data()
        save_training_data(examples)
        
        print("\n" + "=" * 60)
        print("✨ Processing complete!")
        
        if examples:
            print("\nNext steps:")
            print("  1. Review training_data_feedback.csv")
            print("  2. Run: chu ml train intent")
            print("  3. Model will auto-merge feedback examples")
        else:
            print("\nNo feedback to process. Record feedback with:")
            print("  chu feedback bad --task='your task' --agent=editor")
        
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
        exit(1)

if __name__ == "__main__":
    main()
