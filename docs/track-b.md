# Track B - Forum Content and Retained Forum UX

## Mission

Track B owns the non-chat forum migration into the SPA:

- feed
- post detail
- create/edit post flows
- activity view
- notifications
- reactions
- drafts

This track should treat Track A as its main upstream dependency and otherwise move as independently as possible from the realtime chat work.

## Source Mapping

- `TB01` -> `RTF-10`
- `TB02` -> `RTF-11`
- `TB03` -> `RTF-12`
- `TB04` -> `RTF-13`
- `TB05` -> `RTF-14`
- `TB06` -> `RTF-29`
- `TB07` -> `RTF-31`
- `TB08` -> `RTF-30`

## Suggested Execution Order

1. `TB01`
2. `TB02`, `TB03`
3. `TB04`, `TB05`
4. `TB06`, `TB07`, `TB08`

## Tickets

### TB01 - Feed Route in the SPA

Source:

- `RTF-10`

Phase:

- `P1`

Work:

- move the feed into the SPA outlet
- preserve category filtering and pagination
- keep post card behavior compatible with SPA navigation

Depends on:

- `TA03`
- `TA04`
- `TA08`

Blocks:

- `TB02`
- `TB03`
- `TB06`
- `TB07`

Verification Gate:

- the feed works inside the SPA shell
- feed filtering and pagination still function
- opening a post uses SPA navigation instead of full document navigation

### TB02 - Remove Feed Comment Rendering

Source:

- `RTF-11`

Phase:

- `P1`

Work:

- stop fetching comment previews for feed cards
- preserve clean post-card rendering in the feed

Depends on:

- `TB01`

Blocks:

- `TD05`

Verification Gate:

- feed cards no longer render comments
- feed loading no longer requests per-post comment lists

### TB03 - Post Detail Route and Comment Flow

Source:

- `RTF-12`

Phase:

- `P1`

Work:

- move post detail into a SPA route
- load comments only on post detail
- preserve comment creation and comment image upload behavior

Depends on:

- `TB01`

Blocks:

- `TD05`
- `TB06`
- `TB07`

Verification Gate:

- opening a post renders the post-detail route inside the SPA
- comments are loaded only on post detail
- comment submission still works
- comment image uploads still work

### TB04 - Create and Edit Post SPA Flows

Source:

- `RTF-13`

Phase:

- `P1`

Work:

- move create-post into the SPA shell
- move edit-post into the SPA shell
- preserve existing image upload and category behavior

Depends on:

- `TA04`
- `TA08`

Blocks:

- `TD05`
- `TB08`

Verification Gate:

- create and edit post screens render inside the SPA shell
- post create and update behaviors still work
- logout remains available while using these views

### TB05 - Activity View in the SPA

Source:

- `RTF-14`

Phase:

- `P1`

Work:

- move the activity screen into the SPA
- preserve activity data loading
- replace hard page navigation with SPA route transitions

Depends on:

- `TA03`
- `TA04`
- `TA08`

Blocks:

- `TD05`
- `TD07`

Verification Gate:

- activity is accessible inside the shared shell
- activity data loads correctly
- activity navigation no longer depends on a standalone template

### TB06 - Notification Behavior in the SPA

Source:

- `RTF-29`

Phase:

- `P2`

Work:

- preserve notification polling in the SPA
- preserve unread badge and unread-count behavior
- preserve mark-one-read and mark-all-as-read behavior
- preserve notification click and deep-link navigation

Depends on:

- `TA04`
- `TB01`
- `TB03`

Blocks:

- `TD05`
- `TD07`

Verification Gate:

- notification polling works from the SPA shell
- unread badge/count behavior is preserved
- mark-read and mark-all-read still work
- notification click behavior still navigates correctly

### TB07 - Reaction Behavior in the SPA

Source:

- `RTF-31`

Phase:

- `P2`

Work:

- preserve post reaction bindings after SPA migration
- preserve comment reaction bindings after SPA migration
- remove reaction boot assumptions tied to standalone templates

Depends on:

- `TB01`
- `TB03`

Blocks:

- `TD05`
- `TD07`

Verification Gate:

- post reactions work in SPA-rendered views
- comment reactions work in SPA-rendered views
- reaction wiring no longer assumes old page boot logic

### TB08 - Draft Workflows in the SPA

Source:

- `RTF-30`

Phase:

- `P2`

Work:

- preserve save-draft behavior in SPA create-post
- preserve draft edit and publish flows
- keep draft navigation inside the SPA

Depends on:

- `TB04`

Blocks:

- `TD05`
- `TD07`

Verification Gate:

- saving a draft still works from the SPA create-post view
- draft editing and publish flows still work
- draft entry points no longer leave the SPA
