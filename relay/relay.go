package relay

import (
	"encoding/binary"
	"net"
	"relayesp/protocol/iptransmit/types"
	"relayesp/wifi"
	"time"
)

type RelayTableRow struct {
	ConnOne, ConnTwo     net.Conn
	IpInfoOne, IpInfoTwo net.IP
}

type UnorderedClient struct {
	Client net.Conn
	Ipaddr net.IP
}

func (rtw RelayTableRow) Sync() {

}

func (rtw RelayTableRow) Reject() {

}

func FindClientPairs(clients []UnorderedClient) (rtwList []RelayTableRow) {

	rtwList = make([]RelayTableRow, 0)
	for key, client := range clients {
		for k, v := range clients {
			if client.Ipaddr.String() == v.Ipaddr.String() &&
				key != k {
				rtwList = append(rtwList, RelayTableRow{
					ConnOne:   client.Client,
					ConnTwo:   v.Client,
					IpInfoOne: client.Ipaddr,
					IpInfoTwo: v.Ipaddr,
				})
			}
		}
	}
	return

}

func GetUnorderedClients() []UnorderedClient {
	clients := make([]UnorderedClient, 0)
	maxRetries := 0
	for conn := range wifi.ConnChan {
		conn.SetReadDeadline(time.Now().Add(time.Second * 30))
		var staticBuf = [512]byte{}
		buf := staticBuf[:]
	READ:
		_, err := conn.Read(buf)
		if err != nil && maxRetries <= 3 {
			println("Couldnt read from a client, retrying... (client ip:", conn.LocalAddr().String(), ")")
			maxRetries++
			goto READ

		} else if err != nil && maxRetries > 3 {
			println("Couldnt read from a client, retries threshold reached")
			if resp := respondToClient(conn); !resp {
				println("skipping problematic client, client ip:", conn.LocalAddr().String())
				continue
			}
		}
		if buf[1] != types.ClientSyncReq {
			println("unmatching packet type, required ClientSyncReq, client ", conn.LocalAddr().String())
			conn.Close()
			continue
		}
		len := binary.BigEndian.Uint32(buf[10:14])
		ipaddr := string(buf[14:len])
		ip := net.ParseIP(ipaddr)
		if ip == nil {
			println("incorrect ip from the client", conn.LocalAddr().String())
		}
		clients = append(clients, UnorderedClient{
			Client: conn,
			Ipaddr: ip,
		})
	}
	return clients

}

func respondToClient(conn net.Conn) (success bool) {
	outgoungBuf := make([]byte, 14)
	outgoungBuf[0] = types.ServerError
	outgoungBuf[9] = 1
	binary.BigEndian.Uint32(outgoungBuf[10:])
	conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
	_, err := conn.Write(outgoungBuf)
	if err != nil {
		println("unable to send a responce, skipping")
		return false
	}
	return true
}
