# 🧠 RAG-SearchBot (Backend)

RAG-SearchBot เป็นระบบ Chatbot สำหรับตอบคำถามจากบทความบนบล็อก โดยใช้เทคนิค **RAG (Retrieval-Augmented Generation)** ที่ผสานการดึงข้อมูลที่เกี่ยวข้องจากฐานความรู้ แล้วส่งให้ LLM (เช่น LLaMA3) เพื่อตอคำตอบที่แม่นยำ

---

## ✨ Features

- 🧾 **PDF Upload**: รองรับอัปโหลดบทความในรูปแบบ PDF
- 📚 **Text Extraction**: ใช้ Flask (PyMuPDF + OCR) แปลง PDF เป็นข้อความ
- 🧠 **Text Chunking + Embedding**: แบ่งข้อความเป็น Chunk แล้วฝัง (Embed) ด้วย Ollama API
- 🔍 **Context Retrieval**: ดึง Context ที่เกี่ยวข้องด้วย Cosine Similarity
- 🤖 **LLM Answering**: ใช้ LLaMA3 (via Ollama) ตอบคำถามจาก Context
- 🗃️ **PostgreSQL + Redis**: จัดการฐานข้อมูลผู้ใช้, โพสต์, Embedding, และแคช
- 🐳 **Dockerized**: รองรับ Dev/Prod ด้วย Docker Compose

---

## 📦 Tech Stack

| Layer        | Tech                                |
| ------------ | ----------------------------------- |
| Language     | Go 1.22+, Python 3.10+              |
| Backend      | [Gin](https://gin-gonic.com/), GORM |
| Vector Embed | Ollama (LLaMA3, Typhoon)            |
| Database     | PostgreSQL + pgAdmin                |
| Caching      | Redis                               |
| Extraction   | Flask + PyMuPDF + pytesseract (OCR) |
| Dev Tools    | Air (Hot Reload), Docker Compose    |

---

## 🏁 Getting Started

### 1. Clone Project

```bash
git clone https://github.com/boytur/rag-searchbot.git

cd rag-searchbot
```

📂 Structure:
```
backend/
├── cmd/server # Main app entry
├── internal/ # Business logic
├── handlers/ # Gin route handlers
├── models/ # GORM models
├── storage/ # Embedding in-memory store
├── config/ # Configs
├── utils/  # helpers
├── air.toml # Hot reload config
extractor/
├── extractor.py # Flask OCR & Text extraction

```