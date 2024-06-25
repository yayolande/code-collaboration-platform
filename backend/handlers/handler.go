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

type RouteHandler struct {
	Router *chi.Mux
	Bucket *DatabaseBucket
	Cookie *scs.SessionManager
}

func NewRouteHandler() *RouteHandler {
	r := &RouteHandler{}

	return r
}

var cookieKeyUserID string
var cookieKeyUserName string
var cookieKeyUserEmail string

var cookieKeyMessageErrorLogin string
var cookieKeyMessageSuccessLogin string
var cookieKeyMessageErrorRegistration string
var cookieKeyMessageSuccessRegistration string
var cookieKeyMessageErrorNewPost string
var cookieKeyMessageErrorHome string

func init() {
	cookieKeyUserID = "user_id"
	cookieKeyUserName = "user_name"
	cookieKeyUserEmail = "user_email"

	cookieKeyMessageErrorLogin = "error_message_login"
	cookieKeyMessageSuccessLogin = "success_message_login"

	cookieKeyMessageErrorRegistration = "error_message_registration"
	cookieKeyMessageSuccessRegistration = "success_message_registration"

	cookieKeyMessageErrorNewPost = "error_message_new_post"
	cookieKeyMessageErrorHome = "error_message_home"
}

func getLoggedUserID(cookie *scs.SessionManager, req *http.Request) int {
	id := -1

	if cookie.Exists(req.Context(), cookieKeyUserID) {
		id = cookie.GetInt(req.Context(), cookieKeyUserID)
	}

	return id
}

func getLoggedUsername(cookie *scs.SessionManager, req *http.Request) string {
	name := "PLACEHOLDER"

	if cookie.Exists(req.Context(), cookieKeyUserName) {
		name = cookie.GetString(req.Context(), cookieKeyUserName)
	}

	return name
}

func getLoggedUserEmail(cookie *scs.SessionManager, req *http.Request) string {
	email := "PLACEHOLDER@toto.cm"

	if cookie.Exists(req.Context(), cookieKeyUserEmail) {
		email = cookie.GetString(req.Context(), cookieKeyUserEmail)
	}

	return email
}

func setUserDataForTemplateEngine(data map[string]interface{}, cookie *scs.SessionManager, req *http.Request) map[string]interface{} {
	userID := getLoggedUsername(cookie, req)
	userName := getLoggedUsername(cookie, req)
	userEmail := getLoggedUserEmail(cookie, req)

	data["user_id"] = userID
	data["user_name"] = userName
	data["user_email"] = userEmail

	return data
}

func ServeStaticAssets(w http.ResponseWriter, req *http.Request) {
	path := views.PathStaticFiles
	fs := http.FileServer(http.Dir(path))

	fs.ServeHTTP(w, req)
}

func (s *RouteHandler) GetHomePage() http.HandlerFunc {
	queries := s.Bucket.Queries
	dbContext := s.Bucket.DBContext
	sessionManager := s.Cookie

	return func(w http.ResponseWriter, req *http.Request) {
		req.ParseForm()

		basePath := views.PathStaticFiles

		pathToSqueletonPage := views.PathToSqueletonPage
		pathToHomeContent := filepath.Join(basePath, "index.html")
		pathToComponentPage := views.PathToComponentPage

		entryName := views.NameSqueletonPage
		tmpl := template.New(entryName)

		tmpl.Funcs(map[string]interface{}{
			"dict": views.CreateDictionaryFuncTemplate,
		})

		tmpl, err := tmpl.ParseFiles(pathToSqueletonPage, pathToHomeContent, pathToComponentPage)
		// tmpl, err := template.ParseFiles(pathToSqueletonPage, pathToHomeContent, pathToComponentPage)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error parsing html/template file -- " + err.Error()

			log.Println(message)
			http.Error(w, message, http.StatusInternalServerError)
			return
		}

		// TODO: Implement Search features
		searchText := req.FormValue("search")
		searchText = "%" + searchText + "%"
		searchParam := database.SearchPostsParams{
			Code:    searchText,
			Comment: searchText,
		}

		log.Printf("SearchParam: %#v", searchParam)

		posts, err := queries.SearchPosts(*dbContext, searchParam)

		// TODO:  Fetch Posts with Search and Without search
		// posts, err := queries.GetRecentPosts(*dbContext)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while fetch recent posts from DB -- " + err.Error()
			log.Println(message)

			message = "Error while fetching recents posts"
			sessionManager.Put(req.Context(), cookieKeyMessageErrorHome, message)
			posts = []database.SearchPostsRow{}
			// posts = []database.GetRecentPostsRow{}
		}

		data := map[string]interface{}{}
		data["Posts"] = posts
		data["CodeLanguages"] = storage.CodeLanguages
		data["OriginalPost"] = views.Post{}

		data = setUserDataForTemplateEngine(data, sessionManager, req)

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

func (s *RouteHandler) GetNewPostPage() http.HandlerFunc {
	sessionManager := s.Cookie

	return func(w http.ResponseWriter, req *http.Request) {
		basePath := views.PathStaticFiles
		pathToNewPostContent := filepath.Join(basePath, "new", "index.html")
		pathToSqueletonPage := views.PathToSqueletonPage
		pathToComponentPage := views.PathToComponentPage

		tmpl := template.New(views.NameSqueletonPage)

		tmpl.Funcs(map[string]interface{}{
			"dict": views.CreateDictionaryFuncTemplate,
		})

		tmpl, err := tmpl.ParseFiles(pathToSqueletonPage, pathToNewPostContent, pathToComponentPage)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error parsing html/template file -- " + err.Error()

			log.Println(message)
			// TODO: Send message to user instead of nuking the app
			http.Error(w, message, http.StatusInternalServerError)
			return
		}

		messageError := sessionManager.PopString(req.Context(), cookieKeyMessageErrorNewPost)
		langs := storage.CodeLanguages

		data := map[string]interface{}{
			"EmptyPost":              views.Post{},
			"OriginalPost":           views.Post{},
			"AnswersPost":            []views.Post{},
			"CodeLanguages":          langs[:],
			"error_message_new_post": messageError,
		}

		data = setUserDataForTemplateEngine(data, sessionManager, req)

		err = tmpl.Execute(w, data)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while executing html/template file -- " + err.Error()
			log.Println(message)

			sessionManager.Put(req.Context(), cookieKeyMessageErrorNewPost, messageError)

			http.Error(w, message, http.StatusInternalServerError)
			return
		}
	}
}

func (s *RouteHandler) SavePost() http.HandlerFunc {
	DB := s.Bucket.DB
	queries := s.Bucket.Queries
	dbContext := s.Bucket.DBContext

	sessionManager := s.Cookie

	return func(w http.ResponseWriter, req *http.Request) {
		previousUrl := req.Header["Referer"][0]
		indexStart := strings.Index(previousUrl, "/code/")
		previousUrl = previousUrl[indexStart:]

		req.ParseForm()

		var formLangCode string = req.FormValue("language")
		lang, _ := storage.GetLanguageDetailsFromCode(formLangCode)

		// WARNING: Not sure this is the right place for user authentication
		if !sessionManager.Exists(req.Context(), cookieKeyUserID) {
			message := "[/" + req.Method + " " + req.URL.Path + "] "
			message += "Error, User not logged to allow saving snippet into DB"

			log.Println(message)

			errorMessage := "Error, You are not a User! Unable to create Post"
			sessionManager.Put(req.Context(), cookieKeyMessageErrorNewPost, errorMessage)

			nextUrl := previousUrl
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		userID := sessionManager.GetInt(req.Context(), cookieKeyUserID)
		log.Println("userID = ", userID)

		input := database.AddPostParams{
			Code:       req.FormValue("code"),
			Comment:    req.FormValue("comment"),
			UserID:     int64(userID),
			LanguageID: int64(lang.ID),
			PostDate:   time.Now().String(),
		}

		if strings.Trim(input.Code, " \n") == "" || strings.Trim(input.Comment, " \n") == "" {
			message := "[" + req.URL.Path + "] "
			message += "Cannot save 'Post' with empty 'Code Snippet' or empty 'Comment' !"
			log.Println(message)

			message = "Cannot save 'Post' with empty 'Code Snippet' or comment !"
			sessionManager.Put(req.Context(), cookieKeyMessageErrorNewPost, message)

			nextUrl := previousUrl
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		parentPostID, err := strconv.Atoi(req.FormValue("parent_post_id"))
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Unable to parse 'parent post id' to integer -- " + err.Error()
			log.Println(message)

			message = "Internal error while processing the 'Post' data (Contact the admin)"
			sessionManager.Put(req.Context(), cookieKeyMessageErrorNewPost, message)

			nextUrl := previousUrl
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
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

			message = "Error while saving Post to database"
			sessionManager.Put(req.Context(), cookieKeyMessageErrorNewPost, message)

			nextUrl := previousUrl
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		defer tx.Rollback()
		transaction := queries.WithTx(tx)

		post, err := transaction.AddPost(*dbContext, input)

		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while Inserting Post into DB -- " + err.Error()
			log.Println(message)

			message = "Error while inserting Post into Database"
			sessionManager.Put(req.Context(), cookieKeyMessageErrorNewPost, message)

			nextUrl := previousUrl
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
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

			message = "Error while inserting Post Tree into Database"
			sessionManager.Put(req.Context(), cookieKeyMessageErrorNewPost, message)

			nextUrl := previousUrl
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		tx.Commit()

		if parentPostID == -1 {
			parentPostID = int(post.PostID)
		}

		nextUrl := "/code/" + strconv.Itoa(parentPostID)
		http.Redirect(w, req, nextUrl, http.StatusSeeOther)
	}
}

func (s *RouteHandler) GetPostPage(urlParamName string) http.HandlerFunc {
	queries := s.Bucket.Queries
	dbContext := s.Bucket.DBContext
	sessionManager := s.Cookie

	return func(w http.ResponseWriter, req *http.Request) {
		basePath := views.PathStaticFiles

		pathToSqueletonPage := views.PathToSqueletonPage
		pathToPostContent := filepath.Join(basePath, "post", "index.html")
		pathToComponentPage := views.PathToComponentPage

		tmpl := template.New(views.NameSqueletonPage)

		tmpl.Funcs(map[string]interface{}{
			"dict": views.CreateDictionaryFuncTemplate,
		})

		tmpl, err := tmpl.ParseFiles(pathToSqueletonPage, pathToPostContent, pathToComponentPage)
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

		data := map[string]interface{}{}
		data["EmptyPost"] = views.Post{}
		data["OriginalPost"] = orginalPost
		data["AnswersPost"] = answersPost
		data["CodeLanguages"] = langs[:]

		data = setUserDataForTemplateEngine(data, sessionManager, req)

		err = tmpl.Execute(w, data)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while executing template -- " + err.Error()
			log.Println(message)

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *RouteHandler) GetLoginPage() http.HandlerFunc {
	sessionManager := s.Cookie

	return func(w http.ResponseWriter, req *http.Request) {
		basePath := views.PathStaticFiles

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

		messageError := sessionManager.PopString(req.Context(), cookieKeyMessageErrorLogin)
		messageSuccess := sessionManager.PopString(req.Context(), cookieKeyMessageSuccessRegistration)

		data := map[string]interface{}{
			"error_message_login":          messageError,
			"success_message_registration": messageSuccess,
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while executing template -- " + err.Error()
			log.Println(message)

			sessionManager.Put(req.Context(), cookieKeyMessageErrorLogin, messageError)            // Save 'error_message_login' in order to display at least once the message to user
			sessionManager.Put(req.Context(), cookieKeyMessageSuccessRegistration, messageSuccess) // Save 'success_message_registration' in order to display at least once the message to user

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *RouteHandler) LoginUser() http.HandlerFunc {
	queries := s.Bucket.Queries
	dbContext := s.Bucket.DBContext
	sessionManager := s.Cookie

	return func(w http.ResponseWriter, req *http.Request) {
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
			sessionManager.Put(req.Context(), cookieKeyMessageErrorLogin, messageErrorLogin)

			nextUrl := req.RequestURI
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		log.Printf("user logged : %#v\n", userLogged)

		sessionManager.Put(req.Context(), cookieKeyUserID, int(userLogged.UserID))
		sessionManager.Put(req.Context(), cookieKeyUserName, userLogged.Username)
		sessionManager.Put(req.Context(), cookieKeyUserEmail, userLogged.Email)

		nextUrl := "/"
		http.Redirect(w, req, nextUrl, http.StatusSeeOther)
	}
}

func (s *RouteHandler) LogoutUser() http.HandlerFunc {
	sessionManager := s.Cookie

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

func (s *RouteHandler) GetRegistrationPage() http.HandlerFunc {
	sessionManager := s.Cookie

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

		msg := sessionManager.PopString(req.Context(), cookieKeyMessageErrorRegistration)
		data := map[string]interface{}{
			"error_message_registration": msg,
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			message := "[" + req.URL.Path + "] "
			message += "Error while executing template -- " + err.Error()
			log.Println(message)

			sessionManager.Put(req.Context(), cookieKeyMessageErrorRegistration, msg) // Save error message so that user see it at least once
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *RouteHandler) RegisterUser() http.HandlerFunc {
	dbContext := s.Bucket.DBContext
	queries := s.Bucket.Queries
	sessionManager := s.Cookie

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
			sessionManager.Put(req.Context(), cookieKeyMessageErrorRegistration, registrationMessage)

			nextUrl := req.RequestURI
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		if receivedPassword == "" || receivedUsername == "" {
			message := "[" + req.URL.Path + "] "
			message += "Error, Empty password or username for User Registratoin"
			log.Println(message)

			registrationMessage := "Error, empty username or password !"
			sessionManager.Put(req.Context(), cookieKeyMessageErrorRegistration, registrationMessage)

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
			sessionManager.Put(req.Context(), cookieKeyMessageErrorRegistration, registrationMessage)

			nextUrl := req.RequestURI
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		log.Printf("Registration successful: %#v", userRegistered)

		registrationMessage := "Registration was successful"
		sessionManager.Put(req.Context(), cookieKeyMessageSuccessRegistration, registrationMessage)

		nextUrl := "/login"
		http.Redirect(w, req, nextUrl, http.StatusSeeOther)
	}
}

func (s *RouteHandler) Nohup() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {}
}

func (s *RouteHandler) UserOnly(fn http.HandlerFunc) http.HandlerFunc {
	sessionManager := s.Cookie

	return func(w http.ResponseWriter, req *http.Request) {
		if !sessionManager.Exists(req.Context(), cookieKeyUserID) {
			message := "[" + req.URL.Path + "] "
			message += "Error, user not authenticated !"
			log.Println(message)

			message = "Unauthorized access ! You must login first"
			sessionManager.Put(req.Context(), cookieKeyMessageErrorLogin, message)

			nextUrl := "/login"
			http.Redirect(w, req, nextUrl, http.StatusSeeOther)
			return
		}

		fn(w, req)
	}
}
