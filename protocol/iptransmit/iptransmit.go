package iptransmit

type Packet struct {
	MsgType byte
	MsgLen  int32
	Message string
}
