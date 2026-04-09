# Real-Time Forum PRD

## 1. Purpose

This document defines the product requirements for evolving the current forum project into a real-time forum that satisfies the 01-edu real-time-forum exercise while preserving useful existing functionality where it does not conflict with the new requirements.

The target product is an authenticated single-page forum with real-time private messaging between users.

## 2. Current Product Summary

The current project already supports:

- user registration and login
- session-based authentication
- posts, comments, categories, reactions, and image uploads
- drafts and a personal activity area
- polling-based notifications

The current project does not yet satisfy the target exercise because:

- it is a multi-page application, not a SPA
- guests can still access parts of the forum
- registration is missing age, gender, first name, and last name
- comments are rendered in the feed instead of only on post detail
- logout is not consistently reachable from every screen
- there is no private messaging, presence tracking, or WebSocket layer

## 3. Product Goal

Transform the forum into a real-time single-page application where registered users can browse posts, open a post to view comments, and exchange live private messages with other online users without refreshing the page.

## 4. Product Scope

### In Scope

- one HTML entry point for the application
- client-side navigation after initial load
- registration and login required for all forum usage
- extended registration form with:
  - username
  - email
  - password
  - age
  - gender
  - first name
  - last name
- logout available from the persistent authenticated shell on every app screen
- comments hidden from the feed and shown only when a user opens a post
- persistent chat area visible on all authenticated views
- online and offline user presence in chat
- direct private messaging between users in real time
- initial load of the latest 10 messages in a chat
- incremental loading of 10 older messages when scrolling upward
- throttled or debounced history loading to avoid scroll-event spam
- message display including sender username and sent timestamp

### Kept from the Existing Project

These remain in scope unless they directly block the real-time forum work:

- posts and categories
- comments and comment reactions
- post reactions
- image uploads
- drafts
- user activity views
- current notification system

### Out of Scope

- guest browsing or guest posting
- OAuth login as a product requirement
- group chat
- typing indicators
- message edits or deletes
- read receipts
- replacing the existing notification system with WebSockets in this phase

## 5. Users

### Primary User

A registered forum member who logs in, reads posts, comments on posts, and uses direct messages to talk to another user who is currently online.

### Secondary User

A returning user who wants to reopen previous private conversations and continue from the latest history.

## 6. Product Requirements

### 6.1 Authentication

- Unauthenticated users may access only the login and registration entry states.
- All forum content requires a valid authenticated session.
- Login supports username or email plus password.
- Registration automatically creates a valid session on success.
- Logout must be reachable from every authenticated route.

### 6.2 SPA

- The app must be served through a single HTML shell.
- Navigation between feed, post detail, post create/edit, activity, and auth states must happen without full document reloads.
- Shared UI such as navigation and chat must persist across route changes.

### 6.3 Registration Data

- Registration must collect age, gender, first name, and last name in addition to the existing fields.
- Age must be stored as a numeric value.
- Gender, first name, and last name must be stored as user profile data.

### 6.4 Post and Comment Visibility

- The main feed shows posts only.
- Comment previews must not appear in feed cards.
- Comments load only in the post-detail view.
- Posting a comment still happens from the post-detail view.

### 6.5 Private Messaging

- The chat section is visible at all times for authenticated users.
- The chat roster lists all other users.
- The roster shows whether each user is online or offline.
- Users with existing message history are ordered by most recent message activity.
- Users with no history are ordered alphabetically after active conversations.
- A user may only send a new message to a user who is currently online.
- A user may still open and read prior conversation history with offline users.
- Opening a conversation loads the latest 10 messages first.
- Scrolling upward loads 10 older messages at a time.
- New incoming messages appear in real time in the active conversation.
- When a new message arrives from another user, the roster ordering updates immediately.

## 7. User Experience Requirements

### 7.1 Persistent Shell

- Authenticated users see a consistent app shell with:
  - navigation
  - logout control
  - main content panel
  - persistent chat panel

### 7.2 Chat Behavior

- Selecting a user opens the conversation in the main chat panel.
- The composer must clearly indicate when sending is unavailable because the selected user is offline.
- Message history must preserve chronological reading order.
- History loading must feel smooth and must not trigger repeated duplicate loads while scrolling.

### 7.3 Real-Time Feedback

- Online/offline changes must update without refresh.
- New private messages must appear without refresh for both sender and recipient.

## 8. Non-Functional Requirements

- Message delivery should feel immediate under normal local-development conditions.
- The application must remain usable with multiple simultaneous connected users.
- Chat history loading must avoid excessive backend calls caused by raw scroll events.
- Existing forum features kept in scope must continue to function after the SPA migration.

## 9. Success Criteria

The product is considered complete for this phase when:

- the forum is accessible through one HTML shell
- forum content is restricted to authenticated users
- the extended registration fields are present and persisted
- comments no longer appear in the feed
- logout is reachable from every authenticated screen
- users can see online/offline status
- users can open prior direct-message history
- users can send and receive private messages in real time
- older messages load in batches of 10 during upward scroll

## 10. Risks and Dependencies

### Risks

- SPA migration touches navigation, auth state, and screen composition at the same time
- WebSocket presence introduces state that is not handled by the current polling model
- direct-message ordering and pagination can become inconsistent if history queries are not designed carefully

### Dependencies

- session authentication must be reusable for WebSocket connections
- database schema must support new user fields and private messages
- frontend routing and persistent layout must be in place before the chat UX is complete

## 11. Assumptions

- the current `username` field remains the forum nickname
- the current split frontend/backend server topology will remain in place
- local email or username plus password is the required auth method
- existing non-conflicting forum features remain unless later implementation planning explicitly removes them
