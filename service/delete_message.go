package service

import (
	"database/sql"
	"log"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 第一个bool表示是否有错误
// 第二个bool表示是否有这条msgid的条目
// 第三个bool表示是否删除成功
func TryDeleteMessage(tx *sql.Tx, msgid int64, currentUserid int64, admin bool) (bool, bool, bool) {
	// 先上锁, 然后再删除
	row := tx.QueryRow("SELECT sender_user_id  FROM message WHERE id = ? AND deleted = 0 FOR UPDATE", msgid)

	var senderUserID int64
	err := row.Scan(&senderUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, false, false
		}
		log.Printf("failed to QueryRow in TryDeleteMessage: %v", err)
		return false, false, false
	}

	// 没有删除成功, 原因是用户没有权限
	if senderUserID != currentUserid && !admin {
		return true, true, false
	}

	// 已经上锁, 执行删除操作
	_, err = tx.Exec("UPDATE message SET deleted = 1 WHERE id = ?", msgid)
	if err != nil {
		log.Printf("failed to Exec in TryDeleteMessage: %v\n", err)
		return false, false, false
	}

	// OK
	return true, true, true
}

func RespNoSuchMessageEntryToDelete(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorNoSuchMessageEntryToDeleteCode,
		Msg:       errorcodes.ErrorNoSuchMessageEntryToDeleteMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}

func RespDeletedOK(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorOKCode,
		Msg:       errorcodes.ErrorOKMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}

func RespNoDeletePermission(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorNoPermissionToDeleteCode,
		Msg:       errorcodes.ErrorNoPermissionToDeleteMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}
