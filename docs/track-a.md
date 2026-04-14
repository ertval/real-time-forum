# Track A - Platform, SPA Shell, and Auth

## Mission

Track A owns the platform and auth foundation for the real-time forum:

- single-page shell
- route bootstrapping
- persistent app layout
- user-profile schema changes
- registration and login UI
- authenticated-only forum access
- global logout
- user profile page (bonus)

This is the main upstream track. The other tracks should assume Track A is the first shared dependency to land.

## Source Mapping

This track rebrands the following source tickets from `docs/tickets-by-section.md`:

- `TA01` -> `RTF-01`
- `TA03` -> `RTF-03`
- `TA04` -> `RTF-04`
- `TA08` -> `RTF-08`
- `TA09` -> `RTF-09`
- `TA10` -> `RTF-33` (bonus)

## Suggested Execution Order

1. `TA01`, `TA03`
2. `TA04`, `TA08`
3. `TA09`
4. `TA10` (bonus)

## Tickets

### TA01 - Single SPA Shell Entry

Source:

- `RTF-01`

Phase:

- `P0`

Work:

- serve one root HTML shell for the app
- stop relying on standalone templates for app routes
- keep static assets and existing API proxy behavior working

Depends on:

- None

Blocks:

- `TD09`
- `TA03`

Verification Gate:

- the app is served through one HTML document
- direct navigation to supported SPA routes resolves successfully
- static assets load correctly
- `/api/` proxy behavior remains intact

### TA03 - SPA Boot and Client Routing

Source:

- `RTF-03`

Phase:

- `P0`

Work:

- add client-side route resolution for auth, feed, post detail, create/edit, and activity
- add initial SPA boot logic
- support browser history navigation and deep-link entry

Depends on:

- `TA01`

Blocks:

- `TA04`
- `TD10`
- `TA08`
- `TB01`
- `TB05`

Verification Gate:

- route changes happen without full document reloads
- browser back and forward work for supported routes
- pasted route URLs load the expected screen

### TA04 - Persistent App Shell Layout

Source:

- `RTF-04`

Phase:

- `P1`

Work:

- build the shared authenticated app frame
- add persistent header and logout location
- add main content outlet and reserved chat sidebar containers

Depends on:

- `TA03`

Blocks:

- `TA09`
- `TB01`
- `TB04`
- `TB05`
- `TD01`
- `TB06`

Verification Gate:

- authenticated routes render inside a shared shell
- navigation and logout remain visible during route changes
- layout reserves a stable location for chat

### TA08 - Authenticated-Only Forum Access

Source:

- `RTF-08`

Phase:

- `P1`

Work:

- require auth for forum content endpoints
- gate SPA route access using `GET /api/v1/users/me`
- define authenticated vs unauthenticated app boot behavior

Depends on:

- `TA03`

Blocks:

- `TA09`
- `TB01`
- `TB04`
- `TB05`
- `TD04`
- `TC08`

Verification Gate:

- unauthenticated users cannot access feed, posts, comments, activity, or chat
- authenticated users enter the forum shell directly
- app boot correctly routes users based on session state

### TA09 - Global Logout Across the Forum

Source:

- `RTF-09`

Phase:

- `P1`

Work:

- centralize logout inside the persistent shell
- remove dependence on page-specific layouts
- preserve server-side session invalidation behavior

Depends on:

- `TA04`
- `TA08`

Blocks:

- `TD05`

Verification Gate:

- logout is visible and usable from every authenticated route
- logout clears the session and returns the user to auth flow
- logout no longer depends on create/edit page layout

### TA10 - User Profile Page (Bonus)

Source:

- `RTF-33`

Phase:

- `P5` (Bonus)

Work:

- add `GET /api/v1/users/{userID}/profile` endpoint returning public profile data
- add `/profile/:id` SPA route
- build profile page UI showing nickname, first name, last name, age, and gender
- add profile links from roster entries and user references in posts/comments

Depends on:

- `TC10`
- `TA04`
- `TA08`

Blocks:

- `TD07`

Verification Gate:

- navigating to a user profile shows all extended registration fields
- profile is accessible from the chat roster and post/comment author links
- profile data matches what was entered during registration
