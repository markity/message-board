package service

import (
	"database/sql"
	"log"
	boolconvert "message-board/util/bool_convert"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 第一个bool表示是否有错误, 比如网络错误等
// 第二个bool表示是否有这条msgid的条目
// 第三个bool表示是否删除成功(权限问题)
func TryEditMessage(tx *sql.Tx, msgid int64, currentUserid int64,
	content string, anonymous bool) (bool, bool, bool) {
	anonymousInt := boolconvert.BoolToInt(anonymous)

	row := tx.QueryRow("SELECT sender_user_id  FROM message WHERE id = ? AND deleted = 0 FOR UPDATE", msgid)

	var senderUserID int64
	err := row.Scan(&senderUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, false, false
		}
		log.Printf("failed to QueryRow in TryEditMessage: %v", err)
		return false, false, false
	}

	// 没有修改成功, 原因是用户没有权限
	if senderUserID != currentUserid {
		return true, true, false
	}

	// 已经上锁, 执行删除操作
	_, err = tx.Exec("UPDATE message SET content = ?,anonymous = ? WHERE id = ?", content, anonymousInt, msgid)
	if err != nil {
		log.Printf("failed to Exec in TryEditMessage: %v\n", err)
		return false, false, false
	}

	return true, true, true
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
