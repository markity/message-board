package service

import (
	"database/sql"
	"log"
	"message-board/dao"
	boolconvert "message-board/util/bool_convert"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// error != nil时其它返回值没有意义
// 第一个返回值代表是否有该条目
// 第二个返回值代表是否有权限更改此条目(如果第一个返回值为false, 此返回值无意义)
func TryEditMessage(msgid int64, currentUserid int64, content string,
	anonymous bool) (bool, bool, error) {
	tx, err := dao.DB.Begin()
	if err != nil {
		log.Printf("failed to Begin in TryEditMessage: %v\n", err)
		return false, false, err
	}

	anonymousInt := boolconvert.BoolToInt(anonymous)

	row := tx.QueryRow("SELECT sender_user_id  FROM message WHERE id = ? AND deleted = 0 FOR UPDATE", msgid)

	var senderUserID int64
	err = row.Scan(&senderUserID)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return false, false, nil
		}
		log.Printf("failed to QueryRow in TryEditMessage: %v", err)
		return false, false, err
	}

	// 没有修改成功, 原因是用户没有权限
	if senderUserID != currentUserid {
		tx.Rollback()
		return true, false, nil
	}

	// 已经上锁, 执行删除操作
	_, err = tx.Exec("UPDATE message SET content = ?,anonymous = ? WHERE id = ?", content, anonymousInt, msgid)
	if err != nil {
		tx.Rollback()
		log.Printf("failed to Exec in TryEditMessage: %v\n", err)
		return false, false, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("failed to Commit in TryEditMessage: %v\n", err)
		return false, false, err
	}

	return true, true, nil
}

func RespNoSuchMessageEntryToEdit(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorNoSuchMessageEntryToEditCode,
		Msg:       errorcodes.ErrorNoSuchMessageEntryToEditMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}

func RespNoPermissionToEdit(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorNoPermissionToEditCode,
		Msg:       errorcodes.ErrorNoPermissionToEditMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}

func RespEditMessageEntryOK(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorOKCode,
		Msg:       errorcodes.ErrorOKMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}
