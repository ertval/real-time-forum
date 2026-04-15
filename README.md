# Real-Time Forum

A full-stack, real-time single-page forum built with Go and vanilla JavaScript. Users register, log in, create posts, comment, and exchange live private messages through WebSockets — all from a single HTML page.

This project satisfies the [01-edu real-time-forum](docs/requirements.md) exercise requirements.

---

## Features

### Authentication
- Registration with nickname, email, password, age, gender, first name, and last name
- Login with nickname **or** email + password
- Session-based authentication with `HttpOnly` cookies
- Logout available from every page
- All forum content requires authentication — no guest access

### Posts & Comments
- Create, edit, and view posts with category tagging
- Image upload support in posts and comments
- Posts displayed in a paginated feed
- Comments visible **only** on the post detail view (not in the feed)
- Publish and draft state management

### Private Messaging (Real-Time)
- Always-visible chat sidebar with user roster
- Online/offline presence indicators (real-time via WebSocket)
- Roster ordered by last message activity; new users listed alphabetically
- Send private messages to online users
- Read chat history with offline users
- Messages display sender username and timestamp
- Last 10 messages loaded on conversation open
- Scroll up to load 10 more messages (throttled to prevent API spam)
- Messages delivered in real-time without page refresh

### Retained Features
- Post and comment reactions (like/dislike)
- Notification system with polling, badges, and mark-as-read
- My Activity dashboard
- Draft workflows

### Bonus Features
- User profile pages with extended registration data
- Image attachments in private messages
- Concurrency patterns (goroutines/channels, Promises) for performance

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.24+ (standard library) |
| Database | SQLite via `mattn/go-sqlite3` |
| WebSocket | `gorilla/websocket` |
| Auth | `bcrypt` (password hashing), `google/uuid` (sessions) |
| Frontend | Vanilla JavaScript, HTML, CSS |
| Containerization | Docker / Docker Compose (optional) |

### Allowed Packages

Only the following Go packages are permitted:

- All [standard Go packages](https://golang.org/pkg/)
- [gorilla/websocket](https://pkg.go.dev/github.com/gorilla/websocket)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [google/uuid](https://github.com/google/uuid) or [gofrs/uuid](https://github.com/gofrs/uuid)

No frontend frameworks (React, Angular, Vue, etc.) are used.

---

## Architecture

```
cmd/
  backend/           → Backend API server (port 8080)
  frontend/          → Frontend server (port 3000) — serves SPA + proxies API/WS
internal/
  db/                → Persistence layer (SQLite, repository functions, schema)
  handlers/          → HTTP handlers (REST API under /api/v1)
  middleware/        → Request middleware (auth, logging, CORS, recovery)
  router/            → Route registration
  tests/             → Backend integration tests
web/
  static/            → CSS, JS, images, sounds, uploads
  templates/         → SPA HTML shell
data/                → SQLite database file
docs/                → Project documentation (PRD, SDS, tickets)
```

The project uses a **split-server** topology:
- **Frontend server** (`:3000`): serves the single HTML shell, static assets, and proxies `/api/` and `/ws` to the backend.
- **Backend server** (`:8080`): owns business logic, persistence, REST APIs, and WebSocket endpoint.

---

## Quick Start

### Requirements

- Go 1.24+
- Make
- SQLite (bundled via CGo)
- Docker & Docker Compose (optional)

### Install & Run

```bash
make deps          # Install Go dependencies
make run-all       # Start backend (8080) + frontend (3000)
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Individual Commands

```bash
make build-backend    # Build backend binary
make build-frontend   # Build frontend binary
make build-all        # Build both

make run-backend      # Start backend only
make run-frontend     # Start frontend only

make stop-backend     # Stop backend
make stop-frontend    # Stop frontend
make stop-all         # Stop both

make test             # Run all tests
make fmt              # Format code
make vet              # Run Go vet
```

### Docker (Optional)

```bash
make docker-build     # Build Docker image
make docker-run       # Run backend container
make docker-up        # Docker Compose up
make docker-down      # Docker Compose down
```

---

## API Overview

All REST endpoints are under `/api/v1/`:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/users/register` | Register a new user |
| `POST` | `/api/v1/users/login` | Login (nickname or email) |
| `POST` | `/api/v1/users/logout` | Logout |
| `GET` | `/api/v1/users/me` | Current session user |
| `GET` | `/api/v1/posts` | List posts (feed) |
| `POST` | `/api/v1/posts` | Create a post |
| `GET` | `/api/v1/chats` | Chat roster (all users + presence) |
| `GET` | `/api/v1/chats/{userID}/messages` | Chat history (paginated) |
| `GET` | `/api/v1/users/{userID}/profile` | User profile (bonus) |
| `POST` | `/api/v1/chats/{userID}/images` | DM image upload (bonus) |
| `GET` | `/ws` | WebSocket — presence + private messaging |

See [docs/SDS.md](docs/SDS.md) for full API contracts and WebSocket event schemas.

---

## Documentation

| Document | Description |
|----------|-------------|
| [docs/requirements.md](docs/requirements.md) | Exercise specification (source of truth) |
| [docs/audit.md](docs/audit.md) | Audit checklist questions |
| [docs/PRD.md](docs/PRD.md) | Product requirements document |
| [docs/SDS.md](docs/SDS.md) | Software design specification |
| [docs/ticket-tracker.md](docs/ticket-tracker.md) | Implementation progress tracker |
| [architecture.md](architecture.md) | Codebase architecture overview |
| [AGENTS.md](AGENTS.md) | Coding agent guide |

---

## Troubleshooting

### Port 8080 already in use

```bash
fuser -k 8080/tcp          # Linux
lsof -i :8080              # Find PID, then kill -9 <PID>
```

### Port 3000 already in use

```bash
fuser -k 3000/tcp
```

---

## Contributors

- Chris Baikas (chbaikas)
- Alex Smyroglou (asmyrogl)

---

## License

This project is licensed under the [GNU General Public License v3.0](LICENSE).