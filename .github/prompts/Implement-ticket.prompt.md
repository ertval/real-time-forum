---
name: Implement-ticket
description: Orchestrate the implementation of a specific ticket for the Real-Time Forum.
agent: agent
tools: [execute, read, edit, search, web, agent, todo]
---

# Ticket Orchestration: {Ticket}

You are the **Master Orchestrator**. Your goal is to coordinate independent specialized sofware engineering subagents to implement the requested {Ticket} while strictly adhering to the project's **Source of Truth** (`docs/requirements.md`, `docs/audit.md`, `docs/SDS.md`, `AGENTS.md`).

## 0. Rules of Engagement
- **Independent Context**: Every phase MUST be handled by a fresh `spawn subagent` call to ensure zero context bleeding.
- **Iteration Loop**: You MUST NOT proceed to the next phase until the current agent's output is 100% compliant and tests are green.
- **Source of Truth**: Requirements in `docs/` are absolute. If a subagent contradicts them, you must command a retry.
- **Premium Design**: Force the use of `frontend-design` skill for all UI tasks.
- **Test-Driven Baseline**: Every implementation MUST include corresponding integration or E2E tests.
- **Commit** After each phase, every code change MUST be committed with a clear message referencing the ticket ID.

---

## 1. Workflow Phases

### Phase 1: Planning (Spawn Research Agent)
**Task**: `spawn subagent` to analyze the codebase and define the technical contract.
1. **Output**: A `PLAN.md` in `.agents/scratch/` containing:
   - Specific files to modify.
   - API/WS contracts (JSON shapes).
   - DB Schema migrations.
   - A pass/fail checklist derived from the ticket's **Verification Gate**.
2. **Loop**: If `PLAN.md` is ambiguous or missing requirements, command the Research Agent to refine it.

### Phase 2: Parallel Execution (Spawn Implementation Agents)
**Task**: `spawn subagent` for Backend and `spawn subagent` for Frontend in parallel.
- **Backend Agent**:
  - Implement in `internal/`.
  - **Loop**: Must run `make test` and iterate until all ticket-related tests pass.
- **Frontend Agent**:
  - Implement in `SPA/` using ES2026+ Vanilla JS and `frontend-design` skill.
  - **Loop**: Must run `bun run policy` (Vitest) and iterate until all feature tests in `SPA/tests/` pass.
- **Orchestrator Note**: Do not move to Phase 3 until both Implementation Agents report "Ready for Audit" with green tests.

### Phase 3: Independent Audit (Spawn Audit Agent)
**Task**: `spawn subagent` for a cold-start audit. This agent MUST NOT be the same as the implementation agents.
1. **Scope**: The Audit Agent **MUST NOT** read `PLAN.md` or implementation logs. It must evaluate the work solely against the **Source of Truth** (`docs/audit.md`, `docs/requirements.md`, `docs/SDS.md`, and `docs/PRD.md`).
2. **Quality Audit**: Review code logic, design patterns, and quality of implementation. Ensure it matches the high standards defined in `AGENTS.md`.
3. **Compliance Audit**: Check for 100% compliance with the ticket requirements, correctness, and the verification gate.
4. **Automated Testing**: Run **ONLY** `make test` (backend) and `bun run policy` (frontend/SPA) to validate the QA gate.
5. **Verification**: Confirm the ticket's **Verification Gate** is fully satisfied based on the documentation.
6. **Loop**: If the Audit Agent finds ANY defect, violation of requirements, or failing test:
   - Identify the failure clearly referencing the authoritative documentation.
   - Send the failure back to the relevant Implementation Agent (Phase 2).
   - Restart the Audit once implementation is fixed.
7. **Audit Report**: Return a detailed audit summary (the "Verification Manifest" content) to the Orchestrator, including proof of all passing checks and an evaluation of implementation quality.

### Phase 4: Closure (Spawn Documentation Agent)
**Task**: `spawn subagent` to finalize the ticket.
1. Update `docs/ticket-tracker.md` to `[x]`.
2. Create a PR summary in `docs/pr-message/` using the `{TicketID}-{Description}-pr.md` filename and `pr-template.md` structure.
3. **Integration**: Ensure the Audit Report from Phase 3 is fully integrated into the PR's "Verification Gate Satisfaction" and "Testing & Validation Verified" sections.
4. Commit the PR and link it to the ticket in `docs/ticket-tracker.md`.

---

## 2. Technical Standards (ES2026+ Baseline)
- **Temporal API**: Use for all time handling. No `new Date()`.
- **Resource Management**: Use `using` for cleanup (WebSockets, etc.).
- **Vanilla SPA**: Single HTML file, client-side routing, `Proxy` state management.

---

## 3. Ticket Context: {Ticket}
> [!IMPORTANT]
> {TicketDefinition}

**Execution Directive**: Spawn Phase 1 → Review PLAN.md → Spawn Phase 2 (Parallel) → Review Implementation Outputs → Spawn Phase 3 (Audit) → Review Audit Report → Spawn Phase 4 (Closure).
