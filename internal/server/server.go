package server

import (
	"fmt"
	"log"
	"net/http"

	"forum/internal/db"
)

func Start() {
	// Initialize database
	database, err := db.InitDB("internal/db/forum.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database initialized successfully.")
	defer database.Close()

	// Example placeholder handler that uses the DB
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Print all table contents to console (for debugging)
		if err := db.InspectAllTables(database, w); err != nil {
			log.Println("Inspect error:", err)
		}
	})

	port := ":8080"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
