//go:build !tests

// sudo tinygo flash -target=esp32s3-generic -serial=usb -panic=print .

package main

import (
	"machine"
	"relayesp/webserver"
	"relayesp/wifi"
	"time"
)

var (
	key, pass, address string
)

func main() {
	machine.Serial.Configure(machine.UARTConfig{BaudRate: 115200})
	time.Sleep(time.Second * 5)
	println("[INFO] ESP loaded")
	wifi.LoadStack()
	println("[INFO] stack loaded")
	wifi.StartAP()
	println("[INFO] AP started!")
	println("ssid: ", wifi.Ssid, " pass: ", wifi.Pass, "port: ", wifi.LocalNetPort)
	println("[INFO] ip is ", wifi.ApIP)
	time.Sleep(time.Second)
	webserver.InitServer()

	for {
		time.Sleep(time.Second)
	}
}
