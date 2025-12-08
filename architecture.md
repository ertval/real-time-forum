
# Forum Project — Architecture Overview

This document provides an overview of the backend and frontend architecture of the Forum project.  
It describes how the application is structured, how components interact, and how responsibilities are organized within the codebase.

---

## 1. High-Level Architecture

The project follows a **clean modular structure** separating backend and frontend concerns:

- `/cmd` — Entry points for running backend and frontend servers.
- `/internal` — All backend logic (not importable externally, per Go’s `internal/` rule).
- `/web` — Static frontend assets (HTML, JS, CSS).
- Root-level scripts and configuration files (Docker, Makefile, etc.)

Communication between frontend and backend occurs via a **REST API** served under `/api/v1/...`.

---

## 2. Backend Architecture

The backend is a Go HTTP server using standard library components and SQLite as the data store.

### Key directories under `/internal`:

### **2.1 `internal/db`**
Contains all database-related logic:

- `forum_schema.sql` — SQLite schema defining tables and constraints.
- `users.go` — CRUD logic for users, registration, login validation.
- `sessions.go` — Session management using cookies (create, validate, invalidate).
- `posts.go`, `comments.go`, `categories.go`, `reactions.go` — Database operations for forum resources.
- `db.go` — Database initialization, connection creation.
- `error.go` — Shared DB error helpers.

This layer contains **no HTTP logic** — it only handles persistence and data validation related to the database.

---

### **2.2 `internal/handlers`**
These files define HTTP handlers for API endpoints.  
Each handler translates incoming HTTP requests into DB operations and returns JSON responses.

Examples:
- `users.go` — Register, Login, Get User, Me.
- `posts.go` — List posts, create post, view post.
- `comments.go` — create & list comments.
- `categories.go` — list categories.
- `health.go` — health check endpoint.
- `respond.go` — standardized JSON response envelope.

Handlers must remain stateless, using only:
- `http.ResponseWriter`
- `*http.Request`
- DB connections passed via composition (`NewUsers(db)`)

---

### **2.3 `internal/middleware`**
Cross-cutting HTTP middleware:

- `auth.go` — Validates session cookie and attaches `userID` to the request context.
- `logger.go` — Logs requests.
- `recoverer.go` — Recovers from panics and prevents server crashes.
- `cors.go` — Enables frontend communication (localhost:3000 during development).

Middleware wraps handlers inside the router.

---

### **2.4 `internal/router`**
Defines all API routes and attaches handlers + middleware.

Examples:

```
/api/v1/users/register     → public
/api/v1/users/login        → public
/api/v1/users/me           → requires auth
/api/v1/posts              → GET public, POST requires auth
/api/v1/categories         → public
```

The router builds the entire HTTP handler tree.

---

### **2.5 `internal/server`**
Contains:

- `server.go` — Starts the backend server, initializes DB, mounts router.
- `api_test.go` — Full API test suite using httptest + in‑memory SQLite.

This package contains all top-level orchestration required to boot the backend.

---

## 3. Frontend Architecture (`/web`)

This directory contains static assets served independently:

- `index.html` — Main page.
- `app.js` — Frontend logic (calls backend API endpoints).
- `styles.css` — UI styling.

A minimal Go server (under `/cmd/frontend`) can serve these assets locally for development.

Frontend communicates exclusively via **fetch() to /api/v1/...**

---

## 4. Authentication Architecture

The project implements secure cookie-based sessions:

### Login flow:
1. User submits username/email + password.
2. Password is checked with bcrypt.
3. A session is created in DB with:
   - UUID token
   - user_id
   - expires_at
   - ip, user_agent
4. Server sends:
   ```
   Set-Cookie: session_token=<uuid>
   ```

### Authenticated requests:
- Middleware reads the cookie → validates session → attaches userID to request context → handler executes.

### Logout:
- Session is invalidated in DB.
- Cookie is cleared from client.

---

## 5. Database Schema Overview

Key tables:

- `users`
- `sessions`
- `posts`
- `comments`
- `categories`
- `reactions`

Important constraints:
- One active session per user (unique partial index).
- Reactions enforce one like/dislike per user per target.
- Comments support nesting (parent_comment_id).
- Cascading foreign keys ensure consistent cleanup.

---

## 6. Testing Architecture

The file `internal/server/api_test.go` includes full integration tests:

- Health checks  
- Posts listing & creation  
- Comment creation & listing  
- Like toggling  
- User registration  
- Login  
- Authenticated `/me`  
- Full session flow  

Tests run against **in-memory SQLite** with schema autoload.

---

## 7. Future Extensions (not yet implemented)

- Edit/Delete posts
- Full category filtering
- Pagination abstraction
- Email verification
- Admin panel
- Real frontend UI integration

---

## Summary

The project is organized into clear layers:

```
cmd/           → entrypoints
internal/db    → persistence layer
internal/handlers → HTTP controllers
internal/middleware → cross‑cutting concerns
internal/router → routing configuration
internal/server → server startup + tests
web/           → static frontend
```

This modular architecture ensures:
- maintainability  
- clarity  
- separation of concerns  
- ability to add features safely  

---

