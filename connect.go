package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Connect struct {
	// proto holds the Docker Engine REST API protocol i.e. http.
	proto string
	// node holds the host of docker Node
	node string
	// port holds the port of the Docker Engine REST API service
	port int
	// container holds the Docker Container id which be generated by docker
	container string
	// version of the server to talk to.
	version string
	// shell holds the cmd interperer which can run in the container of the 'host' docker node
	shell string
	// 一个可执行实例 ID
	execId string
	// 外部应用来连接
	token     string
	tokenLock *sync.Mutex
}

func NewConnectItemWithOpts(ops ...func(*Connect) error) (*Connect, error) {

	c := &Connect{
		proto:   "http",
		port:    2375,
		version: "v1.25",
		shell:   "sh",
	}

	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func withVersion(version string) func(*Connect) error {
	return func(c *Connect) error {
		c.version = version
		return nil
	}
}

func withNode(node string) func(*Connect) error {
	return func(c *Connect) error {
		c.node = node
		return nil
	}
}

func withPort(port int) func(*Connect) error {
	return func(c *Connect) error {
		c.port = port
		return nil
	}
}

func withContainer(container string) func(*Connect) error {
	return func(c *Connect) error {
		c.container = container
		return nil
	}
}

func encodeData(data interface{}) (*bytes.Buffer, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if err := json.NewEncoder(params).Encode(data); err != nil {
			return nil, err
		}
	}
	return params, nil
}

// getAPIPath returns the versioned request path to call the api.
// It appends the query parameters to the path if they are not empty.
func (c *Connect) getAPIPath(p string, query url.Values) string {
	var apiPath string
	if c.version != "" {
		v := strings.TrimPrefix(c.version, "v")
		apiPath = path.Join("/v"+v, p)
	} else {
		apiPath = p
	}
	host := c.node + ":" + strconv.Itoa(c.port)
	return (&url.URL{Scheme: c.proto, Host: host, Path: apiPath, RawQuery: query.Encode()}).String()
}

func (c *Connect) Exec() error {
	var response IDResponse

	req := ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		DetachKeys:   "ctrl-p,ctrl-q",
		Tty:          true,
		Cmd:          []string{c.shell},
		Privileged:   true,
	}

	reqJson, _ := encodeData(req)

	api := c.getAPIPath("containers/"+c.container+"/exec", url.Values{})
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Post(api, "application/json", reqJson)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return &ServerError{resp.StatusCode, response.Message}
	}

	c.execId = response.Id
	return nil
}

func (c *Connect) Start() (*HijackedResponse, error) {
	opts := ExecStartCheck{
		Detach: false,
		Tty:    false,
	}
	encodeOpts, err := encodeData(opts)
	if err != nil {
		return nil, err
	}

	api := c.getAPIPath("exec/"+c.execId+"/start", url.Values{})
	req, err := http.NewRequest("POST", api, encodeOpts)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")

	conn, err := net.Dial("tcp", c.node+":"+strconv.Itoa(c.port))
	if err != nil {
		return nil, err
	}

	// When we set up a TCP connection for hijack, there could be long periods
	// of inactivity (a long running command with no output) that in certain
	// network setups may cause ECONNTIMEOUT, leaving the client in an unknown
	// state. Setting TCP KeepAlive on the socket connection will prohibit
	// ECONNTIMEOUT unless the socket connection truly is broken
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	if err = req.Write(conn); err != nil {
		return nil, err
	}

	br := bufio.NewReader(conn)
	// Ignore upgrade response header
	for {
		s := ""
		if s, err = br.ReadString('\n'); err != nil {
			return nil, err
		} else {
			s = strings.TrimSpace(s)
		}
		if len(s) <= 0 {
			break
		}
	}
	return &HijackedResponse{conn, br}, nil
}

func (c *Connect) Resize(w, h int) error {
	var response IDResponse

	query := url.Values{
		"w": []string{strconv.Itoa(w)},
		"h": []string{strconv.Itoa(h)},
	}
	api := c.getAPIPath("containers/"+c.container+"/exec", query)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Post(api, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		return &ServerError{resp.StatusCode, response.Message}
	}

	return nil
}

func (c *Connect) String() string {
	return fmt.Sprintf("node: %s container: %s execId: %s", c.node, c.container, c.execId)
}

func (c *Connect) generateToken() string {
	// generate token
	ns := strconv.FormatInt(time.Now().UnixNano(), 10)
	b := md5.Sum([]byte(c.String() + ns))
	token := hex.EncodeToString(b[:])

	c.token = token
	c.tokenLock = new(sync.Mutex)
	return token
}

func (c *Connect) emptyToken() {
	c.token = ""
}
