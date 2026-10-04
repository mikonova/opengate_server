package ntp

import (
	"encoding/binary"
	"log"
	"net"
	"slices"
	"time"

	"github.com/mikonova/OpenGate/errdef"
	"github.com/mikonova/OpenGate/ticker"
)

type NTPResponse struct {
	Header             byte
	Stratum            byte
	Interval           byte
	Precision          byte
	Delay              [4]byte
	Dispersion         [4]byte
	ReferenceID        [4]byte
	ReferenceTimestamp [8]byte
	OriginTimestamp    [8]byte
	ReceiveTimestamp   [8]byte
	TransmitTimestamp  [8]byte
}

var ntpServers = []string{
	"time.google.com",
	"time1.google.com",
	"time2.google.com",
	"time3.google.com",
	"time4.google.com",
	"time.android.com",
	"time.cloudflare.com",
	"time.windows.com",
	"0.pool.ntp.org",
	"1.pool.ntp.org",
	"2.pool.ntp.org",
	"3.pool.ntp.org",
	"ntp0.ntp-servers.net",
	"ntp1.ntp-servers.net",
	"ntp2.ntp-servers.net",
	"ntp3.ntp-servers.net",
	"ntp4.ntp-servers.net",
	"ntp5.ntp-servers.net",
	"ntp6.ntp-servers.net",
	"ntp7.ntp-servers.net",
}

const ntpEpochDifference int64 = 2208988800
const epochSize int64 = 1 << 32

/*
makes a call to NTP time server and returns a precision ticker delayed for the offset between a system clock and an NTP server
The ticker works in a goroutine, yet it is important to use each ticker ONCE as each use consumes a value from its time channel
*/
func GetNtpTime(interval time.Duration) *ticker.PrecisionTicker {
	compareBuf := make([]byte, 48)
	for {
		buf := ntpConnLoop()
		if slices.Equal(buf, compareBuf) {
			continue
		}
		t := parseNTPResp(buf)
		return t

	}

}

func ntpConnLoop() []byte {
	var req = make([]byte, 48)
	req[0] = 0b00011011
	buffer := make([]byte, 48)
	for _, v := range ntpServers {
		conn, err := net.Dial("udp", v+":123")
		if err != nil {
			log.Println(errdef.WarnBase + "dialing error on ntp server \"" + v + "\", switching server")
			continue
		}
		conn.SetDeadline(time.Now().Add(time.Second * 3))

		if _, err = conn.Write(req); err != nil {
			log.Println(errdef.WarnBase + "writing timeout on ntp server \"" + v + "\", switching server")
			conn.Close()
			continue
		}
		if n, err := conn.Read(buffer); err != nil {
			log.Println(errdef.WarnBase + "reading timeout on ntp server \"" + v + "\", switching server")
			conn.Close()
			continue
		} else if n != 48 {
			log.Println(errdef.WarnBase + "response malformed, switching server")
			conn.Close()
			continue
		} else {
			return buffer
		}
	}
	return make([]byte, 48)
}

func parseNTPResp(buffer []byte) *ticker.PrecisionTicker {
	resp := NTPResponse{
		Header:             buffer[0],
		Stratum:            buffer[1],
		Interval:           buffer[2],
		Precision:          buffer[3],
		Delay:              [4]byte(buffer[4:8]),
		Dispersion:         [4]byte(buffer[8:12]),
		ReferenceID:        [4]byte(buffer[12:16]),
		ReferenceTimestamp: [8]byte(buffer[16:24]),
		OriginTimestamp:    [8]byte(buffer[24:32]),
		ReceiveTimestamp:   [8]byte(buffer[32:40]),
		TransmitTimestamp:  [8]byte(buffer[40:]),
	}

	if int16(resp.Stratum) == 0 {
		log.Println(errdef.WarnBase + "STRATUM 0 MESSAGE, PROBABLE KoD")
		log.Println(" Kiss code: ", resp.ReferenceID)
		return &ticker.PrecisionTicker{}
	}
	seconds := binary.BigEndian.Uint32(resp.TransmitTimestamp[:4])
	unixSeconds := int64(seconds) - ntpEpochDifference
	diff := unixSeconds - time.Now().Unix()
	if diff > epochSize/2 {
		unixSeconds -= epochSize
	} else if diff < -epochSize/2 {
		unixSeconds += epochSize
	}
	unixSubSeconds := int64(binary.BigEndian.Uint32(resp.TransmitTimestamp[4:]))
	unixSubSeconds = int64(float64(unixSubSeconds) * 1_000_000_000 / float64(epochSize)) // division to get normal fractions

	ntpTime := time.Unix(unixSeconds, unixSubSeconds)
	pt := ticker.NewPTicker(ntpTime, time.Millisecond)
	return pt
}
