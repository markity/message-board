package main

import (
	"message-board/service"
	"message-board/util/jwt"
	"time"
)

func main() {
	jwtSignaturer := jwt.NewUserJWTSignaturer(jwt.NewRsaSHA256Cryptor())
	sign := jwtSignaturer.Signature(10, true, time.Second)
	// time.Sleep(time.Second * 3)
	println(jwtSignaturer.Check(sign))

	service.MustPrepareTables()
}
