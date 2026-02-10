
# Forum Project — Architecture Overview

This document reflects the **current architecture** of the Forum project,
including drafts, autosave, middleware refactors, and test strategy.

---

## 1. High-Level Overview

The project follows a **clean, layered Go architecture**:

```
cmd/
internal/
web/
```

Frontend communicates with backend via:
```
/api/v1/*
```

---

## 2. Backend Architecture

Built with Go standard library + SQLite.

```
internal/
 ├── db
 ├── handlers
 ├── middleware
 ├── router
 ├── server
 └── tests
```

---

## 2.1 internal/db — Persistence Layer

Responsibilities:
- SQL queries
- Transactions
- Context-aware execution
- Schema ownership

Key modules:
- users.go
- sessions.go
- posts.go
- drafts.go
- comments.go
- categories.go
- reactions.go
- errors.go
- db.go

Rules:
- No HTTP imports
- No JSON
- No request parsing

---

## 2.2 internal/handlers — HTTP Layer

Responsibilities:
- HTTP validation
- Request parsing
- JSON responses
- Status codes

Key handlers:
- users.go
- posts.go
- drafts.go
- comments.go
- categories.go
- health.go

Patterns:
- HandlePosts → collection
- HandlePost → item
- req structs for payloads

---

## 2.3 Middleware

```
Request
 → Logger
 → Recoverer
 → CORS
 → Auth (optional)
 → Handler
```

Files:
- auth.go
- middleware.go

Auth:
- Validates session cookie
- Injects userID into context

---

## 2.4 Router

- Uses http.ServeMux
- Explicit routing
- Auth applied per-route

Examples:
```
GET  /api/v1/posts
POST /api/v1/posts        (auth)
POST /api/v1/drafts       (auth)
GET  /api/v1/users/me     (auth)
```

---

## 2.5 Server

- Initializes DB
- Builds router
- Starts HTTP server

No business logic.

---

## 3. Frontend Architecture

Located under:
```
web/
 ├── static/js
 ├── static/css
 └── index.html
```

Frontend logic split into:
- auth.js
- create-post.js
- drafts.js
- api.js
- ui-messages.js

No framework. Pure JS.

---

## 4. Authentication Model

- Cookie-based sessions
- HttpOnly cookies
- One session per user
- Server-side validation

---

## 5. Database Model

Tables:
- users
- sessions
- posts
- drafts
- comments
- categories
- reactions

Constraints:
- One reaction per user per post
- Drafts owned by user
- FK enforced

---

## 6. Testing Strategy

- Full API integration tests
- httptest
- In-memory SQLite
- Schema loaded dynamically

Tests assert:
- Status codes
- Cookies
- JSON responses

---

## 7. Design Principles

- Explicit over implicit
- No magic frameworks
- Stateless handlers
- SRP-compliant files
- Predictable naming

---

## Summary

```
cmd/                → entrypoints
internal/db         → persistence
internal/handlers   → HTTP
internal/middleware → middleware
internal/router     → routing
internal/server     → bootstrap
internal/tests      → API tests
web/                → frontend
```