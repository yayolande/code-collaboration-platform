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
	queries, ctx := setupDatabase()

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	sessionManager := scs.New()
	sessionManager.Store = memstore.New()

	setupRoute(router, queries, ctx, sessionManager)

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

func setupDatabase() (*database.Queries, *context.Context) {
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

	return queries, &ctx
}

func setupRoute(router *chi.Mux, queries *database.Queries, dbContext *context.Context, sessionManager *scs.SessionManager) {
	dbBucket := &handlers.DatabaseBucket{
		DB:        DB,
		Queries:   queries,
		DBContext: dbContext,
	}

	router.Get("/assets/*", handlers.ServeStaticAssets)
	// router.Post("/new/save", handlers.SavePost(dbBucket))
	router.Get("/", handlers.GetHomePage(dbBucket, sessionManager))
	router.Get("/login", handlers.GetLoginPage(sessionManager))
	router.Post("/login", handlers.LoginUser(dbBucket, sessionManager))
	router.Get("/logout", handlers.LogoutUser(sessionManager))

	router.Get("/register", handlers.GetRegistrationPage(sessionManager))
	router.Post("/register", handlers.RegisterUser(dbBucket, sessionManager))

	router.Route("/code/", func(r chi.Router) {
		r.Get("/new", handlers.CreateNewPost)
		// TODO: Modify the route paht below to "/new" but using method "POST"
		r.Post("/new/save", handlers.SavePost(dbBucket))

		r.Get("/{snippet_id}", handlers.GetPost(dbBucket, "snippet_id"))

		r.Get("/test", func(w http.ResponseWriter, req *http.Request) {
			http.ServeFile(w, req, "../dist/index.html")
		})
	})
}
