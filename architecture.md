# Forum Project — Architecture Overview

This document provides an up-to-date overview of the backend and frontend architecture of the Forum project,
reflecting the latest refactors, naming conventions, and structural decisions.

---

## 1. High-Level Architecture

The project follows a **clean, modular Go architecture** with strict separation of concerns:

- `/cmd` — Application entry points (backend & frontend).
- `/internal` — Core application logic (enforced by Go `internal/` visibility rules).
- `/web` — Static frontend assets.
- Root-level configuration and tooling (Docker, Makefile, CI helpers).

The frontend communicates with the backend exclusively through a **REST API** exposed under:

```
/api/v1/...
```

---

## 2. Backend Architecture

The backend is a Go HTTP server built on the standard library, using **SQLite** for persistence and a layered design.

### Directory overview:

```
internal/
 ├── db          → Persistence layer (SQL + domain data)
 ├── handlers    → HTTP handlers (controllers)
 ├── middleware  → Cross-cutting HTTP concerns
 ├── router      → Route definitions & middleware wiring
 ├── server      → Server bootstrap
 └── tests       → Integration & API tests
```

---

## 2.1 `internal/db` — Persistence Layer

This package contains **all database-related logic** and **no HTTP code**.

### Responsibilities:
- SQL queries & transactions
- Domain validation at persistence level
- Timeouts via `context.Context`
- SQLite schema ownership

### Key files:
- `forum_schema.sql` — Full database schema.
- `users.go` — User creation, login validation, retrieval.
- `sessions.go` — Cookie-based session lifecycle (create, validate, invalidate).
- `posts.go` — Posts CRUD logic.
- `comments.go` — Comment creation and listing.
- `categories.go` — Category queries.
- `reactions.go` — Like/dislike logic.
- `errors.go` — Shared DB error helpers.
- `db.go` — DB initialization & helpers.
- `dbContentsPrinter.go` — **Development-only debug utility** (not used in production).

### Design rules:
- No `http.*` imports
- No JSON marshaling
- No handler logic
- Uses `context.Context` consistently

---

## 2.2 `internal/handlers` — HTTP Handlers

Handlers translate HTTP requests into DB operations and format JSON responses.

### Responsibilities:
- Parse request paths & bodies
- Validate input (HTTP-level)
- Call `internal/db`
- Return standardized JSON responses

### Key files:
- `users.go` — Register, Login, Logout, Me, User item.
- `posts.go` — Posts collection & single-post routing.
- `posts_public.go` — Public-only post listing.
- `comments.go` — Comment creation & listing.
- `categories.go` — Categories collection & item.
- `health.go` — Health check endpoint.
- `response.go` — Unified JSON response envelope.
- `helpers.go` / `posts_helpers.go` — Path parsing, pagination helpers.

### Naming conventions:
- `HandlePosts` → collection-level handler
- `HandlePost` → item-level handler
- Internal helpers like `listPosts`, `createPost`, etc.
- Request payloads named `req` (instead of `in`)

Handlers are **stateless** and depend only on:
- `http.ResponseWriter`
- `*http.Request`
- Injected DB connection

---

## 2.3 `internal/middleware` — HTTP Middleware

Reusable middleware components applied globally or per-route.

### Files:
- `auth.go` — Session validation, injects `userID` into request context.
- `middleware.go`:
  - Logger
  - Recoverer (panic safety)
  - CORS (`EnableCORS`)

### Middleware flow:
```
Request → Logger → Recoverer → CORS → Auth (optional) → Handler
```

---

## 2.4 `internal/router` — Routing

The router defines **all API endpoints** and wires handlers with middleware.

### Example routes:
```
/api/v1/health
/api/v1/users/register
/api/v1/users/login
/api/v1/users/me        (auth)
/api/v1/posts           (GET public, POST auth)
/api/v1/posts/{id}
/api/v1/posts/{id}/comments
/api/v1/categories
```

### Design:
- Uses `http.ServeMux`
- Explicit routing (no magic frameworks)
- Auth middleware applied selectively

---

## 2.5 `internal/server` — Server Bootstrap

Responsible for application startup.

### Contains:
- `server.go` — Initializes DB, builds router, starts HTTP server.

This layer **coordinates** components but contains no business logic.

---

## 3. Frontend Architecture (`/web`)

The frontend is a lightweight static application.

### Contents:
- `index.html`
- `app.js`
- `styles.css`

A minimal Go server (`/cmd/frontend`) can serve these assets locally.

The frontend communicates only via:
```
fetch("/api/v1/...")
```

---

## 4. Authentication Architecture

The project uses **secure cookie-based sessions**.

### Login flow:
1. User submits credentials.
2. Password verified with bcrypt.
3. Session created in DB:
   - UUID token
   - user_id
   - expires_at
   - ip, user_agent
4. Server responds with:
```
Set-Cookie: session_token=<uuid>; HttpOnly
```

### Authenticated requests:
- Auth middleware validates cookie
- Loads session from DB
- Injects `userID` into request context

### Logout:
- Session invalidated
- Cookie cleared

---

## 5. Database Overview

Core tables:
- `users`
- `sessions`
- `posts`
- `comments`
- `categories`
- `reactions`

Key constraints:
- One active session per user
- One reaction per user per post
- Nested comments via `parent_comment_id`
- Foreign keys ensure integrity

---

## 6. Testing Architecture

Located under:
```
internal/tests
```

### Characteristics:
- Full API-level integration tests
- Uses `httptest`
- Uses **in-memory SQLite**
- Schema loaded automatically
- Tests include:
  - Auth flow
  - Comments
  - Posts
  - Categories
  - Likes
  - Health check

Tests validate **real HTTP behavior**, not internal functions.

---

## 7. Design Principles

- Clear separation of concerns
- Explicit routing
- Context-aware DB operations
- Stateless handlers
- Predictable naming
- SRP-compliant files

---

## Summary

```
cmd/                → entrypoints
internal/db         → persistence
internal/handlers   → HTTP handlers
internal/middleware → middleware
internal/router     → routing
internal/server     → startup
internal/tests      → API tests
web/                → frontend
```

This architecture provides:
- clarity
- testability
- maintainability
- safe extensibility
