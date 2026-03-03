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

If you want a clean SQLite database:

make reset-db

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

## ✔️ Recommended Workflow

make deps\
make reset-db (optional)\
make run-all

Frontend will open automatically at:\
http://localhost:3000

------------------------------------------------------------------------

# Project Title: Forum Project

## Description

The **Forum project** is a full-stack web-based application that allows
users to register, log in, create posts, comment, react, and receive
real-time notifications.

The system is built with a clean backend API architecture and a dynamic
frontend UI that communicates through REST endpoints.

------------------------------------------------------------------------

## 🚀 Core Features

### 🔐 Authentication

-   Traditional email/password authentication
-   Secure password hashing with bcrypt
-   Session management using UUID and cookies
-   OAuth authentication with:
    -   Google Login
    -   GitHub Login
-   Protected routes with middleware-based access control

------------------------------------------------------------------------

### 📝 Posts

-   Create, update, delete, and view posts
-   Image upload support in posts
-   Category tagging
-   Publish / Draft state management

------------------------------------------------------------------------

### 💬 Comments

-   Add, update, and delete comments
-   Image upload support in comments
-   Highlight new or specific comments
-   Automatic scroll to newest comment

------------------------------------------------------------------------

### 👍 Reactions

-   Like / Dislike posts
-   Like / Dislike comments
-   Mutual exclusion logic (cannot like and dislike simultaneously)
-   Instant UI updates after reaction

------------------------------------------------------------------------

### 🔔 Real-Time Notifications

-   Notifications for:
    -   Post reactions
    -   Comment reactions
    -   New comments on your posts
-   Live polling system
-   Sound effects for new notifications
-   Notification badge counter
-   Dropdown notification panel
-   Auto-mark as read behavior

------------------------------------------------------------------------

### 📊 My Activity Dashboard

All user-related actions are organized and accessible through **My
Activity**, including:

-   User's posts
-   User's comments
-   Reactions received
-   Notifications history
-   Activity-based navigation

------------------------------------------------------------------------

## 🧩 Technologies Used

-   Go (Backend)
-   SQLite (Database)
-   bcrypt (Password hashing)
-   uuid (Session management)
-   OAuth2 (Google & GitHub authentication)
-   Vanilla JavaScript (Frontend)
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

fuser -k 8080/tcp

#### Option 2 (Linux / macOS - using lsof)

lsof -i :8080\
kill -9 `<PID>`{=html}

#### Option 3 (Cross-platform alternative)

Change the backend port in your configuration if killing the process is
not desired.

After freeing the port, run again:

make run-all

------------------------------------------------------------------------

## Known Issues or Limitations

-   Nested threaded replies are limited to a single level.
-   Advanced user profile customization is not yet implemented.

------------------------------------------------------------------------

## Contributors / Authors

-   Chris Baikas (chbaikas)
-   Alex Smyroglou (asmyrogl)

------------------------------------------------------------------------

## License

This project is licensed under the MIT License.