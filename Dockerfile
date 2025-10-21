FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY src/ ./src/

# Create logs directory
RUN mkdir -p /app/logs

# Create repos directory (will be mounted as volume)
RUN mkdir -p /repos

# Expose webhook port
EXPOSE 8080

# Set Python path
ENV PYTHONPATH=/app

# Run the application
CMD ["python", "-m", "src.main"]
