package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"online_code_platform_server/handlers"
	"online_code_platform_server/sqlc/database"
	"online_code_platform_server/views"

	"github.com/BurntSushi/toml"
	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
	_ "github.com/mattn/go-sqlite3"
)

type (
	Config struct {
		ServerName    string `toml:"name"`
		RootServeFile string `toml:"root_serve_file"`
		Port          int    `toml:"port"`
		Address       string `toml:"address"`
	}
)

//go:embed sqlc/schema.sql
var schema string

var (
	DB *sql.DB
)

func main() {
	config := getUserConfiguration()
	views.SetPathToStaticFiles(config.RootServeFile)

	dbBucket := setupDatabase()

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	sessionManager := scs.New()
	sessionManager.Store = memstore.New()

	clients := make(map[*websocket.Conn]bool)
	upgrader := setupNewUpgrader(clients)

	routeHandler := handlers.NewRouteHandler()
	routeHandler.Router = router
	routeHandler.Cookie = sessionManager
	routeHandler.Bucket = dbBucket
	routeHandler.WebSocketUpgrader = upgrader
	routeHandler.ConnectedClients = &clients

	setupRoute(routeHandler)

	port := ":" + strconv.Itoa(config.Port)
	/*
				server := http.Server{
					Addr:    port,
					Handler: sessionManager.LoadAndSave(router),
				}

		    defer server.Close()
	*/
	engine := nbhttp.NewEngine(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{port},
		Handler: sessionManager.LoadAndSave(routeHandler.Router),
		// WriteTimeout:  time.Minute * 1,
		// KeepaliveTime: time.Minute * 10,
	})

	// err := server.ListenAndServe()
	err := engine.Start()
	if err != nil {
		log.Println("[Error] Server shuting down : ", err.Error())
		panic(err)
	}

	log.Println("[Server] Running server at " + port)
	defer engine.Stop()

	<-make(chan int)
}

func getUserConfiguration() *Config {
	var config Config

	filename := "config.toml"
	_, err := toml.DecodeFile(filename, &config)

	if err != nil {
		log.Println(
			"[Warning] Server configuration file not found or malformated ! \n",
			err.Error(),
			"\n --- New file will be created with default settings",
		)

		config = Config{
			ServerName:    "Online Code Platform Server",
			RootServeFile: "../dist/",
			Port:          3500,
		}

		file, err := os.Create(filename)
		if err != nil {
			log.Println("[Error] Unable to create configuration file -- ", filename, " -- on system: ", err.Error())

			goto FINISHED_CONFIG_LABEL
		}

		configData, err := toml.Marshal(config)
		if err != nil {
			log.Println("[Error] Unable to parse 'toml' data (Marshal): ", err.Error())

			goto FINISHED_CONFIG_LABEL
		}

		_, err = file.Write(configData)
		if err != nil {
			log.Println("[Error] Unable to write 'toml' data to disk: ", err.Error())

			goto FINISHED_CONFIG_LABEL
		}
	}

FINISHED_CONFIG_LABEL:

	log.Printf("[Info] %#v \n", config)

	return &config
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

func setupNewUpgrader(clients map[*websocket.Conn]bool) *websocket.Upgrader {
	upgrader := websocket.NewUpgrader()
	upgrader.KeepaliveTime = 5 * time.Minute

	var mu sync.Mutex

	upgrader.OnOpen(func(conn *websocket.Conn) {
		log.Println("On Open: ", conn.RemoteAddr().String())

		mu.Lock()
		clients[conn] = true
		mu.Unlock()

		log.Printf("clients : %#v \n", clients)
	})

	upgrader.OnMessage(func(conn *websocket.Conn, mt websocket.MessageType, data []byte) {
		log.Println("On Message WS: ", string(data), "-- data type: ", mt)

		counter := 1

		for client := range clients {
			if client == conn {
				continue
			}

			client.WriteMessage(mt, data)
			counter++
			log.Printf("[client %d] %v ===> send data : %s", counter, client, string(data))
		}

		log.Printf("clients : %#v \n", clients)
	})

	upgrader.OnClose(func(conn *websocket.Conn, err error) {
		log.Println("On Close WS: ", conn.RemoteAddr().String(), " -- error: ")

		mu.Lock()
		delete(clients, conn)
		mu.Unlock()

		log.Printf("clients : %#v \n", clients)
		if err != nil {
			log.Println("[Error] An unexpected error occured for ", conn, " : ", err.Error())
			return
		}
	})

	return upgrader
}

func setupRoute(server *handlers.RouteHandler) {
	router := server.Router

	router.Get("/assets/*", handlers.ServeStaticAssets)

	router.Get("/login", server.GetLoginPage())
	router.Post("/login", server.LoginUser())
	router.Get("/logout", server.LogoutUser())

	router.Get("/register", server.GetRegistrationPage())
	router.Post("/register", server.RegisterUser())

	router.Mount("/debug", middleware.Profiler())

	router.Group(func(r chi.Router) {
		r.Use(server.UserOnly)

		r.Get("/", (server.GetHomePage()))
	})

	router.Route("/code/", func(r chi.Router) {
		r.Use(server.UserOnly)

		r.Get("/new", server.GetNewPostPage())
		r.Post("/new", server.SavePost())

		r.Get("/{snippet_id}", server.GetPostPage("snippet_id"))

		r.Get("/ws", server.GetEditorWebSocket())
	})
}
