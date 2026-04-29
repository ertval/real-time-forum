# B03: Post Detail Route and Comment Flow
<!-- Filename: docs/pr-message/B03-Post-Detail-Route-and-Comment-Flow-pr.md -->

This PR implements and closes ticket B03 by introducing a dedicated post-detail route and isolating comment behavior to that route, completing the separation initiated in B02 and aligning the SPA with the forum interaction model. The forum now uses the canonical `/posts/:id` route to render post detail inside the authenticated shell, while comments are fetched, submitted, and image-uploaded only from that route.

## Summary of Changes

### 1. SPA Post-Detail Route
- Added the protected SPA route `/posts/:id` for post detail.
- Wired the route into the SPA lifecycle so route entry initializes post-detail data loading without a full page reload.
- Preserved legacy path normalization while keeping `/posts/:id` as the canonical in-app route.

### 2. Post Detail Rendering and Data Loading
- Implemented a dedicated post-detail view that renders full post content, categories, author/date metadata, comment list container, and comment form.
- Added `getPostById(fetchRef, postId)` and `getPostComments(fetchRef, postId)` helpers for independent post and comment loading.
- Added loading and error states for post-detail initialization.

### 3. API Layer
- **Post and comment endpoints used**: `/api/v1/posts/:id` and `/api/v1/posts/:id/comments`
- **Comment submission endpoint**: `POST /api/v1/posts/:id/comments`
- **Multipart upload preserved**: image upload handled via existing API contract

### 4. Comment Flow Preservation
- Implemented comment rendering only inside the post-detail route.
- Implemented comment submission with post-submit refresh so the new comment appears immediately in the UI.
- Preserved comment image upload behavior by reusing the existing image picker and multipart request path.

### 5. Feed and Post-Detail Separation
- Kept the feed limited to post-only data and SPA navigation into post detail.
- Confirmed comments are not fetched or rendered from the feed path.
- Isolated all `/posts/:id/comments` traffic to the post-detail screen.

## Deliverables

- SPA post-detail route implemented.
- Comments moved to post detail.
- Post-detail comment submission implemented.
- Comment image upload preserved in SPA post detail.
- Feed remains post-only and decoupled from comment fetching.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **B03**:
> - opening a post renders the post-detail route inside the SPA
> - comments are loaded only on post detail
> - comment submission still works
> - comment image uploads still work

Evidence:
- **Post detail inside SPA**: opening a feed post navigates to `/posts/:id` through SPA history and renders the post-detail screen in the shared shell.
- **Comments only in post detail**: comment fetches are issued only from `initPostDetailPage()` and not from the feed flow.
- **Comment submission works**: successful comment creation clears the form and reloads the rendered comment list.
- **Image upload works**: comment submission reuses the existing image picker and multipart upload contract with the `image` form field.

## Testing & Validation Verified

### Automated Test Suite
- [x] `make test` — Backend tests passing
- [x] `bun run policy` — Frontend tests and lint passing
- [x] `bun x biome check .` — No lint or format issues

### QA Checklist
- [x] Verified feed navigation uses the SPA post-detail route `/posts/:id`.
- [x] Verified feed rendering contains no comment preview markup.
- [x] Verified post-detail initialization fetches post content and comments separately.
- [x] Verified comment submission refreshes the comment list after success.
- [x] Verified image-only comment submission works through the shared picker flow.

### Manual E2E Verification
- [ ] Open DevTools Network on the feed route and confirm there are no `/comments` requests.
- [ ] Open a post-detail route and confirm `/api/v1/posts/:id/comments` appears there.
- [ ] Submit a text comment and confirm the list updates without full page reload.
- [ ] Submit a comment with an image and confirm upload and rendered attachment behavior.

## Requirements / Design Consistency

- **requirements.md**: aligned with the rule that comments are visible only after opening a post.
- **PRD.md**: aligned with feed showing posts only and post detail handling comments.
- **SDS.md**: aligned with `/posts/:id` as the post-detail route and with post-detail-only comment loading.

## Blockers

- None.

## Key Files Impacted

- `SPA/core/app/create-app.js`
- `SPA/core/router/routes.js`
- `SPA/core/router/render-template.js`
- `SPA/features/feed/feed.page.js`
- `SPA/features/post/post.api.js`
- `SPA/features/post/post.views.js`
- `SPA/features/post/post-detail.views.js`
- `SPA/features/post/post.page.js`
- `SPA/tests/unit/features/post/post.page.test.js`
- `SPA/tests/unit/core/app/create-app.test.js`
- `SPA/tests/unit/core/router/routes.test.js`
- `SPA/tests/integration/frontend_behavior.test.mjs`
- `docs/ticket-tracker.md`
- `docs/track-b.md`
- `docs/PRD.md`
- `docs/SDS.md`
