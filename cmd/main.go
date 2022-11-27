package main

import (
	"message-board/util"
	"time"
)

func main() {
	jwtSignaturer := util.NewUserJWTSignaturer(util.NewRsaSHA256Cryptor())
	sign := jwtSignaturer.Signature(10, true, time.Second)
	// time.Sleep(time.Second * 3)
	println(jwtSignaturer.Check(sign))
}
