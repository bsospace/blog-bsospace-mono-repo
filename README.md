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
| Proxy         | Nginx (Reverse Proxy + Auto HTTPS)          |
| Image Storage | Chibisafe                                   |

---

## 🏛️ Architecture Diagram
![BSO Space Blog Architecture](blog-arch-diagram.png 'Optional Title')

```
Key Technologies:
├─ Frontend: Next.js 15 (App Router), TypeScript, ShadCN UI, Tailwind CSS 4
│ ├─ TipTap Editor (Rich-text AI-aware content editing)
│ ├─ Streaming Chat (SSE / WebSocket)
│ └─ OAuth 2.0 Authentication (Google / GitHub / Discord)

├─ Backend: Go (Gin Framework), GORM ORM, Clean Architecture
│ ├─ REST + WebSocket APIs
│ ├─ JWT Authentication & Rate Limiting
│ └─ RAG Orchestration + AI Mode Management

├─ Database: PostgreSQL 17 + PGVector (384-dim vector embeddings)
│ ├─ Stores posts, embeddings, AI logs
│ └─ Supports cosine similarity search for RAG

├─ Cache Layer: Redis 7 (In-memory cache & rate-limit store)

├─ AI / LLM Stack:
│ ├─ Intent Classification: GPT-4o-mini (via OpenRouter)
│ ├─ Answer Generation: LLaMA 3.1 70B Instruct (via OpenRouter)
│ ├─ Embedding Generation: Amazon Titan v2 / nomic-embed-text
│ └─ Web Search Fallback: SearXNG (privacy-preserving meta search)

├─ Proxy & Gateway: Nginx (HTTPS termination, Load balancing)
│ └─ Previously used Caddy for auto SSL during early deployment

├─ Queue & Workers: Go channels + Background jobs (asynchronous AI tasks)

├─ Real-time Communication: WebSocket (via go-socket.io)
│ └─ Used for live chat and AI response streaming

├─ CI/CD & Quality Assurance:
│ ├─ Jenkins (Automated CI/CD Pipeline)
│ ├─ SonarQube (Code Quality & Security Scan)
│ ├─ Jest (Unit Testing)
│ └─ Docker Build + Automated Deployment

└─ Deployment & Infrastructure:
├─ Docker Compose (Multi-container orchestration)
├─ Chibisafe (Object Storage for images / CDN)
├─ pgAdmin (Database management)
└─ SearXNG + OpenRouter self-host integration (optional for RAG stack)

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
