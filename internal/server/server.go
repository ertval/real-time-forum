package server

import (
	"fmt"
	"log"
	"net/http"

	"forum/internal/db"
)

// Start initializes the database and starts the HTTP server.
func Start() {
	// Initialize database
	database, err := db.InitDB("internal/db/forum.db", "internal/db/forum_schema.sql")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database initialized successfully.")
	defer database.Close()

	// Example handler that uses the DB
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Print all table contents to console (for debugging)
		if err := db.InspectAllTables(database); err != nil {
			log.Println("Inspect error:", err)
		}

		// Start HTTP server
		port := ":8080"
		fmt.Printf("Server running on http://localhost%s\n", port)
		log.Fatal(http.ListenAndServe(port, nil))
	})
}
