package stun

import (
	"encoding/binary"
	"errors"
	"math/rand/v2"
	"net"
	"slices"
	"time"

	stattr "relayesp/protocol/stun/stunattr"
	sterr "relayesp/protocol/stun/stunerrors"
	"relayesp/protocol/stun/stunlist"
)

type AddrInfo struct {
	isInitialised bool
	IsIPv6        bool
	Port          uint16
	Address       net.IP
}

type attrInfo struct {
	AttrType   int
	AttrValue  []byte
	FullLength uint16
}

var magicCookie []byte = []byte{0x21, 0x12, 0xA4, 0x42}

func StunDial() AddrInfo {
	for {
		if addrinfo := dialingLoop(); addrinfo.isInitialised {
			println("[WARN] " + "STUN sesponse received!")
			return addrinfo
		}
	}
}

func dialingLoop() (info AddrInfo) {
	var conn net.Conn
	transactionID := make([]byte, 12)
	inputBuf := make([]byte, 4096)
	buf := []byte{
		0x00, 0x01, // binding request
		0x00, 0x00, // message length
	}
	buf = append(buf, magicCookie...)

	var seed []byte
	for try := range 4 {
		timeBin, err := time.Now().MarshalBinary()
		if err == nil {
			seed = timeBin
			break
		}
		println("[ERR] error getting the seed, retrying")
		if try == 3 {
			panic("[ERR] couldnt get time/marshall it to bytes")
		}
	}
	chacha8 := rand.NewChaCha8([32]byte())
	_, err := rand.Read(transactionID)
	if err != nil {
		panic(err)
	}

	buf = append(buf, transactionID...)

	for _, v := range stunlist.ServerList {
	RETRY:
		conn, err = net.Dial("udp", v)
		if err != nil {
			println("[WARN] " + "dialing error on server \"" + v + "\", switching server")
			continue
		}
		conn.SetDeadline(time.Now().Add(time.Second * 3))
		if _, err := conn.Write(buf); err != nil {
			println("[WARN] " + "writing timeout on server \"" + v + "\", switching server")
			conn.Close()
			continue
		}
		if _, err := conn.Read(inputBuf); err != nil {
			println("[WARN] " + "reading timeout on server \"" + v + "\", switching server")
			conn.Close()
			continue
		}

		err, n, info := decodeResp(inputBuf, transactionID)
		if err != nil && n == sterr.Alternate {
			conn.Close()
			continue
		} else if err != nil && n == sterr.Retry {
			goto RETRY
		} else if err != nil && n == sterr.Other {
			conn.Close()
			continue
		} else if err == nil && n == sterr.Success {
			conn.Close()
			return info
		} else {
			conn.Close()
			continue
		}

	}
	return info

}

/*
err != nil if any sort of error occured. "n" represents what kind of action should be taken.
n = 0 means there's no error, n = 1 means that server should not be switched, n = 2 means alternate server should be tried
n = 3 means any other error, n = 4 means an error is supposed to be fatal (should not happen once configured)
*/
func decodeResp(buf []byte, tID []byte) (err error, n int, info AddrInfo) {
	argument := attrInfo{}
	header := buf[:20]
	args := buf[20:]
	msgLength := binary.BigEndian.Uint16(header[2:4])
	cookie := header[4:8]
	transactionID := header[8:]

	if n := slices.Compare(transactionID, tID); n != 0 {
		err = errors.New("[ERR] " + "unmatching STUN transactionID")
		println(err.Error())
		return err, sterr.Alternate, info
	}

	if n := slices.Compare(cookie, magicCookie); n != 0 {
		err = errors.New("[ERR] " + "unmatching STUN magic cookie")
		println(err.Error())
		return err, sterr.Alternate, info
	}

	for i := 0; i < int(msgLength); {
		args, argument = extractArg(args)
		i += int(argument.FullLength)
		err, n, info = parseArgument(argument.AttrValue, header, argument.AttrType, tID)
		if info.isInitialised != false {
			break
		}

	}
	return err, n, info

}

func extractArg(argList []byte) (args []byte, argument attrInfo) {
	argType := argList[:2]
	argLen := binary.BigEndian.Uint16(argList[2:4])
	var argumentType int

	if n := slices.Compare(argType, []byte{0x00, 0x20}); n == 0 {
		argumentType = stattr.XORMappedAddr
	} else if n := slices.Compare(argType, []byte{0x00, 0x01}); n == 0 {
		argumentType = stattr.MappedAddr
	} else if n := slices.Compare(argType, []byte{0x00, 0x09}); n == 0 {
		argumentType = stattr.Error
	} else if n := slices.Compare(argType, []byte{0x00, 0x08}); n == 0 {
		argumentType = stattr.MessageIntegrity
	} else {
		argumentType = stattr.Unimportant
	}

	padding := 4 - ((4 + argLen) % 4)

	info := attrInfo{
		AttrType:   argumentType,
		AttrValue:  argList[4 : argLen+4],
		FullLength: 4 + argLen + padding,
	}
	argList = argList[:argLen+padding]
	return argList, info

}

// TODO: доделать парсинг IP
func parseArgument(argument []byte, header []byte, argtype int, tID []byte) (err error, n int, addrinfo AddrInfo) {
	successResp := []byte{0x01, 0x01}
	errResp := []byte{0x01, 0x11}

	reqType := header[:2] // 0x01, 0x01 if success response, 0x01, 0x11 if an error response

	if n := slices.Compare(reqType, successResp); n == 0 && argtype == stattr.XORMappedAddr {
		_ = argument[:1]                                                                          // always 0x00
		family := argument[1]                                                                     // IPv4 or IPv6
		port := binary.BigEndian.Uint16(argument[2:4]) ^ binary.BigEndian.Uint16(magicCookie[:2]) // MAPPED
		xIP := argument[4:]                                                                       // XOR-MAPPED
		ipType, ip := decodeIP(xIP, tID, family)

		addrinfo = AddrInfo{
			isInitialised: true,
			IsIPv6:        ipType,
			Port:          port,
			Address:       ip,
		}
		return nil, sterr.Success, addrinfo

	} else if n := slices.Compare(reqType, errResp); n == 0 {
		_ = argument[:2]
		class := int(argument[2])
		number := int(argument[3])
		reason := string(argument[4:])

		declineStatus := class*100 + number
		switch declineStatus {
		case 300:
			err = errors.New("[ERR] " + "try alternate STUN server: ")
			println(err.Error(), declineStatus, ", reason: ", reason)
			return err, sterr.Alternate, addrinfo
		case 400:
			err = errors.New("[ERR] " + "malformed STUN request: ")
			println(err.Error(), declineStatus, ", reason: ", reason)
			return err, sterr.Other, addrinfo
		case 500:
			err = errors.New("[ERR] " + "temporary server error: ")
			println(err.Error(), declineStatus, ", reason: ", reason)
			return err, sterr.Retry, addrinfo
		default:
			err = errors.New("[ERR] " + "unexpected error: ")
			println(err.Error(), declineStatus, ", reason: ", reason)
			return err, sterr.Alternate, addrinfo
		}
	} else {
		err = errors.New("[ERR] " + "unmatching STUN response type")
		//log.Println(err.Error())
		return err, sterr.Other, addrinfo
	}
}

func decodeIP(xorMappedAddr []byte, transactionID []byte, fam byte) (bool, net.IP) {
	if fam == byte(0x02) {
		ipUnmapped := make([]byte, 16)
		compositeCookie := append(magicCookie, transactionID...)
		for key, value := range compositeCookie {
			ipUnmapped[key] = xorMappedAddr[key] ^ value
		}
		ip := net.IP(ipUnmapped)
		return false, ip
	} else if fam == byte(0x01) {
		ipUnmapped := make([]byte, 4)
		for key, value := range magicCookie {
			ipUnmapped[key] = xorMappedAddr[key] ^ value
		}
		ip := net.IPv4(ipUnmapped[0], ipUnmapped[1], ipUnmapped[2], ipUnmapped[3])
		return false, ip
	} else {
		println("[ERR] " + "unrecognized IP format in STUN response")
	}
	return false, nil
}
