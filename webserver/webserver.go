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

func InitServer() {
	http.HandleFunc("GET /{$}", generalHandler)
	http.HandleFunc("POST /api/wifi/setup", parseParams)
	addr := net.JoinHostPort(wifi.ApIP, wifi.LocalNetPort)
	err := http.ListenAndServe(addr, nil)
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
	bufArr := [8192]byte{}
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
