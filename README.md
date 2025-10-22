# ✨ BSO Blog – Be Simple but Outstanding 📝

**BSO Blog** is a collaborative blogging platform created by Software Engineering students, aimed at sharing knowledge, cutting-edge techniques, and real-world experiences. The system is designed with professionalism in mind, featuring CI/CD pipelines, clean backend architecture, and a modern user interface.

> “Be Simple but Outstanding.”

---

## 📌 Features

- 📰 **Write & Share Blog Posts** – Supports Markdown with syntax highlighting
- 🪄 **Real-Time Editing** – Built with Tiptap Editor + Image Upload with Preview
- 🧠 **RAG-powered Search** – Search blog content using Retrieval-Augmented Generation (Coming Soon)
- 🔐 **Authentication** – Supports OAuth (Google/GitHub, Discord)
- 🚀 **CI/CD** – Jenkins, Jest, SonarQube, and Docker for deployment
- 📊 **Dashboard** – Admin panel for posts and analytics
- 🌐 **Multilang Ready** – Supports both EN and TH

---

## 🏗 Tech Stack

| Layer         | Tech Stack                                  |
| ------------- | ------------------------------------------- |
| Frontend      | Next.js 15, TypeScript, ShadCN UI, Tailwind |
| Editor        | Tiptap (Custom Nodes & Image Upload)        |
| Backend       | GO, Gin                                     |
| Database      | PostgreSQL17 + GORM + PGVector              |
| AI/ML         | Ollama (LLaMA3, nomic-embed-text)           |
| Search        | SearXNG (Meta Search Engine)                |
| CI/CD         | Jenkins + Docker                            |
| Lint & Scan   | SonarQube                                   |
| Deployment    | Docker Compose (Multi-container)            |
| Auth          | OAuth (Google, GitHub, Discord)             |
| Cache         | Redis                                       |
| Proxy         | Caddy (Reverse Proxy + Auto HTTPS)          |
| Image Storage | Chibisafe                                   |

---

## 🏛️ Architecture Diagram

```
```mermaid
graph TD
    A[Client] --> B(Caddy Reverse Proxy);
    B --> C[Frontend (Next.js)];
    B --> D[Backend (Go/Gin API)];
    D --> E[PostgreSQL + PGVector];
    D --> F[Redis Cache];
    D --> G[AWS Bedrock LLM];
    D --> H[Extractor (Python)];
    D --> I[SearXNG];
```

Key Technologies:
├─ Frontend: Next.js 15, TypeScript, ShadCN UI, Tailwind CSS
├─ Backend: Go, Gin Framework, GORM
├─ Database: PostgreSQL 17 + PGVector (Vector DB)
├─ Cache: Redis 7
├─ AI/ML: Ollama (LLaMA3), nomic-embed-text (384-dim embeddings)
├─ Search: SearXNG (Privacy-focused meta-search)
├─ Proxy: Caddy (Auto HTTPS, Reverse Proxy)
├─ Queue: Go channels + Background workers
├─ Real-time: WebSocket (Gorilla WebSocket)
├─ Auth: OAuth 2.0 (Google, GitHub, Discord)
├─ CI/CD: Jenkins, Docker, SonarQube, Jest
└─ Deployment: Docker Compose (Multi-container orchestration)
```

---

## 🧪 CI/CD Pipeline

- ✅ **Test**: Run with Jest
- 🧹 **Lint & Scan**: SonarQube
- 🐳 **Deploy**: Docker + Jenkins
- 🐾 **Auto Deploy**: Pull Request -> Merge -> Deploy

---

## 🧠 Future Plans

- [ ] ✍️ AI Assistant: Auto-summarize / Suggest blog topics
- [ ] 🧠 RAG Search: LLM (LLaMA3)
- [ ] 🧪 Enhanced Analytics Dashboard
- [ ] 📱 Mobile-first UX Improvements

---

## 🤝 Contributors

Powered by BSO Club, Burapha University SE Students ❤️  
Maintained by: [@yamroll](https://github.com/LordEaster) and team.

---

> "Be Simple but Outstanding." – A platform born from the passion of those who love sharing knowledge.

---

## 🤖 AI-Generated Notice

This README file was generated with assistance from **ChatGPT-4o** (OpenAI) based on the project details and requirements provided by the BSO Blog team.
