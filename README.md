
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

---

## Project Title: Forum Project

### Description
The **Forum project** is a web-based application that allows users to register, log in, post, comment, and like posts. It also includes features like categories, filters, and session management.

---

## Features
- **Authentication**: User registration, login, logout, and session management.
- **Posts**: Create, update, delete, and view posts.
- **Comments**: Add, update, and delete comments on posts.
- **Likes**: Like/dislike posts.
- **Filters**: Filter posts by categories.

---

## Technologies Used
- **Go**: Backend development using Go language.
- **SQLite**: Lightweight, serverless database for persistence.
- **bcrypt**: Password hashing for security.
- **uuid**: Universally unique identifiers for user sessions.
- **Docker**: For containerization of the backend service.
- **Docker Compose**: To simplify running the backend in a containerized environment.

---

## Database Schema / ER Diagram
The database includes the following main tables:
- **users**: Stores user credentials and session information.
- **posts**: Stores user-generated content (posts).
- **comments**: Stores comments associated with posts.
- **categories**: Stores categories for posts.
- **reactions**: Stores like/dislike reactions on posts.

(Optionally, you can include a diagram here showing relationships between these tables).

---

## Setup Instructions

### How to Build and Run with Go
1. Clone the repository.
2. Run `make deps` or `go mod tidy` to install dependencies.
3. To build the project:
   - Backend: `make build-backend`
   - Frontend: `make build-frontend`
4. To run the servers:
   - Backend: `make run-backend`
   - Frontend: `make run-frontend`
5. Visit [http://localhost:3000](http://localhost:3000) to use the frontend.

### How to Run with Docker or Docker Compose
1. Build the Docker image: `make docker-build`
2. Run the Docker container: `make docker-run`
3. For Docker Compose (to run both frontend and backend):
   - Start Compose: `make docker-up`
   - Stop Compose: `make docker-down`

---

## Usage Instructions
- **Register**: Navigate to the registration page, fill in your details, and submit.
- **Login**: Enter your credentials and log in to your account.
- **Create Posts**: Use the "New Post" button to create a post with a title and content.
- **Add Comments**: Comments can be added to posts by clicking the "Comment" button.
- **Like Posts**: Posts can be liked/disliked by clicking the thumbs-up/thumbs-down button.

---

## Folder Structure
```
/cmd
  /backend  → Go backend entry point
  /frontend → Frontend entry point (served via Go or static server)
/internal
  /db       → Database-related code
  /handlers → HTTP request handlers (API endpoints)
/templates    → HTML templates for frontend rendering
/static       → Static assets like CSS, JS, images
/db           → SQLite database schema
```

---

## Known Issues or Limitations
- Frontend does not yet support advanced user profile management.
- The comment system lacks the ability to reply to specific comments.

---

## Contributors / Authors
- **Your Name**: Lead Developer
- **Teammates**: Other contributors to the project.

---

## License
This project is licensed under the MIT License — see the [LICENSE](./LICENSE) file for details.