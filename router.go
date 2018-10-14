package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

var (
	upgrader websocket.Upgrader
)

func init() {
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
}

// Signature Algorithm
func genSignature(params map[string]interface{}, appSecret string) (sign string) {

	// sort keys
	sortedKeys := make([]string, 0)
	for k, _ := range params {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// join sorted keys to string
	var paramStr string
	for _, k := range sortedKeys {
		value := fmt.Sprintf("%v", params[k])
		paramStr = paramStr + k + value
	}

	// do md5 hash
	checksum := md5.Sum([]byte(paramStr))
	sign = hex.EncodeToString(checksum[:])
	sign += appSecret
	checksum = md5.Sum([]byte(sign))
	sign = hex.EncodeToString(checksum[:])
	return
}

// Create an exec instance
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

	responseHandle := func(statusCode int, r *reply) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(r)
	}

	var e exec
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		responseHandle(400, &reply{Message: "request body must be json"})
		return
	}

	// Enable signature verification
	if Config.Enable_sign {
		m := make(map[string]interface{}, 0)
		m["container_id"] = e.Container_id
		m["host"] = e.Host
		appKey := e.App_key

		if appKey == "" || e.Sign == "" {
			responseHandle(401, &reply{Message: "Invalid visit, miss parameter 'app_key' or 'sign'"})
			return
		}

		m["app_key"] = appKey

		appSecret, ok := Config.App_keys[appKey]
		if ok == false {
			responseHandle(401, &reply{Message: fmt.Sprintf("Invalid visit, app_key '%s' not exist", appKey)})
			return
		}

		if genSignature(m, appSecret) != e.Sign {
			responseHandle(401, &reply{Message: "Invalid visit"})
			return
		}
	}

	if e.Container_id == "" || e.Host == "" {
		responseHandle(400, &reply{Message: "bad parameter: miss host or container_id"})
		return
	}

	c, err := NewConnectItemWithOpts(withNode(e.Host), withContainer(e.Container_id), withPort(Config.Docker_serve_port))
	if err != nil {
		responseHandle(500, &reply{Message: fmt.Sprintf("%s", err)})
		return
	}

	execId, err := c.Exec()
	if err != nil {
		responseHandle(500, &reply{Message: fmt.Sprintf("%s", err)})
		return
	}

	// save connect instance to Connects store
	Connects[execId] = c

	log.Printf("ToTal: %d newId: %s\n", len(Connects), execId)
	responseHandle(200, &reply{Token: execId})
}

// http upgrade to websocket and  open a communication with docker client
func wsTerminal(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Detetermine whether the 'token' connect instance exists in the Connects store
	// If does not exist, response 404
	token := ps.ByName("token")
	item, ok := Connects[token]
	if ok == false {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Not Found or token '%s' has expired", token)
		return
	}
	// Avoid tokens being used again, so delete connect instance from Connects stor
	delete(Connects, token)

	// websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	hj, err := item.Start()
	if err != nil {
		ws.CloseHandler()(1001, fmt.Sprintf("%s", err))
		log.Println(err)
		return
	}
	defer hj.Conn.Close()

	// read data from docker and send it to websocket
	go readFromDockerToWS(ws, hj.Br)

	// read data from websocket and send it to docker
	readFromWSToDocker(ws, hj.Conn, item)
}

// read data from websocket and send it to docker
func readFromWSToDocker(ws *websocket.Conn, docker net.Conn, c *Connect) {
	for {
		var wm WebsocketMessage
		if ws.ReadJSON(&wm) != nil {
			break
		}

		switch wm.Type {
		case DataMessage:
			data := wm.Data
			fmt.Fprintf(docker, "%s", data)
			fmt.Printf("websocket: %s %d\n", data, len(data))
		case ResizeMessage:
			if err := c.Resize(wm.W, wm.H); err != nil {
				wm = WebsocketMessage{
					Type: ResizeMessage,
					Msg:  err.Error(),
				}

				if v, ok := err.(*ServerError); ok {
					wm.Errno = v.StatusCode
				}
				ws.WriteJSON(wm)
			}
		}
	}
}

// read data from docker and send it to websocket
func readFromDockerToWS(ws *websocket.Conn, docker *bufio.Reader) {
	for {
		data, err := docker.ReadString('\n')
		if err != nil {
			break
		}
		wm := WebsocketMessage{
			Type: DataMessage,
			Data: data,
		}
		ws.WriteJSON(wm)
		fmt.Printf("docker: %s\n", data)
	}
}
