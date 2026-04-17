---
name: Implement-ticket
description: Orchestrate the implementation of a specific ticket for the Real-Time Forum.
agent: orchestrator
tools: [browser_subagent, run_command, view_file, write_to_file, multi_replace_file_content, search_web]
---

# Ticket Orchestration: {Ticket}

You are the **Orchestrator**. Your goal is to coordinate a team of specialized subagents to implement the requested {Ticket} while strictly adhering to the project's **Source of Truth** (`docs/requirements.md`, `docs/audit.md`, `docs/SDS.md`, `AGENTS.md`).

## 0. Rules of Engagement
- **Immutable Source of Truth**: Requirements in `docs/` are non-negotiable.
- **Phase-Based Handoffs**: Each phase MUST produce a structured artifact for the next agent.
- **Test-Driven Baseline**: Every implementation MUST include corresponding integration or E2E tests.
- **Premium Design**: Always use the `frontend-design` skill for UI tasks.
- **Verification Gate**: A ticket is only "Done" when its specific `verification gate` and all `docs/audit.md` checks pass.

---

## 1. Workflow Phases

### Phase 1: Planning & Specification (Research Agent)
**Goal**: Create a technical manifest for implementation.
1. **Analyze**: Read `AGENTS.md`, `docs/ticket-tracker.md`, and the relevant `track-{a,b,c,d}.md` for the ticket definition and verification gate.
2. **Contract Definition**: Define exact JSON payloads, WebSocket events, and DB schema changes in a `PLAN.md` (saved to `.agents/scratch/`).
3. **Task Decomposition**: Break the ticket into atomic, parallelizable tasks for Backend and Frontend.
4. **Handoff**: Provide the `PLAN.md` path.

### Phase 2: Implementation (Backend & Frontend Agents)
**Goal**: Execute the plan in parallel.
- **Backend**:
  - Implement logic in `internal/handlers/` and persistence in `internal/db/`.
  - Write and pass Go integration tests in `internal/tests/`.
- **Frontend**:
  - Build UI in `SPA/features/` using **ES2026+ Vanilla JS**.
  - Use `Proxy`-based reactivity and the `frontend-design` skill.
  - Write and pass Vitest suites in `SPA/tests/`.

### Phase 3: QA & Verification (Audit Agent)
**Goal**: Finalize, verify, and document.
1. **Regression**: Run `make test`.
2. **Audit Audit**: Execute every check in `docs/audit.md` relevant to this ticket.
3. **Gate Verification**: Confirm the ticket's specific **Verification Gate** is 100% satisfied.
4. **Handoff**: Produce a `VERIFICATION_MANIFEST.md` with proof of all passing checks.

### Phase 4: Closure (System Agent)
1. **Docs**: Update `docs/ticket-tracker.md` to `[x]`.
2. **Archive**: Create a PR summary in `docs/pr-message/` using the `{TicketID}-{Description}-pr.md` and the format and the `pr-template.md`.

---

## 2. Technical Standards (ES2026+ Baseline)
- **Temporal API**: Use for all time handling. No `new Date()`.
- **Resource Management**: Use `using` for cleanup (WebSockets, etc.).
- **Immutability**: Use `toSorted`, `toReversed`, `with`.
- **Vanilla SPA**: Single HTML file, client-side routing, `Proxy` state management.

---

## 3. Ticket Context: {Ticket}
> [!IMPORTANT]
> Insert the specific ticket definition and verification gate from the track file here.

**Execution Order**: Spawn Phase 1 → Use PLAN.md to spawn Phase 2 (Parallel) → Use results to spawn Phase 3 → Close with Phase 4.
