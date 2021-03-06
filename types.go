package main

import (
	"bufio"
	"net"
)

// IDResponse Response to an API call that returns just an Id
type IDResponse struct {

	// The id of the newly created object.
	// Required: true
	Id      string `json:"Id"`
	Message string `json:"message"`
}

// ExecConfig is a small subset of the Config struct that holds the configuration
// for the exec feature of docker.
type ExecConfig struct {
	User         string   // User that will run the command
	Privileged   bool     // Is the container in privileged mode
	Tty          bool     // Attach standard streams to a tty.
	AttachStdin  bool     // Attach the standard input, makes possible user interaction
	AttachStderr bool     // Attach the standard error
	AttachStdout bool     // Attach the standard output
	Detach       bool     // Execute in detach mode
	DetachKeys   string   // Escape keys for detach
	Env          []string // Environment variables
	WorkingDir   string   // Working directory
	Cmd          []string // Execution commands and args
}

// ExecStartCheck is a temp struct used by execStart
// Config fields is part of ExecConfig in runconfig package
type ExecStartCheck struct {
	// ExecStart will first check if it's detached
	Detach bool
	// Check if there's a tty
	Tty bool
}

type ServerError struct {
	StatusCode int
	Message    string
}

func (e *ServerError) Error() string {
	return e.Message
}

type HijackedResponse struct {
	Conn net.Conn
	Br   *bufio.Reader
}

type Configuration struct {
	Quick_Start        bool
	Debug              bool
	App_keys           map[string]string
	Enable_sign        bool
	Docker_proto       string
	Docker_serve_port  int
	Docker_api_version string
	Listen             string
}

const (
	// DataMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	DataMessage = "data"

	// ResizeMessage denotes a resize control message
	ResizeMessage = "resize"
)

type WebsocketMessage struct {
	Type string `json:"type"`

	Data string `json:"data,omitempty"`

	W int `json:"w,omitempty"`
	H int `json:"h,omitempty"`

	Errno int    `json:"errno,omitempty"`
	Msg   string `json:"msg,omitempty"`
}
