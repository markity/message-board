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

type Message struct {
	// Anonymous == true时忽略SenderUsername
	MessageID      int64        `json:"message_id"`
	MessageContent string       `json:"message_content"`
	SenderUsername *string      `json:"sender_user_name,omitempty"`
	CreatedAtStr   string       `json:"created_at"`
	ThumbsUp       int64        `json:"thumps_up"`
	Anonymous      bool         `json:"anonymous"`
	SonMessages    [](*Message) `json:"son_messages"`
}

// 如果第一个返回值为nil, 则代表没有此msgid的条目
func TryGetSingleMessage(msgid int64) (*Message, error) {
	tx, err := dao.DB.Begin()
	if err != nil {
		log.Printf("failed to Begin in TryGetSingleMessage: %v\n", err)
		return nil, err
	}

	var messageContent string
	var messageThumbsUp int64
	var createdAtStr string
	var anonymousInt int
	var senderUsername string
	// 开启事务, 有一致性读, 不用担心一致性问题, 直接逐层查询
	query := `
	SELECT 	message.content, message.thumbs_up, message.created_at, message.anonymous,
			user.username FROM message
	LEFT JOIN user ON message.sender_user_id = user.id
	WHERE message.deleted = 0 AND message.id = ?
	`
	row := tx.QueryRow(query, msgid)
	err = row.Scan(&messageContent, &messageThumbsUp, &createdAtStr, &anonymousInt, &senderUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			tx.Commit()
			return nil, nil
		}

		// 查第一条就出现了错误, 直接告知上层重试
		tx.Rollback()
		log.Printf("failed to Scan in TryGetSingleMessage: %v\n", err)
		return nil, err
	}

	msg := Message{
		MessageID:      msgid,
		MessageContent: messageContent,
		SenderUsername: &senderUsername,
		CreatedAtStr:   createdAtStr,
		ThumbsUp:       messageThumbsUp,
		Anonymous:      boolconvert.MustItob(anonymousInt),
		SonMessages:    nil,
	}

	// 当为匿名的时候, 隐藏sender_user_name字段
	if msg.Anonymous {
		msg.SenderUsername = nil
	}

	// 开始扫子层
	err = toolScanSonComments(tx, &msg)
	if err != nil {
		tx.Rollback()
		log.Printf("failed to toolScanSonComments in TryGetSingleMessage: %v\n", err)
		return nil, err
	}

	// 因为东西已经查完了, 不关心事务提交是否出错了, 因此这里不检查错误
	tx.Commit()

	return &msg, nil
}

// 扫子层评论的工具函数, 此处递归, 为保证整洁, 不打印报错信息
func toolScanSonComments(tx *sql.Tx, msg *Message) error {

	var msgID int64
	var messageContent string
	var senderUserID int64
	var messageThumbsUp int64
	var createdAtStr string
	var anonymousInt int
	var senderUsername string

	sonComments := make([](*Message), 0)

	query := `
	SELECT 	message.id, message.content, message.sender_user_id, message.thumbs_up,
	 		message.created_at, message.anonymous, user.username FROM message
	LEFT JOIN user ON message.sender_user_id = user.id
	WHERE message.deleted = 0 AND message.parent_message_id = ?
	`

	rows, err := tx.Query(query, msg.MessageID)
	if err != nil {
		return err
	}

	cnt := 0
	for rows.Next() {
		// TODO 是否应当忽略这个错误
		_ = rows.Scan(&msgID, &messageContent, &senderUserID, &messageThumbsUp, &createdAtStr,
			&anonymousInt, &senderUsername)
		newMsg := Message{
			MessageID:      msgID,
			MessageContent: messageContent,
			SenderUsername: &senderUsername,
			CreatedAtStr:   createdAtStr,
			ThumbsUp:       messageThumbsUp,
			Anonymous:      boolconvert.MustItob(anonymousInt),
			SonMessages:    nil,
		}

		// 匿名时隐藏sender_user_name
		if newMsg.Anonymous {
			newMsg.SenderUsername = nil
		}

		sonComments = append(sonComments, &newMsg)
		cnt++
	}

	// 用计数器优化, 比len快
	if cnt == 0 {
		msg.SonMessages = nil
	} else {
		msg.SonMessages = sonComments
	}

	// 进入递归, 开始扫更子层的东西
	for _, v := range sonComments {
		err := toolScanSonComments(tx, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func TryGetMultipleMessages(entryNum int64, pageNum int64) ([](*Message), error) {
	var msgID int64
	var messageContent string
	var senderUserID int64
	var messageThumbsUp int64
	var createdAtStr string
	var anonymousInt int
	var senderUsername string

	tx, err := dao.DB.Begin()
	if err != nil {
		log.Printf("failed to Begin in TryGetMultipleMessages: %v\n", err)
		return nil, err
	}

	query := `
	SELECT 	message.id, message.content, message.sender_user_id, message.thumbs_up,
			message.created_at, message.anonymous, user.username FROM message
	LEFT JOIN user ON message.sender_user_id = user.id
	WHERE 	message.deleted = 0 
			AND
			message.parent_message_id IS NULL
	ORDER BY message.id DESC
	LIMIT ?,?
	`
	rows, err := tx.Query(query, (pageNum-1)*entryNum, entryNum)
	if err != nil {
		tx.Rollback()
		log.Printf("failed to Query in TryGetMultipleMessages: %v\n", err)
		return nil, err
	}

	messages := make([](*Message), 0)

	for rows.Next() {
		// TODO 是否忽略该错误
		_ = rows.Scan(&msgID, &messageContent, &senderUserID, &messageThumbsUp, &createdAtStr,
			&anonymousInt, &senderUsername)

		newMsg := Message{
			MessageID:      msgID,
			MessageContent: messageContent,
			SenderUsername: &senderUsername,
			CreatedAtStr:   createdAtStr,
			ThumbsUp:       messageThumbsUp,
			Anonymous:      boolconvert.MustItob(anonymousInt),
			SonMessages:    nil,
		}

		// 匿名
		if newMsg.Anonymous {
			newMsg.SenderUsername = nil
		}

		messages = append(messages, &newMsg)
	}

	// 开始扫子层
	for _, v := range messages {
		err := toolScanSonComments(tx, v)
		if err != nil {
			tx.Rollback()
			log.Printf("failed to toolScanSonComments in TryGetMultipleMessages: %v\n", err)
			return nil, err
		}
	}

	// 成功, 直接commit, 不管是否成功, 因为这只是查询操作
	tx.Commit()

	return messages, nil
}

func RespNoSuchMessageEntryToFetch(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorNoSuchMessageEntryToFetchCode,
		Msg:       errorcodes.ErrorNoSuchMessageEntryToFetchMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}

type MessageResp struct {
	errorcodes.BasicErrorResp
	Message *Message `json:"message"`
}

type MessagesResp struct {
	errorcodes.BasicErrorResp
	Messages [](*Message) `json:"messages"`
}

func RespGetSingleMessageOK(ctx *gin.Context, msg *Message) {
	resp := MessageResp{
		BasicErrorResp: errorcodes.BasicErrorResp{
			ErrorCode: errorcodes.ErrorOKCode,
			Msg:       errorcodes.ErrorOKMsg,
		},
		Message: msg,
	}

	ctx.JSON(http.StatusOK, &resp)
}

func RespGetMessagesOK(ctx *gin.Context, msgs [](*Message)) {
	resp := MessagesResp{
		BasicErrorResp: errorcodes.BasicErrorResp{
			ErrorCode: errorcodes.ErrorOKCode,
			Msg:       errorcodes.ErrorOKMsg,
		},
		Messages: msgs,
	}

	ctx.JSON(http.StatusOK, &resp)
}
