package server

import (
	"fmt"
	"forum/internal/router"
	"net/http"

	"forum/internal/db"
)

func Start() {
	// Initialize database
	database, err := db.InitDB("internal/db/forum.db")
	if err != nil {
		db.HandleInitError(err, "initializing database")
		return
	}
	fmt.Println("✅ Database initialized successfully.")
	defer func() {
		if cerr := database.Close(); cerr != nil {
			db.HandleRuntimeError(cerr, "closing database")
		}
	}()

	r := router.NewRouter(database)

	port := ":8080"
	fmt.Printf("✅ Server running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		db.HandleFatalError(err, "starting HTTP server")
	}
}
