package wifi

import (
	"net"
	"time"

	"github.com/soypat/lneto/x/netdev"
	nl "tinygo.org/x/drivers/netlink"
	"tinygo.org/x/espradio/netlink"
)

var (
	ConnChan    chan net.Conn = make(chan net.Conn, 0)
	ControlChan chan int      = make(chan int, 0)
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

func Listen(address string) net.Listener {
	listener, err := net.Listen("udp", address)
	if err != nil {
		println("Unable to create a listener")
		return listener
	}
	return listener
}

func InitConnHandler(listner net.Listener) {
	go func() {
		for {
			if n := <-ControlChan; n == -1 {
				return
			}
			conn, err := listner.Accept()
			if err != nil {
				println("unable to accept connection, retrying")
			}
			ConnChan <- conn
			time.Sleep(time.Second * 5)
		}
	}()
}
