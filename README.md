
# Forum Project — How to Run (Backend + Frontend)

This document explains how to set up and run the Forum project in its **current state**,
including backend, frontend, database, authentication, drafts, and Docker support.

---

## 🚀 1. Requirements

- Go 1.24+
- Make
- SQLite (recommended, auto-managed)
- Node.js (only for frontend tooling if needed)
- Docker & Docker Compose (optional — backend only)

---

## 🛠️ 2. Install Dependencies

### Using Make:
make deps

### Without Make:
go mod tidy

---

## 🗄️ 3. Database Initialization

The project uses **SQLite**.

### Reset database (development only):
make reset-db

> ⚠️ This deletes `forum.db` and recreates it from `forum_schema.sql`.

---

## 🧱 4. Build Targets

### Backend:
make build-backend

### Frontend:
make build-frontend

### Everything:
make build-all

---

## 🌍 5. Running the Application

Two separate servers are used:

| Service   | URL |
|----------|-----|
| Backend API | http://localhost:8080 |
| Frontend UI | http://localhost:3000 |

### Run backend:
make run-backend

### Run frontend:
make run-frontend

### Run both:
make run-all

---

## ✍️ Drafts & Autosave

- Posts support **draft mode**
- Drafts are autosaved client-side and synced with backend
- Categories are preserved during autosave
- Draft endpoints are protected (auth required)

---

## 🔐 Authentication

- Cookie-based sessions
- HttpOnly cookies
- One active session per user
- Session invalidation on logout

---

## 🧪 Testing

Run all backend integration tests:
make test

Tests:
- Use in-memory SQLite
- Load schema automatically
- Validate real HTTP behavior

---

## 🐳 6. Docker (Backend Only)

### Build image:
make docker-build

### Run container:
make docker-run

### Docker Compose:
make docker-up
make docker-down

- Non-root container user
- Persistent volume for SQLite
- Multi-stage build (Go → slim runtime)

---

## 🧹 7. Cleanup

make clean
make clean-all

---

## ✔️ Recommended Workflow

make deps
make reset-db
make run-all

Frontend opens automatically at:
http://localhost:3000

---

## 📌 Notes

- Frontend is **not containerized**
- Backend exposes REST API under:
  http://localhost:8080/api/v1
- CORS enabled for local development