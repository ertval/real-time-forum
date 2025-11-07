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

	//// Example placeholder handler that uses the DB
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	if err := db.InspectAllTables(database, w); err != nil {
	//		db.HandleRuntimeError(err, "inspecting all tables")
	//		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//		return
	//	}
	//})
	r := router.NewRouter()

	port := ":8080"
	fmt.Printf("🌐 Server running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {
		db.HandleFatalError(err, "starting HTTP server")
	}
}
