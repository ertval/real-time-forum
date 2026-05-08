# Audit Coverage & Testing Strategy Analysis (2026-04-18)

## 1. Functional Audit Status vs. Completed Tickets

This table maps the `audit.md` requirements to the currently implemented tickets (A01-A06, A10) and their functional capacity.

| Audit Category | Primary Ticket(s) | Status | Capacity | Rationale |
| :--- | :--- | :--- | :--- | :--- |
| **Infrastructure / Packages** | `A01` | Done | **Fully Functional** | CI/CD gates and dev tools are set up to enforce project standards and allowed packages. |
| **Auth Gating** | `A05` | Done | **Fully Functional** | The system correctly identifies unauthenticated users and redirects them to the login flow. Verified by E2E negative testing. |
| **Global Logout** | `A06` | Done | **Fully Functional** | The logout button is present in the global shell and correctly invalidates sessions across all routes. Verified by E2E. |
| **Registration Flow** | `D10, C11` | Not Started* | **Partly Functional** | **Visual implementation** of the registration form (including Age, Gender, etc.) exists in `auth.views.js`, but backend persistence is pending `C10`. |
| **Login Flow** | `D10, C11` | Not Started* | **Partly Functional** | Login logic is functional via API (used in tests); UI exists in `auth.views.js` but final wiring to the new contract is pending. |
| **Forum (Posts/Comments)** | `B01-B04` | Not Started | **Nothing** | Routes and UI for creating/viewing posts and comments are pending Track B. |
| **Chat (Roster/Presence)** | `D01, C05` | Not Started | **Nothing** | Roster UI and presence broadcasting logic are pending Track C/D. |
| **Private Messaging** | `D02, D04, C06` | Not Started | **Nothing** | WebSocket-based real-time messaging and UI are pending Track C/D. |
| **History / Scrolling** | `D03` | Not Started | **Nothing** | Pagination and throttled scroll loading are pending Track D. |

## 2. Testing Strategy for Foundations (Track A)

A common concern is how the Foundations (Track A) were verified before the User Interface (Track D) or the full Backend logic (Track C) were completed.

### 2.1 Headless Authentication (API Injection)
The E2E test suite ([tickets.test.js](file:///home/ertval/code/zone-modules/real-time-forum/SPA/tests/e2e/tickets.test.js)) bypasses the missing login forms by communicating directly with the backend API.
* **Method**: The tests use Playwright's `page.request` to send raw POST requests to `/api/v1/users/register`.
* **Verification**: This allows the test browser to obtain a session cookie, enabling the verification of the **Authenticated App Shell** (`A04`) and **Routing** (`A03`) without a manual login step.

### 2.2 Auth-Gate Verification (Negative Testing)
To verify **A05 (Authenticated-Only Access)** without a login form:
* The test suite clears all cookies.
* It attempts to access protected routes (e.g., `/`, `/activity`).
* It verifies that the SPA correctly detects the "Unauthorized" state and redirects the browser to the `/login` route.

### 2.3 Structural Integrity
Track A verification focused on the **DOM structure and State Management** rather than feature completeness:
* **Shell Stability**: Checking for the existence of persistent layout containers (`[data-auth-shell]`, `[data-chat-roster]`) that remain stable across route changes.
* **Global Logout**: Verifying that the logout action (triggered via a shell button) successfully invalidates the session cookie and updates the application state.

### 2.4 UI Readiness vs. Integration Status
While `D10` (SPA Login and Registration Views) is marked as "Not Started" in the tracker, the **visual views** for these screens were implemented during the early SPA shell development (`A03/A05`) to ensure a premium look and feel.
* **Registration**: The form already includes `Age`, `Gender`, `First Name`, and `Last Name` as required by the audit.
* **Login**: The form supports `Username/Email` entry.
* **Gap**: These views are currently using legacy API shapes or are not yet saving the extra profile metadata to the database, which is why the tickets remain open.

## 3. Future Verification Milestones

| Ticket | Goal |
| :--- | :--- |
| **D05** | Regression coverage for Forum features (Posts, Comments, Activity). |
| **D06** | Regression coverage for Chat features (Real-time, Roster, History). |
| **D07** | **Final Acceptance Validation**: A comprehensive audit pass using the full `audit.md` checklist. |

## 4. Discrepancy Analysis

| Feature | Audit Requirement | Current State | Risk / Mitigation |
| :--- | :--- | :--- | :--- |
| **User Profile Data** | Age, Gender, Name fields. | Fields visible in UI; missing in DB. | Mitigation: **C10** is prioritized in Wave 1/2. |
| **Chat Ordering** | Last message / Alpha. | Logic not implemented. | Mitigation: **C05** and **D01** will implement this together. |
| **Real-time Notify** | Notifications on new DM. | WebSocket event defined in SDS. | Mitigation: **C06** and **D04** will hook this up in Wave 3. |
| **Scroll Throttle** | No spam on scroll. | Technique identified (Throttle). | Mitigation: **D03** implementation will use the CSS-Tricks pattern. |

