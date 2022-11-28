package api

import (
	"message-board/service"
	fieldcheck "message-board/util/field_check"
	"message-board/util/md5"
	"message-board/util/retry"
	"time"

	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	// 进行基本检查
	username, ok := ctx.GetPostForm("username")
	if !ok {
		service.InvalidParaError(ctx)
		return
	}
	if !fieldcheck.CheckUsernameValid(username) {
		service.InvalidParaError(ctx)
		return
	}
	password, ok := ctx.GetPostForm("password")
	if !ok {
		service.InvalidParaError(ctx)
		return
	}
	if !fieldcheck.CheckPasswordValid(password) {
		service.InvalidParaError(ctx)
		return
	}

	var loginOK bool
	var userID int64
	var userAdmin bool

	ok = retry.RetryFrame(func() bool {
		exist, id, admin, err := service.TryCheckLoginInfo(username, md5.ToMD5(password))
		if err != nil {
			// 重试
			return false
		}
		loginOK = exist
		userID = id
		userAdmin = admin

		return true
	}, 3)

	if !ok {
		service.ServiceNotAvailabelError(ctx)
		return
	}

	if !loginOK {
		service.UserLoginInfoWrong(ctx)
		return
	}

	// 登录成功, 签发jwt
	ctx.SetCookie("authtoken", JwtSignaturer.Signature(userID, userAdmin, time.Hour*2),
		7200, "/", "localhost", false, false)

	service.UserLoginSuccess(ctx)
}
