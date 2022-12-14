package api

import "github.com/gin-gonic/gin"

func InitGroup(engine *gin.Engine) {
	engine.POST("/user", CreateUser)
	engine.PUT("/user", MiddleWareJWTVerify, DispatchUserPut)
	engine.POST("/user/auth", Login)
	engine.DELETE("/user/auth", MiddleWareJWTVerify, Logout)
	engine.GET("/user/info/:username", GetUserinfo)
	engine.POST("/message", MiddleWareJWTVerify, CreateMessage)
	engine.POST("/message/:msgid", MiddleWareJWTVerify, CreateComment)
	engine.DELETE("/message/:msgid", MiddleWareJWTVerify, DeleteMessage)
	engine.PUT("/message/:msgid", MiddleWareJWTVerify, DispatchMessagePut)
	engine.GET("/message/:msgid", GetSingleMessage)
	engine.GET("/message", GetMultipleMessages)
}
