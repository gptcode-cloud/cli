#!/usr/bin/env python3
"""
Train task complexity classification model.

This script:
1. Loads training data from CSV
2. Extracts TF-IDF features
3. Trains a logistic regression classifier
4. Evaluates with cross-validation
5. Exports model weights as JSON for Go embedding
"""

import json
import pandas as pd
from datetime import datetime, timezone
from pathlib import Path
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import cross_val_score, train_test_split
from sklearn.metrics import classification_report, confusion_matrix
import numpy as np

SCRIPT_DIR = Path(__file__).parent
PROJECT_ROOT = SCRIPT_DIR.parent
DATA_PATH = PROJECT_ROOT / "data" / "training_data.csv"
MODEL_OUTPUT = PROJECT_ROOT / "models" / "complexity_model.json"

LABELS = {
    0: "simple",
    1: "complex", 
    2: "multistep"
}

def load_data():
    """Load and validate training data."""
    print(f"📂 Loading data from {DATA_PATH}")
    df = pd.read_csv(DATA_PATH)
    
    required_cols = ['message', 'label']
    if not all(col in df.columns for col in required_cols):
        raise ValueError(f"CSV must contain columns: {required_cols}")
    
    valid_labels = set(LABELS.keys())
    if not set(df['label'].unique()).issubset(valid_labels):
        raise ValueError(f"Labels must be in {valid_labels}")
    
    print(f"✅ Loaded {len(df)} examples")
    print(f"   Simple: {(df['label'] == 0).sum()}")
    print(f"   Complex: {(df['label'] == 1).sum()}")
    print(f"   Multistep: {(df['label'] == 2).sum()}")
    
    return df

def train_model(df):
    """Train TF-IDF + Logistic Regression model."""
    print("\n🔧 Training model...")
    
    X = df['message'].values
    y = df['label'].values
    
    vectorizer = TfidfVectorizer(
        max_features=1000,
        ngram_range=(1, 3),
        min_df=2,
        max_df=0.85,
        lowercase=True,
        strip_accents='unicode'
    )
    
    X_tfidf = vectorizer.fit_transform(X)
    print(f"   Vocabulary size: {len(vectorizer.vocabulary_)}")
    print(f"   Feature matrix shape: {X_tfidf.shape}")
    
    clf = LogisticRegression(
        solver='lbfgs',
        max_iter=1000,
        class_weight='balanced',
        random_state=42
    )
    
    clf.fit(X_tfidf, y)
    print(f"   Trained classifier on {len(X)} examples")
    
    return vectorizer, clf

def evaluate_model(vectorizer, clf, df):
    """Evaluate model performance."""
    print("\n📊 Evaluating model...")
    
    X = df['message'].values
    y = df['label'].values
    X_tfidf = vectorizer.transform(X)
    
    cv_scores = cross_val_score(clf, X_tfidf, y, cv=5, scoring='accuracy')
    print(f"   Cross-validation accuracy: {cv_scores.mean():.3f} (+/- {cv_scores.std() * 2:.3f})")
    
    X_train, X_test, y_train, y_test = train_test_split(
        X_tfidf, y, test_size=0.2, random_state=42, stratify=y
    )
    
    clf_eval = LogisticRegression(
        multi_class='multinomial',
        solver='lbfgs',
        max_iter=1000,
        class_weight='balanced',
        random_state=42
    )
    clf_eval.fit(X_train, y_train)
    
    y_pred = clf_eval.predict(X_test)
    
    print("\n📈 Classification Report:")
    print(classification_report(
        y_test, 
        y_pred,
        target_names=[LABELS[i] for i in sorted(LABELS.keys())],
        digits=3
    ))
    
    print("🔀 Confusion Matrix:")
    cm = confusion_matrix(y_test, y_pred)
    print("           Predicted")
    print("         ", " ".join(f"{LABELS[i]:>8}" for i in sorted(LABELS.keys())))
    for i, row in enumerate(cm):
        print(f"Actual {LABELS[i]:>8}", " ".join(f"{val:>8}" for val in row))
    
    return cv_scores.mean()

def export_model(vectorizer, clf, accuracy, num_examples):
    """Export model weights as JSON for Go embedding."""
    print(f"\n💾 Exporting model to {MODEL_OUTPUT}")
    
    MODEL_OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    
    vocabulary = {str(k): int(v) for k, v in vectorizer.vocabulary_.items()}
    
    model_data = {
        "metadata": {
            "version": "1.0.0",
"trained_at": datetime.now(timezone.utc).isoformat(),
            "accuracy": float(accuracy),
            "num_examples": int(num_examples),
            "vocabulary_size": int(len(vectorizer.vocabulary_)),
            "num_features": int(clf.coef_.shape[1])
        },
        
        "tfidf": {
            "vocabulary": vocabulary,
            "idf_weights": vectorizer.idf_.tolist()
        },
        
        "classifier": {
            "coefficients": clf.coef_.tolist(),
            "intercepts": clf.intercept_.tolist(),
            "classes": [LABELS[int(c)] for c in clf.classes_]
        }
    }
    
    with open(MODEL_OUTPUT, 'w') as f:
        json.dump(model_data, f, indent=2)
    
    file_size_kb = MODEL_OUTPUT.stat().st_size / 1024
    print(f"✅ Model exported successfully")
    print(f"   File size: {file_size_kb:.1f} KB")
    print(f"   Vocabulary: {len(vectorizer.vocabulary_)} terms")
    print(f"   Features: {clf.coef_.shape[1]}")
    
    print("\n🔍 Top predictive features per class:")
    feature_names = vectorizer.get_feature_names_out()
    for i, class_name in enumerate([LABELS[c] for c in sorted(LABELS.keys())]):
        top_indices = np.argsort(clf.coef_[i])[-5:][::-1]
        top_features = [feature_names[idx] for idx in top_indices]
        print(f"   {class_name:>10}: {', '.join(top_features)}")

def main():
    """Main training pipeline."""
    print("=" * 60)
    print("🤖 Task Complexity Detection - Model Training")
    print("=" * 60)
    
    df = load_data()
    
    vectorizer, clf = train_model(df)
    
    accuracy = evaluate_model(vectorizer, clf, df)
    
    export_model(vectorizer, clf, accuracy, len(df))
    
    print("\n" + "=" * 60)
    print("✅ Training complete!")
    print("=" * 60)
    print("\n📝 Next steps:")
    print(f"   1. Review model at: {MODEL_OUTPUT}")
    print(f"   2. Copy to Go embed location if satisfied with accuracy")
    print(f"   3. Implement Go inference code")
    print(f"   4. Rebuild binary: go build -o gptcode cmd/gptcode/main.go")

if __name__ == "__main__":
    main()
