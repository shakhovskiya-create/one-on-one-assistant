FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    libldap2-dev \
    libsasl2-dev \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application files
COPY *.py .
COPY config.yaml.example config.yaml
COPY start.sh .
RUN chmod +x start.sh

# Create logs directory
RUN mkdir -p /app/logs

CMD ["python", "-u", "connector.py"]
