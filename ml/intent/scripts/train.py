import json
import pandas as pd
from datetime import datetime, timezone
from pathlib import Path
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import cross_val_score, train_test_split
from sklearn.metrics import classification_report, confusion_matrix
import numpy as np

# Paths
SCRIPT_DIR = Path(__file__).parent
PROJECT_ROOT = SCRIPT_DIR.parent
DATA_PATH = PROJECT_ROOT / "data" / "training_data.csv"
MODEL_OUTPUT = PROJECT_ROOT / "models" / "router_model.json"

# Labels
LABELS = {
    0: "router",
    1: "query",
    2: "editor",
    3: "research",
    4: "review"
}

LABEL_TO_ID = {v: k for k, v in LABELS.items()}

def load_data():
    """Load and validate training data."""
    print(f"📂 Loading data from {DATA_PATH}")
    df = pd.read_csv(DATA_PATH)
    
    required_cols = ['message', 'label']
    if not all(col in df.columns for col in required_cols):
        raise ValueError(f"CSV must contain columns: {required_cols}")
        
    # Map string labels to IDs
    if df['label'].dtype == 'O':
        df['label_id'] = df['label'].map(LABEL_TO_ID)
        # Check for invalid labels
        if df['label_id'].isnull().any():
            invalid = df[df['label_id'].isnull()]['label'].unique()
            raise ValueError(f"Invalid labels found: {invalid}. Allowed: {list(LABELS.values())}")
        df['label'] = df['label_id']
    
    print(f"   Loaded {len(df)} examples")
    print(f"   Class distribution:\n{df['label'].map(LABELS).value_counts()}")
    return df

def train_model(df):
    """Train TF-IDF + Logistic Regression model."""
    print("\n🧠 Training model...")
    
    X = df['message'].values
    y = df['label'].values
    
    vectorizer = TfidfVectorizer(
        max_features=1000,
        ngram_range=(1, 3),
        min_df=1,
        max_df=0.9,
        lowercase=True,
        strip_accents='unicode'
    )
    
    X_tfidf = vectorizer.fit_transform(X)
    print(f"   Vocabulary size: {len(vectorizer.vocabulary_)}")
    
    clf = LogisticRegression(
        multi_class='multinomial',
        solver='lbfgs',
        max_iter=1000,
        class_weight='balanced',
        C=1.0,
        random_state=42
    )
    
    clf.fit(X_tfidf, y)
    print("   Training complete")
    
    return vectorizer, clf

def evaluate_model(vectorizer, clf, df):
    """Evaluate model performance."""
    print("\n📊 Evaluating model...")
    
    X = df['message'].values
    y = df['label'].values
    X_tfidf = vectorizer.transform(X)
    
    cv_scores = cross_val_score(clf, X_tfidf, y, cv=3, scoring='accuracy')
    print(f"   Cross-validation accuracy: {cv_scores.mean():.3f} (+/- {cv_scores.std() * 2:.3f})")
    
    X_train, X_test, y_train, y_test = train_test_split(
        X_tfidf, y, test_size=0.2, random_state=42, stratify=y
    )
    
    clf_eval = LogisticRegression(
        multi_class='multinomial',
        solver='lbfgs',
        max_iter=1000,
        class_weight='balanced',
        C=1.0,
        random_state=42
    )
    clf_eval.fit(X_train, y_train)
    y_pred = clf_eval.predict(X_test)
    
    print("\n📈 Classification Report:")
    print(classification_report(
        y_test, 
        y_pred, 
        target_names=[LABELS[i] for i in sorted(set(y_test) | set(y_pred))],
        digits=3
    ))
    
    return cv_scores.mean()

def export_model(vectorizer, clf, accuracy, num_examples):
    """Export model weights as JSON for Go embedding."""
    print(f"\n💾 Exporting model to {MODEL_OUTPUT}")
    
    MODEL_OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    
    vocabulary = {str(k): int(v) for k, v in vectorizer.vocabulary_.items()}
    
    model_data = {
        "metadata": {
            "version": "1.0.0",
            "model_type": "router_agent",
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
    
    print(f"✅ Model exported successfully")

if __name__ == "__main__":
    print("🤖 Router Agent - Model Training")
    print("=" * 60)
    
    try:
        df = load_data()
        vectorizer, clf = train_model(df)
        accuracy = evaluate_model(vectorizer, clf, df)
        export_model(vectorizer, clf, accuracy, len(df))
        print("\n" + "=" * 60)
        print("✨ Done!")
    except Exception as e:
        print(f"\n❌ Error: {e}")
        exit(1)
