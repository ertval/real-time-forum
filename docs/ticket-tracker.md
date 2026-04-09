# Ticket Progress Tracker

This file tracks delivery progress for the final `A / B / C / D` track split.

Detailed ticket definitions live in:

- `docs/track-a.md`
- `docs/track-b.md`
- `docs/track-c.md`
- `docs/track-d.md`

The original coverage and source-ticket mapping remain canonical in:

- `docs/tickets-by-section.md`
- `docs/PRD.md`
- `docs/SDS.md`

## Update Rules

1. Keep each ticket in the line format: status + ticket ID + short description + dependency fields.
2. Use `[x]` only when the verification gate in the owning track file is satisfied.
3. Use `[-]` only when a meaningful subset of that ticket already exists in code.
4. Keep `Depends on` and `Blocks` synchronized with the owning track file when ticket definitions change.
5. Do not remove completed tickets from the tracker.

## Status Legend

- `[ ]` = Not Started
- `[-]` = Partially Implemented / In Progress
- `[x]` = Done

## Execution Policy (Low-Blocking First)

1. Respect the canonical phase order: `P0 -> P1 -> P2 -> P3 -> P4`.
2. Inside each phase, prioritize tickets that unblock the most other tracks.
3. Track `A` owns platform, SPA shell, auth, and shared boot behavior.
4. Track `B` owns forum content migration and retained non-chat forum UX.
5. Track `C` owns realtime backend, persistence, migrations, and backend validation.
6. Track `D` owns browser realtime chat integration, frontend regression coverage, and final acceptance.

## Summary Snapshot

- Total tickets: `31`
- Done: `0`
- Partially Implemented: `0`
- Not Started: `31`

## Low-Blocking Claim Queue (Global)

Use this as the default claim order for the next wave of work:

1. **Q0 P0 Shared Foundations**: `TA01`, `TA03`, `TA05`, `TC01`
2. **Q1 P1 Shell/Auth Completion + Forum Base**: `TA04`, `TA06`, `TA07`, `TA08`, `TA09`, `TB01`, `TB04`, `TB05`, `TC02`
3. **Q2 P2 Independent Build-Out**: `TA02`, `TB02`, `TB03`, `TB06`, `TB07`, `TB08`, `TC03`, `TC04`, `TC05`, `TC06`
4. **Q3 P3 Browser Integration**: `TD01`, `TD02`, `TD03`, `TD04`, `TC07`
5. **Q4 P4 Verification**: `TC08`, `TD05`, `TD06`

## Ticket ID Index

- Track A: `TA01` through `TA09`
- Track B: `TB01` through `TB08`
- Track C: `TC01` through `TC08`
- Track D: `TD01` through `TD06`

## Ordered Tickets By Track

### Track A

- [ ] **TA01** P0 - Single SPA Shell Entry | Serve one HTML app shell and route SPA paths through it while keeping static assets and `/api` proxying intact. (Depends on: None) | Blocks: TA02; TA03
- [ ] **TA02** P2 - Frontend WebSocket Proxy | Proxy `/ws` through the frontend server with authenticated cookies preserved. (Depends on: TA01) | Blocks: TD04
- [ ] **TA03** P0 - SPA Boot and Client Routing | Build app boot, client-side routes, browser history support, and deep-link entry. (Depends on: TA01) | Blocks: TA04; TA07; TA08; TB01; TB05
- [ ] **TA04** P1 - Persistent App Shell Layout | Build the shared authenticated shell with navigation, logout, main outlet, and chat container. (Depends on: TA03) | Blocks: TA09; TB01; TB04; TB05; TD01; TB06
- [ ] **TA05** P0 - User Profile Schema Extension | Add and persist `age`, `gender`, `first_name`, and `last_name` in the users schema and repository layer. (Depends on: None) | Blocks: TA06; TC02; TC07
- [ ] **TA06** P1 - Registration API Contract | Extend registration payload validation and keep valid existing auth rules intact. (Depends on: TA05) | Blocks: TC08
- [ ] **TA07** P1 - SPA Login and Registration Views | Build SPA auth views with extended registration fields and username-or-email login UX. (Depends on: TA03) | Blocks: TD05
- [ ] **TA08** P1 - Authenticated-Only Forum Access | Gate forum content behind authenticated app boot and authenticated backend access rules. (Depends on: TA03) | Blocks: TA09; TB01; TB04; TB05; TD04; TC08
- [ ] **TA09** P1 - Global Logout Across the Forum | Make logout reachable from every authenticated screen through the shared shell. (Depends on: TA04; TA08) | Blocks: TD05

### Track B

- [ ] **TB01** P1 - Feed Route in the SPA | Move the feed into the SPA outlet with filtering, pagination, and SPA-style post navigation. (Depends on: TA03; TA04; TA08) | Blocks: TB02; TB03; TB06; TB07
- [ ] **TB02** P1 - Remove Feed Comment Rendering | Remove comment previews from feed cards and stop feed-level comment fetching. (Depends on: TB01) | Blocks: TD05
- [ ] **TB03** P1 - Post Detail Route and Comment Flow | Build SPA post detail with comments, comment creation, and comment image uploads. (Depends on: TB01) | Blocks: TD05; TB06; TB07
- [ ] **TB04** P1 - Create and Edit Post SPA Flows | Move create/edit post flows into the SPA while preserving image and category behavior. (Depends on: TA04; TA08) | Blocks: TD05; TB08
- [ ] **TB05** P1 - Activity View in the SPA | Move the activity screen into the shared shell and preserve activity data loading. (Depends on: TA03; TA04; TA08) | Blocks: TD05; TD06
- [ ] **TB06** P2 - Notification Behavior in the SPA | Preserve polling, unread counts, mark-read behavior, and click/deep-link navigation after SPA migration. (Depends on: TA04; TB01; TB03) | Blocks: TD05; TD06
- [ ] **TB07** P2 - Reaction Behavior in the SPA | Preserve post and comment reaction behavior after SPA migration. (Depends on: TB01; TB03) | Blocks: TD05; TD06
- [ ] **TB08** P2 - Draft Workflows in the SPA | Preserve save-draft, edit-draft, and publish-draft behavior inside SPA routes. (Depends on: TB04) | Blocks: TD05; TD06

### Track C

- [ ] **TC01** P0 - Authenticated WebSocket Endpoint and Connection Manager | Add the backend WebSocket transport with session validation and per-user connection tracking. (Depends on: None) | Blocks: TC04; TC05; TC06; TD04; TC08
- [ ] **TC02** P1 - Private Messages Schema and Repository Layer | Add `private_messages` persistence and pair-based history lookup. (Depends on: TA05) | Blocks: TC03; TC05; TC06; TC07
- [ ] **TC03** P2 - Chat History API | Add latest-10 and older-than paginated history API for direct messages. (Depends on: TC02) | Blocks: TD02; TC08
- [ ] **TC04** P2 - Presence Broadcasting | Add presence snapshots and online/offline transition updates from active socket state. (Depends on: TC01) | Blocks: TC06; TD04; TC08
- [ ] **TC05** P2 - Chat Roster API | Add roster API with presence and last-message metadata plus required ordering rules. (Depends on: TC02; TC01) | Blocks: TD01; TC08
- [ ] **TC06** P2 - Realtime DM Send and Delivery | Validate, persist, and broadcast `dm.send` events plus error responses. (Depends on: TC02; TC01; TC04) | Blocks: TD04; TC08
- [ ] **TC07** P3 - Database Migration Strategy | Define and implement migration handling for new user fields and direct messages. (Depends on: TA05; TC02) | Blocks: TD06
- [ ] **TC08** P4 - Backend Test Coverage for Auth, Messaging, and Presence | Add backend coverage for auth modes, auth gating, roster/history, websocket auth, presence, and direct-message integration. (Depends on: TA06; TA08; TC01; TC03; TC04; TC05; TC06) | Blocks: TD06

### Track D

- [ ] **TD01** P3 - Persistent Chat Roster UI | Build the always-visible roster UI with presence, ordering, previews, and offline-user selection. (Depends on: TA04; TC05) | Blocks: TD02; TD04; TD05
- [ ] **TD02** P3 - Active Conversation Panel and Composer | Build selected-conversation rendering, offline-aware composer state, and empty conversation state. (Depends on: TC03; TD01) | Blocks: TD03; TD04; TD05
- [ ] **TD03** P3 - Incremental History Loading | Add throttled or debounced upward history loading in batches of `10`. (Depends on: TD02) | Blocks: TD05
- [ ] **TD04** P3 - Browser WebSocket Chat Integration | Wire browser-side WebSocket connect, send, receive, presence, ordering, and error handling. (Depends on: TA02; TA08; TC01; TC04; TC06; TD01; TD02) | Blocks: TD05; TD06
- [ ] **TD05** P4 - Frontend Regression Coverage | Add frontend coverage for auth, SPA navigation, content migration, retained forum UX, offline chat behavior, and live messaging. (Depends on: TA07; TA09; TB02; TB03; TB04; TB05; TB06; TB07; TB08; TD01; TD02; TD03; TD04) | Blocks: TD06
- [ ] **TD06** P4 - Final Acceptance Validation | Execute the final PRD/SDS acceptance sweep and record any remaining gaps as follow-up work. (Depends on: TB05; TB06; TB07; TB08; TC07; TC08; TD04; TD05) | Blocks: None

## Cross-Document References

- Source backlog and canonical mapping: `docs/tickets-by-section.md`
- Product requirements: `docs/PRD.md`
- Technical design: `docs/SDS.md`
- Track A definitions: `docs/track-a.md`
- Track B definitions: `docs/track-b.md`
- Track C definitions: `docs/track-c.md`
- Track D definitions: `docs/track-d.md`
