# Track D - Realtime Frontend, Browser Integration, and Final Verification

## Mission

Track D owns the browser-side realtime chat experience and the final frontend integration layer:

- persistent chat roster UI
- active conversation panel
- incremental history loading
- browser WebSocket integration
- SPA and forum regression coverage
- chat-specific frontend regression coverage
- final acceptance validation
- DM image rendering frontend (bonus)

This track is the main integration consumer. It depends on Track A for shell/auth and Track C for chat contracts.

## Source Mapping

- `TD01` -> `RTF-21`
- `TD02` -> `RTF-22`
- `TD03` -> `RTF-23`
- `TD04` -> `RTF-24`
- `TD05` -> `RTF-27`
- `TD06` -> `RTF-32`
- `TD07` -> `RTF-28`
- `TD08` -> `RTF-35` (bonus)

## Suggested Execution Order

1. `TD01`
2. `TD02`
3. `TD03`
4. `TD04`
5. `TD05`
6. `TD06`
7. `TD08` (bonus)
8. `TD07`

`TD05` is intentionally split so SPA/forum regression work can start once Track B stabilizes; it does not need to wait for `TD04`.

## Tickets

### TD01 - Persistent Chat Roster UI

Source:

- `RTF-21`

Phase:

- `P3`

Work:

- render the always-visible user roster in the authenticated shell
- show online/offline state and last-message preview metadata
- keep all rostered users selectable, including offline users

Depends on:

- `TA04`
- `TC05`

Blocks:

- `TD02`
- `TD04`
- `TD06`

Verification Gate:

- the roster is visible on authenticated routes
- presence state is shown per user
- roster order matches backend ordering
- all rostered users remain selectable

### TD02 - Active Conversation Panel and Composer

Source:

- `RTF-22`

Phase:

- `P3`

Work:

- render selected conversation history
- display sender username and message timestamp
- allow selecting any rostered user
- render empty-state conversations when there is no history
- disable sending when the selected user is offline

Depends on:

- `TC03`
- `TD01`

Blocks:

- `TD03`
- `TD04`
- `TD06`

Verification Gate:

- selecting a user loads the latest 10 messages
- offline history remains readable
- selectable users with no history show a valid empty state
- sending is disabled in the UI when the selected user is offline

### TD03 - Incremental History Loading

Source:

- `RTF-23`

Phase:

- `P3`

Work:

- load older history in batches of 10
- use throttle or debounce for scroll-triggered loading
- preserve stable scroll position when prepending history

Depends on:

- `TD02`

Blocks:

- `TD06`

Verification Gate:

- older messages load in batches of 10
- repeated scroll events do not cause burst requests
- the viewport remains usable after prepending history

### TD04 - Browser WebSocket Chat Integration

Source:

- `RTF-24`

Phase:

- `P3`

Work:

- open the WebSocket after authenticated app boot
- emit `dm.send` from the composer
- consume presence and message events
- update active conversation and roster ordering live
- surface send and delivery errors in the UI
- explicitly defer automatic reconnect, backoff, and disconnected-state recovery for this phase

Depends on:

- `TA02`
- `TA08`
- `TC01`
- `TC04`
- `TC06`
- `TD01`
- `TD02`

Blocks:

- `TD06`
- `TD07`

Verification Gate:

- presence changes update the roster without refresh
- outbound messages are emitted through the agreed WebSocket contract
- incoming messages appear live in the active conversation
- message activity reorders the roster correctly
- send and validation errors are visible in the UI
- automatic reconnect and connection-loss recovery are out of scope for this ticket

### TD05 - SPA and Forum Frontend Regression Coverage

Source:

- `RTF-27`

Phase:

- `P4`

Work:

- cover auth routes, auth boot, and logout visibility
- cover feed vs post-detail comment visibility
- cover create/edit post, activity, drafts, notifications, and reactions in the SPA
- cover post and comment image-upload flows

Depends on:

- `TA07`
- `TA09`
- `TB02`
- `TB03`
- `TB04`
- `TB05`
- `TB06`
- `TB07`
- `TB08`

Blocks:

- `TD07`

Verification Gate:

- frontend regression coverage exists for the main SPA transitions
- logout visibility is explicitly protected
- image-upload flows are protected
- retained legacy SPA flows covered by this phase are protected

### TD06 - Chat Frontend Regression Coverage

Source:

- `RTF-32`

Phase:

- `P4`

Work:

- cover offline composer behavior
- cover offline-history readability and empty conversation states
- cover live message rendering and chat send-error handling
- cover roster presence, roster reordering, and incremental history loading behavior

Depends on:

- `TD01`
- `TD02`
- `TD03`
- `TD04`

Blocks:

- `TD07`

Verification Gate:

- offline composer behavior is explicitly protected
- offline history readability and empty-state conversations are tested
- live message rendering and send-error handling are covered
- presence rendering, roster reordering, and history-loading behavior are tested

### TD07 - Final Acceptance Validation

Source:

- `RTF-28`

Phase:

- `P4`

Work:

- execute the final acceptance checklist against the PRD and SDS
- verify retained legacy features still function
- record any remaining gaps as follow-up work instead of hidden work

Depends on:

- `TB05`
- `TB06`
- `TB07`
- `TB08`
- `TC07`
- `TC08`
- `TD04`
- `TD05`
- `TD06`

Blocks:

- None

Verification Gate:

- the build satisfies the documented product success criteria
- retained features are confirmed working or explicitly flagged
- remaining gaps are listed as follow-up items

### TD08 - DM Image Rendering Frontend (Bonus)

Source:

- `RTF-35`

Phase:

- `P5` (Bonus)

Work:

- add image attachment button to the DM composer
- upload selected image via `POST /api/v1/chats/{userID}/images`
- include returned `image_url` in the `dm.send` WebSocket event
- render images inline in the conversation panel (both sent and received)
- display images in chat history loaded from the API
- preserve image rendering during incremental history loading

Depends on:

- `TC09`
- `TD02`
- `TD04`

Blocks:

- `TD07`

Verification Gate:

- a user can select and upload an image in the DM composer
- the image is displayed inline in the conversation for both sender and recipient
- images are visible in chat history on page reload
- image rendering works correctly with incremental history loading
