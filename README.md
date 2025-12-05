# Forum Project — How to Run (Backend + Frontend)

This document explains step-by-step how to set up and run the Forum project, including both the backend and frontend servers.

---

## 🚀 1. Requirements

Ensure the following are installed:

- Go 1.24+
- Make
- SQLite (optional)
- Docker & Docker Compose (optional — backend only)

---

## 🛠️ 2. Install Dependencies

### Using Make:
make deps

### Without Make:
go mod tidy

---

## 🗄️ 3. Initialize / Reset Database

If you want a clean SQLite database:
make reset-db

---

## 🧱 4. Build the Project

### Backend only:
make build-backend

### Frontend only:
make build-frontend

### Build everything:
make build-all

---

## 🌍 5. Run the Application

The project uses two separate servers:

- Backend API → http://localhost:8080
- Frontend UI → http://localhost:3000

### Run backend:
make run-backend

### Run frontend (auto-opens browser):
make run-frontend

### Run both:
make run-all

---

## 🛑 6. Stopping the Servers

### Stop backend:
make stop-backend

### Stop frontend:
make stop-frontend

### Stop both:
make stop-all

---

## 🐳 7. Running Backend with Docker (Optional)

### Build Docker image:
make docker-build

### Run backend container:
make docker-run

### Using Docker Compose:
make docker-up

### Stop Compose:
make docker-down

---

## 🧹 8. Clean Up

### Remove binaries:
make clean

### Clean everything:
make clean-all

---

## ✔️ Recommended Workflow

make deps
make reset-db   (optional)
make run-all

Frontend will open automatically at:
http://localhost:3000

---

## 📌 Notes

- Backend is the only service that runs inside Docker.
- Frontend is designed to run locally (not containerized).
- Frontend communicates with backend using:
  http://localhost:8080/api/v1