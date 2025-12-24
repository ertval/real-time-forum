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
	mux.HandleFunc("/forgotpassword", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/templates/ForgotPassword.html")
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/templates/Login.html")
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/templates/Register.html")
	})
	mux.HandleFunc("/viewpost", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/templates/ViewPost.html")
	})
	mux.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/templates/Home.html")
	})
	log.Println("Frontend running at http://localhost:3000")
	http.ListenAndServe(":3000", mux)
}
