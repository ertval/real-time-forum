# Real-Time Forum Tickets by Section

## Purpose

This document breaks the PRD and SDS into implementation tickets grouped by the agreed six sections:

1. Foundation and SPA Shell
2. Authentication and User Profile
3. Forum Content Migration
4. Direct Messaging Backend
5. Direct Messaging Frontend
6. Migration, Testing, and Hardening

These tickets are written as an intermediate planning artifact. They are meant to be issue-ready enough for decomposition and assignment, while still preserving the system-level dependency map needed for later track planning.

## Sequencing Notes

- Section 1 and Section 2 start first because the rest of the work depends on the SPA shell and the auth model.
- Section 3 depends on the SPA shell and authenticated route model.
- Section 4 can start early because it is mostly backend-isolated once the user/profile schema direction is fixed.
- Section 5 depends on both the SPA shell and the chat backend contracts.
- Section 6 runs throughout the project, but final acceptance happens last.
- `Depends on` shows direct prerequisites only.
- `Blocks` shows direct downstream tickets only.

## 1. Foundation and SPA Shell

### RTF-01: Replace Multi-Page Template Serving with a Single SPA Shell

Goal:
Serve one HTML entry point for the application while preserving static assets and backend API proxying.

Scope:

- introduce one root HTML shell for the frontend
- stop relying on separate HTML templates for home, login, register, create-post, edit-post, activity, and view-post
- update frontend-server route handling so SPA routes resolve to the shell

Depends on:

- none

Blocks:

- RTF-02
- RTF-03

Acceptance Criteria:

- the app is served through one HTML document
- direct navigation to supported SPA routes resolves successfully
- static assets still load correctly
- `/api/` proxy behavior remains intact

### RTF-02: Add Frontend Support for WebSocket Proxying

Goal:
Allow the frontend server to proxy WebSocket traffic to the backend.

Scope:

- add proxy handling for `/ws`
- preserve existing backend proxy behavior for REST APIs
- ensure the proxy works with authenticated cookies

Depends on:

- RTF-01

Blocks:

- RTF-24

Acceptance Criteria:

- a browser client can connect to `/ws` through the frontend server
- the proxied request reaches the backend with the session cookie intact
- REST proxy behavior is not regressed

### RTF-03: Build SPA Bootstrapping and Client-Side Routing

Goal:
Create the JavaScript routing and app boot flow that replaces full-page navigation.

Scope:

- create route resolution for login, register, feed, post detail, create post, edit post, and activity
- add initial app boot logic
- support browser history and back or forward navigation

Depends on:

- RTF-01

Blocks:

- RTF-04
- RTF-07
- RTF-08
- RTF-10
- RTF-14

Acceptance Criteria:

- route changes happen without full document reloads
- browser back and forward work for supported routes
- route entry from a pasted URL loads the expected screen

### RTF-04: Build the Persistent App Shell Layout

Goal:
Provide the shared application frame used by all authenticated routes.

Scope:

- add persistent header and navigation area
- add persistent logout access
- add the main route outlet
- add reserved chat sidebar containers for later chat work

Depends on:

- RTF-03

Blocks:

- RTF-09
- RTF-10
- RTF-13
- RTF-14
- RTF-21
- RTF-29

Acceptance Criteria:

- authenticated routes render inside a shared shell
- navigation and logout remain visible while moving between routes
- layout reserves a stable location for the chat panel

## 2. Authentication and User Profile

### RTF-05: Extend the Users Schema for Required Profile Fields

Goal:
Persist the additional registration fields required by the target product.

Scope:

- add `age`, `gender`, `first_name`, and `last_name` to the users schema
- update repository-layer user models and row scanning
- update user creation logic to persist the new fields

Depends on:

- none

Blocks:

- RTF-06
- RTF-15
- RTF-25

Acceptance Criteria:

- new users persist all required profile fields
- existing user reads do not break
- repository-level tests cover the new fields

### RTF-06: Extend Registration Validation and API Contracts

Goal:
Require and validate the new profile fields in the registration API.

Scope:

- extend the registration request payload
- validate age and required text fields
- keep existing username, email, and password validation behavior where still valid

Depends on:

- RTF-05

Blocks:

- RTF-26

Acceptance Criteria:

- registration rejects missing required profile fields
- registration accepts valid extended payloads
- successful registration still creates a valid session

### RTF-07: Build the SPA Login and Registration Views

Goal:
Own the authenticated entry UI in the SPA, including the extended registration form required by the product.

Scope:

- build the SPA login view
- build the SPA registration view
- add registration inputs for age, gender, first name, and last name
- support login with username or email plus password
- remove guest and OAuth entry choices from the auth UI

Depends on:

- RTF-03

Blocks:

- RTF-27

Acceptance Criteria:

- the SPA exposes login and registration routes
- the registration form includes all required profile fields
- the login UI supports username or email entry
- guest and OAuth entry options are removed from the product flow
- auth views are wired to the agreed endpoint shapes

### RTF-08: Enforce Authenticated-Only Forum Access

Goal:
Prevent unauthenticated users from reading forum content.

Scope:

- convert forum content endpoints from optional auth to required auth
- gate SPA route access at app boot with `GET /api/v1/users/me`
- define authenticated and unauthenticated route behavior in the frontend

Depends on:

- RTF-03

Blocks:

- RTF-09
- RTF-10
- RTF-13
- RTF-14
- RTF-24
- RTF-26

Acceptance Criteria:

- unauthenticated users cannot access feed, posts, comments, activity, or chat
- authenticated users enter the forum shell directly
- app boot correctly routes users based on session state

### RTF-09: Make Logout Reachable from Every Authenticated Screen

Goal:
Fix the current logout accessibility gap and keep logout behavior consistent in the SPA.

Scope:

- centralize logout into the persistent shell
- remove dependence on per-page layout quirks
- preserve server-side session invalidation behavior

Depends on:

- RTF-04
- RTF-08

Blocks:

- RTF-27

Acceptance Criteria:

- logout is visible and usable from every authenticated route
- logout clears the session and returns the user to the auth flow
- logout behavior no longer depends on create or edit page layout

## 3. Forum Content Migration

### RTF-10: Migrate the Feed View into the SPA

Goal:
Move the current home-page feed into a SPA route without full reloads.

Scope:

- render the posts feed inside the SPA outlet
- preserve category filtering and pagination behavior
- keep post card behavior compatible with SPA navigation

Depends on:

- RTF-03
- RTF-04
- RTF-08

Blocks:

- RTF-11
- RTF-12
- RTF-29
- RTF-31

Acceptance Criteria:

- the feed works inside the SPA shell
- feed filtering and pagination still function
- opening a post uses SPA navigation instead of full document navigation

### RTF-11: Remove Comment Rendering from Feed Cards

Goal:
Meet the requirement that comments are visible only when a post is opened.

Scope:

- remove comment-preview fetching from feed rendering
- preserve post-card rendering without embedded comments
- keep comment rendering for post detail only

Depends on:

- RTF-10

Blocks:

- RTF-27

Acceptance Criteria:

- feed cards no longer render comments
- feed loading no longer requests comment lists for each post
- post cards still render correctly without comments

### RTF-12: Build the SPA Post-Detail Route with Comment Loading

Goal:
Move the current post-detail behavior into a dedicated SPA route.

Scope:

- render a single post in the SPA outlet
- load comments only in this route
- preserve comment creation and existing post-detail interactions
- preserve comment image-upload behavior in the post-detail flow

Depends on:

- RTF-10

Blocks:

- RTF-27
- RTF-29

Acceptance Criteria:

- opening a post renders the post-detail route inside the SPA
- comments are loaded only on post detail
- comment submission still works from the post-detail view
- comment image uploads still work from the post-detail view

### RTF-13: Migrate Create and Edit Post Flows into SPA Routes

Goal:
Move post creation and editing into SPA views so they live inside the persistent shell.

Scope:

- migrate create-post flow
- migrate edit-post flow
- preserve image uploads, categories, and existing form behavior where still in scope

Depends on:

- RTF-04
- RTF-08

Blocks:

- RTF-27
- RTF-30

Acceptance Criteria:

- create and edit post screens render inside the SPA shell
- post create and update behaviors still work
- logout remains available while using these views

### RTF-14: Migrate the Activity View into a SPA Route

Goal:
Move the current activity screen into the SPA without bundling unrelated legacy integrations into one ticket.

Scope:

- render the activity view inside the SPA outlet
- preserve activity data loading
- replace hard page navigation with SPA route transitions

Depends on:

- RTF-03
- RTF-04
- RTF-08

Blocks:

- RTF-27
- RTF-28

Acceptance Criteria:

- activity is accessible inside the SPA
- activity data loads correctly inside the shared shell
- activity navigation no longer depends on a standalone template

### RTF-29: Reconnect Retained Notification Behavior in the SPA

Goal:
Preserve retained notification behavior after the move to the SPA shell.

Scope:

- preserve notification polling inside the SPA shell
- preserve unread badge and unread-count behavior
- preserve per-notification read behavior
- preserve mark-all-as-read behavior
- preserve notification click navigation and deep-link behavior
- remove remaining notification bootstrapping assumptions that depended on standalone page templates

Depends on:

- RTF-04
- RTF-10
- RTF-12

Blocks:

- RTF-27
- RTF-28

Acceptance Criteria:

- notifications still function from the SPA shell
- notification polling still functions from the SPA shell
- unread badge and unread-count behavior are preserved
- marking one notification or all notifications as read still works
- notification click behavior still navigates to the correct destination
- notification behavior no longer assumes page-specific template bootstrapping

### RTF-31: Reconnect Retained Reaction Behavior in the SPA

Goal:
Preserve retained post and comment reaction behavior after the move to the SPA shell.

Scope:

- preserve post reaction bindings after SPA migration
- preserve comment reaction bindings after SPA migration
- remove reaction bootstrapping assumptions that depended on standalone page templates

Depends on:

- RTF-10
- RTF-12

Blocks:

- RTF-27
- RTF-28

Acceptance Criteria:

- post reactions still work in SPA-rendered views
- comment reactions still work in SPA-rendered views
- reaction behavior no longer assumes page-specific template bootstrapping

### RTF-30: Preserve Draft Workflows inside SPA Routes

Goal:
Keep drafts usable after moving post flows into the SPA.

Scope:

- preserve save-draft behavior in the SPA create-post flow
- preserve draft editing and publish flows
- ensure draft entry points route correctly inside the SPA

Depends on:

- RTF-13

Blocks:

- RTF-27
- RTF-28

Acceptance Criteria:

- saving a draft still works from the SPA create-post view
- draft editing and publish flows still work
- users can navigate into draft-related flows without leaving the SPA

## 4. Direct Messaging Backend

### RTF-15: Add the Private Messages Schema and Repository Layer

Goal:
Introduce persistent direct-message storage for one-to-one chat.

Scope:

- add the `private_messages` table
- add repository functions for message creation and history lookup
- add indexes needed for pair-based history queries

Depends on:

- RTF-05

Blocks:

- RTF-16
- RTF-17
- RTF-20
- RTF-25

Acceptance Criteria:

- messages can be persisted between two distinct users
- message history can be queried by user pair
- schema and repository tests cover the new table

### RTF-16: Implement the Chat Roster Query and API

Goal:
Expose the user roster with both presence and last-message metadata needed by the chat sidebar.

Scope:

- add `GET /api/v1/chats`
- include all other users
- include `is_online` in the response contract
- compute latest-message metadata
- support the required ordering rules

Depends on:

- RTF-15
- RTF-18

Blocks:

- RTF-21
- RTF-26

Acceptance Criteria:

- the roster returns all users except the current user
- each roster entry includes presence and last-message metadata
- users with history sort by latest message timestamp descending
- users without history sort alphabetically after active conversations

### RTF-17: Implement the Chat History API with 10-Message Pagination

Goal:
Expose paginated direct-message history for the selected conversation.

Scope:

- add `GET /api/v1/chats/{userID}/messages`
- support latest-10 fetch
- support older-than pagination using `before_id`
- return messages in render-ready chronological order

Depends on:

- RTF-15

Blocks:

- RTF-22
- RTF-26

Acceptance Criteria:

- the first request returns the latest 10 messages
- paginated history returns older messages in batches of 10
- the response includes whether older history remains available

### RTF-18: Add an Authenticated WebSocket Endpoint and Connection Manager

Goal:
Establish the backend real-time transport used for presence and messaging.

Scope:

- add `GET /ws`
- validate authenticated session during upgrade
- manage connection lifecycle per user
- support multiple connections for the same user

Depends on:

- none

Blocks:

- RTF-16
- RTF-19
- RTF-20
- RTF-24
- RTF-26

Acceptance Criteria:

- authenticated users can open WebSocket connections
- unauthenticated connections are rejected
- connection lifecycle is tracked reliably for multiple tabs

### RTF-19: Implement Presence Broadcasting

Goal:
Provide real-time online or offline visibility for the chat roster.

Scope:

- keep in-memory presence state based on active connections
- emit `presence.snapshot` on connect
- emit `presence.update` on actual online-state transitions

Depends on:

- RTF-18

Blocks:

- RTF-20
- RTF-24
- RTF-26

Acceptance Criteria:

- a newly connected client receives a presence snapshot
- users transition to online when their first connection opens
- users transition to offline only when their last connection closes

### RTF-20: Implement Real-Time Direct Message Send and Delivery

Goal:
Handle `dm.send` events and broadcast new messages in real time.

Scope:

- validate outgoing message events
- reject invalid recipient states and invalid payloads
- persist valid messages
- emit `dm.message` to both sender and recipient
- emit `chat.error` for rejected sends

Depends on:

- RTF-15
- RTF-18
- RTF-19

Blocks:

- RTF-24
- RTF-26

Acceptance Criteria:

- valid direct messages are persisted and delivered to both parties
- empty, self-directed, and offline-recipient sends are rejected
- delivered events match the agreed SDS payload shape

## 5. Direct Messaging Frontend

### RTF-21: Build the Persistent Chat Roster UI

Goal:
Render the always-visible user list in the authenticated app shell.

Scope:

- render roster entries for all other users
- show online or offline state
- display last-message preview metadata where available
- reflect the required ordering rules in the UI
- keep all rostered users selectable, including offline users

Depends on:

- RTF-04
- RTF-16

Blocks:

- RTF-22
- RTF-24
- RTF-27

Acceptance Criteria:

- the chat roster is visible on authenticated routes
- presence state is shown per user
- roster order matches backend ordering rules
- all rostered users remain selectable, including offline users

### RTF-22: Build the Active Conversation Panel and Composer UI

Goal:
Render the selected direct-message conversation and sending surface.

Scope:

- load and render message history for the selected user
- display message timestamp and sender username
- add a composer UI for sending direct messages
- allow opening a conversation for any rostered user
- allow opening and reading conversation history for offline users
- render an empty conversation state when a selectable user has no prior history
- disable direct-message sending in the UI when the selected user is offline
- clearly indicate when the selected user is offline

Depends on:

- RTF-17
- RTF-21

Blocks:

- RTF-23
- RTF-24
- RTF-27

Acceptance Criteria:

- selecting a user loads the latest 10 messages
- messages display sender username and timestamp
- the composer and conversation panel render correctly for the selected user
- any rostered user can be selected, including offline users
- offline conversation history remains readable
- selectable users with no prior history render a valid empty conversation state
- direct-message sending is disabled in the UI for offline selected users
- offline state is communicated clearly in the chat UI

### RTF-23: Add Incremental History Loading with Scroll Protection

Goal:
Load older messages in controlled batches without spamming requests.

Scope:

- detect upward history loading threshold
- load 10 older messages at a time
- apply throttle or debounce to prevent duplicated rapid fetches
- preserve stable scroll behavior while prepending history

Depends on:

- RTF-22

Blocks:

- RTF-27

Acceptance Criteria:

- older messages load in batches of 10
- repeated scroll events do not produce duplicate burst requests
- the viewport remains usable when older messages are prepended

### RTF-24: Wire WebSocket Presence and Messaging into the Chat UI

Goal:
Connect the frontend chat interface to the real-time backend for both outbound and inbound flows.

Scope:

- open the WebSocket connection after authenticated app boot
- emit `dm.send` from the composer
- consume `presence.snapshot` and `presence.update`
- consume `dm.message` and `chat.error`
- update the active conversation and roster ordering live
- surface send and delivery errors in the UI

Depends on:

- RTF-02
- RTF-08
- RTF-18
- RTF-19
- RTF-20
- RTF-21
- RTF-22

Blocks:

- RTF-27
- RTF-28

Acceptance Criteria:

- presence changes update the roster without refresh
- outbound direct messages are emitted through the WebSocket contract
- incoming messages appear live in the active conversation
- new message activity reorders the roster correctly
- send and validation errors are surfaced clearly to the user

## 6. Migration, Testing, and Hardening

### RTF-25: Define and Implement the Database Migration Strategy

Goal:
Safely evolve existing data to the new schema.

Scope:

- introduce migration logic for new user fields and private messages
- preserve existing retained tables and data
- define how pre-existing users are handled after profile schema changes

Depends on:

- RTF-05
- RTF-15

Blocks:

- RTF-28

Acceptance Criteria:

- schema changes can be applied to an existing database
- existing user records remain usable
- retained forum data is preserved

### RTF-26: Add Backend Test Coverage for Auth, Messaging, and Presence

Goal:
Protect the new backend behavior with automated tests.

Scope:

- extended registration tests
- login-by-username tests
- login-by-email tests
- auth-gating tests for forum routes
- roster tests including presence in the API response
- message history pagination tests
- WebSocket authentication tests
- presence transition tests
- direct-message integration tests

Depends on:

- RTF-06
- RTF-08
- RTF-16
- RTF-17
- RTF-18
- RTF-19
- RTF-20

Blocks:

- RTF-28

Acceptance Criteria:

- backend tests cover the major auth and chat flows
- login by both supported identifiers is explicitly tested
- edge cases around offline recipients and multi-connection presence are covered
- the new test suite is stable and repeatable

### RTF-27: Add Frontend Regression Coverage for SPA, Auth, and Chat Flows

Goal:
Protect the new SPA behavior and real-time chat UX against regressions.

Scope:

- auth route rendering coverage for login and registration
- logout visibility coverage across authenticated routes
- SPA route-navigation coverage
- auth-boot and redirect coverage
- feed-versus-post-detail comment visibility coverage
- post create flow coverage
- post edit flow coverage
- activity view coverage
- post image-upload flow coverage
- comment image-upload flow coverage
- offline composer-state coverage
- offline-history readability coverage
- live message rendering coverage
- chat send-error handling coverage
- chat history-loading behavior coverage
- retained notification, reaction, and draft flow coverage

Depends on:

- RTF-07
- RTF-09
- RTF-11
- RTF-12
- RTF-13
- RTF-14
- RTF-21
- RTF-22
- RTF-23
- RTF-24
- RTF-29
- RTF-31
- RTF-30

Blocks:

- RTF-28

Acceptance Criteria:

- frontend behavior coverage exists for the major SPA transitions
- logout visibility is explicitly protected by tests
- comment visibility rules are protected by tests
- post and comment image-upload flows are protected by tests
- offline composer behavior, offline-history readability, and live message rendering are tested
- retained legacy flows covered by this phase are protected by tests

### RTF-28: Final Regression Sweep and Acceptance Validation

Goal:
Validate that the new real-time forum works end to end without breaking retained features.

Scope:

- execute a final acceptance checklist against the PRD and SDS
- verify retained legacy features still function
- identify cleanup items before merge or handoff

Depends on:

- RTF-14
- RTF-24
- RTF-25
- RTF-26
- RTF-27
- RTF-29
- RTF-31
- RTF-30

Blocks:

- none

Acceptance Criteria:

- the build satisfies the documented product success criteria
- retained features are confirmed working or explicitly flagged
- any remaining gaps are listed as follow-up tickets instead of hidden work

## Suggested Next Step

Use this file as the source for the next planning artifact:

- split the tickets into four developer tracks
- preserve ticket IDs
- keep the section grouping as the dependency map
