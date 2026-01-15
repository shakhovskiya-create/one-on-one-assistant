#!/bin/bash
# Start On-Prem Connector

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check Python version
python3 --version || { echo "Python 3 is required"; exit 1; }

# Create virtual environment if not exists
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
source venv/bin/activate

# Install dependencies
pip install -q -r requirements.txt

# Run connector
echo "Starting On-Prem Connector..."
echo "Press Ctrl+C to stop"
echo ""
python connector.py
