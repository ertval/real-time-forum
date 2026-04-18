# Manual E2E Automation Viability (Implemented Tickets)

## Objective

Assess whether the Manual E2E Verification checklist items in implemented ticket PRs can be automated and enforced in CI.

## Scope

Implemented tickets in the current tracker:

- A01
- A02
- A10
- A03
- A04
- A05
- A06

## Executive Verdict

Yes, this is viable.

- Most Manual E2E checks for A03, A04, A05, A06, and A10 are already covered by existing automated tests.
- A01 and A02 checks are also automatable, but need dedicated infra/server smoke tests to remove remaining manual verification.
- A practical near-term target is >= 90% automation coverage of current Manual E2E checklist items without introducing new third-party runtime dependencies.

## Ticket-by-Ticket Viability Matrix

| Ticket | Manual E2E Items in PR | Automation Viability | Current State | Gap to Close |
|---|---|---|---|---|
| A01 | `make deps` bootstrap from empty env; `make run-all` + `make stop-all` process safety; `make test-frontend` behavior | High (partial for process lifecycle edge cases) | Command-level checks exist but no dedicated infra smoke test suite | Add deterministic shell smoke tests for bootstrap, lifecycle, and cleanup assertions |
| A02 | Root access, SPA access path, static asset reachability | High | Partially covered by frontend server tests | Add route contract tests that encode expected non-breaking behavior for this migration step |
| A10 | Single shell serving, deep-link direct navigation, static assets, `/api/` proxy intact | Very High | Already covered by `cmd/frontend` tests | Add explicit CI gate label for these tests (optional hardening) |
| A03 | No-reload transitions, back/forward behavior, deep-link screen resolution | Very High | Already covered by SPA unit tests + frontend route fallback tests | Mostly complete; add one browser-level no-document-reload sentinel check if desired |
| A04 | Shared shell persistence, nav/logout visibility, reserved chat layout regions | Very High | Already covered by SPA unit tests | No critical gap |
| A05 | Unauthenticated access blocked, authenticated shell entry, boot session routing | Very High | Already covered by backend auth matrix + SPA unit tests | No critical gap |
| A06 | Logout usable from all protected routes, logout returns to auth flow, layout-independent behavior | Very High | Already covered by SPA unit tests and backend logout contract test | No critical gap |

## What Is Already Automated (Evidence)

- SPA route/auth/logout/shell behavior has dedicated unit coverage in `SPA/tests/unit/core/app/create-app.test.js`.
- Route matcher normalization behavior is covered in `SPA/tests/unit/core/router/routes.test.js`.
- Frontend server SPA fallback and deep-link behavior is covered in `cmd/frontend/server_test.go`.
- Frontend API proxy cookie/path behavior is covered in `cmd/frontend/routes_proxy_test.go`.
- Backend auth-gating behavior is covered in `internal/tests/forum_auth_access_test.go` (referenced by A05 PR notes).

## Remaining Gaps and How to Automate Them

### 1) Infrastructure smoke checks (A01)

Add a CI-safe smoke script (or Go test wrapper) that:

1. Removes `node_modules` in a controlled temp workspace clone.
2. Runs `make deps` and verifies Bun + lock/install success.
3. Starts services (`make run-all`) and probes health/routes.
4. Runs `make stop-all` and asserts no lingering processes/ports.

Notes:

- This can be kept deterministic with bounded startup polling and explicit process-name checks.
- Keep it in a separate CI job so failures are isolated from unit/integration suites.

### 2) Route contract migration checks (A02)

Encode A02 migration invariants as frontend server tests:

1. Shell/route availability expectations for that milestone.
2. Static file accessibility contract.
3. Backward-compat behavior expected at that phase.

Notes:

- Some A02 "manual" expectations were phase-specific and later superseded by A10. Treat these as historical migration contracts, not permanent product contracts.

### 3) Optional browser-level sentinel (A03)

Current no-reload guarantees are strong in unit tests. If desired, add one lightweight browser-driven sentinel test to assert that internal nav does not trigger full document navigation under real browser semantics.

## Recommended Automation Architecture

Use three layers and map each ticket check to one of them:

1. Unit (Vitest Node): route rules, shell rendering, auth gating, logout flow.
2. Integration (Go `httptest` + API tests): server routing/proxy/auth contracts.
3. Environment smoke (scripted CI step): toolchain bootstrapping and process lifecycle.

This avoids introducing heavy browser tooling for checks that are already deterministic in existing unit/integration suites.

## CI Gate Proposal

Promote "Manual E2E" PR bullets into mandatory automated gates:

1. `make test` (existing)
2. `make lint` (existing)
3. `make verify-infra-smoke` (new target for A01-like checks)
4. `go test ./cmd/frontend -run SPA|Proxy` (explicit server routing/proxy gate)
5. `bun run test -- --run` scoped smoke subset for A03-A06 shell/auth/logout checks (optional split job)

## Effort Estimate

- A01 infra automation: Medium (1-2 days)
- A02 contract test hardening: Low-Medium (0.5-1 day)
- CI gate wiring + PR template alignment: Low (0.5 day)

Total: ~2-3.5 days for robust baseline automation of all implemented-ticket Manual E2E checks.

## Risks and Mitigations

- Risk: flaky process lifecycle checks in CI.
  - Mitigation: bounded retries, explicit readiness probes, strict cleanup traps.
- Risk: migration-era expectations conflict with newer ticket behavior.
  - Mitigation: classify tests as "historical migration contracts" vs "current product contracts".
- Risk: duplicated assertions across suites.
  - Mitigation: keep one ownership layer per behavior (unit vs integration vs infra smoke).

## Final Viability Decision

Automation of Manual E2E checks for implemented tickets is not only possible but mostly already achieved for feature behavior. Completing A01/A02 infra and route-contract smoke coverage will provide a consistent "manual-to-automated" PR verification model going forward.