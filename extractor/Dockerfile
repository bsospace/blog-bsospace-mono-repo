FROM mcr.microsoft.com/playwright/python:v1.47.0-jammy

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# OCR + PDF tools เพิ่มเติม
USER root
RUN apt-get update && apt-get install -y --no-install-recommends \
    tesseract-ocr tesseract-ocr-eng tesseract-ocr-tha \
    poppler-utils \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
# ภาพนี้มี playwright อยู่แล้ว ไม่ต้องลง browser ซ้ำ
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 5002
CMD ["python", "extractor.py"]