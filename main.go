// sudo tinygo flash -target=esp32s3-generic -serial=usb -panic=print -opt=z .

package main

import (
	"machine"
	"net"
	"relayesp/boardio"
	"relayesp/protocol/stun"
	"relayesp/relay"
	"relayesp/wifi"
	"runtime"
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

	ssid, pass := askCreds()
	wifi.WifiConnect(ssid, pass)
	addrInfo := stun.StunDial()
	ip := net.IP(addrInfo.Address[:])
	println("IMPORTANT: Connection data: ", ip.String(), ", use any forwarded port")
	relay.CreateHTTP()

	for {
		time.Sleep(time.Second)
		runtime.Gosched()
	}
}

func askCreds() (string, string) {
	println("[INFO] enter network ssid")
	ssid := boardio.GetInput()
	println("[INFO] enter network password")
	pass := boardio.GetInput()
	println(ssid)
	println(pass)
	return ssid, pass
}
