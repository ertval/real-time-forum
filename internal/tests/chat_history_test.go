// internal/tests/chat_history_test.go
package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"forum/internal/db"
)

func TestChatHistory_Unauthorized(t *testing.T) {
	h, _ := newTestAPI(t)

	tests := []struct {
		name string
		path string
		want int
	}{
		{"no messages suffix", "/api/v1/chats/1", http.StatusUnauthorized},
		{"empty user id", "/api/v1/chats//messages", http.StatusTemporaryRedirect},
		{"missing user id", "/api/v1/chats/messages", http.StatusUnauthorized},
		{"invalid user id", "/api/v1/chats/abc/messages", http.StatusUnauthorized},
		{"negative user id", "/api/v1/chats/-1/messages", http.StatusUnauthorized},
		{"zero user id", "/api/v1/chats/0/messages", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, _ := doRequest(t, h, http.MethodGet, tt.path, nil)
			if rec.Code != tt.want {
				t.Errorf("expected %d, got %d body=%s", tt.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestChatHistory_AuthenticatedNoMessages(t *testing.T) {
	h, sqlDB := newTestAPI(t)
	defer sqlDB.Close()

	ctx := context.Background()
	now := time.Now().UnixNano()

	aliceID, err := db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "alice",
		Email:     fmt.Sprintf("alice_%d@example.com", now),
		Password:  "password123",
		Age:       20,
		Gender:    "other",
		FirstName: "Alice",
		LastName:  "Test",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "bob",
		Email:     fmt.Sprintf("bob_%d@example.com", now),
		Password:  "password123",
		Age:       20,
		Gender:    "other",
		FirstName: "Bob",
		LastName:  "Test",
	})
	if err != nil {
		t.Fatal(err)
	}

	token := loginTestUserByEmail(t, h, fmt.Sprintf("alice_%d@example.com", now))

	rec, body := doRequestWithToken(t, h, http.MethodGet, fmt.Sprintf("/api/v1/chats/%d/messages", aliceID+1), token, nil)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", rec.Code, body)
	}

	var resp struct {
		Data struct {
			Messages []any `json:"messages"`
			HasMore  bool  `json:"has_more"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(resp.Data.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(resp.Data.Messages))
	}
	if resp.Data.HasMore != false {
		t.Errorf("expected has_more=false, got %v", resp.Data.HasMore)
	}
}

func TestChatHistory_WithMessages(t *testing.T) {
	h, sqlDB := newTestAPI(t)
	defer sqlDB.Close()

	ctx := context.Background()
	now := time.Now().UnixNano()

	aliceID, err := db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "alice",
		Email:     fmt.Sprintf("alice_%d@example.com", now),
		Password:  "password123",
		Age:       20,
		Gender:    "other",
		FirstName: "Alice",
		LastName:  "Test",
	})
	if err != nil {
		t.Fatal(err)
	}
	bobID, err := db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "bob",
		Email:     fmt.Sprintf("bob_%d@example.com", now),
		Password:  "password123",
		Age:       20,
		Gender:    "other",
		FirstName: "Bob",
		LastName:  "Test",
	})
	if err != nil {
		t.Fatal(err)
	}
	charlieID, err := db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "charlie",
		Email:     fmt.Sprintf("charlie_%d@example.com", now),
		Password:  "password123",
		Age:       20,
		Gender:    "other",
		FirstName: "Charlie",
		LastName:  "Test",
	})
	if err != nil {
		t.Fatal(err)
	}

	token := loginTestUserByEmail(t, h, fmt.Sprintf("alice_%d@example.com", now))

	// Alice sends messages to Bob
	for i := 0; i < 5; i++ {
		_, err := sqlDB.ExecContext(ctx, `INSERT INTO private_messages (sender_id, recipient_id, body) VALUES (?, ?, ?)`,
			aliceID, bobID, "hello from alice")
		if err != nil {
			t.Fatalf("insert message alice->bob: %v", err)
		}
	}
	// Bob sends messages to Alice
	for i := 0; i < 7; i++ {
		_, err := sqlDB.ExecContext(ctx, `INSERT INTO private_messages (sender_id, recipient_id, body) VALUES (?, ?, ?)`,
			bobID, aliceID, "hi from bob")
		if err != nil {
			t.Fatalf("insert message bob->alice: %v", err)
		}
	}

	// Charlie sends messages to Alice (shouldn't appear in alice-bob conversation)
	_, err = sqlDB.ExecContext(ctx, `INSERT INTO private_messages (sender_id, recipient_id, body) VALUES (?, ?, ?)`,
		charlieID, aliceID, "charlie to alice")
	if err != nil {
		t.Fatalf("insert message charlie->alice: %v", err)
	}

	// Alice queries messages with Bob (userID = bobID)
	rec, body := doRequestWithToken(t, h, http.MethodGet, fmt.Sprintf("/api/v1/chats/%d/messages", bobID), token, nil)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", rec.Code, body)
	}

	var resp struct {
		Data struct {
			Messages []struct {
				ID             int64  `json:"id"`
				SenderID       int64  `json:"sender_id"`
				RecipientID    int64  `json:"recipient_id"`
				SenderUsername string `json:"sender_username"`
				Body           string `json:"body"`
				CreatedAt      string `json:"created_at"`
			} `json:"messages"`
			HasMore bool `json:"has_more"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	// Should return 10 messages (latest 10) - 5 from alice + 5 from bob = 10
	if len(resp.Data.Messages) != 10 {
		t.Errorf("expected 10 messages, got %d", len(resp.Data.Messages))
	}

	// Should be in chronological order (oldest first)
	for i := 1; i < len(resp.Data.Messages); i++ {
		if resp.Data.Messages[i].ID <= resp.Data.Messages[i-1].ID {
			t.Errorf("messages not in chronological order at index %d", i)
		}
	}

	// has_more should be true (5+7=12 total, we got 10, so more exists)
	if !resp.Data.HasMore {
		t.Errorf("expected has_more=true, got false")
	}
	_ = rec
}

func TestChatHistory_ResponseFormat(t *testing.T) {
	h, sqlDB := newTestAPI(t)
	defer sqlDB.Close()

	ctx := context.Background()
	now := time.Now().UnixNano()

	_, _ = db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "alice2",
		Email:     fmt.Sprintf("alice2_%d@example.com", now),
		Password:  "password123",
		Age:       20,
		Gender:    "other",
		FirstName: "Alice",
		LastName:  "Test",
	})
	_, _ = db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "bob2",
		Email:     fmt.Sprintf("bob2_%d@example.com", now),
		Password:  "password123",
		Age:       20,
		Gender:    "other",
		FirstName: "Bob",
		LastName:  "Test",
	})

	token := loginTestUserByEmail(t, h, fmt.Sprintf("alice2_%d@example.com", now))

	rec, body := doRequestWithToken(t, h, http.MethodGet, "/api/v1/chats/2/messages", token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, body)
	}

	// Verify the response has the "data" wrapper as per SDS
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if _, ok := resp["data"]; !ok {
		t.Errorf("response missing 'data' wrapper")
	}
}

func TestChatHistory_SenderUsernameIncluded(t *testing.T) {
	h, sqlDB := newTestAPI(t)
	defer sqlDB.Close()

	ctx := context.Background()
	now := time.Now().UnixNano()

	aliceID, aliceEmail, err := createUserForTest(ctx, sqlDB, "alice", now)
	if err != nil {
		t.Fatal(err)
	}
	bobID, _, err := createUserForTest(ctx, sqlDB, "bob", now)
	if err != nil {
		t.Fatal(err)
	}

	token := loginTestUserByEmail(t, h, aliceEmail)

	// Alice sends message to Bob
	_, err = sqlDB.ExecContext(ctx, `INSERT INTO private_messages (sender_id, recipient_id, body) VALUES (?, ?, ?)`,
		aliceID, bobID, "test message")
	if err != nil {
		t.Fatal(err)
	}

	rec, body := doRequestWithToken(t, h, http.MethodGet, fmt.Sprintf("/api/v1/chats/%d/messages", bobID), token, nil)

	var resp struct {
		Data struct {
			Messages []struct {
				SenderID       int64  `json:"sender_id"`
				SenderUsername string `json:"sender_username"`
			} `json:"messages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Data.Messages) == 0 {
		t.Fatal("expected at least one message")
	}

	if resp.Data.Messages[0].SenderUsername == "" {
		t.Error("sender_username should not be empty")
	}
	_ = rec
}

func TestChatHistory_InvalidBeforeID(t *testing.T) {
	h, sqlDB := newTestAPI(t)
	defer sqlDB.Close()

	ctx := context.Background()
	now := time.Now().UnixNano()

	_, aliceEmail, err := createUserForTest(ctx, sqlDB, "alice", now)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = createUserForTest(ctx, sqlDB, "bob", now)
	if err != nil {
		t.Fatal(err)
	}

	token := loginTestUserByEmail(t, h, aliceEmail)

	tests := []struct {
		name       string
		beforeID   string
		wantStatus int
	}{
		{"negative", "-1", http.StatusBadRequest},
		{"zero", "0", http.StatusBadRequest},
		{"invalid", "abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/v1/chats/2/messages?before_id=" + tt.beforeID
			rec, _ := doRequestWithToken(t, h, http.MethodGet, path, token, nil)
			if rec.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

// Helper to create a user and return ID + email
func createUserForTest(ctx context.Context, sqlDB *sql.DB, username string, now int64) (int64, string, error) {
	email := fmt.Sprintf("%s_%d@example.com", username, now)
	id, err := db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  username,
		Email:     email,
		Password:  "password123",
		Age:       20,
		Gender:    "other",
		FirstName: username,
		LastName:  "Test",
	})
	return id, email, err
}
