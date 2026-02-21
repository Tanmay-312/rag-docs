# 🚀 Turbo-Ingest: High-Speed Ephemeral RAG

**Turbo-Ingest** is a privacy-first, zero-login RAG (Retrieval-Augmented Generation) engine designed for lightning-fast document intelligence. Built with a **Golang** concurrency-first backend and a **Next.js** agentic frontend, it processes massive PDFs 10x faster than traditional Python-based pipelines.

## 🌟 Key Features

- **Parallel Ingestion Engine:** Utilizes Golang Goroutines and Worker Pools to parse and chunk PDFs concurrently.
- **Privacy-First (No-Login):** Temporary session-based vector storage. No personal data is ever saved; one click wipes everything.
- **AI Sanitization Agent:** Automatically redacts PII (emails, keys, phone numbers) using Gemini 1.5 Flash before data hits the vector store.
- **Real-time SSE Streaming:** Live ingestion progress updates and typewriter-style AI responses.
- **Auto-Summarization:** Instant generation of executive summaries and key takeaway cards upon upload.

## 🛠️ The Stack

| Layer         | Technology                                                      |
| ------------- | --------------------------------------------------------------- |
| **Frontend**  | Next.js 15 (App Router), Tailwind CSS, Shadcn/UI, Framer Motion |
| **Backend**   | Golang (deployed as Vercel Serverless Functions)                |
| **AI / LLM**  | Gemini 1.5 Flash (Processing & Embeddings)                      |
| **Vector DB** | Upstash Vector (Serverless & Ephemeral)                         |
| **DevOps**    | Vercel CLI, Monorepo Architecture                               |

## 🚀 Quick Start (Local Development)

### 1. Prerequisites

- [Vercel CLI](https://vercel.com/download) installed: `npm i -g vercel`
- Go 1.21+ installed.

### 2. Environment Setup

Create a `.env.local` in the root:

```bash
UPSTASH_VECTOR_REST_URL="your_upstash_url"
UPSTASH_VECTOR_REST_TOKEN="your_upstash_token"
GEMINI_API_KEY="your_google_api_key"

```

### 3. Run Locally

```bash
# Install dependencies
npm install

# Run both Next.js and Go functions simultaneously
vercel dev

```

Your app is now live at `http://localhost:3000` and your Go API at `http://localhost:3000/api`.

## 🏗️ Architecture

The project follows a **Monorepo** structure optimized for Vercel:

- `/app`: Next.js frontend logic and UI components.
- `/api`: Golang serverless functions handling the ingestion worker pool.
- `/components`: Reusable Shadcn + Framer Motion UI elements.

## 🛡️ Privacy & Security

This project is designed for **maximum privacy**:

1. **Redaction:** The Sanitization Agent scans all text chunks before embedding.
2. **Ephemeral:** Vectors are stored in a session-specific namespace.
3. **Nuclear Option:** The "Wipe All" button clears the session's vector namespace and local storage instantly.

---
