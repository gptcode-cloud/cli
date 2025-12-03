#!/usr/bin/env python3
"""
Context Detection Model Training

Classifies project context based on language breakdown from go-enry:
- pure_code: Single language dominant (>80%)
- polyglot_balanced: Multiple languages (30-50% each)
- polyglot_scripted: Main language + scripts
- documentation: Markdown/docs dominant (>60%)
- infrastructure: Terraform/Docker/K8s configs
- data_science: Jupyter/R/Python with data files
"""

import json
import os
import sys
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.ensemble import RandomForestClassifier
from sklearn.model_selection import train_test_split
from sklearn.metrics import classification_report, confusion_matrix

# Features from go-enry output
FEATURE_NAMES = [
    "language_count",        # number of languages with >1%
    "primary_ratio",         # percentage of dominant language
    "secondary_ratio",       # percentage of 2nd language
    "has_docs",             # markdown/rst present
    "has_tests",            # test files present
    "has_scripts",          # shell/bash present
    "has_infrastructure",   # docker/terraform present
    "has_data",             # csv/json data files present
]

CONTEXTS = [
    "pure_code",
    "polyglot_balanced",
    "polyglot_scripted",
    "documentation",
    "infrastructure",
    "data_science",
]

def load_dataset(data_dir):
    """Load training data from CSV"""
    csv_path = Path(data_dir) / "training_data.csv"
    if not csv_path.exists():
        print(f"Error: {csv_path} not found")
        print("Creating sample dataset...")
        create_sample_dataset(data_dir)
    
    df = pd.read_csv(csv_path)
    return df

def create_sample_dataset(data_dir):
    """Create initial sample dataset"""
    samples = []
    
    # Pure code examples (10 samples)
    for i in range(10):
        ratio = 0.85 + (i * 0.01)
        samples.append([1, ratio, 0.05, 1, 1, 0, 0, 0, "pure_code"])
    
    # Polyglot balanced (10 samples)
    for i in range(10):
        primary = 0.35 + (i * 0.02)
        secondary = 0.30 + (i * 0.01)
        samples.append([3, primary, secondary, 1, 1, 1, 0, 0, "polyglot_balanced"])
    
    # Polyglot scripted (10 samples)
    for i in range(10):
        primary = 0.60 + (i * 0.01)
        samples.append([2, primary, 0.25, 1, 1, 1, 0, 0, "polyglot_scripted"])
    
    # Documentation (10 samples)
    for i in range(10):
        ratio = 0.75 + (i * 0.01)
        samples.append([1, ratio, 0.10, 1, 0, 0, 0, 0, "documentation"])
    
    # Infrastructure (10 samples)
    for i in range(10):
        primary = 0.45 + (i * 0.02)
        samples.append([2, primary, 0.35, 1, 0, 1, 1, 0, "infrastructure"])
    
    # Data science (10 samples)
    for i in range(10):
        primary = 0.60 + (i * 0.01)
        samples.append([2, primary, 0.25, 1, 1, 0, 0, 1, "data_science"])
    
    df = pd.DataFrame(samples, columns=FEATURE_NAMES + ["context"])
    
    csv_path = Path(data_dir) / "training_data.csv"
    df.to_csv(csv_path, index=False)
    print(f"Created sample dataset: {csv_path}")
    print(f"Samples: {len(df)}")

def train_model(data_dir, model_dir):
    """Train Random Forest classifier"""
    df = load_dataset(data_dir)
    
    X = df[FEATURE_NAMES].values
    y = df["context"].values
    
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42
    )
    
    print(f"Training samples: {len(X_train)}")
    print(f"Test samples: {len(X_test)}")
    
    clf = RandomForestClassifier(
        n_estimators=100,
        max_depth=10,
        random_state=42,
        class_weight="balanced"
    )
    
    clf.fit(X_train, y_train)
    
    y_pred = clf.predict(X_test)
    
    print("\nClassification Report:")
    print(classification_report(y_test, y_pred))
    
    print("\nConfusion Matrix:")
    print(confusion_matrix(y_test, y_pred))
    
    print("\nFeature Importances:")
    for i, importance in enumerate(clf.feature_importances_):
        print(f"  {FEATURE_NAMES[i]:20s}: {importance:.3f}")
    
    # Export model to JSON
    export_model(clf, model_dir)
    
    return clf

def export_model(clf, model_dir):
    """Export model to JSON for Go embedding"""
    model_data = {
        "type": "random_forest",
        "n_estimators": clf.n_estimators,
        "classes": clf.classes_.tolist(),
        "feature_names": FEATURE_NAMES,
        "trees": []
    }
    
    for tree in clf.estimators_:
        tree_data = {
            "feature": tree.tree_.feature.tolist(),
            "threshold": tree.tree_.threshold.tolist(),
            "value": tree.tree_.value.tolist(),
            "children_left": tree.tree_.children_left.tolist(),
            "children_right": tree.tree_.children_right.tolist(),
        }
        model_data["trees"].append(tree_data)
    
    model_path = Path(model_dir) / "context_detection.json"
    with open(model_path, "w") as f:
        json.dump(model_data, f, indent=2)
    
    print(f"\nModel exported: {model_path}")
    print(f"Size: {model_path.stat().st_size / 1024:.1f} KB")

def main():
    script_dir = Path(__file__).parent
    project_root = script_dir.parent
    data_dir = project_root / "data"
    model_dir = project_root / "models"
    
    model_dir.mkdir(exist_ok=True)
    
    print("Training Context Detection Model")
    print("=" * 50)
    
    clf = train_model(str(data_dir), str(model_dir))
    
    print("\nTraining complete!")

if __name__ == "__main__":
    main()
