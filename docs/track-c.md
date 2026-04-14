# Track C - Realtime Backend, Persistence, and Backend Validation

## Mission

Track C owns the realtime server spine:

- authenticated WebSocket transport
- direct-message persistence
- roster and history APIs
- presence broadcasting
- realtime message delivery
- backend migration strategy
- backend test coverage
- DM image upload backend (bonus)

This track should stay backend-focused and avoid owning browser UI except where API and event contracts must be defined.

## Source Mapping

- `TC01` -> `RTF-18`
- `TC02` -> `RTF-15`
- `TC03` -> `RTF-17`
- `TC04` -> `RTF-19`
- `TC05` -> `RTF-16`
- `TC06` -> `RTF-20`
- `TC07` -> `RTF-25`
- `TC08` -> `RTF-26`
- `TC09` -> `RTF-34` (bonus)

## Suggested Execution Order

1. `TC01`, `TC02`
2. `TC03`, `TC04`
3. `TC05`, `TC06`
4. `TC07`, `TC08`
5. `TC09` (bonus)

## Tickets

### TC01 - Authenticated WebSocket Endpoint and Connection Manager

Source:

- `RTF-18`

Phase:

- `P0`

Work:

- add `GET /ws`
- validate session cookies during upgrade
- manage connection lifecycle per user
- support multiple connections for the same user

Depends on:

- None

Blocks:

- `TC04`
- `TC05`
- `TC06`
- `TD04`
- `TC08`

Verification Gate:

- authenticated users can open WebSocket connections
- unauthenticated connections are rejected
- multi-tab connection lifecycle is tracked reliably

### TC02 - Private Messages Schema and Repository Layer

Source:

- `RTF-15`

Phase:

- `P1`

Work:

- add the `private_messages` table
- add repository functions for message creation and pair-based history lookup
- add the required indexes

Depends on:

- `TA05`

Blocks:

- `TC03`
- `TC05`
- `TC06`
- `TC07`

Verification Gate:

- messages can be persisted between two users
- pair-based history queries work
- schema and repository tests cover the new table

### TC03 - Chat History API

Source:

- `RTF-17`

Phase:

- `P2`

Work:

- add `GET /api/v1/chats/{userID}/messages`
- support latest-10 fetch
- support `before_id` pagination for older history
- return render-ready chronological order

Depends on:

- `TC02`

Blocks:

- `TD02`
- `TC08`

Verification Gate:

- the first request returns the latest 10 messages
- older history returns in batches of 10
- the API exposes whether more history exists

### TC04 - Presence Broadcasting

Source:

- `RTF-19`

Phase:

- `P2`

Work:

- keep in-memory presence state from active connections
- emit `presence.snapshot` on connect
- emit `presence.update` on actual online-state transitions

Depends on:

- `TC01`

Blocks:

- `TC06`
- `TD04`
- `TC08`

Verification Gate:

- new clients receive a presence snapshot
- first connection marks a user online
- last disconnect marks a user offline

### TC05 - Chat Roster API

Source:

- `RTF-16`

Phase:

- `P2`

Work:

- add `GET /api/v1/chats`
- include all other users
- include presence and last-message metadata
- enforce required roster ordering

Depends on:

- `TC02`
- `TC01`

Blocks:

- `TD01`
- `TC08`

Verification Gate:

- roster entries include `is_online` and last-message metadata
- users with history sort by latest activity
- users without history sort alphabetically after active conversations

### TC06 - Realtime DM Send and Delivery

Source:

- `RTF-20`

Phase:

- `P2`

Work:

- validate outbound `dm.send`
- reject invalid, self-directed, or offline-recipient sends
- persist valid messages
- emit `dm.message` and `chat.error`

Depends on:

- `TC02`
- `TC01`
- `TC04`

Blocks:

- `TD04`
- `TC08`

Verification Gate:

- valid direct messages are persisted and delivered to sender and recipient
- invalid sends are rejected with the agreed contract
- delivered payloads match the SDS event shape

### TC07 - Database Migration Strategy

Source:

- `RTF-25`

Phase:

- `P3`

Work:

- define migration logic for new user fields and private messages
- preserve existing retained tables and data
- define handling for pre-existing users after schema change

Depends on:

- `TA05`
- `TC02`

Blocks:

- `TD07`

Verification Gate:

- schema changes can be applied to an existing database
- existing users remain usable
- retained forum data is preserved

### TC08 - Backend Test Coverage for Auth, Messaging, and Presence

Source:

- `RTF-26`

Phase:

- `P4`

Work:

- add backend tests for extended registration
- add login-by-username and login-by-email tests
- add auth-gating, roster, history, websocket auth, presence, and DM integration tests

Depends on:

- `TA06`
- `TA08`
- `TC01`
- `TC03`
- `TC04`
- `TC05`
- `TC06`

Blocks:

- `TD07`

Verification Gate:

- backend tests cover the major auth and chat flows
- both supported login modes are explicitly tested
- offline-recipient and multi-connection presence cases are covered

### TC09 - DM Image Upload Backend (Bonus)

Source:

- `RTF-34`

Phase:

- `P5` (Bonus)

Work:

- add `image_path TEXT DEFAULT NULL` column to `private_messages`
- add `POST /api/v1/chats/{userID}/images` endpoint for DM image uploads
- reuse existing image upload validation (max size, allowed types)
- save uploaded images to `web/static/uploads/dm/`
- extend `dm.send` WebSocket event to accept optional `image_url` field
- extend `dm.message` event to include `image_url` when present
- extend history API responses to include `image_url`

Depends on:

- `TC02`
- `TC06`

Blocks:

- `TD08`
- `TD07`

Verification Gate:

- DM image upload saves the file to disk and returns a valid URL
- messages with images are persisted with the `image_path` column populated
- `dm.message` events include `image_url` when an image is attached
- history API returns `image_url` for messages that have images
