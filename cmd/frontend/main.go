package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// Static assets only
	mux.Handle("/css/",
		http.StripPrefix("/css/",
			http.FileServer(http.Dir("./web/css")),
		),
	)

	// Pages
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	mux.HandleFunc("/createpost", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/templates/CreatePost.html")
	})
	log.Println("Frontend running at http://localhost:3000")
	http.ListenAndServe(":3000", mux)
}
