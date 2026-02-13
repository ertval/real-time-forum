// internal/router/router.go

package router

import (
	"database/sql"
	"net/http"

	"forum/internal/handlers"
	"forum/internal/middleware"
)

const (
	apiPrefix      = "/api/v1"
	frontendOrigin = "http://localhost:3000"
)

func NewRouter(database *sql.DB) http.Handler {
	mux := http.NewServeMux()

	/*-----------
	  HANDLERS
	-----------*/
	health := handlers.NewHealthHandler()
	posts := handlers.NewPostsHandler(database)
	users := handlers.NewUsersHandler(database)
	categories := handlers.NewCategoriesHandler(database)

	/*------------
	  MIDDLEWARE
	------------*/
	auth := middleware.Auth(database)

	/*--------
	  HEALTH
	--------*/
	mux.Handle(
		apiPrefix+"/health",
		middleware.AllowMethods(
			http.HandlerFunc(health.Health),
			http.MethodGet,
		),
	)
	/*---------------------
	  CATEGORIES (PUBLIC)
	---------------------*/
	mux.Handle(
		apiPrefix+"/categories",
		middleware.AllowMethods(
			http.HandlerFunc(categories.HandleCategories),
			http.MethodGet,
		),
	)

	mux.Handle(
		apiPrefix+"/categories/",
		middleware.AllowMethods(
			http.HandlerFunc(categories.HandleCategory),
			http.MethodGet,
		),
	)

	mux.Handle(
		apiPrefix+"/categories/view",
		middleware.AllowMethods(
			http.HandlerFunc(categories.ListCategoriesWithPosts),
			http.MethodGet,
		),
	)

	/*----------------------------
	  POSTS - PUBLIC COLLECTIONS
	----------------------------*/

	mux.Handle(
		apiPrefix+"/posts/public",
		middleware.AllowMethods(
			http.HandlerFunc(posts.ListPublicPosts),
			http.MethodGet,
		),
	)

	/*-------------
	  POSTS DRAFT
	-------------*/
	mux.Handle(
		apiPrefix+"/posts/draft",
		middleware.AllowMethods(
			auth(http.HandlerFunc(posts.HandleDraft)),
			http.MethodGet,
			http.MethodPost,
		),
	)

	mux.Handle(
		apiPrefix+"/posts/draft/",
		middleware.AllowMethods(
			auth(http.HandlerFunc(posts.HandleDraftByID)),
			http.MethodPut,
			http.MethodDelete,
		),
	)

	/*-------------------
	  POSTS COLLECETION
	-------------------*/
	// GET  /posts → list
	// POST /posts → create (auth)
	mux.HandleFunc(apiPrefix+"/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			posts.HandlePosts(w, r)
		case http.MethodPost:
			auth(http.HandlerFunc(posts.HandlePosts)).ServeHTTP(w, r)
		default:
			handlers.MethodNotAllowed(w, r)
		}
	})

	/*------------------------------------
	  POSTS ITEM + COMMENTS + REACTIONS)
	------------------------------------*/
	// Handles:
	// GET    /posts/{id}
	// PATCH  /posts/{id}
	// DELETE /posts/{id}
	// POST   /posts/{id}/like
	// POST   /posts/{id}/dislike
	// GET    /posts/{id}/comments
	// POST   /posts/{id}/comments
	mux.HandleFunc(apiPrefix+"/posts/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			posts.HandlePost(w, r)
		case http.MethodPost, http.MethodPatch, http.MethodDelete:
			auth(http.HandlerFunc(posts.HandlePost)).ServeHTTP(w, r)
		default:
			handlers.MethodNotAllowed(w, r)
		}
	})

	/*-------------
	  USER POSTS
	-------------*/
	mux.Handle(
		apiPrefix+"/posts/mine",
		middleware.AllowMethods(
			auth(http.HandlerFunc(posts.ListMyPosts)),
			http.MethodGet,
		),
	)

	mux.Handle(
		apiPrefix+"/posts/liked",
		middleware.AllowMethods(
			auth(http.HandlerFunc(posts.ListLikedPosts)),
			http.MethodGet,
		),
	)

	/*---------
	   USERS
	---------*/
	mux.Handle(
		apiPrefix+"/users/register",
		middleware.AllowMethods(
			http.HandlerFunc(users.Register),
			http.MethodPost,
		),
	)

	mux.Handle(
		apiPrefix+"/users/login",
		middleware.AllowMethods(
			http.HandlerFunc(users.Login),
			http.MethodPost,
		),
	)

	mux.Handle(
		apiPrefix+"/users/logout",
		middleware.AllowMethods(
			auth(http.HandlerFunc(users.Logout)),
			http.MethodPost,
		),
	)

	/*---------------
	  GOOGLE AUTH
	---------------*/
	mux.Handle(
		apiPrefix+"/auth/google",
		middleware.AllowMethods(
			http.HandlerFunc(users.GoogleStart),
			http.MethodGet,
		),
	)

	mux.Handle(
		apiPrefix+"/auth/google/callback",
		middleware.AllowMethods(
			http.HandlerFunc(users.GoogleCallback),
			http.MethodGet,
		),
	)

	mux.Handle(
		apiPrefix+"/users/me",
		middleware.AllowMethods(
			auth(http.HandlerFunc(users.Me)),
			http.MethodGet,
		),
	)

	mux.Handle(
		apiPrefix+"/users/",
		middleware.AllowMethods(
			auth(http.HandlerFunc(users.HandleUser)),
			http.MethodGet,
		),
	)

	/*---------------------------
	  COMMENTS ITEM + REACTIONS
	---------------------------*/
	// GET    /comments/{id}
	// PATCH  /comments/{id}
	// DELETE /comments/{id}
	// POST   /comments/{id}/like
	// POST   /comments/{id}/dislike
	mux.HandleFunc(apiPrefix+"/comments/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			posts.HandleComment(w, r)
		case http.MethodPost, http.MethodPatch, http.MethodDelete:
			auth(http.HandlerFunc(posts.HandleComment)).ServeHTTP(w, r)
		default:
			handlers.MethodNotAllowed(w, r)
		}
	})

	/*-------------------------
	  API FALLBACK (JSON 404)
	-------------------------*/
	mux.HandleFunc("/api", notFoundJSON)
	mux.HandleFunc("/api/", notFoundJSON)

	/*-------------------
	  GLOBAL MIDDLEWARE
	-------------------*/
	return addMiddlewares(mux)
}

func notFoundJSON(w http.ResponseWriter, r *http.Request) {
	handlers.WriteError(
		w,
		r,
		handlers.NewError("NOT_FOUND", "route not found", http.StatusNotFound),
	)
}

func addMiddlewares(handler http.Handler) http.Handler {
	handler = middleware.Logger(handler)
	handler = middleware.Recoverer(handler)
	handler = middleware.EnableCORS(frontendOrigin)(handler)
	return handler
}
