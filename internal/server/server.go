package server

import (
	"log"
	"net/http"

	"forum/internal/db"
	"forum/internal/router"
)

const (
	dbPath = "./internal/db/forum.db"
	addr   = ":8080"
)

func Start() {
	// ---------------------------------------------------------
	// Initialize database
	// ---------------------------------------------------------
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	log.Println("Database initialized")

	// ---------------------------------------------------------
	// Build HTTP handler (router + middleware)
	// ---------------------------------------------------------
	handler := router.NewRouter(database)

	log.Printf("Server running on http://localhost%s", addr)

	// ---------------------------------------------------------
	// Start HTTP server
	// ---------------------------------------------------------
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
