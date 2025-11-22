import sys
import json
import pandas as pd
from pathlib import Path
from sklearn.metrics import classification_report, confusion_matrix, accuracy_score
import numpy as np

SCRIPT_DIR = Path(__file__).parent
PROJECT_ROOT = SCRIPT_DIR.parent
MODEL_PATH = PROJECT_ROOT / "models" / "router_model.json"
EVAL_DATA_PATH = PROJECT_ROOT / "data" / "eval.csv"

LABELS = {
    0: "router",
    1: "query",
    2: "editor",
    3: "research",
    4: "review"
}
LABEL_TO_ID = {v: k for k, v in LABELS.items()}

def load_model():
    with open(MODEL_PATH) as f:
        return json.load(f)

def tokenize(text):
    import re
    text = text.lower()
    text = re.sub(r'[^a-z0-9\s]', ' ', text)
    return text.split()

def ngrams(tokens, n):
    if n == 1:
        return tokens
    return [' '.join(tokens[i:i+n]) for i in range(len(tokens)-n+1)]

def vectorize(text, vocab, idf):
    tokens = tokenize(text)
    uni = ngrams(tokens, 1)
    bi = ngrams(tokens, 2)
    tri = ngrams(tokens, 3)
    all_grams = uni + bi + tri
    
    counts = {}
    for gram in all_grams:
        if gram in vocab:
            idx = vocab[gram]
            counts[idx] = counts.get(idx, 0) + 1
    
    vec = [0.0] * len(idf)
    for idx, count in counts.items():
        vec[idx] = count * idf[idx]
    
    norm = sum(v*v for v in vec) ** 0.5
    if norm > 0:
        vec = [v/norm for v in vec]
    
    return vec

def predict(text, model):
    vocab = model['tfidf']['vocabulary']
    idf = model['tfidf']['idf_weights']
    coefs = model['classifier']['coefficients']
    intercepts = model['classifier']['intercepts']
    classes = model['classifier']['classes']
    
    vec = vectorize(text, vocab, idf)
    
    logits = []
    for i, coef in enumerate(coefs):
        score = intercepts[i]
        for j, val in enumerate(vec):
            if j < len(coef):
                score += coef[j] * val
        logits.append(score)
    
    exp_logits = [np.exp(x - max(logits)) for x in logits]
    sum_exp = sum(exp_logits)
    probs = [e / sum_exp for e in exp_logits]
    
    best_idx = probs.index(max(probs))
    return classes[best_idx], max(probs)

def evaluate(model, df):
    y_true = []
    y_pred = []
    confidences = []
    
    for _, row in df.iterrows():
        label_str = row['label']
        pred, conf = predict(row['message'], model)
        
        y_true.append(label_str)
        y_pred.append(pred)
        confidences.append(conf)
    
    accuracy = accuracy_score(y_true, y_pred)
    avg_confidence = np.mean(confidences)
    
    print(f"\n📊 Evaluation Results")
    print("=" * 60)
    print(f"Accuracy: {accuracy:.3f}")
    print(f"Average Confidence: {avg_confidence:.3f}")
    print(f"\n{classification_report(y_true, y_pred, digits=3)}")
    
    print("\n🔢 Confusion Matrix:")
    labels = sorted(set(y_true) | set(y_pred))
    cm = confusion_matrix(y_true, y_pred, labels=labels)
    
    print(f"\n{'':>12}", end='')
    for lbl in labels:
        print(f"{lbl:>12}", end='')
    print()
    
    for i, lbl in enumerate(labels):
        print(f"{lbl:>12}", end='')
        for j in range(len(labels)):
            print(f"{cm[i][j]:>12}", end='')
        print()
    
    print("\n" + "=" * 60)
    
    low_conf = [i for i, c in enumerate(confidences) if c < 0.7]
    if low_conf:
        print(f"\n⚠️  {len(low_conf)} predictions with confidence < 0.7:")
        for idx in low_conf[:5]:
            row = df.iloc[idx]
            pred, conf = predict(row['message'], model)
            print(f"  {conf:.2f} | true={row['label']:8s} pred={pred:8s} | \"{row['message']}\"")
        if len(low_conf) > 5:
            print(f"  ... and {len(low_conf) - 5} more")

def main():
    eval_file = sys.argv[1] if len(sys.argv) > 1 else str(EVAL_DATA_PATH)
    
    if not Path(MODEL_PATH).exists():
        print(f"❌ Model not found: {MODEL_PATH}")
        print(f"   Run 'chu ml train router_agent' first")
        return 1
    
    if not Path(eval_file).exists():
        print(f"❌ Evaluation data not found: {eval_file}")
        return 1
    
    print(f"📂 Loading model from {MODEL_PATH}")
    model = load_model()
    
    print(f"📂 Loading eval data from {eval_file}")
    df = pd.read_csv(eval_file)
    
    if 'message' not in df.columns or 'label' not in df.columns:
        print(f"❌ CSV must contain 'message' and 'label' columns")
        return 1
    
    print(f"   Loaded {len(df)} examples")
    print(f"   Distribution: {df['label'].value_counts().to_dict()}")
    
    evaluate(model, df)
    return 0

if __name__ == "__main__":
    sys.exit(main())
