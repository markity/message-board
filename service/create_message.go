package service

import (
	"database/sql"
	"log"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	timeconvert "message-board/util/time_convert"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 不导出这个结构体
type messageInsertedResp struct {
	errorcodes.BasicErrorResp
	MessageID int64 `json:"message_id"`
}

// 如果插入成功, 则返回true, id
// 否则返回nil, 0
func TryCreateTopMessage(userid int64, content string, anonymous bool) (bool, int64) {
	var anonymousInt int
	if anonymous {
		anonymousInt = 1
	} else {
		anonymousInt = 0
	}

	res, err := dao.DB.Exec("INSERT INTO message(content,sender_user_id,anonymous,created_at) VALUES(?,?,?,?)",
		content, userid, anonymousInt, timeconvert.TimeToStr(time.Now()))
	if err != nil {
		log.Printf("failed to Exec in TryCreateTopMessage: %v\n", err)
		return false, 0
	}

	lastInserted, _ := res.LastInsertId()

	// 插入成功
	return true, lastInserted
}

// 先锁住父评论, 然后插入子评论
// 第一个bool指示是否重试, 第二个bool指示操作是否成功
// 如果parentID的父评论不存在, 则返回true, false
// 如果插入成功, 则返回true, true
// 出现任何其它错误, 如网络错误, 返回false, false, 上层应回滚事务
func TryCreateComment(tx *sql.Tx, parentID int64, senderID int64,
	sonContent string, annoymous bool, createdAt time.Time) (bool, bool, int64) {
	// 先锁住父评论, 然后再插入加子评论
	row := tx.QueryRow("SELECT id FROM message WHERE id = ? FOR UPDATE", parentID)
	var discard int64
	err := row.Scan(&discard)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, false, 0
		}
		log.Printf("failed to QueryRow in TryCreateComment: %v\n", err)
		return false, false, 0
	}

	// 存在该评论, 执行插入
	result, err := tx.Exec("INSERT INTO message(content, sender_user_id, parent_message_id, created_at, anonymous) VALUES(?,?,?,?,?)",
		sonContent, senderID, parentID, timeconvert.TimeToStr(createdAt), annoymous)
	if err != nil {
		log.Printf("failed to Exec in TryCreateComment: %v\n", err)
		return false, false, 0
	}
	lastInserted, _ := result.LastInsertId()

	return true, true, lastInserted
}

func RespNoSuchParentComment(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorOKCode,
		Msg:       errorcodes.ErrorOKMsg,
	}
	ctx.JSON(http.StatusOK, &resp)
}

// -----------------------------
func RespCreateMessageOrCommentOK(ctx *gin.Context, insertedMessageID int64) {
	resp := messageInsertedResp{
		BasicErrorResp: errorcodes.BasicErrorResp{
			ErrorCode: errorcodes.ErrorOKCode,
			Msg:       errorcodes.ErrorOKMsg,
		},
		MessageID: insertedMessageID,
	}

	ctx.JSON(http.StatusOK, &resp)
}
