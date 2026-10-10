//go:build tests

package main

import (
	"fmt"
	"relayesp/protocol/stun"
)

func main() {
	test := stun.StunDial()
	fmt.Println(test.Address)

}
