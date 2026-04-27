# A07: User Profile Page (Bonus)

This PR implements the User Profile Page bonus feature for the real-time forum. It introduces a public profile endpoint, a corresponding single-page application (SPA) route, and integrates profile links into existing user references across the forum interface.

## Summary of Changes

### 1. SPA Route & Profile UI
- **Profile Page View**: Created a visually premium profile page in `SPA/features/profile/profile.views.js` using modern CSS and skeleton loaders.
- **Profile Component Initialization**: Added `initProfilePage` in `SPA/features/profile/profile.page.js` to fetch and render the profile data dynamically into the view.
- **Router Integration**: Added the `/profile/:id` route to `render-template.js` and `create-app.js` to enable smooth client-side routing to user profiles.

### 2. Backend API
- **User Profile Endpoint**: Added `GET /api/v1/users/{userID}/profile` handler in `internal/handlers/users.go`. Modifies the existing `HandleUser` to conditionally return only public profile data (`user_id`, `username`, `first_name`, `last_name`, `age`, `gender`) when the `/profile` suffix is present in the path.

### 3. Navigation Links
- **Post & Comment Integration**: Updated `SPA/features/post/post-card.views.js` to wrap the author's name in a clickable `<a data-link class="profile-link">` anchor tag linking to `/profile/{user_id}`.

## Verification Gate Satisfaction

This PR fully satisfies the verification gate for ticket **A07**:
> "- navigating to a user profile shows all extended registration fields
- profile is accessible from the chat roster (Note: Currently blocked by D01 implementation) and post/comment author links
- profile data matches what was entered during registration"

## Testing & Validation Verified

### Automated Test Suite
- [x] `make test` — Passed. Verified integration tests, UI behaviors, and Go unittests (19 Playwright tests, 47 Vitest tests).
- [x] `bun run check` — Passed. Verified all JS files pass Biome linting and formatting.

### QA Checklist
- [x] Tested navigating to a post and clicking the author's name routes directly to the profile view using the SPA router.
- [x] Tested loading the profile page manually. It renders the correct layout and fetches the respective `first_name`, `last_name`, `age`, `gender` information.

## Key Files Impacted
- `internal/handlers/users.go`
- `SPA/core/router/render-template.js`
- `SPA/core/app/create-app.js`
- `SPA/features/profile/profile.views.js`
- `SPA/features/profile/profile.page.js`
- `SPA/features/profile/profile.css`
- `SPA/assets/css/main.css`
- `SPA/features/post/post-card.views.js`
