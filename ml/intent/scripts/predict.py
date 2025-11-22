import sys
import json
import numpy as np
from pathlib import Path
from sklearn.feature_extraction.text import TfidfVectorizer

# Paths
SCRIPT_DIR = Path(__file__).parent
PROJECT_ROOT = SCRIPT_DIR.parent
MODEL_PATH = PROJECT_ROOT / "models" / "intent_model.json"

def load_model():
    """Load model from JSON."""
    if not MODEL_PATH.exists():
        raise FileNotFoundError(f"Model not found at {MODEL_PATH}")
        
    with open(MODEL_PATH, 'r') as f:
        return json.load(f)

def softmax(x):
    """Compute softmax values for each sets of scores in x."""
    e_x = np.exp(x - np.max(x))
    return e_x / e_x.sum()

def predict(model, message):
    """Make prediction using loaded model."""
    # Recreate TF-IDF vectorizer
    vectorizer = TfidfVectorizer(vocabulary=model['tfidf']['vocabulary'])
    vectorizer.idf_ = np.array(model['tfidf']['idf_weights'])
    
    # Transform message
    X = vectorizer.transform([message])
    X_array = X.toarray()[0]
    
    # Get coefficients and intercepts
    coefs = np.array(model['classifier']['coefficients'])
    intercepts = np.array(model['classifier']['intercepts'])
    classes = model['classifier']['classes']
    
    # Compute scores (logits)
    scores = np.dot(coefs, X_array) + intercepts
    
    # Convert to probabilities
    probs = softmax(scores)
    
    # Get prediction
    pred_idx = int(np.argmax(probs))
    
    return {
        'label': classes[pred_idx],
        'confidence': float(probs[pred_idx]),
        'probabilities': {
            classes[i]: float(probs[i])
            for i in range(len(classes))
        }
    }

def interactive_mode(model):
    print("\n🤖 Intent Classifier (Interactive)")
    print("Type 'exit' to quit")
    print("-" * 40)
    
    while True:
        try:
            message = input("\n> ")
            if message.lower() in ['exit', 'quit']:
                break
            
            result = predict(model, message)
            print(f"Label: {result['label']}")
            print(f"Confidence: {result['confidence']:.2f}")
            print("Probabilities:")
            for label, prob in sorted(result['probabilities'].items(), key=lambda x: x[1], reverse=True):
                print(f"  {label}: {prob:.3f}")
                
        except KeyboardInterrupt:
            break
        except Exception as e:
            print(f"Error: {e}")

if __name__ == "__main__":
    try:
        model = load_model()
        
        if len(sys.argv) > 1:
            # Single prediction
            message = " ".join(sys.argv[1:])
            result = predict(model, message)
            print(json.dumps(result, indent=2))
        else:
            # Interactive mode
            interactive_mode(model)
            
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
