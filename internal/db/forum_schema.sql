-- ===============================================================
-- Forum Database Schema (SQLite)
-- ===============================================================
-- Tables:
--   users            → registered members
--   categories       → post grouping
--   posts            → user-created threads
--   comments         → nested replies
--   post_categories  → M:N relation posts ↔ categories
--   reactions        → likes/dislikes on posts or comments
--   sessions         → login tracking with expiration
--
-- Notes:
--   - All timestamps use UTC via strftime('%Y-%m-%dT%H:%M:%SZ','now')
--   - updated_at fields are updated by application logic (no triggers)
--   - Foreign keys enforce cascades for consistent cleanup
-- ===============================================================


-- ===============================================================
-- USERS
-- ===============================================================
CREATE TABLE IF NOT EXISTS users (
  id               INTEGER PRIMARY KEY AUTOINCREMENT,
  username         TEXT NOT NULL UNIQUE CHECK (length(username) BETWEEN 3 AND 30),
  email            TEXT NOT NULL UNIQUE CHECK (instr(email, '@') > 1),
  password_hash    TEXT NOT NULL,
  is_active        INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
  session_version  INTEGER NOT NULL DEFAULT 0,
  created_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  updated_at       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);

-- ===============================================================
-- OAUTH USERS (Google, Github, etc)
-- ===============================================================
-- Linked to the main users table through user_id.
-- OAuth users do NOT require password_hash.
-- ===============================================================

CREATE TABLE IF NOT EXISTS oauth_users (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id           INTEGER NOT NULL,
  provider          TEXT NOT NULL,                -- "google", "github"
  provider_user_id  TEXT NOT NULL,                -- Google 'sub', Github id
  created_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  UNIQUE(provider, provider_user_id),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- ===============================================================
-- CATEGORIES
-- ===============================================================
CREATE TABLE IF NOT EXISTS categories (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  name       TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);


-- ===============================================================
-- POSTS
-- ===============================================================
CREATE TABLE IF NOT EXISTS posts (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  author_id   INTEGER NOT NULL,
  title       TEXT NOT NULL CHECK (length(title) > 0),
  image_url   TEXT,
  body        TEXT NOT NULL CHECK (length(body) > 0),
  status      TEXT NOT NULL DEFAULT 'published'
               CHECK (status IN ('draft','published','archived')),
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);


-- ===============================================================
-- COMMENTS (nested)
-- ===============================================================
CREATE TABLE IF NOT EXISTS comments (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id            INTEGER NOT NULL,
  user_id            INTEGER NOT NULL,
  parent_comment_id  INTEGER,
  body               TEXT NOT NULL CHECK (length(body) > 0),
  created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  updated_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (parent_comment_id) REFERENCES comments(id) ON DELETE CASCADE
);


-- ===============================================================
-- POST ↔ CATEGORY RELATION (M:N)
-- ===============================================================
CREATE TABLE IF NOT EXISTS post_categories (
  post_id     INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  PRIMARY KEY (post_id, category_id),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

-- ===============================================================
-- SESSIONS
-- ===============================================================
CREATE TABLE IF NOT EXISTS sessions (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL,
  token       TEXT NOT NULL UNIQUE,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  expires_at TEXT NOT NULL,
  ip          TEXT,
  user_agent  TEXT CHECK (length(user_agent) <= 512),
  is_valid    INTEGER NOT NULL DEFAULT 1 CHECK (is_valid IN (0, 1)),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);


-- ===============================================================
-- REACTIONS (post or comment)
-- ===============================================================
CREATE TABLE IF NOT EXISTS reactions (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL,
  post_id     INTEGER,
  comment_id  INTEGER,
  value       INTEGER NOT NULL CHECK (value IN (-1, 1)),
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  CHECK ((post_id IS NOT NULL) != (comment_id IS NOT NULL)),
  FOREIGN KEY (user_id)    REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (post_id)    REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
);

-- ===============================================================
-- INDEXES
-- ===============================================================

-- Posts
CREATE INDEX IF NOT EXISTS idx_posts_user ON posts(author_id);

-- Comments
CREATE INDEX IF NOT EXISTS idx_comments_post    ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user    ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent  ON comments(parent_comment_id);

-- Post ↔ Categories
CREATE INDEX IF NOT EXISTS idx_post_categories_category
  ON post_categories(category_id);

-- Reactions
CREATE UNIQUE INDEX IF NOT EXISTS ux_react_user_post
  ON reactions(user_id, post_id)
  WHERE post_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_react_user_comment
  ON reactions(user_id, comment_id)
  WHERE comment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_reactions_post
  ON reactions(post_id);

CREATE INDEX IF NOT EXISTS idx_reactions_comment
  ON reactions(comment_id);

-- Sessions
CREATE UNIQUE INDEX IF NOT EXISTS ux_session_single_active
  ON sessions(user_id)
  WHERE is_valid = 1;

CREATE INDEX IF NOT EXISTS idx_sessions_user
  ON sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_sessions_expires
  ON sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_oauth_user_id
ON oauth_users(user_id);
