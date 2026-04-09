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

This is the main upstream track. The other tracks should assume Track A is the first shared dependency to land.

## Source Mapping

This track rebrands the following source tickets from `docs/tickets-by-section.md`:

- `TA01` -> `RTF-01`
- `TA02` -> `RTF-02`
- `TA03` -> `RTF-03`
- `TA04` -> `RTF-04`
- `TA05` -> `RTF-05`
- `TA06` -> `RTF-06`
- `TA07` -> `RTF-07`
- `TA08` -> `RTF-08`
- `TA09` -> `RTF-09`

## Suggested Execution Order

1. `TA01`, `TA03`, `TA05`
2. `TA04`, `TA06`, `TA07`, `TA08`
3. `TA09`
4. `TA02`

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

- `TA02`
- `TA03`

Verification Gate:

- the app is served through one HTML document
- direct navigation to supported SPA routes resolves successfully
- static assets load correctly
- `/api/` proxy behavior remains intact

### TA02 - Frontend WebSocket Proxy

Source:

- `RTF-02`

Phase:

- `P2`

Work:

- proxy `/ws` traffic from the frontend server to the backend
- preserve existing REST proxy behavior
- ensure authenticated cookies survive proxying

Depends on:

- `TA01`

Blocks:

- `TD04`

Verification Gate:

- a browser client can connect to `/ws` through the frontend server
- the proxied request reaches the backend with the session cookie intact
- REST proxy behavior is not regressed

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
- `TA07`
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

### TA05 - User Profile Schema Extension

Source:

- `RTF-05`

Phase:

- `P0`

Work:

- extend the users schema with `age`, `gender`, `first_name`, and `last_name`
- update repository models and scanning
- preserve existing user reads

Depends on:

- None

Blocks:

- `TA06`
- `TC02`
- `TC07`

Verification Gate:

- new users persist all required profile fields
- existing user reads do not break
- repository coverage exists for the new fields

### TA06 - Registration API Contract

Source:

- `RTF-06`

Phase:

- `P1`

Work:

- extend the registration payload and validation
- validate age and required profile text fields
- keep existing valid username, email, and password rules

Depends on:

- `TA05`

Blocks:

- `TC08`

Verification Gate:

- registration rejects missing required profile fields
- valid extended payloads succeed
- successful registration still creates a session

### TA07 - SPA Login and Registration Views

Source:

- `RTF-07`

Phase:

- `P1`

Work:

- build SPA login and registration screens
- add extended registration fields
- support login by username or email
- remove guest and OAuth entry paths from the UI

Depends on:

- `TA03`

Blocks:

- `TD05`

Verification Gate:

- login and registration routes render inside the SPA
- registration includes all required fields
- login accepts username or email entry
- guest and OAuth entry options are not exposed

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
