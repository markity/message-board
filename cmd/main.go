package main

import (
	"message-board/api"
	"message-board/service"

	"github.com/gin-gonic/gin"
)

func main() {
	service.MustResetTables()
	engine := gin.Default()
	api.InitGroup(engine)

	engine.Run("127.0.0.1:8000")
}
