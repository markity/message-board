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
		service.RespInvalidParaError(ctx)
		return
	}
	if !fieldcheck.CheckUsernameValid(username) {
		service.RespInvalidParaError(ctx)
		return
	}
	password, ok := ctx.GetPostForm("password")
	if !ok {
		service.RespInvalidParaError(ctx)
		return
	}
	if !fieldcheck.CheckPasswordValid(password) {
		service.RespInvalidParaError(ctx)
		return
	}

	var loginOK bool
	var userID int64
	var userAdmin bool

	ok = retry.RetryFrame(func() bool {
		// exist代表是否查询到 username = xxx, password = xxx的条目
		// 如果查到, 就代表密码正确, 否则就返回用户账户或密码不正确
		queryOK, passwordOK, id, admin := service.TryCheckLoginInfo(username, md5.ToMD5(password))
		if !queryOK {
			// 重试
			return false
		}
		loginOK = passwordOK
		userID = id
		userAdmin = admin

		return true
	}, 3)

	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	if !loginOK {
		service.RespUserLoginInfoWrong(ctx)
		return
	}

	// 登录成功, 签发jwt
	jwt := JwtSignaturer.Signature(userID, userAdmin, time.Hour*2)
	ctx.SetCookie("authtoken", jwt, 7200, "/", "localhost", false, false)

	service.RespUserLoginOK(ctx)
}
