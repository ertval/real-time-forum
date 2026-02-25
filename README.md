# Forum Project --- How to Run (Backend + Frontend)

This document explains step-by-step how to set up and run the Forum
project, including both the backend and frontend servers.

------------------------------------------------------------------------

## 🚀 1. Requirements

Ensure the following are installed:

-   Go 1.24+
-   Make
-   SQLite (optional)
-   Docker & Docker Compose (optional --- backend only)

------------------------------------------------------------------------

## 🛠️ 2. Install Dependencies

### Using Make:

make deps

### Without Make:

go mod tidy

------------------------------------------------------------------------

## 🗄️ 3. Initialize / Reset Database

If you want a clean SQLite database: make reset-db

------------------------------------------------------------------------

## 🧱 4. Build the Project

### Backend only:

make build-backend

### Frontend only:

make build-frontend

### Build everything:

make build-all

------------------------------------------------------------------------

## 🌍 5. Run the Application

The project uses two separate servers:

-   Backend API → http://localhost:8080
-   Frontend UI → http://localhost:3000

### Run backend:

make run-backend

### Run frontend (auto-opens browser):

make run-frontend

### Run both:

make run-all

------------------------------------------------------------------------

## 🛑 6. Stopping the Servers

### Stop backend:

make stop-backend

### Stop frontend:

make stop-frontend

### Stop both:

make stop-all

------------------------------------------------------------------------

## 🐳 7. Running Backend with Docker (Optional)

### Build Docker image:

make docker-build

### Run backend container:

make docker-run

### Using Docker Compose:

make docker-up

### Stop Compose:

make docker-down

------------------------------------------------------------------------

## 🧹 8. Clean Up

### Remove binaries:

make clean

### Clean everything:

make clean-all

------------------------------------------------------------------------

## ✔️ Recommended Workflow

make deps\
make reset-db (optional)\
make run-all

Frontend will open automatically at:\
http://localhost:3000

------------------------------------------------------------------------

## 📌 Notes

-   Backend is the only service that runs inside Docker.
-   Frontend is designed to run locally (not containerized).
-   Frontend communicates with backend using:
    http://localhost:8080/api/v1

------------------------------------------------------------------------

## Project Title: Forum Project

### Description

The **Forum project** is a web-based application that allows users to
register, log in, post, comment, and like posts. It also includes
features like categories, filters, and session management.

------------------------------------------------------------------------

## Features

-   Authentication: User registration, login, logout, and session
    management.
-   Posts: Create, update, delete, and view posts.
-   Comments: Add, update, and delete comments on posts.
-   Likes: Like/dislike posts and comments.
-   Filters: Filter posts by categories.
-   Notifications: Real-time notifications for reactions and comments.

------------------------------------------------------------------------

## Technologies Used

-   Go (Backend)
-   SQLite (Database)
-   bcrypt (Password hashing)
-   uuid (Session management)
-   Docker (Containerization)
-   Docker Compose (Orchestration)

------------------------------------------------------------------------

## Folder Structure

/cmd\
/backend\
/frontend\
/internal\
/db\
/handlers\
/templates\
/static\
/data

------------------------------------------------------------------------

## 🧯 Troubleshooting

### Port 8080 already in use

If you see:

listen tcp :8080: bind: address already in use

It means another process is already using the backend port.

#### Option 1 (Linux - using fuser)

``` bash
fuser -k 8080/tcp
```

#### Option 2 (Linux / macOS - using lsof)

``` bash
lsof -i :8080
kill -9 <PID>
```

#### Option 3 (Cross-platform alternative)

Change the backend port in your configuration if killing the process is
not desired.

After freeing the port, run again:

``` bash
make run-all
```

------------------------------------------------------------------------

## Known Issues or Limitations

-   Advanced user profile management is not yet implemented.
-   Nested threaded comment replies are limited.

------------------------------------------------------------------------

## Contributors / Authors

-   Chris Baikas (chbaikas)
-   Alex Smyroglou (asmyrogl)

------------------------------------------------------------------------

## License

This project is licensed under the MIT License.