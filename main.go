// sudo tinygo flash -target=esp32s3-generic -serial=usb -panic=print -opt=z .

package main

import (
	"machine"
	"relayesp/boardio"
	"relayesp/relay"
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

CREDS:
	ssid, pass := askCreds()
	err, _ := wifi.WifiConnect(ssid, pass)
	if err != nil {
		goto CREDS
	}
	relay.CreateHTTP()

	for {
		time.Sleep(time.Second)
	}
}

func askCreds() (string, string) {
	println("[INFO] enter network ssid")
	ssid := boardio.GetInput()
	println("[INFO] enter network password")
	pass := boardio.GetInput()
	println(ssid, " ", pass)
	return ssid, pass
}
