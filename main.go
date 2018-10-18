package main

import (
	"fmt"
	"html/template"
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
	if Config.Quick_Start {
		router.GET("/", Index)
		router.ServeFiles("/static/*filepath", http.Dir("static"))
	}
	router.POST("/api/sign/exec", apiSignExec)
	router.GET("/ws/terminal/:token", wsTerminal)
	fmt.Printf("Listen %s ...\n", Config.Listen)
	log.Fatal(http.ListenAndServe(Config.Listen, router))
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("static/index.html")
	t.Execute(w, Config)
}
