package handlers

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"online_code_platform_server/sqlc/database"
	"online_code_platform_server/storage"
	"online_code_platform_server/views"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type DatabaseBucket struct {
	DB        *sql.DB
	Queries   *database.Queries
	DBContext *context.Context
}

func ServeStaticAssets(w http.ResponseWriter, req *http.Request) {
	path := views.PathStaticFiles
	fs := http.FileServer(http.Dir(path))

	fs.ServeHTTP(w, req)
}

func GetHomePage(bucket *DatabaseBucket, sessionManager *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		basePath := views.PathStaticFiles

		pathToSqueletonPage := views.PathToSqueletonPage
		pathToHomeContent := filepath.Join(basePath, "index.html")

		tmpl, err := template.ParseFiles(pathToSqueletonPage, pathToHomeContent)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error parsing html/template file -- " + err.Error()

			log.Println(message)
			http.Error(w, message, http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{}

		err = tmpl.Execute(w, data)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error executing template file -- " + err.Error()

			log.Println(message)
			http.Error(w, message, http.StatusInternalServerError)
			return
		}

	}
}

// r.Get("/new",
func CreateNewPost(w http.ResponseWriter, req *http.Request) {
	basePath := views.PathStaticFiles

	pathToFile := filepath.Join(basePath, "new", "index.html")
	pathToComponentFile := filepath.Join(basePath, "components.tmpl")

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

	posts := views.PostTree{
		EmptyPost:     views.Post{},
		OriginalPost:  views.Post{},
		AnswersPost:   []views.Post{},
		CodeLanguages: langs[:],
	}

	err = tmpl.Execute(w, posts)
	if err != nil {
		message := "[" + req.URL.Path + "] "
		message += "Error while executing html/template file -- " + err.Error()

		log.Println(message)
		http.Error(w, message, http.StatusInternalServerError)
		return
	}
}

// r.Post("/new/save",
// func SavePost(DB *sql.DB, dbContext *context.Context, queries *database.Queries) http.HandlerFunc {
func SavePost(dbBucket *DatabaseBucket) http.HandlerFunc {
	DB := dbBucket.DB
	queries := dbBucket.Queries
	dbContext := dbBucket.DBContext

	return func(w http.ResponseWriter, req *http.Request) {
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

		if strings.Trim(input.Code, " \n") == "" {
			message := "[" + req.URL.Path + "] "
			message += "Cannot save 'Post' with empty 'Code Snippet' !"
			log.Println(message)

			http.Error(w, message, http.StatusInternalServerError)
			return
		}

		parentPostID, err := strconv.Atoi(req.FormValue("parent_post_id"))
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Unable to parse 'parent post id' to integer -- " + err.Error()
			log.Println(message)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if parentPostID == 0 {
			parentPostID = -1 // Db only understand this for parentless posts
		}

		tx, err := DB.Begin()
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Transaction failed to start -- " + err.Error()
			log.Println(message)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer tx.Rollback()
		transaction := queries.WithTx(tx)

		post, err := transaction.AddPost(*dbContext, input)

		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while Inserting Post into DB -- " + err.Error()
			log.Println(message)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		postTree := database.AddPostIntoTreeParams{
			PostID:       post.PostID,
			ParentPostID: int64(parentPostID),
		}

		_, err = transaction.AddPostIntoTree(*dbContext, postTree)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while Inserting PostTree into DB -- " + err.Error()
			log.Println(message)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tx.Commit()

		// nextUrl := fmt.Sprintf("../%d", post.PostID)
		if parentPostID == -1 {
			parentPostID = int(post.PostID)
		}

		// nextUrl := fmt.Sprintf("../%d", parentPostID)
		nextUrl := "../" + strconv.Itoa(parentPostID)
		http.Redirect(w, req, nextUrl, http.StatusSeeOther)
	}
}

func GetPost(dbBucket *DatabaseBucket, urlParamName string) http.HandlerFunc {
	queries := dbBucket.Queries
	dbContext := dbBucket.DBContext

	return func(w http.ResponseWriter, req *http.Request) {
		basePath := views.PathStaticFiles

		pathToFile := filepath.Join(basePath, "post", "index.html")
		pathToComponentFile := filepath.Join(basePath, "components.tmpl")

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
		idParamString := chi.URLParam(req, urlParamName)
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

		postsConverted := []views.Post{}

		for _, post := range posts {
			tmp := views.Post{}
			tmp.New(post)

			postsConverted = append(postsConverted, tmp)
		}

		orginalPost := views.Post{}
		answersPost := []views.Post{}

		if len(postsConverted) <= 0 {
			message := "[" + req.URL.Path + "] "
			message += "No Post Found"
			log.Println(message)

			w.Write([]byte("Page Not Found"))
			return
		}

		orginalPost = postsConverted[0]

		if len(postsConverted) > 1 {
			answersPost = postsConverted[1:]
		}

		langs := storage.CodeLanguages

		postsFormated := views.PostTree{
			EmptyPost:     views.Post{},
			OriginalPost:  orginalPost,
			AnswersPost:   answersPost,
			CodeLanguages: langs[:],
		}

		err = tmpl.Execute(w, postsFormated)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while executing template -- " + err.Error()
			log.Println(message)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetLoginPage(sessionManager *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		basePath := views.PathStaticFiles

		// pathToSqueleton := filepath.Join(basePath, "partial.tmpl")
		pathToSqueleton := views.PathToSqueletonPage
		pathToLoginContent := filepath.Join(basePath, "login/index.html")

		tmpl, err := template.ParseFiles(pathToSqueleton, pathToLoginContent)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while parsing template -- " + err.Error()
			log.Println(message)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		messageError := sessionManager.PopString(req.Context(), "error_message_login")
		messageSuccess := sessionManager.PopString(req.Context(), "success_message_registration")

		data := map[string]interface{}{
			"error_message_login":          messageError,
			"success_message_registration": messageSuccess,
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while executing template -- " + err.Error()
			log.Println(message)

			sessionManager.Put(req.Context(), "error_message_login", messageError)            // Save 'error_message_login' in order to display at least once the message to user
			sessionManager.Put(req.Context(), "success_message_registration", messageSuccess) // Save 'success_message_registration' in order to display at least once the message to user

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func LoginUser(bucket *DatabaseBucket, sessionManager *scs.SessionManager) http.HandlerFunc {
	queries := bucket.Queries
	dbContext := bucket.DBContext

	return func(w http.ResponseWriter, req *http.Request) {
		/*
			sessionManager.Put(req.Context(), "error_message_login", "Hello Steve, there is an issue with your login")

			nextUrl := req.RequestURI
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
		*/

		req.ParseForm()
		receivedUsername := req.PostFormValue("username")
		receivedPassword := req.PostFormValue("password")

		guest := database.LoginUserParams{
			Username: receivedUsername,
			Password: receivedPassword,
		}

		userLogged, err := queries.LoginUser(*dbContext, guest)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error fetching user from DB -- " + err.Error()
			log.Println(message)

			messageErrorLogin := "Wrong User Credential"
			sessionManager.Put(req.Context(), "error_message_login", messageErrorLogin)

			nextUrl := req.RequestURI
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		// TODO: prep works to mark user as logged
		sessionManager.Put(req.Context(), "user_id", userLogged.UserID)
		sessionManager.Put(req.Context(), "user_name", userLogged.Username)
		sessionManager.Put(req.Context(), "user_password", userLogged.Password)
		sessionManager.Put(req.Context(), "user_email", userLogged.Email)

		nextUrl := "/"
		http.Redirect(w, req, nextUrl, http.StatusSeeOther)
	}
}

func LogoutUser(sessionManager *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		err := sessionManager.Destroy(req.Context())
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while terminating/destroying the user session" + err.Error()
			log.Println(message)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextUrl := "/login"
		http.Redirect(w, req, nextUrl, http.StatusSeeOther)
	}
}

func GetRegistrationPage(sessionManager *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		basePath := views.PathStaticFiles

		pathToSqueleton := views.PathToSqueletonPage
		pathToRegistrationContent := filepath.Join(basePath, "register/index.html")

		tmpl, err := template.ParseFiles(pathToSqueleton, pathToRegistrationContent)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while parsing template -- " + err.Error()
			log.Println(message)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		msg := sessionManager.PopString(req.Context(), "error_message_registration")
		data := map[string]interface{}{
			"error_message_registration": msg,
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while executing template -- " + err.Error()
			log.Println(message)

			sessionManager.Put(req.Context(), "error_message_registration", msg) // Save error message so that user see it at least once
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func RegisterUser(bucket *DatabaseBucket, sessionManager *scs.SessionManager) http.HandlerFunc {
	dbContext := bucket.DBContext
	queries := bucket.Queries

	return func(w http.ResponseWriter, req *http.Request) {
		req.ParseForm()

		receivedPassword := req.PostFormValue("password")
		receivedConfirmedPassword := req.PostFormValue("confirm_password")
		receivedUsername := req.PostFormValue("username")

		receivedPassword = strings.Trim(receivedPassword, " ")
		receivedConfirmedPassword = strings.Trim(receivedConfirmedPassword, " ")
		receivedUsername = strings.Trim(receivedUsername, " ")

		if receivedPassword != receivedConfirmedPassword {
			message := "[" + req.URL.Path + "] "
			message += "Password is not matching the confirmation password !"
			log.Println(message)

			registrationMessage := "Error, password not matching !"
			sessionManager.Put(req.Context(), "error_message_registration", registrationMessage)

			nextUrl := req.RequestURI
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		if receivedPassword == "" || receivedUsername == "" {
			message := "[" + req.URL.Path + "] "
			message += "Error, Empty password or username for User Registratoin"
			log.Println(message)

			registrationMessage := "Error, empty username or password !"
			sessionManager.Put(req.Context(), "error_message_registration", registrationMessage)

			nextUrl := req.RequestURI
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		userToAdd := database.AddUserParams{
			Username: req.PostFormValue("username"),
			Password: req.PostFormValue("password"),
			Email:    req.PostFormValue("email"),
			Status:   int64(1),
		}

		userRegistered, err := queries.AddUser(*dbContext, userToAdd)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error While Inserting User to DB -- " + err.Error()
			log.Println(message)

			registrationMessage := err.Error()
			sessionManager.Put(req.Context(), "error_message_registration", registrationMessage)

			nextUrl := req.RequestURI
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		log.Printf("Registration successful: %#v", userRegistered)

		registrationMessage := "Registration was successful"
		sessionManager.Put(req.Context(), "success_message_registration", registrationMessage)

		nextUrl := "/login"
		http.Redirect(w, req, nextUrl, http.StatusSeeOther)
	}
}
