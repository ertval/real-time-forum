
--TODO updated_at field on comments,users,posts to be set by in app logic

-- forum_schema.sql
-- SQLite schema for a simple forum app
PRAGMA foreign_keys = ON;

-- 1) users
CREATE TABLE IF NOT EXISTS users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  username      TEXT    NOT NULL UNIQUE,
  email         TEXT    NOT NULL UNIQUE,
  password_hash TEXT    NOT NULL,
  is_active     INTEGER NOT NULL DEFAULT 1, --1 means active, 0 means inactive
  created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- 2) posts
CREATE TABLE IF NOT EXISTS posts (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL,
  title       TEXT    NOT NULL,
  body        TEXT    NOT NULL,
  status      TEXT    NOT NULL DEFAULT 'published',
  created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at  TEXT    NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);

-- 3) comments
CREATE TABLE IF NOT EXISTS comments (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id            INTEGER NOT NULL,
  user_id            INTEGER NOT NULL,
  parent_comment_id  INTEGER,
  body               TEXT    NOT NULL,
  created_at         TEXT    NOT NULL DEFAULT (datetime('now')),
  updated_at         TEXT    NOT NULL DEFAULT (datetime('now')),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (parent_comment_id) REFERENCES comments(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_comment_id);


-- 4) categories
CREATE TABLE IF NOT EXISTS categories (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  name       TEXT NOT NULL UNIQUE,
  slug       TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- 5) post_categories
CREATE TABLE IF NOT EXISTS post_categories (
  post_id     INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  PRIMARY KEY (post_id, category_id),
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_categories_category ON post_categories(category_id);

-- 6) reactions (like = +1, dislike = -1) for posts OR comments
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

-- one reaction per user per target
CREATE UNIQUE INDEX IF NOT EXISTS ux_react_user_post
    ON reactions(user_id, post_id)
    WHERE post_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_react_user_comment
    ON reactions(user_id, comment_id)
    WHERE comment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_react_post_id ON reactions(post_id);
CREATE INDEX IF NOT EXISTS idx_react_comment_id ON reactions(comment_id);

-- 7) sessions
CREATE TABLE IF NOT EXISTS sessions (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER NOT NULL,
  token       TEXT    NOT NULL UNIQUE,
  created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
  expires_at  TEXT    NOT NULL,
  ip          TEXT,
  user_agent  TEXT,
  is_valid    INTEGER NOT NULL DEFAULT 1,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_sessions_single_active
    ON sessions(user_id)
    WHERE is_valid = 1;

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
