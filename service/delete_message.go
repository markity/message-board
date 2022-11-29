package service

import (
	"database/sql"
	"log"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// err != nil 时(有可能是网络错误), 其它返回值没有意义
// 第一个返回值代表是否存在该消息
// 第二个返回值代表是否有删除权限, 是否成功删除(第一个返回值为false的之后这个返回值没有意义)
func TryDeleteMessage(msgid int64, currentUserid int64, admin bool) (bool, bool, error) {
	tx, err := dao.DB.Begin()
	if err != nil {
		log.Printf("failed to Begin in TryDeleteMessage: %v\n", err)
		return false, false, err
	}

	// 先上锁, 然后再删除
	row := tx.QueryRow("SELECT sender_user_id  FROM message WHERE id = ? AND deleted = 0 FOR UPDATE", msgid)

	var senderUserID int64
	err = row.Scan(&senderUserID)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return false, false, nil
		}
		log.Printf("failed to QueryRow in TryDeleteMessage: %v", err)
		return false, false, err
	}

	// 没有删除成功, 原因是用户没有权限
	if senderUserID != currentUserid && !admin {
		tx.Rollback()
		return true, true, nil
	}

	// 已经上锁, 执行删除操作
	_, err = tx.Exec("UPDATE message SET deleted = 1 WHERE id = ?", msgid)
	if err != nil {
		tx.Rollback()
		log.Printf("failed to Exec in TryDeleteMessage: %v\n", err)
		return false, false, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("failed to Commit in TryDeleteMessage: %v\n", err)
		return false, false, err
	}

	// OK
	return true, true, nil
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
