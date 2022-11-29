package api

import (
	"log"
	"message-board/service"
	"message-board/util/retry"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateComment(ctx *gin.Context) {
	msgidStr := ctx.Param("msgid")
	msgid, err := strconv.ParseUint(msgidStr, 10, 64)
	if err != nil {
		service.RespInvalidParaError(ctx)
		return
	}

	content, ok := ctx.GetPostForm("content")
	if !ok {
		service.RespInvalidParaError(ctx)
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
			service.RespInvalidParaError(ctx)
			return
		}
	}

	userInfo_, _ := ctx.Get("user")
	userInfo := userInfo_.(*UserAuthInfo)
	var parentEntryExist bool
	var lastInsertedCommentID int64

	ok = retry.RetryFrame(func() bool {
		exist, lastInserted, err := service.TryCreateComment(int64(msgid), userInfo.UserID, content, anonymous, time.Now())
		if err != nil {
			// 一些意料之外的错误, 选择重试
			log.Printf("failed to TryCreateComment: %v\n", err)
			return false
		}

		// 成功
		parentEntryExist = exist
		lastInsertedCommentID = lastInserted
		return true
	}, 3)

	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	// 没有这条父消息
	if !parentEntryExist {
		service.RespNoSuchParentComment(ctx)
		return
	}

	// 成功发布
	service.RespCreateMessageOrCommentOK(ctx, lastInsertedCommentID)
}
