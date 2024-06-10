package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"online_code_platform_server/sqlc/database"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
	"online_code_platform_server/storage"
	"online_code_platform_server/views"
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

	setupRoute(router, queries, ctx)

	port := ":2200"
	server := http.Server{
		Addr:    port,
		Handler: router,
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

func setupRoute(router *chi.Mux, queries *database.Queries, dbContext *context.Context) {

	router.Get("/assets/*", func(w http.ResponseWriter, req *http.Request) {
		path, _ := os.Getwd()
		path = filepath.Join(path, "..", "dist/")

		path = views.PathStaticFiles
		fs := http.FileServer(http.Dir(path))

		fs.ServeHTTP(w, req)
	})

	router.Route("/code-snipets/", func(r chi.Router) {
		r.Get("/new", func(w http.ResponseWriter, req *http.Request) {
			path := views.PathStaticFiles

			pathToFile := filepath.Join(path, "new", "index.html")
			pathToComponentFile := filepath.Join(path, "components.tmpl")

			tmpl := template.New("index.html")

			tmpl.Funcs(map[string]interface{}{
				"dict": views.CreateDictionaryFuncTemplate,
			})

			tmpl, err := tmpl.ParseFiles(pathToFile, pathToComponentFile)
			if err != nil {
				message := "[" + req.URL.Path + "] "
				message += "Error parsing html/template file -- " + err.Error()

				log.Println(message)
				http.Error(w, message, http.StatusInternalServerError)
				return
			}

			langs := storage.CodeLanguages

			orginalPost := views.Post{
				PostId:       0,
				LanguageCode: langs[0].Code,
			}

			answersPost := []views.Post{}

			posts := struct {
				OriginalPost  views.Post
				AnswersPost   []views.Post
				CodeLanguages [len(langs)]storage.CodingLanguage
			}{
				OriginalPost:  orginalPost,
				AnswersPost:   answersPost,
				CodeLanguages: langs,
			}

			err = tmpl.Execute(w, posts)
			if err != nil {
				message := "[" + req.URL.Path + "] "
				message += "Error while executing html/template file -- " + err.Error()

				log.Println(message)
				http.Error(w, message, http.StatusInternalServerError)
				return
			}
		})

		r.Post("/new/save", func(w http.ResponseWriter, req *http.Request) {
			req.ParseForm()
			log.Printf("%#v \n\n", req.Form)

			var formLangCode string = req.FormValue("language")

			lang, _ := storage.GetLanguageDetailsFromCode(formLangCode)

			// TODO: User need proper DB table
			// WARNING: Remove 'defaultUserID' for an appropriate one
			defaultUserID := -3

			input := database.AddPostParams{
				Code:       req.FormValue("code"),
				Comment:    req.FormValue("comment"),
				UserID:     int64(defaultUserID),
				LanguageID: int64(lang.ID),
				PostDate:   time.Now().String(),
			}

			post, err := queries.AddPost(*dbContext, input)

			if err != nil {
				message := "[" + req.URL.Path + "] "
				message += "Error while Inserting Post into DB -- " + err.Error()
				log.Println(message)

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			postTree := database.AddPostIntoTreeParams{
				PostID:       post.PostID,
				ParentPostID: -1,
			}

			_, err = queries.AddPostIntoTree(*dbContext, postTree)
			if err != nil {
				message := "[" + req.URL.Path + "] "
				message += "Error while Inserting PostTree into DB -- " + err.Error()
				log.Println(message)

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			nextUrl := fmt.Sprintf("../%d", post.PostID)

			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
		})

		r.Get("/{snipet_id}", func(w http.ResponseWriter, req *http.Request) {

			path := views.PathStaticFiles

			pathToFile := filepath.Join(path, "index.html")
			pathToComponentFile := filepath.Join(path, "components.tmpl")

			tmpl := template.New("index.html")

			tmpl.Funcs(map[string]interface{}{
				"dict": views.CreateDictionaryFuncTemplate,
			})

			tmpl, err := tmpl.ParseFiles(pathToFile, pathToComponentFile)
			if err != nil {
				message := "[" + req.URL.Path + "] "
				message += "Error while parsing Template file -- " + err.Error()
				log.Println(message)

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// TODO:
			// GET User Posts ...
			idParamString := chi.URLParam(req, "snipet_id")
			snipetId, err := strconv.Atoi(idParamString)

			if err != nil {
				message := "[" + req.URL.Path + "] "
				message += "URL parameter (id) must be an integer -- " + err.Error()
				log.Println(message)

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			orginalPostParam := database.GetPostsFromRootParams{
				PostID:       int64(snipetId),
				ParentPostID: int64(snipetId),
			}

			//
			// WARNING: The current implementation dont fetch the user info, since 'users' table not userd
			// In the future, we should take it into consideration
			//
			posts, err := queries.GetPostsFromRoot(*dbContext, orginalPostParam)
			if err != nil {
				message := "[" + req.URL.Path + "] "
				message += "Unable to get 'posts' from DB -- " + err.Error()
				log.Println(message)

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if len(posts) == 0 {
				// TODO: Show something if there is no posts available for a particular postID
				message := "[" + req.URL.Path + "] "
				message += "No Post Found"
				log.Println(message)

				w.Write([]byte("Page Not Found"))
				return
			}

			postsConverted := []views.Post{}

			for _, post := range posts {
				tmp := views.Post{}
				tmp.New(post)

				postsConverted = append(postsConverted, tmp)
			}

			orginalPost := views.Post{}
			answersPost := []views.Post{}

			if len(postsConverted) <= 0 {
				// TODO: Show something if there is no posts available for a particular postID
				message := "[" + req.URL.Path + "] "
				message += "No Post Found"
				log.Println(message)

				w.Write([]byte("Page Not Found"))
				return
			} else if len(postsConverted) > 0 {
				orginalPost = postsConverted[0]
			}

			if len(postsConverted) > 1 {
				answersPost = postsConverted[1:]
			}

			langs := storage.CodeLanguages

			postsFormated := views.PostTree{
				OriginalPost:  orginalPost,
				AnswersPost:   answersPost,
				CodeLanguages: langs[:],
			}

			fmt.Printf("\n Before Failure \n\n %#v \n\n", posts)

			err = tmpl.Execute(w, postsFormated)
			if err != nil {
				message := "[" + req.URL.Path + "] "
				message += "Error while executing template -- " + err.Error()
				log.Println(message)

				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

		r.Get("/test", func(w http.ResponseWriter, req *http.Request) {
			http.ServeFile(w, req, "../dist/index.html")
		})
	})

}
