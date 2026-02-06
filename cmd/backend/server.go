// cmd/backend/server.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"forum/internal/db"
	"forum/internal/router"
)

const addr = ":8080"

func Start() {
	/* ------------------------------------
	   Resolve DB path (local & Docker)
	 -------------------------------------*/
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/forum.db"
	}

	/* ----------------------------
	  Initialize database
	 -----------------------------*/
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	log.Println("Database initialized at", dbPath)

	/* ----------------------------
	   Background session cleanup
	-----------------------------*/
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := db.CleanupSessions(context.Background(), database); err != nil {
				log.Println("session cleanup error:", err)
			}
		}
	}()

	/* ----------------------------
	   HTTP server
	-----------------------------*/
	handler := router.NewRouter(database)
	log.Println("Server running on http://localhost" + addr)

	log.Fatal(http.ListenAndServe(addr, handler))
}
