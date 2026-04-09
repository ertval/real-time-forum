# Track D - Realtime Frontend, Browser Integration, and Final Verification

## Mission

Track D owns the browser-side realtime chat experience and the final frontend integration layer:

- persistent chat roster UI
- active conversation panel
- incremental history loading
- browser WebSocket integration
- frontend regression coverage
- final acceptance validation

This track is the main integration consumer. It depends on Track A for shell/auth and Track C for chat contracts.

## Source Mapping

- `TD01` -> `RTF-21`
- `TD02` -> `RTF-22`
- `TD03` -> `RTF-23`
- `TD04` -> `RTF-24`
- `TD05` -> `RTF-27`
- `TD06` -> `RTF-28`

## Suggested Execution Order

1. `TD01`
2. `TD02`
3. `TD03`
4. `TD04`
5. `TD05`
6. `TD06`

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
- `TD05`

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
- `TD05`

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

- `TD05`

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

Depends on:

- `TA02`
- `TA08`
- `TC01`
- `TC04`
- `TC06`
- `TD01`
- `TD02`

Blocks:

- `TD05`
- `TD06`

Verification Gate:

- presence changes update the roster without refresh
- outbound messages are emitted through the agreed WebSocket contract
- incoming messages appear live in the active conversation
- message activity reorders the roster correctly
- send and validation errors are visible in the UI

### TD05 - Frontend Regression Coverage

Source:

- `RTF-27`

Phase:

- `P4`

Work:

- cover auth routes, auth boot, and logout visibility
- cover feed vs post-detail comment visibility
- cover create/edit post, activity, drafts, notifications, and reactions in the SPA
- cover image-upload flows, offline composer behavior, offline-history readability, live message rendering, and send-error handling

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
- `TD01`
- `TD02`
- `TD03`
- `TD04`

Blocks:

- `TD06`

Verification Gate:

- frontend regression coverage exists for the main SPA transitions
- logout visibility is explicitly protected
- image-upload flows are protected
- offline composer behavior, offline history readability, and live message rendering are tested
- retained legacy SPA flows covered by this phase are protected

### TD06 - Final Acceptance Validation

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

Blocks:

- None

Verification Gate:

- the build satisfies the documented product success criteria
- retained features are confirmed working or explicitly flagged
- remaining gaps are listed as follow-up items
