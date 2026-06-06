# B08: Draft Workflows in the SPA
<!-- Filename: docs/pr-message/B08-Draft-Workflows-in-the-SPA-pr.md -->

This PR closes ticket `B08` by restoring the draft workflow inside the SPA. The
create-post "Save Draft" button now creates a persisted draft through the
existing `POST /api/v1/posts` contract with `status="draft"`, while normal
publishing continues to create published posts. Draft validation, navigation,
and tests were updated so draft entry points stay inside the SPA with no hard
reloads or backend/API changes.

## Summary of Changes

### 1. Create-Post Draft Submission
- **Submitter-aware form handling**: The create/edit post submit handler now
  reads `event.submitter` so the create-post `Save Draft` button
  (`name="action" value="draft"`) is distinguished from the normal publish
  button.
- **Draft status on create**: `handleCreateSubmit` passes `status: "draft"` for
  draft saves and `status: "published"` for normal publish submits.
- **Keyboard submit remains publish**: Form submits without a submitter still
  follow the existing publish behavior, which keeps the default create-post path
  predictable.

### 2. API Payload and Validation
- **Status forwarding**: `createPost` now accepts an optional `status` and sends
  only whitelisted values (`draft` or `published`) on both JSON and multipart
  requests.
- **Multipart parity**: `buildPostMultipartFormData` appends the `status` field
  when provided, so image-backed draft creation uses the same contract as JSON
  draft creation.
- **Draft-aware validation**: Drafts require a title and body-or-image, but no
  category. Publishing still requires at least one category.

### 3. SPA Navigation and Existing Draft Flows
- **In-SPA draft navigation**: A successful draft save shows `Draft saved.` and
  navigates to `/activity` through the SPA router.
- **No hard reloads**: The draft path remains intercepted by the JS submit
  handler and does not use `window.location` or standalone page boot logic.
- **Existing publish/edit flows preserved**: Draft publishing remains the
  Activity owner status toggle (`PATCH /api/v1/posts/{id}` with
  `status="published"`), and the edit-post flow remains unchanged.

### 4. Backend / Schema / API
- **No backend changes**: The existing create-post endpoint already supports
  draft status.
- **No schema changes**: Draft state continues to use the existing post status
  model.
- **No new endpoints**: The SPA intentionally reuses `POST /api/v1/posts` with
  `status="draft"` and does not depend on the legacy `/api/v1/posts/draft`
  endpoints.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **B08**:
> - saving a draft still works from the SPA create-post view
> - draft editing and publish flows still work
> - draft entry points no longer leave the SPA

## Testing & Validation Verified

### Automated Test Suite
- [x] `make build` — backend and frontend servers compile
- [x] `make test` — Go, Vitest, and Playwright umbrella passing
- [x] `bun x biome check .` — static analysis passing
- [x] `bun run policy` — frontend policy gate passing
- [x] `go test ./internal/...` — backend packages passing
- [x] `bun run test` — Vitest passing, including new B08 draft coverage

### QA Checklist
- [x] Save Draft from `/create-post` sends `status="draft"`
- [x] Publish from `/create-post` sends `status="published"`
- [x] Draft save without categories is allowed
- [x] Publish without categories is still blocked
- [x] Draft save navigates to `/activity` through the SPA router
- [x] Draft publish remains handled by the Activity status toggle
- [x] Edit-post behavior is unchanged
- [x] No backend/API/schema changes were introduced

### Automated Behavioral Verification (E2E)
- [x] Existing Playwright coverage remains passing through `make test`
- [x] Unit coverage in `SPA/tests/unit/features/post/post.page.test.js` verifies
  draft submit, publish submit, category validation differences, and SPA
  navigation behavior
- [ ] Dedicated browser-level draft journey coverage can be added under `D05`
  for additional regression hardening

## Key Files Impacted
- `SPA/features/post/post.page.js`
- `SPA/features/post/post.api.js`
- `SPA/core/shared/utils.js`
- `SPA/tests/unit/features/post/post.page.test.js`
- `docs/track-b.md`
- `docs/ticket-tracker.md`
- `docs/SDS.md`
- `architecture.md`

## Audit Traceability

- **Ticket**: B08 - Draft Workflows in the SPA
- **Source**: RTF-30
- **Phase**: P2
- **Depends on**: B04
- **Blocks**: D05, D07
- **Audit Result**: PASS
- **Backend Changes**: None
- **API Contract Changes**: None
- **Migration Type**: Legacy draft behavior preserved inside SPA

## Follow-up Notes

- Add a backend `httptest` for `POST /api/v1/posts` with `status="draft"` on
  JSON and multipart payloads as part of D05 regression hardening.
- Optional cleanup: remove unused legacy draft endpoints/templates only if a
  later ticket explicitly scopes that deletion.

## Regression Review

Verified:

- No backend regressions
- No API contract regressions
- No routing regressions
- No create-post regressions
- No activity-view regressions
- Existing publish workflow preserved
- Existing edit-post workflow preserved
- Existing attachment upload workflow preserved
