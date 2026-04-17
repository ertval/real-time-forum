# Ticket Progress Tracker

This file tracks delivery progress for the real-time forum project.

## Team Assignment

| Track | Dev | Focus Area |
|-------|-----|------------|
| **A** | Dev 1 | Platform & SPA Foundations (shell, routing, auth-gating) |
| **B** | Dev 2 | Forum MVP & Features (feed, post detail, notifications, activity) |
| **C** | Dev 3 | Realtime Backend & Persistence (schema, WS server, chat APIs) |
| **D** | Dev 4 | Realtime Frontend & Verification (chat UI, WS integration, tests) |

---

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

- Total tickets: `37`
- Done: `4`
- Partially Implemented: `0`
- Not Started: `33`

---

## Implementation Order — MVP-First Strategy

The implementation is organized into **6 waves**. Waves 1–3 deliver a functioning **MVP** that satisfies all mandatory audit requirements. Waves 4–5 deliver retained legacy features and final verification. Wave 6 delivers audit bonus features.

### Wave 1 — Foundations (P0)

> **Goal:** SPA shell, client routing, user schema, and WebSocket transport — the building blocks everything else depends on.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 1 | [x] | **A01** | A | Infrastructure, CI/CD, and Dev Tools Setup | None | A02 |
| 2 | [x] | **A02** | A | Frontend Architecture & Directory Restructuring | A01 | A10, D09 |
| 3 | [x] | **A10** | A | Single SPA Shell Entry | A02 | D09, A03 |
| 4 | [ ] | **C10** | C | User Profile Schema Extension | None | C11, C02, C07 |
| 5 | [ ] | **C01** | C | Authenticated WebSocket Endpoint and Connection Manager | None | C04, C05, C06, D04, C08 |
| 6 | [x] | **A03** | A | SPA Boot and Client Routing | A10 | A04, D10, A05, B01, B05 |

### Wave 2 — Auth + Shell + Core Forum (P1)

> **Goal:** Auth gating, persistent shell, login/register UI, feed, post detail, comments — the core forum experience required by the audit.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 6 | [ ] | **C11** | C | Registration API Contract | C10 | C08 |
| 7 | [ ] | **A04** | A | Persistent App Shell Layout | A03 | A06, B01, B04, B05, D01, B06 |
| 8 | [ ] | **A05** | A | Authenticated-Only Forum Access | A03 | A06, B01, B04, B05, D04, C08 |
| 9 | [ ] | **D10** | D | SPA Login and Registration Views | A03 | D05 |
| 10 | [ ] | **A06** | A | Global Logout Across the Forum | A04, A05 | D05 |
| 11 | [ ] | **B01** | B | Feed Route in the SPA | A03, A04, A05 | B02, B03, B06, B07 |
| 12 | [ ] | **B02** | B | Remove Feed Comment Rendering | B01 | D05 |
| 13 | [ ] | **B03** | B | Post Detail Route and Comment Flow | B01 | D05, B06, B07 |
| 14 | [ ] | **B04** | B | Create and Edit Post SPA Flows | A04, A05 | D05, B08 |
| 15 | [ ] | **C02** | C | Private Messages Schema and Repository Layer | C10 | C03, C05, C06, C07 |

### Wave 3 — Real-Time Chat MVP (P2 + P3)

> **Goal:** Full chat system — roster, history, presence, DM delivery, and browser integration. Completing this wave satisfies **all mandatory audit requirements**.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 16 | [ ] | **D09** | D | Frontend WebSocket Proxy | A02 | D04 |
| 17 | [ ] | **C03** | C | Chat History API | C02 | D02, C08 |
| 18 | [ ] | **C04** | C | Presence Broadcasting | C01 | C06, D04, C08 |
| 19 | [ ] | **C05** | C | Chat Roster API | C02, C01 | D01, C08 |
| 20 | [ ] | **C06** | C | Realtime DM Send and Delivery | C02, C01, C04 | D04, C08 |
| 21 | [ ] | **D01** | D | Persistent Chat Roster UI | A04, C05 | D02, D04, D06 |
| 22 | [ ] | **D02** | D | Active Conversation Panel and Composer | C03, D01 | D03, D04, D06 |
| 23 | [ ] | **D03** | D | Incremental History Loading | D02 | D06 |
| 24 | [ ] | **D04** | D | Browser WebSocket Chat Integration | D09, A05, C01, C04, C06, D01, D02 | D06, D07 |

### Wave 4 — Retained Legacy Features (P2)

> **Goal:** Migrate retained non-chat features (activity, notifications, reactions, drafts) into the SPA. These are preserved features from the previous forum, not strictly required by the real-time-forum audit but part of the product.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 25 | [ ] | **B05** | B | Activity View in the SPA | A03, A04, A05 | D05, D07 |
| 26 | [ ] | **B06** | B | Notification Behavior in the SPA | A04, B01, B03 | D05, D07 |
| 27 | [ ] | **B07** | B | Reaction Behavior in the SPA | B01, B03 | D05, D07 |
| 28 | [ ] | **B08** | B | Draft Workflows in the SPA | B04 | D05, D07 |

### Wave 5 — Bonus Features (P5)

> **Goal:** Audit bonus points — user profiles, DM image attachments, and concurrency patterns. These are optional features that score bonus audit points.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 29 | [ ] | **A07** | A | User Profile Page (Bonus) | C10, A04, A05 | D07 |
| 30 | [ ] | **C09** | C | DM Image Upload Backend (Bonus) | C02, C06 | D08, D07 |
| 31 | [ ] | **D08** | D | DM Image Rendering Frontend (Bonus) | C09, D02, D04 | D07 |

### Wave 6 — Finalization & Acceptance (P3–P4)

> **Goal:** Database migration, full test coverage, and final project acceptance.

| # | Status | Ticket | Track | Description | Depends on | Blocks |
|---|--------|--------|-------|-------------|------------|--------|
| 32 | [ ] | **C07** | C | Database Migration Strategy | C10, C02 | D07 |
| 33 | [ ] | **C08** | C | Backend Test Coverage for Auth, Messaging, and Presence | C11, A05, C01, C03, C04, C05, C06 | D07 |
| 34 | [ ] | **D05** | D | SPA and Forum Frontend Regression Coverage | D10, A06, B02, B03, B04, B05, B06, B07, B08 | D07 |
| 35 | [ ] | **D06** | D | Chat Frontend Regression Coverage | D01, D02, D03, D04 | D07 |
| 36 | [ ] | **D07** | D | Final Acceptance Validation | B05, B06, B07, B08, C07, C08, D04, D05, D06, A07, C09, D08 | None |

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

- Track A: `A01` through `A10`
- Track B: `B01` through `B08`
- Track C: `C01` through `C11`
- Track D: `D01` through `D10`

## Cross-Document References

- Exercise specification: `docs/requirements.md`
- Audit checklist: `docs/audit.md`
- Product requirements: `docs/PRD.md`
- Technical design: `docs/SDS.md`
- Track A definitions: `docs/track-a.md`
- Track B definitions: `docs/track-b.md`
- Track C definitions: `docs/track-c.md`
- Track D definitions: `docs/track-d.md`
