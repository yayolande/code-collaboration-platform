package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"net/http"

	"online_code_platform_server/handlers"
	"online_code_platform_server/sqlc/database"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed sqlc/schema.sql
var schema string

var (
	DB *sql.DB
)

func main() {
	dbBucket := setupDatabase()

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	sessionManager := scs.New()
	sessionManager.Store = memstore.New()

	routeHandler := handlers.NewRouteHandler()
	routeHandler.Router = router
	routeHandler.Cookie = sessionManager
	routeHandler.Bucket = dbBucket

	setupRoute(routeHandler)

	port := ":2200"
	server := http.Server{
		Addr:    port,
		Handler: sessionManager.LoadAndSave(router),
	}

	log.Println("[Server] Running server at " + port)

	err := server.ListenAndServe()
	if err != nil {
		log.Println("Server error : ", err.Error())
	}
}

func setupDatabase() *handlers.DatabaseBucket {
	db, err := sql.Open("sqlite3", "posts.db")
	if err != nil {
		message := "[DB] Error while opening DB -- " + err.Error()
		log.Println(message)

		panic(err)
	}

	DB = db
	ctx := context.Background()

	_, err = db.ExecContext(ctx, schema)
	if err != nil {
		message := "[DB Schema] Error while executing DB sql schema -- " + err.Error()
		log.Println(message)

		panic(err)
	}

	queries := database.New(db)

	dbBucket := &handlers.DatabaseBucket{
		DB:        DB,
		Queries:   queries,
		DBContext: &ctx,
	}

	return dbBucket
}

func setupRoute(server *handlers.RouteHandler) {
	router := server.Router

	router.Get("/assets/*", handlers.ServeStaticAssets)

	router.Get("/", server.UserOnly(server.GetHomePage()))
	router.Get("/login", server.GetLoginPage())
	router.Post("/login", server.LoginUser())
	router.Get("/logout", server.LogoutUser())

	router.Get("/register", server.GetRegistrationPage())
	router.Post("/register", server.RegisterUser())

	router.Route("/code/", func(r chi.Router) {
		r.Get("/new", server.UserOnly(server.GetNewPostPage()))
		r.Post("/new", server.UserOnly(server.SavePost()))

		r.Get("/{snippet_id}", server.GetPostPage("snippet_id"))

		r.Get("/test", func(w http.ResponseWriter, req *http.Request) {
			http.ServeFile(w, req, "../dist/index.html")
		})
	})
}
