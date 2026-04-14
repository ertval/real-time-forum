# PR Description: SPA Migration & Audit Bonus Preparation

This PR streamlines the project's documentation and implementation plan, ensuring absolute alignment with the `real-time-forum` exercise requirements while preparing the codebase for audit bonus points.

## Summary of Changes

### 1. Requirements & Audit Alignment
*   **Source of Truth Correction**: Swapped the contents of `docs/audit.md` and `docs/requirements.md` to correctly map the exercise specification as the source of truth and the audit checklist as the verification tool.
*   Updated all cross-references in `README.md`, `AGENTS.md`, and the `ticket-tracker.md` to reflect this change.

### 2. MVP-First Implementation Strategy
*   **Ticket Tracker Reorganization**: Reorganized the implementation schedule into **6 logical waves**.
    *   **Waves 1-3**: Focused on delivering a core MVP that satisfies all mandatory audit requirements.
    *   **Waves 4-5**: Focused on migrating retained legacy features (notifications, reactions, etc.).
    *   **Wave 6**: Focused on Audit Bonus Requirements.
*   Updated total ticket count to 35 to include mandatory, legacy, and bonus features.

### 3. Bonus Requirements Integration
*   **PRD & SDS Updates**: Explicitly defined requirements and technical designs for:
    *   **User Profiles**: Dedicated profile pages displaying registration data.
    *   **DM Image Attachments**: Enabling image sharing within private messages.
    *   **Performance/Concurrency**: Use of goroutines/channels (Go) and Promises (JS).
*   **New Tickets**: Added tickets `TA10` (Profile UI), `TC09` (Image Backend), and `TD08` (Image Frontend) to the respective track files.

### 4. Developer & Agent Tooling
*   **AGENTS.md**: Created/Refined a comprehensive guide for AI coding assistants, detailing architecture, coding conventions, and common pitfalls specific to this repo.
*   **README.md**: Updated the project features, API overview, and tech stack to reflect the shift from a multi-page app to a Real-Time SPA.

### 5. Validation
*   **Gap Analysis**: Performed a full trace of the 26 mandatory and 6 bonus requirements against the current ticket backlog, confirming 100% coverage.

## Key Files Impacted
- `README.md`
- `AGENTS.md`
- `docs/PRD.md`
- `docs/SDS.md`
- `docs/ticket-tracker.md`
- Track files: `track-a.md`, `track-c.md`, `track-d.md`
- Validation sources: `audit.md`, `requirements.md`
