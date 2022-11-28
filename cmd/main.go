package main

import (
	"message-board/api"

	"github.com/gin-gonic/gin"
)

func main() {
	// service.MustPrepareTables()
	engine := gin.Default()
	api.InitGroup(engine)

	engine.Run("127.0.0.1:8000")
}
