package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./web"))
	mux := http.NewServeMux()
	mux.Handle("/", fs)
	mux.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/errors/404.html")
	})
	log.Println("Frontend running at http://localhost:3000")
	http.ListenAndServe(":3000", mux)
}
