package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/julienschmidt/httprouter"
)

const (
	port = "11245"
	host = "127.0.0.1"
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

	fmt.Printf("Listen %s ...\n", Config.Listen)
	log.Fatal(http.ListenAndServe(Config.Listen, router))
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s := "Welcome\n"
	s += "POST /api/sign/exec\n"
	fmt.Fprint(w, s)
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

	// 签名验证
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

	// 判断必须参数是否存在
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

	Connects[execId] = c // 运行时保存执行对象

	responseHandle(200, &reply{Token: execId})
}

func genSignature(params map[string]interface{}, appSecret string) (sign string) {

	// key 排序
	sortedKeys := make([]string, 0)
	for k, _ := range params {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// 拼接k-v成字符串
	var paramStr string
	for _, k := range sortedKeys {
		value := fmt.Sprintf("%v", params[k])
		paramStr = paramStr + k + value
	}

	// 计算 md5 hash
	checksum := md5.Sum([]byte(paramStr))
	sign = hex.EncodeToString(checksum[:])
	sign += appSecret
	checksum = md5.Sum([]byte(sign))
	sign = hex.EncodeToString(checksum[:])
	return
}
