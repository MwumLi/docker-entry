package main

import (
	"encoding/json"
	"log"
	"os"
)

// ToDo: remove Enable_sign and App_keys when finished
var defaultConf = Configuration{
	Docker_proto:       "http",
	Docker_serve_port:  2375,
	Docker_api_version: "v1.24",
	Listen:             "127.0.0.1:8888",
	Quick_Start:        true,
}

func loadConf(path string) Configuration {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		if os.IsExist(err) {
			log.Fatalf("%s not exist", path)
		}
		return defaultConf
	}
	from := &Configuration{
		Quick_Start: true,
	}
	json.NewDecoder(file).Decode(&from)

	to := defaultConf
	mergeConf(&to, from)
	return to
}

func mergeConf(to, from *Configuration) {
	if from.Enable_sign {
		to.Enable_sign = from.Enable_sign
	}

	if from.Debug {
		to.Debug = from.Debug
	}

	if from.Quick_Start == false {
		to.Quick_Start = from.Quick_Start
	}

	if len(from.App_keys) > 0 {
		to.App_keys = from.App_keys
	}

	if from.Docker_proto != "" {
		to.Docker_proto = from.Docker_proto
	}

	if from.Docker_api_version != "" {
		to.Docker_api_version = from.Docker_api_version
	}

	if from.Listen != "" {
		to.Listen = from.Listen
	}
}
