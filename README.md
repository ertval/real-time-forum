# Forum Project — How to Run (Backend + Frontend)

This document explains step-by-step how to set up and run the Forum project, including both the backend and frontend servers.

---

## 🚀 1. Requirements

Ensure the following are installed:

- Go 1.23+
- Make
- SQLite (optional)
- Docker & Docker Compose (optional)

---

## 🛠️ 2. Install Dependencies

### Using Make:
```
make deps
```

### Without Make:
```
go mod tidy
```

---

## 🗄️ 3. Initialize / Reset Database

If you want a clean SQLite database:

### Using Make:
```
make reset-db
```

---

## 🧱 4. Build the Project

### Build backend only:
```
make build-backend
```

### Build frontend only:
```
make build-frontend
```

### Build both:
```
make build-all
```

---

## 🌍 5. Run the Application

The project uses two separate servers:

- Backend API → http://localhost:8080  
- Frontend UI → http://localhost:3000  

### Run backend only:
```
make run-backend
```

### Run frontend only (auto-opens browser):
```
make run-frontend
```

### Run both backend + frontend (recommended):
```
make run-all
```

The frontend will automatically open in your browser.

---

## 🛑 6. Stopping the Servers

### Stop backend:
```
make stop-backend
```

### Stop frontend:
```
make stop-frontend
```

### Stop all:
```
make stop-all
```

---

## 🐳 7. Running with Docker (Optional)

### Build Docker image:
```
make docker-build
```

### Run container:
```
make docker-run
```

### Docker Compose:
```
make docker-up
```

### Stop Compose:
```
make docker-down
```

---

## 🧹 8. Clean Up

### Remove binaries:
```
make clean
```

### Clean everything (binaries + Docker artifacts):
```
make clean-all
```

---

## ✔️ Recommended Workflow

1. Install dependencies:
```
make deps
```

2. Reset DB (optional):
```
make reset-db
```

3. Run both servers:
```
make run-all
```

Your browser will automatically open at:

➡ http://localhost:3000

---

## 📌 Notes

- The backend and frontend are separate Go servers.
- The frontend communicates with the backend through:
  ```
  http://localhost:8080/api/v1
  ```
- Auto-opening browser works only for the frontend commands.

---