# Forum Project --- Architecture Overview

This document reflects the current architecture of the Forum project,
including authentication (Google & GitHub), image uploads, reactions,
real-time notifications, and the My Activity dashboard.

------------------------------------------------------------------------

## 1. High-Level Overview

The project follows a clean, layered Go architecture with strict
separation between persistence, HTTP logic, middleware, and frontend.

web/
  static/            → Legacy and shared static assets
  templates/         → Legacy HTML templates
SPA/                 → Single Page Application (Modern Vanilla JS ES2026+)
  index.html         → SPA Shell Entrypoint
  main.js            → App Bootstrap
  assets/            → Global CSS & Static Images
  core/              → Router, API client, Global State
  components/        → Shared Reusable UI Components
  features/          → Vertical Domain Slices (Auth, Feed, Chat, etc.)
  tests/             → Unit and Integration Tests

The frontend communicates with the backend via:

/api/v1/*        → REST APIs (CRUD, Auth)
/ws              → WebSockets (Presence, Private Messaging)

Architecture style:
- Layered backend (db, handlers, middleware)
- Single Page Application (SPA) shell
- Screaming Architecture for frontend features
- Event-driven real-time interactions via WebSockets
- Clean separation of concerns

------------------------------------------------------------------------

## 2. Backend Architecture

Built with Go standard library + SQLite.

internal/
├── db/              → Persistence (SQLite, SQL queries)
├── handlers/        → HTTP Request Handlers
├── middleware/      → Request Middleware (Auth, CORS, Logging)
├── router/          → Route Registration
└── tests/           → Backend Integration Tests

------------------------------------------------------------------------

## 3. Frontend Architecture (SPA)

Located under:

SPA/
├── assets/          → CSS & Global Assets
├── components/      → Reusable UI Fragments
├── core/            → Infrastructure (API, Router, State)
├── features/        → Vertical Slices (Domain Logic & Views)
└── main.js          → Entry Point

Characteristics:
- Pure Vanilla JavaScript (ES2026+)
- No frontend framework (React, Vue, etc.)
- Modular ES modules
- Proxy-based Global State management
- Client-side Routing
- Vitest for testing suite (Unit, Integration, E2E)

------------------------------------------------------------------------

## 4. Authentication Model

Supported authentication methods:

-   Email / Password
-   Google OAuth
-   GitHub OAuth

Authentication design:

-   Cookie-based sessions
-   HttpOnly cookies
-   One session per user
-   Server-side session validation
-   OAuth callback flow integrated in handlers

------------------------------------------------------------------------

## 5. Reactions System

-   Like / Dislike for posts
-   Like / Dislike for comments
-   One reaction per user per entity
-   Mutual exclusion enforced
-   Server returns updated counts
-   Frontend updates UI instantly

------------------------------------------------------------------------

## 6. Image Upload System

Supported in: - Posts - Comments

Features: - Multipart form handling - Max size validation - Image
preview in UI - Lazy loading - Transparent PNG detection (frontend
enhancement)

------------------------------------------------------------------------

## 7. Real-Time Notifications

Notifications triggered on:

-   Post reactions
-   Comment reactions
-   New comments on user posts

Architecture:

-   notifications table
-   Polling mechanism (frontend)
-   Badge counter
-   Dropdown panel
-   Sound feedback
-   Mark-as-read endpoints

------------------------------------------------------------------------

## 8. My Activity Dashboard

Aggregates:

-   User posts
-   User comments
-   Reactions received
-   Notifications

Provides centralized user activity tracking and navigation.

------------------------------------------------------------------------

## 9. Database Model

Core tables:

-   users
-   sessions
-   posts
-   drafts
-   comments
-   categories
-   reactions
-   notifications

Constraints:

-   One reaction per user per entity
-   Foreign keys enforced
-   Cascading rules defined
-   Draft ownership enforced

------------------------------------------------------------------------

## 10. Testing Strategy

-   Full API integration tests
-   httptest package
-   In-memory SQLite
-   Dynamic schema loading
-   Authentication flow testing
-   Reaction logic testing

Assertions cover:

-   HTTP status codes
-   Cookies
-   JSON structure
-   Database side-effects

------------------------------------------------------------------------

## 11. Design Principles

-   Explicit over implicit
-   No heavy frameworks
-   Predictable naming conventions
-   Stateless handlers
-   SRP-compliant modules
-   Clear separation between layers
-   Defensive error handling

------------------------------------------------------------------------

## Summary

cmd/ → entrypoints\
internal/db → persistence layer\
internal/handlers → HTTP logic\
internal/middleware → request middleware\
internal/router → routing configuration\
internal/server → application bootstrap\
internal/tests → API integration tests\
web/ → frontend

This architecture ensures maintainability, clarity, testability, and
production-ready structure without relying on external frameworks.