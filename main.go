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
	Config = loadConf(SYSTEM_PATH)
	// `kill -USR1 pid` to reload config
	go reloadConf(SYSTEM_PATH)
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.ServeFiles("/static/*filepath", http.Dir("static"))
	router.POST("/api/sign/exec", apiSignExec)
	router.GET("/ws/terminal/:token", wsTerminal)
	fmt.Printf("Listen %s ...\n", Config.Listen)
	log.Fatal(http.ListenAndServe(Config.Listen, router))
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if Config.Quick_Start == false {
		w.WriteHeader(404)
		w.Write([]byte("Not Found!"))
		return
	}
	t, _ := template.ParseFiles("static/index.html")
	t.Execute(w, Config)
}
