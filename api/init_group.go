package api

import "github.com/gin-gonic/gin"

func InitGroup(engine *gin.Engine) {
	engine.POST("/user", CreateUser)
}
