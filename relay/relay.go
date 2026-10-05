package relay

import (
	"net"
	"relayesp/wifi"
	"slices"
	"time"

	"github.com/mikonova/relayproto/proto"
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
	go func() {
		bufArr := [8192]byte{}
		buf := bufArr[:]
		tries := 0
		for conn := range wifi.ConnChan {
		REREAD:
			conn.SetReadDeadline(time.Now().Add(time.Second * 5))
			_, err := conn.Read(buf)
			if err != nil && tries <= 3 {
				println("[ERR] ", err.Error())
				println("[WARN] connection ", conn.RemoteAddr().String(), " made an error, retrying")
				tries++
				goto REREAD
			} else if err != nil && tries > 3 {
				println("[WARN] connection ", conn.RemoteAddr().String(), " timed out")
				conn.Close()
				tries = 0
				continue
			}
			packet, sig := proto.Decode(buf)
			if n := slices.Compare(sig, proto.Signature); n != 0 {
				println("[WARN] wrong packet signature")

			}
			switch packet.(type) {
			case proto.IpPacket:
				ipPacket := packet.(proto.IpPacket)
				println("[ERR] client cannot send ip packets, aborting", "[", ipPacket.IpString, "]")
				conn.Close()
				continue
			case proto.MessagePacket:
				msgPacket := packet.(proto.MessagePacket)
				//if msgPacket.PacketType == types.
			}
		}
	}()

}

func redirect() {

}
