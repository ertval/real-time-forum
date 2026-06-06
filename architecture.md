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
- Each feature slice follows the `{feature}.api.js` / `{feature}.views.js` / `{feature}.page.js` separation (API client, view markup, page/controller lifecycle)
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

### SPA reaction wiring (B07)

The SPA reaction engine lives in `SPA/features/post/`:

-   `post.reactions.bindings.js` — a single delegated `change` listener
    attached to `document`, guarded by a module-level flag so it is bound
    exactly once regardless of how many feature pages call
    `initReactionBindings`. The listener survives SPA route changes (only
    the shell outlet is swapped) and re-rendered comment lists without
    re-binding.
-   `post.reactions.logic.js` — pure helpers, including
    `normalizeReactionState`, which reconciles the two backend payload
    shapes: list/detail (`likes` / `dislikes` / `my_reaction` as an int)
    and the reaction POST response (`likes_count` / `dislikes_count` /
    `reaction`). It exposes a canonical
    `{ likeCount, dislikeCount, userReaction }` so every view renders
    counts and active state uniformly.
-   `post.reactions.api.js` — the reaction POST client.

Reaction pill markup is shared via `renderPostReactions` /
`renderCommentReactions` (`post-card.views.js`) and reused by the feed,
post-detail, and activity views. Scope resolution matches on the
`data-post-id` / `data-comment-id` attribute (not a fixed tag), so the
feed/activity `<article data-post-id>` and the post-detail
`<section data-post-id>` both resolve correctly.

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
-   Mark-as-read endpoints (`GET /api/v1/notifications`,
    `PATCH /api/v1/notifications/{id}/read`,
    `PATCH /api/v1/notifications/read-all`)

Frontend (SPA feature slice `SPA/features/notification/`, mounted in the
persistent authenticated shell):

-   5s polling lifecycle started after authenticated boot/login and
    stopped on logout or `401` (no duplicate intervals; survives
    client-side route changes)
-   unread badge counter and dropdown panel
-   mark-one-read / mark-all-read with optimistic update reconciled by
    the next poll
-   click → mark-read → SPA-router navigation to
    `/posts/{id}?highlight={comment_id|last}` (no full page reload),
    with comment deep-link highlighting on the post detail view

> Toast and sound feedback from the legacy notification UI are not part
> of the SPA notification slice (out of scope for the B06 migration).

------------------------------------------------------------------------

## 8. My Activity Dashboard

The Activity view is a SPA feature slice mounted at `/activity` inside the
shared authenticated shell. It aggregates the user's forum activity in
collapsible sections:

-   created posts (with owner-only status toggle, edit, delete actions)
-   authored comments (with inline edit and delete)
-   liked posts
-   disliked posts

Implementation:

-   `SPA/features/activity/activity.api.js` — REST client, query-state
    helpers, and post/comment mutation calls
-   `SPA/features/activity/activity.views.js` — view markup
    (sections, post cards, comment entries, inline comment editor)
-   `SPA/features/activity/activity.page.js` — page controller
    (filters, pagination, owner/comment action dispatch, inline edit
    lifecycle)

Lifecycle:

-   The SPA router matches `/activity` ([SPA/core/router/routes.js](SPA/core/router/routes.js))
    and renders `renderActivityView()` into the shell outlet via
    `renderTemplate` ([SPA/core/router/render-template.js](SPA/core/router/render-template.js))
-   `runRouteInitializer` in
    [SPA/core/app/create-app.js](SPA/core/app/create-app.js) invokes
    `initActivityPage()` once the activity markup is mounted
-   Filter changes (status, items-per-section) call
    `history.replaceState`; pagination updates use SPA history APIs;
    activity mutations re-fetch through `refresh()` instead of triggering
    a full document reload

Navigation:

-   `/activity` is reached through SPA route transitions (e.g. the shell
    nav `<a data-link href="/activity">`)
-   Deep-link refresh on `/activity` is served by the SPA shell catch-all
    in `cmd/frontend`
-   Browser back/forward replays `popstate` through the SPA router

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