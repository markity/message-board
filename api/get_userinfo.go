package api

import (
	"message-board/service"
	"message-board/util/retry"

	"github.com/gin-gonic/gin"
)

func GetUserinfo(ctx *gin.Context) {
	username := ctx.Param("username")

	var exist bool
	var userInfo *service.UserInfo

	ok := retry.RetryFrame(func() bool {
		ok, ui := service.TryGetUserinfo(username)
		if !ok {
			return false
		}

		if ui == nil {
			// 不存在此用户
			exist = false
		} else {
			exist = true
			userInfo = ui
		}

		return true
	}, 3)

	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	if !exist {
		service.RespNoSuchUser(ctx)
		return
	}

	service.RespGetUserinfoOK(ctx, userInfo)
}
