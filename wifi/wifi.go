package wifi

import (
	"relayesp/boardio"
	"time"

	"github.com/soypat/lneto/x/netdev"
	nl "tinygo.org/x/drivers/netlink"
	"tinygo.org/x/espradio/netlink"
)

var (
	GlobalPort      string = ":10520"
	isGlobalPortSet bool   = false
)

func WifiConnect(wifiName, pass string) (err error, address string) {
	link := netlink.Esplink{}
	netdev.UseNetdev(&link)
	params := nl.ConnectParams{
		Ssid:            wifiName,
		Passphrase:      pass,
		WatchdogTimeout: time.Second * 30,
	}
	err = link.NetConnect(&params)
	if err != nil {
		println("Connection error: ", err.Error())
		return err, ""
	}
	println("Connected to WIFI")
	time.Sleep(time.Second * 3)
	addr, err := link.Addr()
	if err != nil {
		println(err.Error())
		return err, addr.String()
	}

	println("local ip is: ", addr.String())

	return nil, addr.String()
}

func GetAskPort(port string) {
	if isGlobalPortSet == true {
		println("[WARN] global port is already set, ignoring")
		return
	}
	isGlobalPortSet = true
	if port != "" {
		GlobalPort = port
		return
	}

	port = boardio.GetInput()
	if port == "" {
		return
	}
	GlobalPort = ":" + port
}
