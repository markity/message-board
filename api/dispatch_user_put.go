package api

import (
	"log"
	"message-board/service"
	fieldcheck "message-board/util/field_check"
	"message-board/util/md5"
	"message-board/util/retry"

	"github.com/gin-gonic/gin"
)

// 由于/user PUT设计了两个接口, 用下面的接口统一分发
func DispatchUserPut(ctx *gin.Context) {
	putType, ok := ctx.GetPostForm("put_type")
	if !ok {
		service.RespInvalidParaError(ctx)
		return
	}

	switch putType {
	case "change_password":
		// 分配给ChangePassword
		ChangePassword(ctx)
		return
	case "personal_signature":
		ChangeSignature(ctx)
	default:
		// 不合法的put_type
		service.RespInvalidParaError(ctx)
		return
	}
}

// 修改密码
func ChangePassword(ctx *gin.Context) {
	// 一些基本form-data的获取以及格式检查, 检查完了再查库, 减少数据库的压力
	oldPassword, ok := ctx.GetPostForm("old_password")
	if !ok {
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

	passwordVerify, ok := ctx.GetPostForm("password_verify")
	if !ok {
		service.RespInvalidParaError(ctx)
		return
	}

	if password != passwordVerify {
		service.RespInvalidParaError(ctx)
		return
	}

	var oldPasswordOK bool
	ctxUser_, _ := ctx.Get("user")
	ctxUser := ctxUser_.(*UserAuthInfo)

	// ok, 进行修改密码的数据库操作
	ok = retry.RetryFrame(func() bool {
		ok, err := service.TryChangePassword(ctxUser.UserID, string(md5.ToMD5(oldPassword)),
			string(md5.ToMD5(password)))
		if err != nil {
			log.Printf("failed to TryChangePassword in ChangePassword: %v\n", err)
			return false
		}

		if ok {
			oldPasswordOK = true
		} else {
			oldPasswordOK = false
		}

		// 不再重试
		return true
	}, 3)

	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	if !oldPasswordOK {
		service.RespOldPasswordWrong(ctx)
		return
	}

	service.RespChangePasswordOK(ctx)
}

// 修改签名
func ChangeSignature(ctx *gin.Context) {
	var personalSignature *string
	personalSignature_, ok := ctx.GetPostForm("personal_signature")
	personalSignature = &personalSignature_
	// 允许用户把个性签名删除
	if !ok {
		personalSignature = nil
	} else {
		if personalSignature_ == "" {
			personalSignature = nil
		} else {
			if !fieldcheck.CheckPersonalSignatureValid(personalSignature_) {
				service.RespInvalidParaError(ctx)
				return
			}
		}
	}

	ctxUser_, _ := ctx.Get("user")
	ctxUser := ctxUser_.(*UserAuthInfo)

	// OK, 开始数据库修改的逻辑
	ok = retry.RetryFrame(func() bool {
		ok := service.TryChangeSignature(ctxUser.UserID, personalSignature)
		if !ok {
			return false
		} else {
			return true
		}
	}, 3)

	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	service.RespChangeSignatrueOK(ctx)
}
