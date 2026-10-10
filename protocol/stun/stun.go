package stun

import (
	"encoding/binary"
	"errors"
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
	Address       [16]byte
}

type attrInfo struct {
	AttrType   int
	AttrValue  []byte
	FullLength uint16
}

// readonly
var magicCookie [4]byte = [4]byte{0x21, 0x12, 0xA4, 0x42}

// constant allocations
var (
	errAlternate      = errors.New("[ERR] try alternate STUN server: ")
	errMalformed      = errors.New("[ERR] malformed STUN request: ")
	errTempServer     = errors.New("[ERR] temporary server error: ")
	errUnexpected     = errors.New("[ERR] unexpected error: ")
	errUnmatchingResp = errors.New("[ERR] unmatching STUN response type")
	errTrID           = errors.New("[ERR] unmatching STUN transactionID")
	errCookie         = errors.New("[ERR] unmatching STUN magic cookie")
)

func StunDial() AddrInfo {
	for {
		if addrinfo := dialingLoop(); addrinfo.isInitialised {
			println("[WARN] " + "STUN sesponse received!")
			return addrinfo
		}
	}
}

func dialingLoop() (info AddrInfo) {
	var conn *net.UDPConn
	transactionBufArr := [12]byte{}
	inputbufArr := [4096]byte{}
	outputBufArr := [20]byte{}

	transactionID := transactionBufArr[0:0]
	inputBuf := inputbufArr[:]
	outputBuf := outputBufArr[0:0]

	outputBuf = append(outputBuf,
		0x00, 0x01, // binding request
		0x00, 0x00, // message length
	)
	outputBuf = append(outputBuf, magicCookie[:]...)

	var err error
	var raddr *net.UDPAddr
	timeNano := time.Now().UnixNano()
	transactionID = binary.BigEndian.AppendUint64(transactionID, uint64(timeNano))
	transactionID = binary.BigEndian.AppendUint32(transactionID, uint32(timeNano))

	outputBuf = append(outputBuf, transactionID...)

	for _, v := range stunlist.ServerList {
	RETRY:
		raddr, err = net.ResolveUDPAddr("udp", v)
		if err != nil {
			println("[WARN] dialing error on server \"", v, "\", switching server. Full report below:")
			println(err.Error(), "\n---\n")
			continue
		}
		conn, err = net.DialUDP("udp", nil, raddr)
		if err != nil {
			println("[WARN] dialing error on server \"", v, "\", switching server")
			println(err.Error(), "\n---\n")

			continue
		}
		println("ssdsds", "\n---\n")
		conn.SetDeadline(time.Now().Add(time.Second * 5))
		if _, err := conn.Write(outputBuf); err != nil {
			println("[WARN] writing timeout on server \"", v, "\", switching server")
			conn.Close()
			continue
		}
		if _, err := conn.Read(inputBuf); err != nil {
			println("[WARN] reading timeout on server \"", v, "\", switching server")
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
		println(errTrID.Error())
		return errTrID, sterr.Alternate, info
	}

	if n := slices.Compare(cookie, magicCookie[:]); n != 0 {
		println(errCookie.Error())
		return errCookie, sterr.Alternate, info
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

	padding := (4 - (argLen % 4)) % 4
	//padding := 4 - ((4 + argLen) % 4)

	info := attrInfo{
		AttrType:   argumentType,
		AttrValue:  argList[4 : argLen+4],
		FullLength: 4 + argLen + padding,
	}
	argList = argList[:argLen+padding]
	return argList, info

}

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
			println(errAlternate.Error(), declineStatus, ", reason: ", reason)
			return errAlternate, sterr.Alternate, addrinfo
		case 400:
			println(errMalformed, declineStatus, ", reason: ", reason)
			return errMalformed, sterr.Other, addrinfo
		case 500:
			println(errTempServer, declineStatus, ", reason: ", reason)
			return errTempServer, sterr.Retry, addrinfo
		default:
			println(errUnexpected, declineStatus, ", reason: ", reason)
			return errUnexpected, sterr.Alternate, addrinfo
		}
	} else {
		println(errUnmatchingResp.Error())
		return errUnmatchingResp, sterr.Other, addrinfo
	}
}

func decodeIP(xorMappedAddr []byte, transactionID []byte, fam byte) (family bool, ipArr [16]byte) {
	if fam == byte(0x02) {
		ipUnmappedArr := [16]byte{}
		ipUnmapped := ipUnmappedArr[:]
		compositeCookieArr := [16]byte{}
		compositeCookie := compositeCookieArr[:]
		compositeCookie = append(compositeCookie, magicCookie[:]...)
		compositeCookie = append(compositeCookie, transactionID...)
		for key, value := range compositeCookie {
			ipUnmapped[key] = xorMappedAddr[key] ^ value
		}
		copy(ipArr[:], ipUnmapped)
		return true, ipArr
	} else if fam == byte(0x01) {
		ipUnmappedArr := [4]byte{}
		ipUnmapped := ipUnmappedArr[:]
		for key, value := range magicCookie {
			ipUnmapped[key] = xorMappedAddr[key] ^ value
		}
		ipArr = [16]byte{}
		ip := ipArr[:]
		copy(ip, net.IPv4(ipUnmapped[0], ipUnmapped[1], ipUnmapped[2], ipUnmapped[3]))
		return false, ipArr
	} else {
		println("[ERR] unrecognized IP format in STUN response")
	}
	return false, [16]byte{}
}
