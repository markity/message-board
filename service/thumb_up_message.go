package service

import (
	"database/sql"
	"log"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 第一个参数指示是否有该条目
// 第二个参数告知是否点赞成功
func TryThumbUpMessage(currentUserid int64, destMessageID int64) (bool, bool, error) {
	tx, err := dao.DB.Begin()
	if err != nil {
		return false, false, err
	}

	var discardInt64 int64
	var discarString string
	var thumbsUp int64

	// 先拿目标消息的锁, 判断消息是否存在
	row := tx.QueryRow("SELECT thumbs_up FROM message WHERE id = ? FOR UPDATE", destMessageID)
	err = row.Scan(&thumbsUp)
	if err != nil {
		// 没有此消息, 先尝试释放锁
		if err == sql.ErrNoRows {
			// 这里不做检查, 本意是释放锁, 错误值为nil也告知外部不要重试
			tx.Rollback()
			return false, false, nil
		}
		log.Printf("failed to QueryRow in TryThumbUpMessage: %v\n", err)
		return false, false, err
	}

	// 写操作, 先拿分布式锁, 然后再查询中间表, 所有访问中间表的操作先
	// 从分布式表锁拿锁, tbname字段unique, 有索引, 搜索也很快
	row = tx.QueryRow("SELECT tbname FROM distributed_lock WHERE tbname = 'thumb_message_user' FOR UPDATE")
	err = row.Scan(&discarString)
	if err != nil {
		// 这里尝试拿锁, 要求分布式表锁的表存在预先准备好的数据, 否则回发生致命错误
		tx.Rollback()
		log.Printf("failed to QueryRow in TryThumbUpMessage: %v\n", err)
		return false, false, err
	}

	// 拿到全局的表锁了, 现在执行查询, 判断用户是否已经点赞了该评论
	row = tx.QueryRow("SELECT id FROM thumb_message_user WHERE user_id = ? AND message_id = ?",
		currentUserid, destMessageID)
	err = row.Scan(&discardInt64)
	if err != nil {
		// 该用户没有点赞过该评论, 执行点赞
		if err == sql.ErrNoRows {
			// 插入中间表表项
			_, err := tx.Exec("INSERT INTO thumb_message_user(message_id, user_id) VALUES(?,?)",
				destMessageID, currentUserid)
			if err != nil {
				tx.Rollback()
				log.Printf("failed to Exec in TryThumbUpMessage: %v\n", err)
				// 插入失败, 应当重试
				return false, false, err
			}

			// 更新被锁住的message的thumbs_up字段
			_, err = tx.Exec("UPDATE message SET thumbs_up = ? WHERE id = ?", thumbsUp+1,
				destMessageID)
			if err != nil {
				tx.Rollback()
				log.Printf("failed to Exec in TryThumbUpMessage: %v\n", err)
				return false, false, err
			}

			// 全部完毕, 尝试commit事务
			err = tx.Commit()
			if err != nil {
				log.Printf("failed to Commit in TryThumbUpMessage: %v\n", err)
				return false, false, err
			}

			// 成功
			return true, true, nil
		}

		// 其它错误
		log.Printf("failed to QueryRow in TryThumbUpMessage: %v\n", err)
		return false, false, err
	}

	// 查到了, 说明已存在, 简单地尝试释放表锁, 不检查rollback的错误, 告知外部不必重试
	tx.Rollback()
	return true, false, nil
}

func RespNoSuchMessageEntryToThumbUp(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorNoSuchMessageEntryToThumbUpCode,
		Msg:       errorcodes.ErrorNoSuchMessageEntryToThumbUpMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}

func RespYouAlreadyLikedIt(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorYouAlreadyLikedItCode,
		Msg:       errorcodes.ErrorYouAlreadyLikedItMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}

func RespThumbUpMessageOK(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorOKCode,
		Msg:       errorcodes.ErrorOKMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}
