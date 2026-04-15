# Forum Project --- Architecture Overview

This document reflects the current architecture of the Forum project,
including authentication (Google & GitHub), image uploads, reactions,
real-time notifications, and the My Activity dashboard.

------------------------------------------------------------------------

## 1. High-Level Overview

The project follows a clean, layered Go architecture with strict
separation between persistence, HTTP logic, middleware, and frontend.

Structure:

cmd/\
internal/\
web/

Frontend communicates with backend via:

/api/v1/\*

Architecture style: - Layered architecture - Clear separation of
concerns - Stateless HTTP handlers - Context-aware database operations

------------------------------------------------------------------------

## 2. Backend Architecture

Built with Go standard library + SQLite.

internal/\
├── db\
├── handlers\
├── middleware\
├── router\
├── server\
└── tests

------------------------------------------------------------------------

## 2.1 internal/db --- Persistence Layer

Responsibilities: - SQL queries - Transactions - Context-aware
execution - Schema ownership - Reaction aggregation - Draft
persistence - Notification storage

Key modules: - users.go - sessions.go - posts.go - drafts.go -
comments.go - categories.go - reactions.go - notifications.go -
errors.go - db.go

Rules: - No HTTP imports - No JSON encoding - No request parsing - No
business logic leakage

------------------------------------------------------------------------

## 2.2 internal/handlers --- HTTP Layer

Responsibilities: - Request validation - JSON parsing - Status code
handling - Response formatting - OAuth callback handling

Key handlers: - users.go - posts.go - drafts.go - comments.go -
categories.go - notifications.go - health.go

Patterns: - HandlePosts → collection - HandlePost → single resource -
Structured request DTOs - Consistent error response schema

------------------------------------------------------------------------

## 2.3 Middleware

Request Flow:

Request\
→ Logger\
→ Recoverer\
→ CORS\
→ OptionalAuth\
→ Auth (when required)\
→ Handler

Files: - auth.go - middleware.go

Authentication middleware: - Validates session cookie - Injects userID
into context - Supports optional auth for public endpoints

------------------------------------------------------------------------

## 2.4 Router

-   Uses http.ServeMux
-   Explicit route definitions
-   Middleware applied per-route
-   API versioning (/api/v1)

Examples:

GET /api/v1/posts\
POST /api/v1/posts (auth)\
POST /api/v1/drafts (auth)\
GET /api/v1/users/me (auth)\
GET /api/v1/notifications (auth)\
PATCH /api/v1/notifications/read-all (auth)

------------------------------------------------------------------------

## 2.5 Server

-   Initializes SQLite database
-   Loads schema
-   Builds router
-   Applies middleware stack
-   Starts HTTP server

No business logic inside server package.

------------------------------------------------------------------------

## 3. Frontend Architecture

Located under:

web/\
├── static/\
│   ├── js/\
│   │   ├── core/         (api wrapper, router, store etc.)\
│   │   ├── features/     (vertical slices by domain: auth, chat, feed, etc.)\
│   │   └── components/   (shared reusable UI elements)\
│   ├── css/\
│   ├── images/\
│   └── sounds/\
└── templates/\
    └── index.html        (single SPA shell)

Characteristics: 
- Pure Vanilla JavaScript (ES2026+) 
- No React/Angular/Vue
- Single HTML document (SPA format)
- Screaming Architecture / Clean Vertical Slices for source-code organization inside `js/features/`
- Modular ES modules
- Event delegation for dynamic DOM
- API-driven UI state with REST + WebSockets

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