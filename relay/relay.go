package relay

import (
	"net/http"
	"relayesp/wifi"
	"slices"
	"time"

	"github.com/mikonova/relayproto/proto"
	"github.com/mikonova/relayproto/proto/server"
	"github.com/mikonova/relayproto/types"

	"golang.org/x/net/websocket"
)

var connections []*websocket.Conn = make([]*websocket.Conn, 0)

func Listen() {
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(handler))
	err := http.ListenAndServe(wifi.GlobalPort, mux)
	if err != nil {
		println("error creating a websocket listener")
	}
}

func handler(ws *websocket.Conn) {
	signatureArr := []byte{42, 34, 204, 149}
	packetArr := [8192]byte{}
	packet := packetArr[:]
	ipBuf := [20]byte{}
	_, err := ws.Read(packet)

	if err != nil {
		println("[ERR] data reading error form the client ", ws.RemoteAddr())
	}

	server.GetIp(packetArr, &ipBuf)
	sigRecvd := ipBuf[16:]
	if n := slices.Compare(sigRecvd, signatureArr[:]); n != 0 {
		println("possible data fragmentation")
		var s proto.SignalingPacket
		packet := s.Encode(types.OnReceiveFail)
		ws.SetWriteDeadline(time.Now().Add(time.Second * 5))
		_, err := ws.Write(packet[:])
		if err != nil {

		}
	}
}
