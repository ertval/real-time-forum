# B07: Reaction Behavior in the SPA
<!-- Filename: docs/pr-message/B07-Reaction-Behavior-in-the-SPA-pr.md -->

This PR closes ticket `B07` by completing the reaction system migration into the
SPA runtime. Post and comment reactions now work in every SPA-rendered view —
the feed, the post-detail route, and the activity view — driven by a single
delegated reaction engine with no `DOMContentLoaded` or standalone-template boot
assumptions. Reaction rendering is unified across both backend payload shapes,
the audited scope-selector mismatch is resolved, and the backend reaction API
contracts are reused unchanged.

## Summary of Changes

### 1. Unified Reaction Rendering (both API shapes)
- **`normalizeReactionState` (single source of truth)**: Added to
  `SPA/features/post/post.reactions.logic.js`. It reconciles the two backend
  payload shapes into a canonical `{ likeCount, dislikeCount, userReaction }`:
  - list/detail responses: `likes` / `dislikes` / `my_reaction` (int `1`/`-1`/`0`)
  - reaction POST responses: `likes_count` / `dislikes_count` / `reaction`
  `userReaction` is normalized to the string `'like'` / `'dislike'` / `''` so
  every view drives `checked` state and counts uniformly.
- **Shared pill markup**: `post-card.views.js` now exposes
  `renderPostReactions` and `renderCommentReactions` built on one private
  `renderReactions(item, scope)` helper — no duplicated reaction templates.
- **Consumers unified**: the feed post card, post-detail post + comments, and
  the activity view (`getReactionMetrics` now delegates to
  `normalizeReactionState`) all render from the same normalization, fixing the
  prior feed mismatch where `likes`/`my_reaction` were ignored in favor of
  `likes_count`/`current_user_reaction`.

### 2. Post Detail Reactions
- **Post reactions**: `renderPostDetailView` (`post-detail.views.js`) now renders
  a `post-actions` section with `renderPostReactions(post)` inside the
  `<section data-post-id>` detail root.
- **Comment reactions**: `renderCommentsMarkup` (`post.page.js`) now renders
  `renderCommentReactions(comment)` inside each
  `<article class="comment" data-comment-id>`.
- **Wiring**: `initPostDetailPage` calls `initReactionBindings({ documentRef, fetchRef })`
  after the detail view mounts, reusing the existing engine — no second engine.
- **Scope-selector mismatch resolved**: the audit flagged that the reaction
  listener resolved post scope via `button.closest('article[data-post-id]')`,
  but the post-detail root is `<section data-post-id>`. The reaction `<input>`
  no longer carries `data-post-id` / `data-comment-id` of its own, so it can
  never collide with the unique `[data-post-id]` container selector
  (`[data-post-id="N"]` now selects exactly one element on the detail page).
  Instead, `resolveReactionTarget` (`post.reactions.bindings.js`) reads the
  `.reactions` wrapper's `data-reaction-scope` hook (`post` / `comment`), then
  resolves the id from the nearest container via `closest('[data-post-id]')`
  or `closest('[data-comment-id]')`. Matching on the attribute rather than a
  fixed `article` tag, the feed/activity `<article data-post-id>` and the
  post-detail `<section data-post-id>` both resolve.

### 3. Backend / API Compatibility
- **Backend Contracts Preserved**: Existing reaction endpoints
  (`POST /api/v1/posts/:id/{like,dislike}`,
  `POST /api/v1/comments/:id/{like,dislike}`) are reused unchanged.
- **No Schema Changes**: Both existing reaction payload shapes — list/detail
  (`likes`/`dislikes`/`my_reaction`) and the reaction POST response
  (`likes_count`/`dislikes_count`/`reaction`) — are supported through frontend
  normalization, requiring no backend or schema modification.

### 4. No Duplicate Listeners, No Standalone Boot
- `initReactionBindings` attaches **one** delegated `change` listener to
  `document`, guarded by a module-level `reactionsInitialized` flag, so the
  feed, post-detail, and activity pages can each call it without producing
  duplicate listeners.
- Because the listener is bound at the document level, it survives SPA route
  transitions (only the shell outlet is swapped) and `reloadComments()`
  re-renders without re-binding.
- Added a defensive `typeof documentRef.addEventListener !== 'function'` guard so
  the binding no-ops on degenerate document refs instead of throwing.
- No per-feature `DOMContentLoaded` and no legacy `web/static/js/reactions.js` /
  `view-post.js` boot path is involved.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **B07**:
> - post reactions work in SPA-rendered views
> - comment reactions work in SPA-rendered views
> - reaction wiring no longer assumes old page boot logic

## Testing & Validation Verified

### Automated Test Suite
- [x] `npx vitest run` — 164/164 tests passing
- [x] New `SPA/tests/unit/features/post/post.reactions.bindings.test.js`
  (10 tests) covering:
  - post like / dislike in both `<article data-post-id>` and
    `<section data-post-id>` scopes
  - switch reaction (like → dislike clears the opposite)
  - toggle-off revert on failed request and on unauthenticated user
  - comment like / dislike to the comment endpoints
  - switch comment reaction
  - re-rendered comments still react via the delegated listener (no duplicate
    listener)
- [x] `SPA/tests/unit/policy/spa-import-paths.test.js` — passing (confirms no
  module re-introduces legacy `/web/static/*` imports or
  `DOMContentLoaded`-style standalone boot logic)
- [x] `npx biome check SPA/features SPA/tests/unit/features/post` — clean
- [x] `go test ./internal/tests/ ./internal/db/` — passing (confirms backend
  reaction contracts unchanged)

### QA Checklist
- [x] Post reactions work in the feed
- [x] Post reactions work on the post-detail route
- [x] Comment reactions work on the post-detail route
- [x] Reactions work in the activity view (posts and comments)
- [x] Like, dislike, toggle-same, and switch-reaction behaviors verified
- [x] Counts update in place from the POST response with no full page reload
- [x] Active reaction state (checked input) updates correctly and clears the
  opposite
- [x] Rendering handles both list/detail (`likes`/`my_reaction`) and POST
  (`likes_count`/`reaction`) payload shapes
- [x] Reaction bindings survive SPA navigation and comment re-renders
- [x] Exactly one delegated listener is attached (module-level guard)
- [x] Scope-selector mismatch (`section` vs `article`) resolved
- [x] Optimistic state reverts on request failure or when unauthenticated

### Automated Behavioral Verification (E2E)
- [ ] Real-browser Playwright coverage for the full reaction chain (toggle →
  POST → count update → active-state) is not yet added; recommended under D05
  (SPA and Forum Frontend Regression Coverage). Current coverage is
  unit/integration with faithful DOM mocks.

## Audit Traceability

- Ticket: B07 - Reaction Behavior in the SPA
- Source: RTF-31
- Phase: P2
- Verification Gate: PASS
- Backend Changes: None
- API Contract Changes: None
- Migration Type: Legacy → SPA

## Key Files Impacted
- `SPA/features/post/post.reactions.logic.js` *(add `normalizeReactionState`)*
- `SPA/features/post/post.reactions.bindings.js` *(scope-selector fix + addEventListener guard)*
- `SPA/features/post/post-card.views.js` *(shared `renderPostReactions` / `renderCommentReactions`)*
- `SPA/features/post/post-detail.views.js` *(render post reactions)*
- `SPA/features/post/post.page.js` *(render comment reactions + wire `initReactionBindings`)*
- `SPA/features/activity/activity.views.js` *(delegate to shared normalization, string user-reaction)*
- `SPA/tests/unit/features/post/post.reactions.bindings.test.js` *(new — 10 tests)*
- `architecture.md`
- `docs/SDS.md`
- `docs/ticket-tracker.md`
- `docs/track-b.md`

## Follow-up Notes
- **Sound (deferred):** the legacy `web/static/js/reactions.js` played a
  `playReaction` sound on toggle; this was not migrated, consistent with B06
  deferring toast/sound. It can be added as a small follow-up if product wants
  full parity.
- **Auth prompt UX:** the legacy path used `Auth.requireOrPrompt()` (auth modal
  on unauthenticated reaction); the SPA engine silently reverts via a
  `GET /users/me` check. If a prompt is desired, it can be layered on later.
- **E2E coverage:** add a Playwright scenario for the reaction chain under D05.

## Regression Review

Verified:

- No backend regressions
- No API contract regressions
- No authentication regressions
- No routing regressions
- No full page reloads introduced
- No duplicate reaction listeners introduced
- Existing reaction behavior preserved across feed, activity, and post detail
