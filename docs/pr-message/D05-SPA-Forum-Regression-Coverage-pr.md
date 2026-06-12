# D05: SPA and Forum Frontend Regression Coverage

This PR closes the regression-coverage gaps for the core SPA and forum flows delivered in Waves 2–4. Existing suites already protected auth submission, app boot/auth-gating, logout behavior, post create/edit, drafts, activity, notifications, and reactions. D05 adds the missing view-layer and image-upload coverage so the audit-critical transitions — the feed↔post-detail comment split, logout visibility, and post/comment image uploads — are explicitly protected against regression. No production code changed; this is a pure test ticket.

## Summary of Changes

### 1. Image-upload flow coverage (audit gate: "image-upload flows are protected")
- **`core/shared/utils.test.js`**: covers `buildImageRequestOptions` (JSON branch vs. multipart branch — verifies the manual `Content-Type` is dropped so the browser sets the multipart boundary) and `buildPostMultipartFormData` (title/body, repeated `category_ids`, image file, draft `status`, trimmed `image_url`, `remove_image` flag), plus the `MAX_IMAGE_BYTES` / accept-list constants.
- **`core/shared/image-picker.test.js`**: covers `setupImagePicker` end to end with DOM stubs — native dialog trigger, file selection (name label + object-URL preview + clear button), oversize rejection (`onTooLarge` + input reset), clearing a selected file, persisted-image edit mode (`hasAnyImage`), persisted-image clearing (`onClearPersisted`), and listener teardown via `destroy()`.
- **`features/post/post.api.test.js`**: covers `createPost` (JSON with normalized category ids, multipart with image + draft status, unsupported-status omission, fetch-rejection safety), `updatePost` (`remove_image` JSON path and multipart replacement), `loadEditablePost` (category-id normalization + image-url trimming/null-coercion), and `createPostComment` (JSON body vs. multipart image, rejection safety).

### 2. Feed vs. post-detail comment visibility (B02/B03 regression)
- **`features/feed/feed.views.test.js`**: asserts the feed scaffold (filters, posts output, pagination) renders and explicitly asserts the feed contains **no** comment affordances (`data-comments`, `comment-form`, `name="body"`, `<textarea>`, "Post comment").
- **`features/post/post-detail.views.test.js`**: asserts the comment region/composer and comment image-upload input render **only** on post detail, that the author profile link and timestamps render, and that attacker-controlled title/body/username are HTML-escaped.

### 3. Logout visibility (A06/D05 regression)
- **`features/shell/shell.views.test.js`**: asserts the authenticated shell always renders the `data-action="logout"` control, the persistent nav (Feed / Create Post / Activity), the reserved chat roster/active regions, and the injected page content.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **D05**:
> - frontend regression coverage exists for the main SPA transitions
> - logout visibility is explicitly protected
> - image-upload flows are protected
> - retained legacy SPA flows covered by this phase are protected

Mapping: logout visibility → `shell.views.test.js` (+ existing `create-app.test.js`); image-upload flows → `utils.test.js`, `image-picker.test.js`, `post.api.test.js`; main SPA transitions and retained legacy flows → feed/post-detail view tests plus the existing auth, create-app, post.page, activity, notification, and reactions suites.

## Testing & Validation Verified

### Automated Test Suite
- [x] `bun run policy` (Biome check + full Vitest) — **PASS**: 27 files, **237 tests** passing (199 baseline + 38 new). Biome reported no diagnostics.
- [x] `bun x biome check --write` on the 6 new files — **PASS**: formatting normalized, no lint errors.
- [x] `make verify-infra` — **PASS**: bootstrap, dual-server build, run→stop process management, and API proxy all green.

### QA Checklist
- [x] New suites run in isolation (`bun x vitest run SPA/tests/unit/core/shared SPA/tests/.../post.api.test.js ...`) — 38/38 pass.
- [x] Full suite re-run confirms zero regressions in pre-existing tests (199 prior tests still pass).
- [x] No production source files modified — diff is test-only plus tracker/PR docs.

### Automated Behavioral Verification (E2E)
- [x] Feed-vs-detail comment split (B02/B03) protected at the view layer via `feed.views.test.js` / `post-detail.views.test.js`.
- [x] Logout visibility protected via `shell.views.test.js` and existing `create-app.test.js` route-transition tests.
- [x] Infrastructure check: process management and API proxy verified via `make verify-infra`.

## Key Files Impacted
- `SPA/tests/unit/core/shared/utils.test.js` (new)
- `SPA/tests/unit/core/shared/image-picker.test.js` (new)
- `SPA/tests/unit/features/post/post.api.test.js` (new)
- `SPA/tests/unit/features/post/post-detail.views.test.js` (new)
- `SPA/tests/unit/features/feed/feed.views.test.js` (new)
- `SPA/tests/unit/features/shell/shell.views.test.js` (new)
- `docs/ticket-tracker.md` (D05 marked Done)
