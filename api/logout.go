package api

import (
	"message-board/service"

	"github.com/gin-gonic/gin"
)

func Logout(ctx *gin.Context) {
	ctx.SetCookie("authtoken", "", -1, "/", "localhost", false, false)
	service.RespUserLogoutOK(ctx)
}
