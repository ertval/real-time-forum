# A04 Verification Manifest

## Scope
Ticket: **A04 - Persistent App Shell Layout**
Gate source: `docs/track-a.md` (A04 verification gate)

## Commands Run + Output Summary
1. `make test`
- Result: **PASS** (`EXIT:0`)
- Summary: Go test suite passed and invoked frontend checks; Vitest reported all frontend test files/tests passed.

2. `bun run test`
- Result: **PASS**
- Summary: `vitest run` passed with `2` test files and `12` tests passing.

3. `bun run check`
- Result: **PASS**
- Summary: `biome check .` completed with "Checked 66 files ... No fixes applied."

## Gate-by-Gate Evidence
### Gate 1: Authenticated routes render inside a shared shell
- Gate text: `docs/track-a.md:105`
- Implementation evidence:
  - Shared authenticated shell renderer: `SPA/main.js:141`
  - Auth shell wrapper marker: `SPA/main.js:143`
  - Protected route rendering path wraps route content in shell: `SPA/main.js:309`
- Test evidence:
  - Explicit test for shared shell on authenticated routes: `SPA/tests/unit/main.test.js:271`
  - Assert auth shell exists: `SPA/tests/unit/main.test.js:281`

### Gate 2: Navigation and logout remain visible during route changes
- Gate text: `docs/track-a.md:106`
- Implementation evidence:
  - Persistent forum navigation inside shell header: `SPA/main.js:131`
  - Persistent logout control in shell header: `SPA/main.js:147`
- Test evidence:
  - Route-change persistence test: `SPA/tests/unit/main.test.js:289`
  - Logout visibility assertions after navigation: `SPA/tests/unit/main.test.js:303`

### Gate 3: Layout reserves a stable location for chat
- Gate text: `docs/track-a.md:107`
- Implementation evidence:
  - Dedicated chat roster and active-chat containers in shell: `SPA/main.js:152`, `SPA/main.js:155`
  - Stable two-column shell layout with reserved sidebar track: `SPA/assets/css/main.css:174`, `SPA/assets/css/main.css:176`
  - Sticky chat region placement: `SPA/assets/css/main.css:194`, `SPA/assets/css/main.css:199`
- Test evidence:
  - Authenticated shell test asserts both chat regions exist: `SPA/tests/unit/main.test.js:284`, `SPA/tests/unit/main.test.js:285`

## Residual Risks
- The A04 checks validate shell persistence and layout reservation, but do not validate functional logout behavior (expected to be covered by A06 scope).
- Unit tests verify DOM output/state transitions; they do not exercise real browser rendering/layout edge cases under all viewport/device combinations.

VERIFIED: YES
