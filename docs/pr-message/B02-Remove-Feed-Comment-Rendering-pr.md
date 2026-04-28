# B02: Remove Feed Comment Rendering

This PR closes ticket B02 by removing comment preview fetching and rendering from the SPA feed. Feed cards are now strictly post-only, and comment loading responsibility is reserved for the post-detail flow.

## Summary of Changes

### 1. Feed Data Flow Cleanup
- Removed feed-side comment preview fetching from `SPA/features/feed/feed.page.js`.
- Feed cards are now rendered directly from post payloads without `previewComments` or any comment-related mapping.
- Feed loading no longer triggers per-post `GET /posts/:id/comments` requests.

### 2. Post Card View Cleanup
- Removed comment preview helpers and preview markup from `SPA/features/post/post-card.views.js`.
- Simplified `renderPostCard(...)` so it only accepts the remaining active options used by feed rendering.
- Preserved the existing post card layout, reaction controls, and SPA click-through behavior.

### 3. Preview API Removal
- Removed the obsolete preview helper from `SPA/features/post/post.api.js`.
- Eliminated the last dedicated frontend helper whose only purpose was feed comment previews.

### 4. Route Consistency Alignment
- Updated feed navigation to use the canonical SPA post-detail path `/posts/:id`.
- Updated route normalization so legacy `/view-post/:id` and prior `/post/:id` links still resolve safely to `/posts/:id`.
- Updated route and app tests to reflect the plural canonical path without breaking old deep links.

### 5. Regression Coverage Updates
- Updated integration coverage to assert that feed/post-card rendering does not emit comment preview markup.
- Removed stale tests that referenced the deleted preview helper and preview props.
- Kept SPA routing coverage green after the path normalization change.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **B02**:
> - feed cards no longer render comments
> - feed loading no longer requests per-post comment lists

Evidence:
- **No feed comment rendering**: post cards no longer include preview comment DOM or comment-preview helpers.
- **No feed comment fetching**: feed rendering no longer imports or calls preview comment APIs.
- **Clean separation for B03**: comment loading is now isolated to future post-detail work rather than partially booting from the feed.

## Testing & Validation Verified

### Automated Test Suite
- [x] `npx vitest run SPA/tests/unit/core/router/routes.test.js SPA/tests/unit/core/app/create-app.test.js SPA/tests/integration/frontend_behavior.test.mjs`

### QA Checklist
- [x] Verified there is no `loadPostCommentsPreview` function left in the SPA source.
- [x] Verified there are no `previewComments` or `showCommentPreview` references left in `SPA/`.
- [x] Verified feed navigation still pushes a valid SPA route for post detail.
- [x] Verified post cards render without comment-preview markup.

### Manual Verification To Run In Browser
- [ ] Open DevTools Network.
- [ ] Load the feed route.
- [ ] Confirm there are zero requests matching `/posts/:id/comments`.

## Key Files Impacted
- `SPA/features/feed/feed.page.js`
- `SPA/features/post/post-card.views.js`
- `SPA/features/post/post.api.js`
- `SPA/core/router/routes.js`
- `SPA/tests/integration/frontend_behavior.test.mjs`
- `SPA/tests/unit/core/router/routes.test.js`
- `SPA/tests/unit/core/app/create-app.test.js`
- `docs/track-b.md`
- `docs/ticket-tracker.md`
- `docs/SDS.md`
