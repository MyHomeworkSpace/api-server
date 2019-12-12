package main

import (
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
	http.Handle("/", router)

	log.Printf("Listening on port %d", config.GetCurrent().Server.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.GetCurrent().Server.Port), nil)
	if err != nil {
		log.Fatalln(err)
	}
}
