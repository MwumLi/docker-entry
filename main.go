package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var (
	Connects map[string]*Connect
	Config   Configuration
)

func init() {
	Connects = make(map[string]*Connect, 10)
	Config = loadConf("/etc/docker-entry.json")
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/api/sign/exec", apiSignExec)
	router.GET("/ws/terminal/:token", wsTerminal)

	fmt.Printf("Listen %s ...\n", Config.Listen)
	log.Fatal(http.ListenAndServe(Config.Listen, router))
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s := "Welcome\n"
	s += "POST /api/sign/exec\n"
	s += "GET /ws/terminal/:token http -> websocket\n"
	fmt.Fprint(w, s)
}
