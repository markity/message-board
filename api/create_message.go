package api

import (
	"message-board/service"
	errorcodes "message-board/util/error_codes"
	fieldcheck "message-board/util/field_check"
	"message-board/util/retry"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateMessage(ctx *gin.Context) {
	// 基本检查
	content, ok := ctx.GetPostForm("content")
	if !ok {
		service.InvalidParaError(ctx)
		return
	}
	if !fieldcheck.CheckMessageValid(content) {
		service.InvalidParaError(ctx)
		return
	}

	var anonymous bool
	anoymousStr, ok := ctx.GetPostForm("anonymous")
	if !ok {
		anonymous = false
	} else {
		// 只接收false或true, 其它的值视为非法
		if anoymousStr == "true" {
			anonymous = true
		} else if anoymousStr == "false" {
			anonymous = false
		} else {
			service.InvalidParaError(ctx)
			return
		}
	}

	userInfo_, _ := ctx.Get("user")
	userInfo := userInfo_.(*UserAuthInfo)

	var insertedMessageID int64

	// 执行数据库插入
	ok = retry.RetryFrame(func() bool {
		ok, lastInserted := service.TryCreateTopMessage(userInfo.UserID, content, anonymous)
		if !ok {
			return false
		}

		insertedMessageID = lastInserted
		return true
	}, 3)

	if !ok {
		service.ServiceNotAvailabelError(ctx)
		return
	}

	resp := service.MessageInsertedResp{
		BasicErrorResp: errorcodes.BasicErrorResp{
			ErrorCode: errorcodes.ErrorOKCode,
			Msg:       errorcodes.ErrorOKMsg,
		},
		MessageID: insertedMessageID,
	}

	ctx.JSON(http.StatusOK, &resp)
}
