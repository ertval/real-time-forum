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
- A01 -> RTF-01, A03 -> RTF-03, A04 -> RTF-04, A08 -> RTF-08, A09 -> RTF-09, A10 -> RTF-33 (bonus)

## Suggested Execution Order
1. A01, A03 | 2. A04, A08 | 3. A09 | 4. A10 (bonus)

## Tickets

### A01 - Single SPA Shell Entry
Source: RTF-01 | Phase: P0
Depends on: None
Blocks: D09, A03

Work:
- serve one root HTML shell for the app
- stop relying on standalone templates for app routes
- keep static assets and existing API proxy behavior working

Verification Gate:
- the app is served through one HTML document
- direct navigation to supported SPA routes resolves successfully
- static assets load correctly
- `/api/` proxy behavior remains intact

### A03 - SPA Boot and Client Routing
Source: RTF-03 | Phase: P0
Depends on: A01
Blocks: A04, D10, A08, B01, B05

Work:
- add client-side route resolution for auth, feed, post detail, create/edit, and activity
- add initial SPA boot logic
- support browser history navigation and deep-link entry

Verification Gate:
- route changes happen without full document reloads
- browser back and forward work for supported routes
- pasted route URLs load the expected screen

### A04 - Persistent App Shell Layout
Source: RTF-04 | Phase: P1
Depends on: A03
Blocks: A09, B01, B04, B05, D01, B06

Work:
- build the shared authenticated app frame
- add persistent header and logout location
- add main content outlet and reserved chat sidebar containers

Verification Gate:
- authenticated routes render inside a shared shell
- navigation and logout remain visible during route changes
- layout reserves a stable location for chat

### A08 - Authenticated-Only Forum Access
Source: RTF-08 | Phase: P1
Depends on: A03
Blocks: A09, B01, B04, B05, D04, C08

Work:
- require auth for forum content endpoints
- gate SPA route access using `GET /api/v1/users/me`
- define authenticated vs unauthenticated app boot behavior

Verification Gate:
- unauthenticated users cannot access feed, posts, comments, activity, or chat
- authenticated users enter the forum shell directly
- app boot correctly routes users based on session state

### A09 - Global Logout Across the Forum
Source: RTF-09 | Phase: P1
Depends on: A04, A08
Blocks: D05

Work:
- centralize logout inside the persistent shell
- remove dependence on page-specific layouts
- preserve server-side session invalidation behavior

Verification Gate:
- logout is visible and usable from every authenticated route
- logout clears the session and returns the user to auth flow
- logout no longer depends on create/edit page layout

### A10 - User Profile Page (Bonus)
Source: RTF-33 | Phase: P5 (Bonus)
Depends on: C10, A04, A08
Blocks: D07

Work:
- add `GET /api/v1/users/{userID}/profile` endpoint returning public profile data
- add `/profile/:id` SPA route
- build profile page UI showing nickname, first name, last name, age, and gender
- add profile links from roster entries and user references in posts/comments

Verification Gate:
- navigating to a user profile shows all extended registration fields
- profile is accessible from the chat roster and post/comment author links
- profile data matches what was entered during registration
