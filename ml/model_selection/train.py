#!/usr/bin/env python3
"""
Train ML model to predict best LLM model for a given task.

Input features:
- action (edit, review, plan, research)
- language (go, python, typescript, etc)
- complexity (simple, medium, complex)

Output: model_id (best model to use)

Training data comes from feedback history.
"""
import json
import os
from pathlib import Path
from collections import Counter

import numpy as np
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import LabelEncoder
from sklearn.metrics import accuracy_score, classification_report

def load_feedback_data():
    """Load all feedback events from ~/.gptcode/feedback/"""
    home = Path.home()
    feedback_dir = home / ".gptcode" / "feedback"
    
    if not feedback_dir.exists():
        print(f"No feedback directory found at {feedback_dir}")
        return []
    
    all_events = []
    for file_path in feedback_dir.glob("*.json"):
        try:
            with open(file_path) as f:
                events = json.load(f)
                all_events.extend(events)
        except Exception as e:
            print(f"Error loading {file_path}: {e}")
    
    return all_events

def convert_to_training_data(events):
    """Convert feedback events to ML training format"""
    data = []
    
    for event in events:
        # Skip events without required fields
        if not event.get('model') or not event.get('agent') or not event.get('task'):
            continue
        
        # Map agent to action
        agent = event['agent'].lower()
        action_map = {
            'editor': 'edit',
            'reviewer': 'review',
            'validator': 'review',
            'planner': 'plan',
            'research': 'research'
        }
        action = action_map.get(agent)
        if not action:
            continue
        
        # Extract language from task
        task = event.get('task', '').lower()
        language = 'unknown'
        if '.go' in task:
            language = 'go'
        elif '.py' in task:
            language = 'python'
        elif '.ts' in task or '.js' in task:
            language = 'typescript'
        elif '.ex' in task or '.exs' in task:
            language = 'elixir'
        
        # Determine complexity
        complexity = 'simple'
        if any(word in task for word in ['refactor', 'reorganize', 'complex', 'system']):
            complexity = 'complex'
        elif any(word in task for word in ['multiple', 'all', 'entire']):
            complexity = 'medium'
        
        # Success from sentiment
        success = event.get('sentiment') == 'good'
        
        # Only learn from successful executions
        # (We know what works, not just what doesn't)
        if success:
            data.append({
                'action': action,
                'language': language,
                'complexity': complexity,
                'model': event['model'],
                'backend': event.get('backend', 'unknown')
            })
    
    return data

def train_model_selector(data):
    """Train random forest to predict best model"""
    if len(data) < 10:
        print(f"Not enough training data: {len(data)} samples")
        print("Need at least 10 successful task executions with feedback")
        return None
    
    # Prepare features
    X = []
    y = []
    
    for sample in data:
        features = [
            sample['action'],
            sample['language'],
            sample['complexity']
        ]
        X.append(features)
        y.append(sample['model'])
    
    # Encode categorical features
    action_encoder = LabelEncoder()
    language_encoder = LabelEncoder()
    complexity_encoder = LabelEncoder()
    model_encoder = LabelEncoder()
    
    X_encoded = np.column_stack([
        action_encoder.fit_transform([x[0] for x in X]),
        language_encoder.fit_transform([x[1] for x in X]),
        complexity_encoder.fit_transform([x[2] for x in X])
    ])
    
    y_encoded = model_encoder.fit_transform(y)
    
    # Train/test split
    X_train, X_test, y_train, y_test = train_test_split(
        X_encoded, y_encoded, test_size=0.2, random_state=42
    )
    
    # Train model
    clf = RandomForestClassifier(n_estimators=100, random_state=42)
    clf.fit(X_train, y_train)
    
    # Evaluate
    y_pred = clf.predict(X_test)
    accuracy = accuracy_score(y_test, y_pred)
    
    print(f"\n=== Model Selection Classifier ===")
    print(f"Training samples: {len(X_train)}")
    print(f"Test samples: {len(X_test)}")
    print(f"Accuracy: {accuracy:.2%}")
    
    print(f"\nModel distribution:")
    model_counts = Counter(y)
    for model, count in model_counts.most_common():
        print(f"  {model}: {count} samples")
    
    # Save model and encoders
    output_dir = Path(__file__).parent / "models"
    output_dir.mkdir(exist_ok=True)
    
    model_data = {
        'feature_importances': clf.feature_importances_.tolist(),
        'action_classes': action_encoder.classes_.tolist(),
        'language_classes': language_encoder.classes_.tolist(),
        'complexity_classes': complexity_encoder.classes_.tolist(),
        'model_classes': model_encoder.classes_.tolist(),
        'accuracy': float(accuracy),
        'n_samples': len(data)
    }
    
    output_file = output_dir / "model_selector.json"
    with open(output_file, 'w') as f:
        json.dump(model_data, f, indent=2)
    
    print(f"\nModel saved to: {output_file}")
    
    return clf, model_data

def main():
    print("Loading feedback data...")
    events = load_feedback_data()
    print(f"Loaded {len(events)} feedback events")
    
    print("\nConverting to training data...")
    training_data = convert_to_training_data(events)
    print(f"Extracted {len(training_data)} training samples")
    
    if len(training_data) < 10:
        print("\n⚠️  Not enough data to train!")
        print("Execute more tasks successfully and they will be automatically recorded.")
        print("After ~20-30 successful tasks, run this again.")
        return
    
    print("\nTraining model...")
    train_model_selector(training_data)
    print("\n✅ Done!")

if __name__ == "__main__":
    main()
