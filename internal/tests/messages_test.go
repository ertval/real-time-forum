// internal/tests/messages_test.go
package tests

import (
	"context"
	"database/sql"
	"testing"

	"forum/internal/db"
)

// seedTwoUsers creates alice and bob in the given DB and returns their IDs.
func seedTwoUsers(t *testing.T, sqlDB *sql.DB) (aliceID, bobID int64) {
	t.Helper()
	ctx := context.Background()

	aliceID, err := db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "alice",
		Email:     "alice@example.com",
		Password:  "password123",
		Age:       25,
		Gender:    "female",
		FirstName: "Alice",
		LastName:  "Smith",
	})
	if err != nil {
		t.Fatalf("create alice: %v", err)
	}

	bobID, err = db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "bob",
		Email:     "bob@example.com",
		Password:  "password123",
		Age:       25,
		Gender:    "male",
		FirstName: "Bob",
		LastName:  "Jones",
	})
	if err != nil {
		t.Fatalf("create bob: %v", err)
	}

	return aliceID, bobID
}

/*--------------------------
  CreateMessage
---------------------------*/

func TestCreateMessage_Persists(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, bobID := seedTwoUsers(t, sqlDB)

	msg, err := db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{
		SenderID:    aliceID,
		RecipientID: bobID,
		Body:        "hello bob",
	})
	if err != nil {
		t.Fatalf("CreateMessage: %v", err)
	}
	if msg.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if msg.SenderID != aliceID {
		t.Errorf("sender: got %d want %d", msg.SenderID, aliceID)
	}
	if msg.RecipientID != bobID {
		t.Errorf("recipient: got %d want %d", msg.RecipientID, bobID)
	}
	if msg.Body != "hello bob" {
		t.Errorf("body: got %q want %q", msg.Body, "hello bob")
	}
	if msg.CreatedAt == "" {
		t.Error("expected non-empty created_at")
	}
}

func TestCreateMessage_SelfSendRejected(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, _ := seedTwoUsers(t, sqlDB)

	_, err := db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{
		SenderID:    aliceID,
		RecipientID: aliceID,
		Body:        "talking to myself",
	})
	if err == nil {
		t.Fatal("expected error for self-send")
	}
}

func TestCreateMessage_EmptyBodyRejected(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, bobID := seedTwoUsers(t, sqlDB)

	_, err := db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{
		SenderID:    aliceID,
		RecipientID: bobID,
		Body:        "",
	})
	if err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestCreateMessage_WhitespaceBodyRejected(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, bobID := seedTwoUsers(t, sqlDB)

	_, err := db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{
		SenderID:    aliceID,
		RecipientID: bobID,
		Body:        "   ",
	})
	if err == nil {
		t.Fatal("expected error for whitespace-only body")
	}
}

/*--------------------------
  GetMessageHistory
---------------------------*/

func TestGetMessageHistory_BothDirections(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, bobID := seedTwoUsers(t, sqlDB)

	db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{SenderID: aliceID, RecipientID: bobID, Body: "hi"})
	db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{SenderID: bobID, RecipientID: aliceID, Body: "hey"})
	db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{SenderID: aliceID, RecipientID: bobID, Body: "how are you"})

	msgs, _, err := db.GetMessageHistory(ctx, sqlDB, aliceID, bobID, 0)
	if err != nil {
		t.Fatalf("GetMessageHistory: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if msgs[0].Body != "hi" || msgs[1].Body != "hey" || msgs[2].Body != "how are you" {
		t.Errorf("unexpected order: %v", msgs)
	}
}

func TestGetMessageHistory_LimitsTen(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, bobID := seedTwoUsers(t, sqlDB)

	for i := 0; i < 15; i++ {
		db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{SenderID: aliceID, RecipientID: bobID, Body: "msg"})
	}

	msgs, _, err := db.GetMessageHistory(ctx, sqlDB, aliceID, bobID, 0)
	if err != nil {
		t.Fatalf("GetMessageHistory: %v", err)
	}
	if len(msgs) != 10 {
		t.Fatalf("expected 10 messages (latest), got %d", len(msgs))
	}
}

func TestGetMessageHistory_BeforeIDPagination(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, bobID := seedTwoUsers(t, sqlDB)

	for i := 0; i < 15; i++ {
		db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{SenderID: aliceID, RecipientID: bobID, Body: "msg"})
	}

	latest, _, _ := db.GetMessageHistory(ctx, sqlDB, aliceID, bobID, 0)
	if len(latest) != 10 {
		t.Fatalf("expected 10 latest, got %d", len(latest))
	}

	older, _, err := db.GetMessageHistory(ctx, sqlDB, aliceID, bobID, latest[0].ID)
	if err != nil {
		t.Fatalf("GetMessageHistory with beforeID: %v", err)
	}
	if len(older) != 5 {
		t.Fatalf("expected 5 older messages, got %d", len(older))
	}
	for _, m := range older {
		if m.ID >= latest[0].ID {
			t.Errorf("expected id < %d, got %d", latest[0].ID, m.ID)
		}
	}
}

func TestGetMessageHistory_LimitsTenLatest(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, bobID := seedTwoUsers(t, sqlDB)

	var lastID int64
	for i := 0; i < 15; i++ {
		msg, _ := db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{SenderID: aliceID, RecipientID: bobID, Body: "msg"})
		lastID = msg.ID
	}

	msgs, _, err := db.GetMessageHistory(ctx, sqlDB, aliceID, bobID, 0)
	if err != nil {
		t.Fatalf("GetMessageHistory: %v", err)
	}
	if len(msgs) != 10 {
		t.Fatalf("expected 10 messages, got %d", len(msgs))
	}
	if msgs[len(msgs)-1].ID != lastID {
		t.Errorf("expected last message to be newest (id %d), got id %d", lastID, msgs[len(msgs)-1].ID)
	}
}

func TestGetMessageHistory_ConversationIsolation(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, bobID := seedTwoUsers(t, sqlDB)

	charlieID, err := db.CreateUser(ctx, sqlDB, db.CreateUserRequest{
		Username:  "charlie",
		Email:     "charlie@example.com",
		Password:  "password123",
		Age:       25,
		Gender:    "male",
		FirstName: "Charlie",
		LastName:  "Brown",
	})
	if err != nil {
		t.Fatalf("create charlie: %v", err)
	}

	db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{SenderID: aliceID, RecipientID: bobID, Body: "alice to bob"})
	db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{SenderID: aliceID, RecipientID: charlieID, Body: "alice to charlie"})
	db.CreateMessage(ctx, sqlDB, db.CreateMessageRequest{SenderID: charlieID, RecipientID: bobID, Body: "charlie to bob"})

	msgs, _, err := db.GetMessageHistory(ctx, sqlDB, aliceID, bobID, 0)
	if err != nil {
		t.Fatalf("GetMessageHistory: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message in alice-bob conversation, got %d", len(msgs))
	}
	if msgs[0].Body != "alice to bob" {
		t.Errorf("unexpected message in alice-bob conversation: %q", msgs[0].Body)
	}
}

func TestGetMessageHistory_EmptyConversation(t *testing.T) {
	sqlDB := setupTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	aliceID, bobID := seedTwoUsers(t, sqlDB)

	msgs, _, err := db.GetMessageHistory(ctx, sqlDB, aliceID, bobID, 0)
	if err != nil {
		t.Fatalf("GetMessageHistory: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected empty slice, got %d messages", len(msgs))
	}
}
