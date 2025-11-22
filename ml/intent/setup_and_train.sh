#!/bin/bash
set -e

# Directory setup
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
VENV_DIR="${SCRIPT_DIR}/venv"
REQUIREMENTS_FILE="${SCRIPT_DIR}/requirements.txt"

echo "🤖 Intent Classifier Model Setup"
echo "================================================"

# Create venv if not exists
if [ ! -d "$VENV_DIR" ]; then
    echo "📦 Creating virtual environment..."
    python3 -m venv "$VENV_DIR"
fi

# Activate venv
source "$VENV_DIR/bin/activate"

# Install dependencies
if [ -f "$REQUIREMENTS_FILE" ]; then
    echo "⬇️  Installing dependencies..."
    pip install -q -r "$REQUIREMENTS_FILE"
else
    echo "❌ Requirements file not found at $REQUIREMENTS_FILE"
    exit 1
fi

# Run training
echo "🚀 Starting training..."
python3 "${SCRIPT_DIR}/scripts/train.py"

# Run prediction test (optional, if predict script exists)
if [ -f "${SCRIPT_DIR}/scripts/predict.py" ]; then
    echo ""
    echo "🧪 Testing model..."
    python3 "${SCRIPT_DIR}/scripts/predict.py"
fi

echo ""
echo "✅ Setup and training complete!"
