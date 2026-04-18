# Viability Report: Automating Manual E2E Verification

This report evaluates the feasibility of automating the "Manual E2E Verification" steps currently documented in the project's Pull Request (PR) messages for Track A tickets.

## 1. Executive Summary
Automating the "Manual E2E Verification" checks is **highly viable** and, in fact, **already partially implemented**. Most behavioral checks (routing, auth-gating, logout-logic) are covered by existing Vitest and Go integration tests. The remaining infrastructure-level checks (process management, fresh environment setup) can be automated using a dedicated verification script.

## 2. Analysis of Current Verification Steps

### 2.1 Behavioral Verification (SPA Logic)
Tickets **A03**, **A04**, **A05**, and **A06** rely heavily on verifying SPA behaviors:
- **Routes & Navigation**: Deep-linking, back/forward, no-reload transitions.
- **Auth Gating**: Redirection on startup based on session state.
- **Persistent UI**: Shell visibility across routes, logout control presence.
- **Logout Flow**: API invocation and state reset.

**Status**: **[ALREADY AUTOMATED]**
These are covered in `SPA/tests/unit/core/app/create-app.test.js` using a custom `browser-mock.js`.
- *Example*: `test('logout button calls API and returns user to login flow', ...)` explicitly verifies the manual check for A06.

### 2.2 Routing & Proxy Verification (Frontend Server)
Tickets **A02** and **A10** list manual checks for:
- Root access serving the SPA shell.
- Sub-routes (e.g., `/login`) falling back to `index.html`.
- Proxying `/api/` and `/ws` to the backend.

**Status**: **[ALREADY AUTOMATED]**
These are covered in Go:
- `cmd/frontend/server_test.go` (e.g., `TestSPARouting`) verifies sub-route fallback.
- `cmd/frontend/routes_proxy_test.go` verifies proxying behavior.

### 2.3 Infrastructure Verification (DevOps)
Tickets **A01** and **A02** list manual checks for:
- `make deps` bootstrapping from an empty state.
- `make run-all` and `make stop-all` managing processes correctly.

**Status**: **[PARTIALLY MANUAL]**
These currently require human intervention to run the commands and check `ps` or `pkill` results.

---

## 3. Automation Roadmap

To fully eliminate manual E2E verification, the following steps are recommended:

### Phase 1: Infrastructure Verification Script
Create a `scripts/verify-infrastructure.sh` that performs the "Manual A01" checks:
1. `rm -rf node_modules`
2. `make deps` (Check exit code 0)
3. `make build-all`
4. `make run-all &` (Background)
5. Wait for ports `:3000` and `:8080` to open.
6. `curl -f http://localhost:3000/` (Check for SPA shell)
7. `make stop-all`
8. Verify no `frontend` or `backend` processes remain.

### Phase 2: Promote Tests to E2E Tier
As defined in `SDS.md` (10.2.1), move the high-level journey tests from `SPA/tests/unit/` to `SPA/tests/e2e/`. Although they use mocks, they represent user-level journeys. Transitioning these to a real headless browser (e.g., Playwright) would provide "True E2E" but might exceed the "Go standard package" constraint if not handled via Bun/NPM.

### Phase 3: Automated PR Component Generation
Update the project's internal ticketing workflow (`.agents/workflows/impl-ticket.md`) to:
1. Run the test suite.
2. Harvest the PASS/FAIL results.
3. Replace "Manual E2E" with "Automated E2E Verification Results" in the PR template.

## 4. Conclusion
The infrastructure for automation is already strong. The project is currently repeating "Manual" steps in PR messages that are actually being validated by the `vitest` run in the CI pipeline. 

> [!TIP]
> **Recommendation**: Immediately replace the "Manual E2E Verification" section with "Automated Behavioral Verification" in upcoming PRs, referencing the specific Vitest/Go tests that satisfy the gate. Create a single shell script for the DevOps/Infrastructure checks to complete the automation.

---
