package main

import (
	"dcard/middleware"
	"dcard/redis_manager"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

const REDIS_ADDR = "localhost:6379"
const HOST_ADDR = ":8081"

func endpointHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint handler called")
}

func main() {
	redis_manager.InitRedis(REDIS_ADDR, 16, 1024, time.Second*300)

	router := mux.NewRouter()
	router.
		Methods("GET").
		Path("/").
		HandlerFunc(endpointHandler)

	n := negroni.New()
	n.Use(&middleware.RateLimiter{})
	n.UseHandler(router)

	err := http.ListenAndServe(HOST_ADDR, n)
	if err != nil {
		panic(err)
	}
}

