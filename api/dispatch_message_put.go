package api

import (
	"message-board/service"
	fieldcheck "message-board/util/field_check"
	"message-board/util/retry"
	"strconv"

	"github.com/gin-gonic/gin"
)

func DispatchMessagePut(ctx *gin.Context) {
	putType, ok := ctx.GetPostForm("put_type")
	if !ok {
		service.RespInvalidParaError(ctx)
		return
	}

	switch putType {
	case "edit":
		ChangeMessage(ctx)
		return
	case "thumb_up":
		// ThumbUpMessage(ctx)
	default:
		service.RespInvalidParaError(ctx)
		return
	}
}

// put_type = edit
func ChangeMessage(ctx *gin.Context) {
	msgidStr := ctx.Param("msgid")
	msgid, err := strconv.ParseUint(msgidStr, 10, 64)
	if err != nil {
		service.RespInvalidParaError(ctx)
		return
	}

	content, ok := ctx.GetPostForm("content")
	if !ok {
		service.RespInvalidParaError(ctx)
	}
	if !fieldcheck.CheckMessageValid(content) {
		service.RespInvalidParaError(ctx)
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
			service.RespInvalidParaError(ctx)
			return
		}
	}

	userInfo_, _ := ctx.Get("user")
	userInfo := userInfo_.(*UserAuthInfo)

	var entryExist bool
	var hasPermission bool

	ok = retry.RetryFrame(func() bool {
		tx, ok := service.NewTX()
		if !ok {
			return false
		}
		queryOK, exist, editOK := service.TryEditMessage(tx, int64(msgid), userInfo.UserID, content, anonymous)
		if !queryOK {
			tx.Rollback()
			return false
		}

		err := tx.Commit()
		if err != nil {
			return false
		}

		entryExist = exist
		hasPermission = editOK
		return true
	}, 3)

	if !ok {
		service.RespServiceNotAvailabelError(ctx)
		return
	}

	if !entryExist {
		service.RespNoSuchMessageEntryToEdit(ctx)
		return
	}

	if !hasPermission {
		service.RespNoPermissionToEdit(ctx)
		return
	}

	service.RespEditMessageEntryOK(ctx)
}

// // put_type = thumb_up
// func ThumbUpMessage(ctx *gin.Context) {
// 	msgidStr := ctx.Param("msgid")
// 	msgid, err := strconv.ParseUint(msgidStr, 10, 64)
// 	if err != nil {
// 		service.InvalidParaError(ctx)
// 		return
// 	}

// }
