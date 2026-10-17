package relay

import (
	"errors"
	"net/http"
	"relayesp/wifi"
	"runtime"
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
	conns       [maxConns]connInfo
	activeConns int
	mut         sync.RWMutex
}

type connInfo struct {
	conn *websocket.Conn
	id   [16]byte
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
		} else {
			var s proto.SignalingPacket
			packet := s.Encode(types.ServerHello)

			for try := range 3 {
				ws.SetWriteDeadline(time.Now().Add(deadLine))
				_, err := ws.Write(packet[:])
				if err != nil {
					continue
				}
				if err != nil && try >= 2 {
					ws.Close()
					println("[WARN] client timed out")
					return
				}

				break
			}
		}

		packet := [8192]byte{}
		ws.SetReadDeadline(time.Now().Add(deadLine * 2))
		_, err := ws.Read(packet[:])
		if err != nil {
			ws.Close()
			println("[WARN] server timed out")
			return
		}
		t := server.GetType(packet)
		if t != types.IdCarrier && !ConnStorage.GetConnByHandle(ws) {
			ws.Close()
			return
		}
		switch t {

		case types.IdCarrier:
			idSig := [20]byte{}
			server.GetId(packet, &idSig)
			idArr := [16]byte{}
			id := idArr[:]
			copy(id, idSig[:])

			_, err = ConnStorage.GetConn(idArr)
			if err != nil {
				println("[INFO] arrived client w/ invalid ID")
				var s proto.SignalingPacket
				packet := s.Encode(types.ServerInternalError)

				for try := range 3 {
					ws.SetWriteDeadline(time.Now().Add(deadLine))
					_, err = ws.Write(packet[:])
					if err != nil {
						continue
					}
					if err != nil && try >= 2 {
						println("[WARN] couldn't tell the client about invalid ID")
						ws.Close()
						return
					}

					break
				}

			} else {
				var s proto.SignalingPacket
				packet := s.Encode(types.ServerHello)

				for try := range 3 {
					ws.SetWriteDeadline(time.Now().Add(deadLine))
					_, err = ws.Write(packet[:])
					if err != nil {
						continue
					}
					if err != nil && try >= 2 {
						ws.Close()
						println("[WARN] client timed out")
						return
					}

					break
				}
				ConnStorage.AddConn(ws, idArr)
			}

		case types.ClientHello:
			var s proto.SignalingPacket
			packet := s.Encode(types.ServerHello)

			for try := range 3 {
				ws.SetWriteDeadline(time.Now().Add(deadLine))
				_, err = ws.Write(packet[:])
				if err != nil {
					continue
				}
				if err != nil && try >= 2 {
					ws.Close()
					println("[WARN] client timed out")
					return
				}

				break
			}
		case types.Message:

			idSig := [20]byte{}
			server.GetId(packet, &idSig)
			idArr := [16]byte{}
			id := idArr[:]
			copy(id, idSig[:])

			wsOther, err := ConnStorage.GetConn(idArr)
			if err != nil {
				println("[INFO] arrived client w/ invalid ID")
				var s proto.SignalingPacket
				packet := s.Encode(types.ServerInternalError)

				for try := range 3 {
					ws.SetWriteDeadline(time.Now().Add(deadLine))
					_, err = ws.Write(packet[:])
					if err != nil {
						continue
					}
					if err != nil && try >= 2 {
						println("[WARN] couldn't tell the client about invalid ID")
						ws.Close()
						return
					}

					break
				}

			} else {
				wsOther.SetWriteDeadline(time.Now().Add(deadLine))
				wsOther.Write()
			}
		}

		// ConnStorage.AddConn(ws)
		runtime.Gosched()
	}

}

func (cs *connStorage) AddConn(connHandle *websocket.Conn, id [16]byte) {
	cs.mut.Lock()
	defer cs.mut.Unlock()
	for i := range maxConns {
		if cs.conns[i].conn == nil {
			cs.conns[i].conn = connHandle
			cs.conns[i].id = id
			cs.activeConns++
			return
		}
	}
	println("[WARN] reached max connections capacity")
}

func (cs *connStorage) DelConn(connHandle *websocket.Conn) {
	cs.mut.Lock()
	defer cs.mut.Unlock()
	for i := range maxConns {
		if connHandle == cs.conns[i].conn {
			cs.conns[i].conn = nil
			cs.conns[i].id = [16]byte{}
			cs.activeConns--
			return
		}
	}
	println("[WARN] connection not in ConnStorage")
}

func (cs *connStorage) GetConnByHandle(ws *websocket.Conn) bool {
	cs.mut.RLock()
	defer cs.mut.RUnlock()
	for i := range maxConns {
		if ws == cs.conns[i].conn {
			return true
		}
	}
	return false
}

func (cs *connStorage) GetConn(id [16]byte) (*websocket.Conn, error) {
	cs.mut.RLock()
	defer cs.mut.RUnlock()

	for i := range maxConns {
		if cs.conns[i].conn == nil {
			continue
		}
		if cs.conns[i].id == id {
			return cs.conns[i].conn, nil
		}
	}
	return nil, errors.New("[WARN] couldn't find client while searchig")
}

func (cs *connStorage) GetActiveConns() int {
	cs.mut.RLock()
	defer cs.mut.RLock()
	return cs.activeConns
}
