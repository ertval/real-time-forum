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
	if err := db.InitDB("forum.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database initialized successfully.")

	// Basic route for testing
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the Forum!")
	})

	// Start HTTP server
	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
