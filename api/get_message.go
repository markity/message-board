package api

import (
	"log"
	"message-board/service"
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

	service.RespGetMessageOK(ctx, msgResp)
}

func GetMultipleMessages(ctx *gin.Context) {

}
