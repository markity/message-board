package api

import (
	"message-board/service"
	"time"

	"github.com/gin-gonic/gin"

	fieldcheck "message-board/util/field_check"
	"message-board/util/md5"
	"message-board/util/retry"
)

func CreateUser(ctx *gin.Context) {
	username, ok := ctx.GetPostForm("username")
	if !ok {
		service.RespInvalidParaError(ctx)
		return
	}
	password, ok := ctx.GetPostForm("password")
	if !ok {
		service.RespInvalidParaError(ctx)
		return
	}

	var personalSignature *string
	personalSignature_, ok := ctx.GetPostForm("personal_signature")
	if !ok {
		personalSignature = nil
	} else {
		if !fieldcheck.CheckPersonalSignatureValid(personalSignature_) {
			service.RespInvalidParaError(ctx)
			return
		}
		if personalSignature_ == "" {
			personalSignature = nil
		} else {
			personalSignature = &personalSignature_
		}
	}

	if !fieldcheck.CheckUsernameValid(username) || !fieldcheck.CheckPasswordValid(password) {
		service.RespInvalidParaError(ctx)
		return
	}

	// 标识该用户名是否已被占用
	isAlreadyRegistered := false

	// 开始尝试插入, 最多重试3次, 如果成功, 外部ok变量为true
	ok = retry.RetryFrame(func() bool {
		err, insertOK := service.TryCreateUser(username, md5.ToMD5(password), personalSignature, time.Now(), false)
		if err != nil {
			// failed, do retry
			return false
		}

		// 数据库操作成功, 更新外部变量
		if insertOK {
			isAlreadyRegistered = false
		} else {
			isAlreadyRegistered = true
		}

		// 不再重试
		return true
	}, 3)

	// 尝试三次均失败, 告知用户目前服务不可用
	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	if isAlreadyRegistered {
		// 告知用户该用户名已被占用
		service.RespUsernameOccupied(ctx)
		return
	}

	service.RespCreateUserOK(ctx)
}
