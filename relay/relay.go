package relay

import (
	"net"
)

type RelayTableRow struct {
	ConnOne, ConnTwo     net.Conn
	IpInfoOne, IpInfoTwo net.IP
}

type UnorderedClient struct {
	Client net.Conn
	Ipaddr net.IP
}

var UnorderedClientList []UnorderedClient = make([]UnorderedClient, 0)

func (rtw RelayTableRow) Sync() {

}

func (rtw RelayTableRow) Reject() {

}

func HandleConn() {

}

func redirect() {

}
