// cmd/backend/server.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"forum/internal/db"
	"forum/internal/env"
	"forum/internal/router"
)

const addr = ":8080"

func Start() {
	/* ----------------------------
	   Load environment variables
	-----------------------------*/
	env.LoadEnv(".env")

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
		log.Fatal("DATABASE INIT ERROR:", err)
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
				log.Println("Session cleanup error:", err)
			}
		}
	}()

	/* ----------------------------
	   HTTP server
	-----------------------------*/
	handler := router.NewRouter(database)

	log.Println("Server running on http://localhost" + addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("SERVER ERROR:", err)
	}
}
