package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

const (
	port = "11245"
	host = "127.0.0.1"
)

func main() {
	router := httprouter.New()
	router.POST("/api/sign/exec", apiSignExec)

	log.Fatal(http.ListenAndServe(host+":"+port, router))
}

func apiSignExec(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type exec struct {
		Container_id string
		Host         string
		App_key      string
		Sign         string
	}

	type reply struct {
		Token   string `json:"token,omitempty"`
		Message string `json:"message,omitempty"`
	}
	var e exec
	setHeader := func(w http.ResponseWriter, statusCode int) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
	}

	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		setHeader(w, 400)
		res := reply{
			Message: "request body must be json",
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	if e.Container_id == "" || e.Host == "" {
		setHeader(w, 400)
		fmt.Fprintf(w, "{ message: '%s'}", "bad parameter: miss host or container_id")
		return
	}

	fmt.Printf("%+v\n", e)
	setHeader(w, 400)
	fmt.Fprintf(w, "{ message: '%s'}", "bad parameter: miss host or container_id")
}
