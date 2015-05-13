package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type ApiMessage struct {
	PackType string `json:"pack_type"`
	Data     []byte `json:"data"`
}

type ApiData struct {
	Service string   `json:"service"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
}

func Warn(msg string) {
	log.Println("Catch an error, error:", msg)
	url := "http://service.192.168.94.26.xip.io/service/call"
	apiData := ApiData{"Api_Warn", "notice", []string{"common", "日志告警: " + msg}}
	apiDataEncoded, err := json.Marshal(apiData)
	if err != nil {
		log.Println(err)
		return
	}
	apiMessage := ApiMessage{"json", apiDataEncoded}
	apiMessageEncoded, err := json.Marshal(apiMessage)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Warning params:", string(apiMessageEncoded))
	bytesReader := bytes.NewReader(apiMessageEncoded)
	resp, err := http.Post(url, "application/json", bytesReader)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
}
