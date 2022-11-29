package api

import (
	"log"
	"message-board/service"
	fieldcheck "message-board/util/field_check"
	"message-board/util/retry"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 获取某条消息本体以及子消息
func GetSingleMessage(ctx *gin.Context) {
	msgidStr := ctx.Param("msgid")
	msgid, err := strconv.ParseUint(msgidStr, 10, 64)
	if err != nil {
		service.RespInvalidParaError(ctx)
		return
	}

	var msgResp *service.Message
	ok := retry.RetryFrame(func() bool {
		msg, err := service.TryGetSingleMessage(int64(msgid))
		if err != nil {
			log.Printf("failed to TryGetSingleMessage in GetSingleMessage: %v\n", err)
			return false
		}

		msgResp = msg

		return true
	}, 3)

	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	if msgResp == nil {
		service.RespNoSuchMessageEntryToFetch(ctx)
		return
	}

	service.RespGetSingleMessageOK(ctx, msgResp)
}

func GetMultipleMessages(ctx *gin.Context) {
	// 获取基本参数, 以及一些基本检查

	var entryNum int64

	entryNumStr, ok := ctx.GetPostForm("entry_num")
	if !ok {
		// 没有此参数的时候, 默认值为10
		entryNum = 20
	}
	entryNum, err := strconv.ParseInt(entryNumStr, 10, 32)
	if err != nil {
		service.RespInvalidParaError(ctx)
		return
	}
	if !fieldcheck.CheckEntryNumValid(entryNum) {
		service.RespInvalidParaError(ctx)
		return
	}

	pageNumStr, ok := ctx.GetPostForm("page_num")
	if !ok {
		service.RespInvalidParaError(ctx)
		return
	}
	pageNum, err := strconv.ParseInt(pageNumStr, 10, 64)
	if err != nil {
		service.RespInvalidParaError(ctx)
		return
	}
	if !fieldcheck.CheckPageNumValid(pageNum) {
		service.RespInvalidParaError(ctx)
		return
	}

	var messages [](*service.Message)

	// 开始执行查询
	ok = retry.RetryFrame(func() bool {
		msgs, err := service.TryGetMultipleMessages(entryNum, pageNum)
		if err != nil {
			log.Printf("failed to TryGetMultipleMessages in GetMultipleMessages: %v\n", err)
			return false
		}

		messages = msgs
		return true
	}, 3)

	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	// 成功
	service.RespGetMessagesOK(ctx, messages)
}
