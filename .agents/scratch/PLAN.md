# A04 Implementation Plan: Persistent App Shell Layout

## Sources Reviewed
- docs/requirements.md
- docs/audit.md
- docs/SDS.md
- AGENTS.md
- docs/track-a.md
- docs/ticket-tracker.md
- SPA/index.html
- SPA/main.js
- SPA/assets/css/main.css
- SPA/tests/unit/main.test.js

## Scope Guard
- Implement only ticket A04 (Persistent App Shell Layout).
- Preserve A03 route behavior (matching, deep links, history navigation).
- Do not implement A05 auth-gating changes beyond current bootstrap behavior.
- Do not implement A06 logout API behavior; include persistent logout UI location only.

## Files To Modify
- SPA/main.js
  - Add authenticated shell rendering that persists across protected route transitions.
  - Keep public routes (login/register) outside the authenticated shell.
  - Add stable shell regions: header, logout control location, content outlet, chat sidebar containers.
- SPA/index.html
  - Keep single SPA entry and app root.
  - Ensure shell can be mounted under the existing root without introducing extra HTML pages.
- SPA/assets/css/main.css
  - Add shell layout and responsive behavior.
  - Reserve stable chat sidebar containers for roster and active chat.
- SPA/tests/unit/main.test.js
  - Add A04 assertions for shell persistence, always-visible nav/logout location, and reserved chat regions.

## API And WS Contracts
- REST used: GET /api/v1/users/me
  - Role in A04: decide whether public view or authenticated shell is rendered on boot.
  - No new request/response shapes introduced in this ticket.
- WebSocket: N/A in A04
  - Reason: A04 only reserves layout containers for chat; no live chat transport wiring in scope.

## DB Schema Migrations
- N/A for A04.
- Reason: this ticket is frontend shell/layout only.

## Verification Checklist (Pass/Fail)
- [ ] authenticated routes render inside a shared shell
  - Pass: protected routes (/, /post/:id, /create-post, /edit-post/:id, /activity) render inside one common shell wrapper.
  - Fail: protected screens render as standalone pages without shared frame.
- [ ] navigation and logout remain visible during route changes
  - Pass: forum navigation and logout control location remain present while navigating protected routes.
  - Fail: nav or logout disappears on some protected routes.
- [ ] layout reserves a stable location for chat
  - Pass: protected routes always include dedicated roster and active-chat containers in a fixed sidebar area.
  - Fail: chat containers are missing or route-dependent.

## Test Plan
1. bun run test SPA/tests/unit/main.test.js
   - Proves existing A03 behavior still passes and new A04 checks pass.
2. bun run test
   - Proves broader frontend test suite regression safety.
3. make test
   - Proves project-wide Go and integration tests remain green after A04 frontend changes.

## Risks And Assumptions
- Assumption: A06 will wire the logout action later; A04 provides persistent placement now.
- Risk: render refactor may accidentally regress route access handling.
- Risk: responsive shell CSS may affect existing spacing in route content.
- Mitigation: keep route matching logic intact and add targeted shell persistence tests.
