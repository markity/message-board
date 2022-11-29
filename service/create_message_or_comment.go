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

// error代表意料之外的错误, 比如网络错误等
// 当error != nil 时, 前两个返回值没有意义, 当error == nil 但 bool == false时, int64返回值没有意义
// 第一个bool代表是否存在父评论, 如果不存在返回false, 此时当然插入失败, 第二个参数为0(没有意义)
// 如果bool值返回的true, 则代表存在父评论且插入成功, 此时第二个参数是插入对象的id
//		(也就是说只有第一个返回值为true时, 第二个参数才有意义)
func TryCreateComment(parentCommmentID int64, senderID int64, sonContent string,
	annoymous bool, createdAt time.Time) (bool, int64, error) {
	tx, err := dao.DB.Begin()
	if err != nil {
		log.Printf("failed to Begin in TryCreateComment: %v\n", err)
		return false, 0, err
	}

	// 先锁住父评论, 然后再插入加子评论
	row := tx.QueryRow("SELECT id FROM message WHERE id = ? FOR UPDATE", parentCommmentID)
	var discard int64
	err = row.Scan(&discard)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return false, 0, nil
		}
		log.Printf("failed to QueryRow in TryCreateComment: %v\n", err)
		return false, 0, err
	}

	// 存在该评论, 执行插入
	result, err := tx.Exec("INSERT INTO message(content, sender_user_id, parent_message_id, created_at, anonymous) VALUES(?,?,?,?,?)",
		sonContent, senderID, parentCommmentID, timeconvert.TimeToStr(createdAt), annoymous)
	if err != nil {
		log.Printf("failed to Exec in TryCreateComment: %v\n", err)
		tx.Rollback()
		return false, 0, err
	}
	lastInserted, _ := result.LastInsertId()

	err = tx.Commit()
	if err != nil {
		log.Printf("failed to Commit in TryCreateComment: %v\n", err)
		return false, lastInserted, err
	}

	return true, lastInserted, nil
}

func RespNoSuchParentComment(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorNoSuchMessageEntryToCommentCode,
		Msg:       errorcodes.ErrorNoSuchMessageEntryToCommentMsg,
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
