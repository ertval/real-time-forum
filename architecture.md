# Forum Project --- Architecture Overview

This document reflects the current architecture of the Forum project,
including extended user profiles, image uploads, reactions,
real-time notifications, and private messaging via WebSockets.

------------------------------------------------------------------------

## 1. High-Level Overview

The project follows a clean, layered Go architecture with strict
separation between persistence, HTTP logic, middleware, and frontend.

web/
  static/            → Legacy and shared static assets
  templates/         → Legacy HTML templates
SPA/                 → Single Page Application (Modern Vanilla JS ES2026+)
  index.html         → SPA Shell Entrypoint
  main.js            → Thin bootstrap + public exports for tests
  assets/            → Global CSS entry + design tokens/base styles
  core/              → App orchestration, router, shared utils
  components/        → Shared Reusable UI Components
  features/          → Vertical Domain Slices (Auth, Feed, Post, Activity, Shell, Chat)
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
├── assets/          → CSS entry point and global assets
├── components/      → Reusable UI Fragments
├── core/            → App lifecycle, router, and utility modules
├── features/        → Vertical slices with route/shell renderers
└── main.js          → Entrypoint and compatibility exports

Characteristics:
- Pure Vanilla JavaScript (ES2026+)
- No frontend framework (React, Vue, etc.)
- Modular ES modules
- Feature slices are implemented for auth, feed, post, activity, shell, and profile routes
- Chat and messaging areas are in active development
- Client-side Routing
- Vitest for testing suite (Unit, Integration, E2E)

------------------------------------------------------------------------

## 4. Authentication Model

Supported authentication methods:

-   Email / Password
-   Google & GitHub OAuth (Legacy/Retained - Auth gating primary focus)

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
-   private_messages
-   oauth_users

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
cmd/backend and cmd/frontend → application bootstrap\
internal/tests → API integration tests\
web/ → frontend

This architecture ensures maintainability, clarity, testability, and
production-ready structure without relying on external frameworks.