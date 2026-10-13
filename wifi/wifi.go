package wifi

import (
	"net/netip"
	"relayesp/boardio"
	"time"

	"github.com/soypat/lneto/x/netdev"
	nl "tinygo.org/x/drivers/netlink"
	"tinygo.org/x/espradio"
	"tinygo.org/x/espradio/netlink"
)

const (
	Ssid         = "ESP_AP"
	Pass         = "88888888"
	ApIP         = "192.168.4.1"
	LocalNetPort = ":5000"
)

var (
	GlobalPort      string = ":10520"
	isGlobalPortSet bool   = false
	Link            *netlink.Esplink
)

func LoadStack() {
	Link = &netlink.Esplink{}
	netdev.UseNetdev(Link)
}

func WifiConnect(wifiName, pass string) (err error, address string) {
	params := nl.ConnectParams{
		Ssid:            wifiName,
		Passphrase:      pass,
		WatchdogTimeout: time.Second * 30,
	}
	err = Link.NetConnect(&params)
	if err != nil {
		println("Connection error: ", err.Error())
		return err, ""
	}
	println("Connected to WIFI")
	time.Sleep(time.Second * 3)
	addr, err := Link.Addr()
	if err != nil {
		println(err.Error())
		return err, addr.String()
	}

	println("local ip is: ", addr.String())

	return nil, addr.String()
}

func getAskPort(port string) {
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

func StartAP() {
	err := Link.NetConnectAP(netlink.APConnectParams{
		APConfig: espradio.APConfig{
			SSID:     Ssid,
			Password: Pass,
			Channel:  6,
		},
		StaticAddr:       netip.MustParseAddr(ApIP),
		EnableDHCPServer: true,
		MaxUDPPorts:      2,
		MaxTCPPorts:      4,
		PassivePeers:     64,
	})
	if err != nil {
		println("[ERR] failed to start AP: ", err)
		return
	}
}
