//go:build tests

package main

import (
	"time"
)

func main() {
	time.Sleep(time.Second * 5)
	println("i work")
	for {
		time.Sleep(time.Second)
	}

}
