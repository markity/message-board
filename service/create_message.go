package service

import (
	"log"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	timeconvert "message-board/util/time_convert"
	"time"
)

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

type MessageInsertedResp struct {
	errorcodes.BasicErrorResp
	MessageID int64 `json:"message_id"`
}
