# Ticket Progress Tracker

This file tracks delivery progress for the real-time forum project.

Detailed ticket definitions live in:

- `docs/track-a.md`
- `docs/track-b.md`
- `docs/track-c.md`
- `docs/track-d.md`

The canonical product and technical requirements are in:

- `docs/requirements.md` — exercise specification (source of truth)
- `docs/audit.md` — audit checklist questions (source of truth)
- `docs/PRD.md` — product requirements document
- `docs/SDS.md` — software design specification

## Update Rules

1. Keep each ticket in the line format: status + ticket ID + short description + dependency fields.
2. Use `[x]` only when the verification gate in the owning track file is satisfied.
3. Use `[-]` only when a meaningful subset of that ticket already exists in code.
4. Keep `Depends on` and `Blocks` synchronized with the owning track file when ticket definitions change.
5. Do not remove completed tickets from the tracker.
6. A test ticket may start early, even if some lower-priority earlier-phase work is still open, only when all of its direct dependencies are already complete.

## Status Legend

- `[ ]` = Not Started
- `[-]` = Partially Implemented / In Progress
- `[x]` = Done

## Summary Snapshot

- Total tickets: `35`
- Done: `0`
- Partially Implemented: `0`
- Not Started: `35`

---

## Implementation Order — MVP-First Strategy

The implementation is organized into **6 waves**. Waves 1–3 deliver a functioning **MVP** that satisfies all mandatory audit requirements. Waves 4–5 deliver retained legacy features and final verification. Wave 6 delivers audit bonus features.

### Wave 1 — Foundations (P0)

> **Goal:** SPA shell, client routing, user schema, and WebSocket transport — the building blocks everything else depends on.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 1 | [ ] | **TA01** | A | Single SPA Shell Entry | None | TD09, TA03 |
| 2 | [ ] | **TC10** | C | User Profile Schema Extension | None | TC11, TC02, TC07 |
| 3 | [ ] | **TC01** | C | Authenticated WebSocket Endpoint and Connection Manager | None | TC04, TC05, TC06, TD04, TC08 |
| 4 | [ ] | **TA03** | A | SPA Boot and Client Routing | TA01 | TA04, TD10, TA08, TB01, TB05 |

### Wave 2 — Auth + Shell + Core Forum (P1)

> **Goal:** Auth gating, persistent shell, login/register UI, feed, post detail, comments — the core forum experience required by the audit.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 5 | [ ] | **TC11** | C | Registration API Contract | TC10 | TC08 |
| 6 | [ ] | **TA04** | A | Persistent App Shell Layout | TA03 | TA09, TB01, TB04, TB05, TD01, TB06 |
| 7 | [ ] | **TA08** | A | Authenticated-Only Forum Access | TA03 | TA09, TB01, TB04, TB05, TD04, TC08 |
| 8 | [ ] | **TD10** | D | SPA Login and Registration Views | TA03 | TD05 |
| 9 | [ ] | **TA09** | A | Global Logout Across the Forum | TA04, TA08 | TD05 |
| 10 | [ ] | **TB01** | B | Feed Route in the SPA | TA03, TA04, TA08 | TB02, TB03, TB06, TB07 |
| 11 | [ ] | **TB02** | B | Remove Feed Comment Rendering | TB01 | TD05 |
| 12 | [ ] | **TB03** | B | Post Detail Route and Comment Flow | TB01 | TD05, TB06, TB07 |
| 13 | [ ] | **TB04** | B | Create and Edit Post SPA Flows | TA04, TA08 | TD05, TB08 |
| 14 | [ ] | **TC02** | C | Private Messages Schema and Repository Layer | TC10 | TC03, TC05, TC06, TC07 |

### Wave 3 — Real-Time Chat MVP (P2 + P3)

> **Goal:** Full chat system — roster, history, presence, DM delivery, and browser integration. Completing this wave satisfies **all mandatory audit requirements**.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 15 | [ ] | **TD09** | D | Frontend WebSocket Proxy | TA01 | TD04 |
| 16 | [ ] | **TC03** | C | Chat History API | TC02 | TD02, TC08 |
| 17 | [ ] | **TC04** | C | Presence Broadcasting | TC01 | TC06, TD04, TC08 |
| 18 | [ ] | **TC05** | C | Chat Roster API | TC02, TC01 | TD01, TC08 |
| 19 | [ ] | **TC06** | C | Realtime DM Send and Delivery | TC02, TC01, TC04 | TD04, TC08 |
| 20 | [ ] | **TD01** | D | Persistent Chat Roster UI | TA04, TC05 | TD02, TD04, TD06 |
| 21 | [ ] | **TD02** | D | Active Conversation Panel and Composer | TC03, TD01 | TD03, TD04, TD06 |
| 22 | [ ] | **TD03** | D | Incremental History Loading | TD02 | TD06 |
| 23 | [ ] | **TD04** | D | Browser WebSocket Chat Integration | TD09, TA08, TC01, TC04, TC06, TD01, TD02 | TD06, TD07 |

### Wave 4 — Retained Legacy Features (P2)

> **Goal:** Migrate retained non-chat features (activity, notifications, reactions, drafts) into the SPA. These are preserved features from the previous forum, not strictly required by the real-time-forum audit but part of the product.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 24 | [ ] | **TB05** | B | Activity View in the SPA | TA03, TA04, TA08 | TD05, TD07 |
| 25 | [ ] | **TB06** | B | Notification Behavior in the SPA | TA04, TB01, TB03 | TD05, TD07 |
| 26 | [ ] | **TB07** | B | Reaction Behavior in the SPA | TB01, TB03 | TD05, TD07 |
| 27 | [ ] | **TB08** | B | Draft Workflows in the SPA | TB04 | TD05, TD07 |

### Wave 5 — Migration, Testing, and Acceptance (P3–P4)

> **Goal:** Database migration, backend and frontend test coverage, and final acceptance validation.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 28 | [ ] | **TC07** | C | Database Migration Strategy | TC10, TC02 | TD07 |
| 29 | [ ] | **TC08** | C | Backend Test Coverage for Auth, Messaging, and Presence | TC11, TA08, TC01, TC03, TC04, TC05, TC06 | TD07 |
| 30 | [ ] | **TD05** | D | SPA and Forum Frontend Regression Coverage | TD10, TA09, TB02, TB03, TB04, TB05, TB06, TB07, TB08 | TD07 |
| 31 | [ ] | **TD06** | D | Chat Frontend Regression Coverage | TD01, TD02, TD03, TD04 | TD07 |
| 32 | [ ] | **TD07** | D | Final Acceptance Validation | TB05, TB06, TB07, TB08, TC07, TC08, TD04, TD05, TD06, TA10, TC09, TD08 | None |

### Wave 6 — Bonus Features (P5)

> **Goal:** Audit bonus points — user profiles, DM image attachments, and concurrency patterns. These are optional features that score bonus audit points.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 33 | [ ] | **TA10** | A | User Profile Page (Bonus) | TC10, TA04, TA08 | TD07 |
| 34 | [ ] | **TC09** | C | DM Image Upload Backend (Bonus) | TC02, TC06 | TD08, TD07 |
| 35 | [ ] | **TD08** | D | DM Image Rendering Frontend (Bonus) | TC09, TD02, TD04 | TD07 |

---

## MVP Boundary

Completing **Waves 1–3** (tickets 1–23) delivers a fully functioning real-time forum that passes **all mandatory audit requirements**:

- ✅ SPA with single HTML file
- ✅ Registration with all required fields
- ✅ Login with nickname or email
- ✅ Auth-gated forum access
- ✅ Logout from any page
- ✅ Posts with categories in feed
- ✅ Comments only on post detail
- ✅ Online/offline user list
- ✅ Roster ordered by last message / alphabetical
- ✅ Private messaging in real time
- ✅ Message format with date and username
- ✅ Last 10 messages loaded initially
- ✅ Scroll-up pagination with throttle/debounce
- ✅ Real-time notification of new messages

---

## Ticket ID Index

- Track A: `TA01` through `TA10`
- Track B: `TB01` through `TB08`
- Track C: `TC01` through `TC11`
- Track D: `TD01` through `TD10`

## Cross-Document References

- Exercise specification: `docs/requirements.md`
- Audit checklist: `docs/audit.md`
- Product requirements: `docs/PRD.md`
- Technical design: `docs/SDS.md`
- Track A definitions: `docs/track-a.md`
- Track B definitions: `docs/track-b.md`
- Track C definitions: `docs/track-c.md`
- Track D definitions: `docs/track-d.md`
