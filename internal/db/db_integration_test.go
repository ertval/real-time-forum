package db_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	forumdb "forum/internal/db"
)

const testDBPath = "test.db"

// setupDB initializes a fresh test database and ensures proper cleanup.
func setupDB(t *testing.T) *sql.DB {
	t.Helper()

	// Clean previous leftovers (db + WAL/SHM)
	for _, f := range []string{testDBPath, testDBPath + "-wal", testDBPath + "-shm"} {
		_ = os.Remove(f)
	}

	forumdb.LogInfo("Initializing temporary test database: %s", testDBPath)
	d, err := forumdb.InitDB(testDBPath)
	if err != nil {
		forumdb.LogWarn("InitDB failed: %v", err)
		t.Fatalf("InitDB failed: %v", err)
	}

	// Safer cleanup — close DB before removing files
	t.Cleanup(func() {
		_ = d.Close()
		for _, f := range []string{testDBPath, testDBPath + "-wal", testDBPath + "-shm"} {
			_ = os.Remove(f)
		}
		forumdb.LogInfo("Cleaned up test database files")
	})

	return d
}

// mustExec runs a SQL statement or fails the test if it errors.
func mustExec(t *testing.T, d *sql.DB, q string, args ...any) sql.Result {
	t.Helper()
	res, err := d.Exec(q, args...)
	if err != nil {
		t.Fatalf("exec failed: %v\nquery: %s", err, q)
	}
	return res
}

// mustQueryRowInt executes a query returning one int result or fails.
func mustQueryRowInt(t *testing.T, d *sql.DB, q string, args ...any) int {
	t.Helper()
	var n int
	if err := d.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatalf("scan int failed: %v\nquery: %s", err, q)
	}
	return n
}

// expectErrorExec ensures a SQL command produces an error (for constraint testing).
func expectErrorExec(t *testing.T, d *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := d.Exec(q, args...); err == nil {
		t.Fatalf("expected error, got nil\nquery: %s", q)
	}
}

// TestIntegration_DB validates the schema, constraints, and cascading behavior.
func TestIntegration_DB(t *testing.T) {
	forumdb.LogInfo("Starting DB integration tests")
	d := setupDB(t)
	defer forumdb.LogInfo("DB integration tests completed successfully")

	t.Run("0_ForeignKeys", func(t *testing.T) {
		mustExec(t, d, `PRAGMA foreign_keys = ON;`)
		var fk int
		if err := d.QueryRow(`PRAGMA foreign_keys;`).Scan(&fk); err != nil {
			t.Fatalf("foreign_keys scan failed: %v", err)
		}
		if fk != 1 {
			t.Fatalf("expected foreign_keys=1, got %d", fk)
		}
		forumdb.LogInfo("Verified PRAGMA foreign_keys enabled")
	})

	t.Run("1_SeedData", func(t *testing.T) {
		mustExec(t, d, `INSERT INTO users (username,email,password_hash) VALUES
			('alex','alex@example.com','h1'),
			('maria','maria@example.com','h2');`)

		mustExec(t, d, `INSERT INTO categories (name,slug) VALUES
			('General Discussion','general'),
			('Technology','technology'),
			('Sports','sports'),
			('Music','music'),
			('News','news');`)

		mustExec(t, d, `INSERT INTO posts (user_id,title,body) VALUES
			(1,'Welcome to the Forum','Hey everyone!'),
			(2,'Best Programming Language?','Your thoughts?'),
			(1,'Latest Football Results','Did you watch it?'),
			(2,'Favorite Bands','What are you into?'),
			(1,'Daily Tech News','Share tech headlines');`)

		mustExec(t, d, `INSERT INTO post_categories (post_id,category_id) VALUES
			(1,1),(2,2),(3,3),(4,4),(5,2),(5,5);`)

		mustExec(t, d, `INSERT INTO comments (post_id,user_id,body) VALUES
			(1,2,'Welcome Alex!'),
			(2,1,'Go is the best!'),
			(2,2,'Python is easier'),
			(3,2,'Crazy match!'),
			(4,1,'Arctic Monkeys!');`)
		mustExec(t, d, `INSERT INTO comments (post_id,user_id,parent_comment_id,body)
			VALUES (1,1,1,'Thanks Maria!');`)

		mustExec(t, d, `INSERT INTO reactions (user_id, post_id, value) VALUES
			(1,2,1),
			(2,2,1),
			(1,3,-1);`)
		mustExec(t, d, `INSERT INTO reactions (user_id, comment_id, value) VALUES
			(2,1,1),
			(1,2,1);`)

		mustExec(t, d, `INSERT INTO sessions (user_id, token, expires_at, ip, user_agent) VALUES
			(1,'abc123', datetime('now','+1 day'), '127.0.0.1','UA'),
			(2,'xyz789', datetime('now','+1 day'), '127.0.0.2','UA');`)

		users := mustQueryRowInt(t, d, `SELECT COUNT(*) FROM users`)
		posts := mustQueryRowInt(t, d, `SELECT COUNT(*) FROM posts`)
		if users != 2 || posts != 5 {
			t.Fatalf("unexpected counts users=%d posts=%d", users, posts)
		}
		forumdb.LogInfo("Seeded initial test data successfully")
	})

	t.Run("2_ReactionChecks", func(t *testing.T) {
		expectErrorExec(t, d, `INSERT INTO reactions (user_id, post_id, comment_id, value) VALUES (1,1,1,1)`)
		expectErrorExec(t, d, `INSERT INTO reactions (user_id, value) VALUES (1, 1)`)
		expectErrorExec(t, d, `INSERT INTO reactions (user_id, post_id, value) VALUES (1, 1, 2)`)
	})

	t.Run("3_ReactionUniquenessAndToggle", func(t *testing.T) {
		expectErrorExec(t, d, `INSERT INTO reactions (user_id, post_id, value) VALUES (2,2,1)`)

		n := mustQueryRowInt(t, d, `SELECT COUNT(*) FROM reactions WHERE user_id=1 AND post_id=1`)
		if n != 0 {
			t.Fatalf("unexpected preexisting reaction for user1/post1")
		}
		res, err := d.Exec(`UPDATE reactions SET value=? WHERE user_id=? AND post_id=?`, -1, 1, 1)
		if err != nil {
			t.Fatalf("update toggle failed: %v", err)
		}
		aff, _ := res.RowsAffected()
		if aff == 0 {
			mustExec(t, d, `INSERT INTO reactions (user_id, post_id, value) VALUES (?,?,?)`, 1, 1, -1)
		}
		val := mustQueryRowInt(t, d, `SELECT value FROM reactions WHERE user_id=1 AND post_id=1`)
		if val != -1 {
			t.Fatalf("expected value -1 after toggle, got %d", val)
		}
	})

	t.Run("4_SessionsSingleActive", func(t *testing.T) {
		expectErrorExec(t, d, `INSERT INTO sessions (user_id, token, expires_at) VALUES (1,'conflict', datetime('now','+1 day'))`)

		mustExec(t, d, `UPDATE sessions SET is_valid=0 WHERE user_id=1`)
		mustExec(t, d, `INSERT INTO sessions (user_id, token, expires_at) VALUES (1,'fresh', datetime('now','+2 days'))`)

		expectErrorExec(t, d, `INSERT INTO sessions (user_id, token, expires_at) VALUES (1,'again', datetime('now','+2 days'))`)
	})

	t.Run("5_CascadeDeleteUser", func(t *testing.T) {
		tx, err := d.Begin()
		if err != nil {
			t.Fatalf("begin tx failed: %v", err)
		}
		if _, err := tx.Exec(`DELETE FROM users WHERE id=2`); err != nil {
			_ = tx.Rollback()
			t.Fatalf("delete user2 failed: %v", err)
		}
		var postsBy2 int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM posts WHERE user_id=2`).Scan(&postsBy2); err != nil {
			_ = tx.Rollback()
			t.Fatalf("scan postsBy2 failed: %v", err)
		}
		if postsBy2 != 0 {
			_ = tx.Rollback()
			t.Fatalf("expected posts by user2 to be 0 after cascade; got %d", postsBy2)
		}
		_ = tx.Rollback()
	})

	t.Run("6_CascadeDeletePost", func(t *testing.T) {
		tx, err := d.Begin()
		if err != nil {
			t.Fatalf("begin tx failed: %v", err)
		}
		if _, err := tx.Exec(`DELETE FROM posts WHERE id=5`); err != nil {
			_ = tx.Rollback()
			t.Fatalf("delete post5 failed: %v", err)
		}
		var post5, post5cats, post5reacts int
		_ = tx.QueryRow(`SELECT COUNT(*) FROM posts WHERE id=5`).Scan(&post5)
		_ = tx.QueryRow(`SELECT COUNT(*) FROM post_categories WHERE post_id=5`).Scan(&post5cats)
		_ = tx.QueryRow(`SELECT COUNT(*) FROM reactions WHERE post_id=5`).Scan(&post5reacts)
		if post5 != 0 || post5cats != 0 || post5reacts != 0 {
			_ = tx.Rollback()
			t.Fatalf("cascade failed: post5=%d cats=%d reacts=%d", post5, post5cats, post5reacts)
		}
		_ = tx.Rollback()
	})

	t.Run("7_Filters", func(t *testing.T) {
		techCount := mustQueryRowInt(t, d, `
			SELECT COUNT(*)
			FROM posts p
			JOIN post_categories pc ON pc.post_id=p.id
			JOIN categories c ON c.id=pc.category_id
			WHERE c.slug='technology'`)
		if techCount != 2 {
			t.Fatalf("expected 2 tech posts, got %d", techCount)
		}

		myPosts := mustQueryRowInt(t, d, `SELECT COUNT(*) FROM posts WHERE user_id=1`)
		if myPosts == 0 {
			t.Fatalf("expected posts for user 1, got 0")
		}

		likedByUser1 := mustQueryRowInt(t, d, `
			SELECT COUNT(*)
			FROM posts p
			JOIN reactions r ON r.post_id=p.id
			WHERE r.user_id=1 AND r.value=1`)
		if likedByUser1 != 1 {
			t.Fatalf("expected 1 liked post by user1, got %d", likedByUser1)
		}
	})

	t.Run("8_UpdatedAtAppManaged", func(t *testing.T) {
		var before string
		if err := d.QueryRow(`SELECT updated_at FROM posts WHERE id=1`).Scan(&before); err != nil {
			t.Fatalf("scan before failed: %v", err)
		}

		mustExec(t, d, `UPDATE posts SET title='Welcome (edit 1)' WHERE id=1`)
		var mid string
		if err := d.QueryRow(`SELECT updated_at FROM posts WHERE id=1`).Scan(&mid); err != nil {
			t.Fatalf("scan mid failed: %v", err)
		}
		if mid != before {
			t.Fatalf("expected updated_at unchanged; before=%s mid=%s", before, mid)
		}

		time.Sleep(1100 * time.Millisecond)
		mustExec(t, d, `UPDATE posts SET title='Welcome (edit 2)', updated_at = datetime('now') WHERE id=1`)
		var after string
		if err := d.QueryRow(`SELECT updated_at FROM posts WHERE id=1`).Scan(&after); err != nil {
			t.Fatalf("scan after failed: %v", err)
		}
		if after == before {
			t.Fatalf("expected updated_at to change; before=%s after=%s", before, after)
		}
	})

	t.Run("9_QuickCheck", func(t *testing.T) {
		var status string
		if err := d.QueryRow(`PRAGMA quick_check;`).Scan(&status); err != nil {
			t.Fatalf("quick_check scan failed: %v", err)
		}
		if status != "ok" {
			t.Fatalf("quick_check not ok: %s", status)
		}
		forumdb.LogInfo("Database integrity verified (PRAGMA quick_check = ok)")
	})

	abs, _ := filepath.Abs(testDBPath)
	forumdb.LogInfo("Test database path: %s", abs)
}
