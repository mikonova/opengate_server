package relay

import (
	"errors"
	"net"
	"net/http"
	"relayesp/wifi"
	"slices"
	"sync"
	"time"

	"github.com/mikonova/relayproto/proto"
	"github.com/mikonova/relayproto/proto/server"
	"github.com/mikonova/relayproto/types"

	"golang.org/x/net/websocket"
)

const (
	maxConns  int           = 10
	queueSize int           = 50
	deadLine  time.Duration = time.Second * 30
)

type connStorage struct {
	conns       [maxConns]*websocket.Conn
	activeConns int
	mut         sync.RWMutex
}

var (
	ConnStorage connStorage = connStorage{}
)

func CreateHTTP() {
	mux := http.NewServeMux()
	mux.Handle("/ws/", websocket.Handler(handler))
	err := http.ListenAndServe(wifi.GlobalPort, mux)
	if err != nil {
		println("[ERR] error creating a websocket listener")
	}
}

func handler(ws *websocket.Conn) {
	for {
		if ConnStorage.GetActiveConns() >= maxConns {
			var sig proto.SignalingPacket
			packet := sig.Encode(types.ServerMaxConnsReached)
			ws.SetWriteDeadline(time.Now().Add(deadLine))
			_, err := ws.Write(packet[:])
			if err != nil {
				println("[INFO] couldn't send error to client, silently closing")
			}
			ws.Close()
			return
		}
		ConnStorage.AddConn(ws)
		signatureArr := []byte{42, 34, 204, 149}
		packetArr := [8192]byte{}
		packet := packetArr[:]
		ipBuf := [20]byte{}

		ws.SetReadDeadline(time.Now().Add(deadLine * 2))
		_, err := ws.Read(packet)
		if err != nil {
			println("[INFO] client timed out ")
			ws.Close()
		}

		t := server.GetIp(packetArr, &ipBuf)
		ipArr := [16]byte{}
		sigRecvd := ipBuf[:]
		copy(sigRecvd, ipBuf[:16])
		switch t {
		case types.Message:
			if n := slices.Compare(sigRecvd, signatureArr[:]); n != 0 {
				println("[WARN] possible data fragmentation from client ")
				var s proto.SignalingPacket
				packet := s.Encode(types.OnReceiveFail)
				ws.SetWriteDeadline(time.Now().Add(deadLine))
				_, err := ws.Write(packet[:])
				if err != nil {
					println("[INFO] client is unreachable")
					ConnStorage.DelConn(ws)
					ws.Close()
				}
			}
			destConn, err := ConnStorage.GetConn(ipArr)
			if err != nil {
				println(err.Error())
				var info proto.SignalingPacket
				packet := info.Encode(types.RelayError)
				ws.SetWriteDeadline(time.Now().Add(deadLine))
				_, err := ws.Write(packet[:])
				if err != nil {
					println("[INFO] client is unreachable")
					ConnStorage.DelConn(ws)
					ws.Close()
				}
			} else {
				destConn.SetWriteDeadline(time.Now().Add(deadLine))
				_, err := destConn.Write(packet)
				if err != nil {
					println("[INFO] target client timed out")
					ConnStorage.DelConn(destConn)
					destConn.Close()
				}
			}
		case types.ClientHello:
			for tries := range 5 {
				var sig proto.SignalingPacket
				packet := sig.Encode(types.ServerHello)
				ws.SetWriteDeadline(time.Now().Add(deadLine))
				_, err := ws.Write(packet[:])
				if err != nil && tries == 3 {
					println("[INFO] client not responding")
					ws.Close()

				}
				if err != nil {
					println("[INFO] timeout, retrying")
					continue
				}
			}
		}
	}

}

func (cs *connStorage) AddConn(connHandle *websocket.Conn) {
	cs.mut.Lock()
	defer cs.mut.Unlock()
	for i := range maxConns {
		if cs.conns[i] == nil {
			cs.conns[i] = connHandle
			cs.activeConns++
			return
		}
	}
	println("[Err] reached max connections capacity")
}

func (cs *connStorage) DelConn(connHandle *websocket.Conn) {
	cs.mut.Lock()
	defer cs.mut.Unlock()
	for i := range maxConns {
		if connHandle == cs.conns[i] {
			cs.conns[i] = nil
			cs.activeConns--
			return
		}
	}
	println("connection not in ConnStorage")
}

func (cs *connStorage) GetConn(ip [16]byte) (*websocket.Conn, error) {
	cs.mut.RLock()
	defer cs.mut.RUnlock()
	ipSlice := ip[:]

	for i := range maxConns {
		if cs.conns[i] == nil {
			continue
		}
		wsaddr, _, err := net.SplitHostPort(cs.conns[i].Request().RemoteAddr)
		if err != nil {
			println("[err] malformed address")
			return nil, err
		}
		if wsip := net.ParseIP(wsaddr); wsip.Equal(ipSlice) {
			return cs.conns[i], nil
		}
	}
	return nil, errors.New("[ERR] couldn't find client while searchig")
}

func (cs *connStorage) GetActiveConns() int {
	cs.mut.RLock()
	defer cs.mut.RLock()
	return cs.activeConns
}
