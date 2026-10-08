//go:build tests

package main

import (
	"relayesp/protocol/stun"
)

func main() {
	_ = stun.StunDial()
}
