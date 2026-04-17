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
- B01 -> RTF-10, B02 -> RTF-11, B03 -> RTF-12, B04 -> RTF-13, B05 -> RTF-14, B06 -> RTF-29, B07 -> RTF-31, B08 -> RTF-30

## Suggested Execution Order
1. B01 | 2. B02, B03 | 3. B04, B05 | 4. B06, B07, B08

## Tickets

### B01 - Feed Route in the SPA
Source: RTF-10 | Phase: P1
Depends on: A03, A04, A05
Blocks: B02, B03, B06, B07

Work:
- move the feed into the SPA outlet
- preserve category filtering and pagination
- keep post card behavior compatible with SPA navigation

Verification Gate:
- the feed works inside the SPA shell
- feed filtering and pagination still function
- opening a post uses SPA navigation instead of full document navigation

### B02 - Remove Feed Comment Rendering
Source: RTF-11 | Phase: P1
Depends on: B01
Blocks: D05

Work:
- stop fetching comment previews for feed cards
- preserve clean post-card rendering in the feed

Verification Gate:
- feed cards no longer render comments
- feed loading no longer requests per-post comment lists

### B03 - Post Detail Route and Comment Flow
Source: RTF-12 | Phase: P1
Depends on: B01
Blocks: D05, B06, B07

Work:
- move post detail into a SPA route
- load comments only on post detail
- preserve comment creation and comment image upload behavior

Verification Gate:
- opening a post renders the post-detail route inside the SPA
- comments are loaded only on post detail
- comment submission still works
- comment image uploads still work

### B04 - Create and Edit Post SPA Flows
Source: RTF-13 | Phase: P1
Depends on: A04, A05
Blocks: D05, B08

Work:
- move create-post into the SPA shell
- move edit-post into the SPA shell
- preserve existing image upload and category behavior

Verification Gate:
- create and edit post screens render inside the SPA shell
- post create and update behaviors still work
- logout remains available while using these views

### B05 - Activity View in the SPA
Source: RTF-14 | Phase: P1
Depends on: A03, A04, A05
Blocks: D05, D07

Work:
- move the activity screen into the SPA
- preserve activity data loading
- replace hard page navigation with SPA route transitions

Verification Gate:
- activity is accessible inside the shared shell
- activity data loads correctly
- activity navigation no longer depends on a standalone template

### B06 - Notification Behavior in the SPA
Source: RTF-29 | Phase: P2
Depends on: A04, B01, B03
Blocks: D05, D07

Work:
- preserve notification polling in the SPA
- preserve unread badge and unread-count behavior
- preserve mark-one-read and mark-all-as-read behavior
- preserve notification click and deep-link navigation

Verification Gate:
- notification polling works from the SPA shell
- unread badge/count behavior is preserved
- mark-read and mark-all-read still work
- notification click behavior still navigates correctly

### B07 - Reaction Behavior in the SPA
Source: RTF-31 | Phase: P2
Depends on: B01, B03
Blocks: D05, D07

Work:
- preserve post reaction bindings after SPA migration
- preserve comment reaction bindings after SPA migration
- remove reaction boot assumptions tied to standalone templates

Verification Gate:
- post reactions work in SPA-rendered views
- comment reactions work in SPA-rendered views
- reaction wiring no longer assumes old page boot logic

### B08 - Draft Workflows in the SPA
Source: RTF-30 | Phase: P2
Depends on: B04
Blocks: D05, D07

Work:
- preserve save-draft behavior in SPA create-post
- preserve draft edit and publish flows
- keep draft navigation inside the SPA

Verification Gate:
- saving a draft still works from the SPA create-post view
- draft editing and publish flows still work
- draft entry points no longer leave the SPA
