//go:build tests

package main

import (
	"machine"
	"time"
)

func main() {
	if err := machine.USBCDC.Configure(machine.UARTConfig{}); err != nil {
		println("CDC configure failed:", err.Error())
		return
	}

	println("введите что-то")

	for {
		b, err := machine.USBCDC.ReadByte()
		if err == nil {
			println("получен байт:", b)
			break
		}
		time.Sleep(time.Millisecond)
	}

}
