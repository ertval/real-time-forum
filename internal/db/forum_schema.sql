
--TODO updated_at field on comments,users,posts to be set by in app logic

-- ===============================================================
-- Forum database schema (SQLite)
-- ===============================================================
-- users: registered forum members
-- posts: threads created by users
-- comments: replies to posts or nested comments
-- categories: thematic grouping of posts
-- post_categories: M:N relation between posts and categories
-- reactions: likes/dislikes by users on posts/comments
-- sessions: active login sessions with expiration
--
-- Notes:
-- - All timestamps use datetime('now') (UTC)
-- - updated_at fields are managed by the application (not triggers)
-- - Foreign keys enforce cascades for data consistency
-- ===============================================================

PRAGMA foreign_keys = ON;

-- users: registered members with unique username/email and bcrypt-hashed password
CREATE TABLE IF NOT EXISTS users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  username      TEXT    NOT NULL UNIQUE CHECK (length(username) BETWEEN 3 AND 30),
  email         TEXT    NOT NULL UNIQUE CHECK (instr(email, '@') > 1),
  password_hash TEXT    NOT NULL, -- bcrypt hashed password
  is_active     INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
  created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- categories: thematic groups used to organize and filter posts
CREATE TABLE IF NOT EXISTS categories (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  name       TEXT NOT NULL UNIQUE,
  slug       TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- posts: user-created threads; deleted with their author (ON DELETE CASCADE)
CREATE TABLE IF NOT EXISTS posts (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  author_id    INTEGER NOT NULL,
  title        TEXT    NOT NULL CHECK (length(title) > 0),
  body         TEXT    NOT NULL CHECK (length(body) > 0),
  status       TEXT    NOT NULL DEFAULT 'published' CHECK (status IN ('draft','published','archived')),
  category_id  INTEGER,
  created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at   TEXT    NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

-- comments: replies to posts or other comments (nested, cascade on delete)
CREATE TABLE IF NOT EXISTS comments (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id            INTEGER NOT NULL,
  user_id            INTEGER NOT NULL,
  parent_comment_id  INTEGER,
  body               TEXT    NOT NULL CHECK (length(body) > 0),
  created_at         TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at         TEXT    NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (parent_comment_id) REFERENCES comments(id) ON DELETE CASCADE
);


--Not implemented yet
-- -- post_categories: many-to-many relation between posts and categories
-- CREATE TABLE IF NOT EXISTS post_categories (
--   post_id     INTEGER NOT NULL,
--   category_id INTEGER NOT NULL,
--   PRIMARY KEY (post_id, category_id),
--   FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
--   FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
-- );


-- reactions: likes/dislikes by users on posts or comments (one per target)
CREATE TABLE IF NOT EXISTS reactions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL,
    post_id     INTEGER,   -- one of (post_id, comment_id) must be non-NULL
    comment_id  INTEGER,
    value       INTEGER NOT NULL CHECK (value IN (-1, 1)), -- +1 like, -1 dislike
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    CHECK ((post_id IS NOT NULL) != (comment_id IS NOT NULL)),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE
    );

-- sessions: active user logins with expiration and one active per user
CREATE TABLE IF NOT EXISTS sessions (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL,
  token       TEXT    NOT NULL UNIQUE,
  created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
  expires_at  TEXT    NOT NULL CHECK (expires_at > created_at),
  ip          TEXT,
  user_agent  TEXT CHECK (length(user_agent) <= 512),
  is_valid    INTEGER NOT NULL DEFAULT 1 CHECK (is_valid IN (0, 1)),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(author_id);

CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_comment_id);

-- CREATE INDEX IF NOT EXISTS idx_post_categories_category ON post_categories(category_id);

-- one reaction per user per target
CREATE UNIQUE INDEX IF NOT EXISTS ux_react_user_post
    ON reactions(user_id, post_id)
    WHERE post_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_react_user_comment
    ON reactions(user_id, comment_id)
    WHERE comment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_react_post_id ON reactions(post_id);
CREATE INDEX IF NOT EXISTS idx_react_comment_id ON reactions(comment_id);

CREATE UNIQUE INDEX IF NOT EXISTS ux_sessions_single_active
    ON sessions(user_id)
    WHERE is_valid = 1;

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- # list
-- curl -s "http://localhost:8080/api/v1/posts?page=1&per_page=10" | jq
--
-- # if empty, seed one row quickly
-- sqlite3 internal/db/forum.db \
-- "INSERT INTO posts (author_id, title, body, created_at) VALUES (1, 'Hello', 'World', datetime('now'));"
--
-- # list again (should show the post)
-- curl -s "http://localhost:8080/api/v1/posts?page=1&per_page=10" | jq
