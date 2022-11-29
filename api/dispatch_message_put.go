package api

import (
	"log"
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
		ThumbUpMessage(ctx)
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
		exist, editOK, err := service.TryEditMessage(int64(msgid), userInfo.UserID, content, anonymous)
		if err != nil {
			log.Printf("failed to TryEditMessage in ChangeMessage: %v\n", err)
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

// put_type = thumb_up
func ThumbUpMessage(ctx *gin.Context) {
	msgidStr := ctx.Param("msgid")
	msgid, err := strconv.ParseUint(msgidStr, 10, 64)
	if err != nil {
		service.RespInvalidParaError(ctx)
		return
	}

	userInfo_, _ := ctx.Get("user")
	userInfo := userInfo_.(*UserAuthInfo)

	var messageEntryExist bool
	var thumbUpOK bool

	retry.RetryFrame(func() bool {
		exist, ok, err := service.TryThumbUpMessage(userInfo.UserID, int64(msgid))
		if err != nil {
			log.Printf("failed to TryThumbUpMessage in ThumbUpMessage: %v\n", err)
			return false
		}

		messageEntryExist = exist
		thumbUpOK = ok
		return true
	}, 3)

	if !messageEntryExist {
		service.RespNoSuchMessageEntryToThumbUp(ctx)
		return
	}

	if !thumbUpOK {
		service.RespYouAlreadyLikedIt(ctx)
		return
	}

	service.RespThumbUpMessageOK(ctx)
}
