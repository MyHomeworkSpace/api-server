package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/MyHomeworkSpace/api-server/schools"
	"github.com/julienschmidt/httprouter"

	"github.com/MyHomeworkSpace/api-server/schools/dalton"
	"github.com/MyHomeworkSpace/api-server/schools/mit"

	"github.com/MyHomeworkSpace/api-server/api"
	"github.com/MyHomeworkSpace/api-server/auth"
	"github.com/MyHomeworkSpace/api-server/calendar"
	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/email"
)

type errorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type csrfResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func main() {
	log.Println("MyHomeworkSpace API Server")

	config.Init()

	initDatabase()
	initRedis()

	migrationName := flag.String("migrate", "", "If specified, the API server will run the migration with the given name.")
	flag.Parse()

	if *migrationName != "" {
		migrate(*migrationName)
		return
	}

	email.Init()

	calendar.InitCalendar()

	api.DB = DB
	api.MainRegistry = schools.MainRegistry
	api.RedisClient = RedisClient

	auth.DB = DB
	auth.RedisClient = RedisClient

	data.DB = DB
	data.MainRegistry = schools.MainRegistry
	data.RedisClient = RedisClient

	schools.MainRegistry.Register(dalton.CreateSchool())
	schools.MainRegistry.Register(mit.CreateSchool())

	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write([]byte("MyHomeworkSpace API Server"))
	})
	router.ServeFiles("/api_tester/*filepath", http.Dir("api_tester/"))

	api.Init(router) // API init delayed because router must be started first
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if it's a preflight, handle it here
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "false")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "authorization")
			w.WriteHeader(http.StatusOK)
			return
		}

		// otherwise, pass it through
		router.ServeHTTP(w, r)
	}))

	log.Printf("Listening on port %d", config.GetCurrent().Server.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.GetCurrent().Server.Port), nil)
	if err != nil {
		log.Fatalln(err)
	}
}
