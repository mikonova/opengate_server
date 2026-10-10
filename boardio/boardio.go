package boardio

import (
	"machine"
)

func GetInput() (output string) {
	for {
		data, err := machine.Serial.ReadByte()
		if err != nil {
			continue
		}
		if data == '\n' {
			break
		} else {
			output += string(data)
		}
	}
	return output
}
