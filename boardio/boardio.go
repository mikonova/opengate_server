package boardio

import (
	"machine"
	"runtime"
)

func GetInput() string {
	buffer := [96]byte{}
	buf := buffer[0:0]
	println("[INFO] waiting input!")
	for {
		if machine.Serial.Buffered() == 0 {
			runtime.Gosched()
			continue
		}
		data, err := machine.Serial.ReadByte()
		if err != nil {
			break
		}
		if rune(data) == '\n' || rune(data) == '\r' || rune(data) == '\t' {
			break
		}
		buf = append(buf, data)
		runtime.Gosched()
	}
	output := string(buf)
	return output
}
