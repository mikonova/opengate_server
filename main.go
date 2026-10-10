//go:build !tests

package main

import (
	"machine"
	"relayesp/boardio"
	"relayesp/wifi"
	"time"
)

var (
	key, pass, address string
	CommChan           chan byte = make(chan byte)
)

func main() {
	machine.Serial.Configure(machine.UARTConfig{BaudRate: 115200})
	time.Sleep(time.Second * 5)
LOGIN:
	println("enter wifi SSID")
	ssid := boardio.GetInput()
	println("enter wifi password")
	pass := boardio.GetInput()
	println("current inputs are: ", ssid, pass)
	println("proceed (y/n)?[y]")
	conf := boardio.GetInput()
	if conf == "n" {
		goto LOGIN
	} else {
		var err error = nil
		err, address = wifi.WifiConnect(ssid, pass)
		if err != nil {
			println("Unable to connect to network")
			println(err.Error())
			goto LOGIN
		}
	}
	for {

	}
}
