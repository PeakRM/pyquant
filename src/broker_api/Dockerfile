# Test API Dockerfile
FROM python:3.12.1

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .
CMD ["uvicorn", "broker_api:app", "--host", "0.0.0.0", "--port", "8000"]