package webserver

import (
	"encoding/json"
	"net"
	"net/http"
	"relayesp/wifi"
)

type Credentials struct {
	Ssid     string `json:"ssid"`
	Password string `json:"password"`
}

// var sockChan = chan

func InitServer() {
	listener, err := net.Listen("tcp", wifi.LocalNetPort)
	if err != nil {
		println("[ERR] unable to create listener")
		return
	}
	http.HandleFunc("GET /{$}", generalHandler)
	http.HandleFunc("POST /api/wifi/setup", parseParams)
	err = http.Serve(listener, nil)
	if err != nil {
		println(err)
	}
}

func generalHandler(w http.ResponseWriter, r *http.Request) {
	println("[INFO] client connected!")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte(MainPage))
	if err != nil {
		println(err)
	}
}

func parseParams(w http.ResponseWriter, r *http.Request) {
	println("[INFO] credentials received!")
	cred := Credentials{}
	bufArr := [1024]byte{}
	buf := bufArr[:]
	r.Body.Read(buf)
	err := json.Unmarshal(buf, &cred)
	if err != nil {
		println("[ERR] couldn't unmarshall request")
		return
	}
	if cred.Password == "" || cred.Ssid == "" {
		println("[INFO] fabricated or empty request, aborting")
		return
	}

	wifi.WifiConnect(cred.Ssid, cred.Password)
}
