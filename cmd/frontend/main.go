package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	// ---------------------------------------------------------
	// Static assets (/static/*)
	// ---------------------------------------------------------
	mux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("./web/static")),
		),
	)

	// ---------------------------------------------------------
	// Pages (HTML templates)
	// ---------------------------------------------------------

	mux.HandleFunc("/", serveTemplate("./web/templates/home.html"))

	mux.HandleFunc("/home", serveTemplate("./web/templates/home.html"))
	mux.HandleFunc("/create-post", serveTemplate("./web/templates/create-post.html"))
	mux.HandleFunc("/forgot-password", serveTemplate("./web/templates/forgotpassword.html"))
	mux.HandleFunc("/login", serveTemplate("./web/templates/login.html"))
	mux.HandleFunc("/register", serveTemplate("./web/templates/register.html"))
	mux.HandleFunc("/view-post", serveTemplate("./web/templates/view-post.html"))

	log.Println("Frontend running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", mux))
}

func serveTemplate(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
