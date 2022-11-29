package api

import (
	"log"
	"message-board/service"
	"message-board/util/retry"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DeleteMessage(ctx *gin.Context) {
	msgidStr := ctx.Param("msgid")
	msgid, err := strconv.ParseUint(msgidStr, 10, 64)
	if err != nil {
		service.RespInvalidParaError(ctx)
		return
	}

	userInfo_, _ := ctx.Get("user")
	userInfo := userInfo_.(*UserAuthInfo)
	var entryExist bool
	var hasPermission bool

	ok := retry.RetryFrame(func() bool {
		tx, ok := service.NewTX()
		if !ok {
			return false
		}

		// 当条目存在确ok == false时, 原因是没有权限
		queryOK, exist, ok := service.TryDeleteMessage(tx, int64(msgid), userInfo.UserID, userInfo.Admin)
		if !queryOK {
			tx.Rollback()
			return false
		}

		err := tx.Commit()
		if err != nil {
			log.Printf("failed to commit: %v\n", err)
			return false
		}

		entryExist = exist
		hasPermission = ok
		return true
	}, 3)

	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	if !entryExist {
		service.RespNoSuchMessageEntryToDelete(ctx)
		return
	}

	if !hasPermission {
		service.RespNoDeletePermission(ctx)
		return
	}

	service.RespDeletedOK(ctx)
}
