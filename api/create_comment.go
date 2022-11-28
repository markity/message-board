package api

import (
	"message-board/service"
	errorcodes "message-board/util/error_codes"
	"message-board/util/retry"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateComment(ctx *gin.Context) {
	msgidStr := ctx.Param("msgid")
	msgid, err := strconv.ParseUint(msgidStr, 10, 64)
	if err != nil {
		service.InvalidParaError(ctx)
		return
	}

	content, ok := ctx.GetPostForm("content")
	if !ok {
		service.InvalidParaError(ctx)
		return
	}

	var anonymous bool
	anonymousStr, ok := ctx.GetPostForm("anonymous")
	if !ok {
		anonymous = false
	} else {
		// 只接收false或true, 其它的值视为非法
		if anonymousStr == "true" {
			anonymous = true
		} else if anonymousStr == "false" {
			anonymous = false
		} else {
			service.InvalidParaError(ctx)
			return
		}
	}

	userInfo_, _ := ctx.Get("user")
	userInfo := userInfo_.(*UserAuthInfo)
	var parentEntryExist bool
	var lastInsertedCommentID int64

	ok = retry.RetryFrame(func() bool {
		tx, ok := service.NewTX()
		if !ok {
			// 开启事务开始, 重试
			return false
		}
		queryOK, exist, lastInserted := service.TryCreateComment(tx, int64(msgid), userInfo.UserID, content, anonymous, time.Now())
		if !queryOK {
			// 一些意料之外的错误, 选择重试
			tx.Rollback()
			return false
		}

		// 提交事务失败也重试
		err := tx.Commit()
		if err != nil {
			return false
		}

		// 成功
		parentEntryExist = exist
		lastInsertedCommentID = lastInserted
		return true
	}, 3)

	if !ok {
		service.ServiceNotAvailabelError(ctx)
		return
	}

	// 没有这条父消息
	if !parentEntryExist {
		service.NoSuchParentComment(ctx)
		return
	}

	// 成功发布
	resp := service.MessageInsertedResp{
		BasicErrorResp: errorcodes.BasicErrorResp{
			ErrorCode: errorcodes.ErrorOKCode,
			Msg:       errorcodes.ErrorOKMsg,
		},
		MessageID: lastInsertedCommentID,
	}
	ctx.JSON(http.StatusOK, &resp)
}