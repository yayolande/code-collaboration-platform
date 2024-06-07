package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	path, _ := os.Getwd()
	pathStaticFiles := filepath.Join(path, "..", "dist/")

	router := chi.NewRouter()

	router.Use(middleware.Logger)

	// router.Handle("/folder", http.RedirectHandler("/", http.StatusTemporaryRedirect))

	router.Get("/assets/*", func(w http.ResponseWriter, req *http.Request) {
		path, _ := os.Getwd()
		path = filepath.Join(path, "..", "dist/")

		path = pathStaticFiles
		fs := http.FileServer(http.Dir(path))

		// log.Printf("fs : %#v\n\n", fs)
		// log.Printf("http.Dir('../dist/') : %#v\n\n", http.Dir(path))
		// log.Printf("request : %#v\n\n", req)

		fs.ServeHTTP(w, req)
	})

	router.Route("/code-snipets/", func(r chi.Router) {
		r.Get("/new", func(w http.ResponseWriter, req *http.Request) {

			// pathToFile := filepath.Join(pathStaticFiles, "new_snippet.html")
			pathToFile := filepath.Join(pathStaticFiles, "new", "index.html")
			pathToComponentFile := filepath.Join(pathStaticFiles, "components.tmpl")

			tmpl, err := template.ParseFiles(pathToFile, pathToComponentFile)
			if err != nil {
				message := "[" + req.URL.Path + "] "
				message += "Error parsing html/template file -- " + err.Error()

				log.Println(message)
				http.Error(w, message, http.StatusInternalServerError)
				return
			}

			type Post struct {
				PostId        int
				Username      string
				CodeSnipet    string
				LanguageLabel string
				Comment       string
				Date          string // time.Now() ?????????????????//
			}

			orginalPost := Post{
				PostId:        0,
				LanguageLabel: "js",
			}

			answersPost := []Post{}

			posts := struct {
				OriginalPost Post
				AnswersPost  []Post
			}{
				OriginalPost: orginalPost,
				AnswersPost:  answersPost,
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
			log.Printf("%#v", req.Form)

			http.Redirect(w, req, "../1234", http.StatusSeeOther)
		})

		r.Get("/{snipet_id}", func(w http.ResponseWriter, req *http.Request) {
			snipetId := chi.URLParam(req, "snipet_id")
			_ = snipetId

			pathToFile := filepath.Join(pathStaticFiles, "index.html")
			pathToComponentFile := filepath.Join(pathStaticFiles, "components.tmpl")

			tmpl, err := template.ParseFiles(pathToFile, pathToComponentFile)
			if err != nil {
				http.Error(w, "File not Found for "+snipetId+" :: "+err.Error(), http.StatusNotFound)
				return
			}

			// tmpl.Execute(w, snipetId)
			// tmpl.Execute(w, template.HTML(`<b>World</b>`))

			type Post struct {
				PostId        int
				Username      string
				CodeSnipet    string
				LanguageLabel string
				Comment       string
				Date          string // time.Now() ?????????????????//
			}

			orginalPost := Post{
				PostId:        2,
				Username:      "Karla",
				CodeSnipet:    "// Hello World Try by KARLA \n console.log('Hello World')\n\nfunction man() {\n\talert('Yeah')\n}",
				LanguageLabel: "js",
				Date:          "March 2024",
				Comment:       "How to do the same thing with Golang instead ? Help !",
			}

			answersPost := []Post{
				{
					PostId:        4,
					Username:      "IamTrolling",
					CodeSnipet:    "# I dont know X) \ndef mouchachou(): \n\tprint('hello friend')",
					LanguageLabel: "python",
					Comment:       "Hope it helped X-)",
					Date:          "May 2024",
				},
				{
					PostId:        8,
					Username:      "steveen",
					CodeSnipet:    "package main \n\nfunc main() { \n\tfmt.Println('Hello World') \n}",
					LanguageLabel: "go",
					Comment:       "Note that I didn't include the 'package' statement, as well as the necessary 'import'",
					Date:          "May 2024",
				},
			}

			posts := struct {
				OriginalPost Post
				AnswersPost  []Post
			}{
				OriginalPost: orginalPost,
				AnswersPost:  answersPost,
			}

			err = tmpl.Execute(w, posts)
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
