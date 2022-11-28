package service

import (
	"log"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 因为这个程序不会删除user, 因此不用额外检测rowsAffected, 之间检测
// Exec的err就行
func TryChangeSignature(userid int64, newSignature *string) bool {
	_, err := dao.DB.Exec("UPDATE user SET personal_signature WHERE id = ?", userid)
	if err != nil {
		log.Printf("failed to Exec in TryChangeSignature: %v\n", err)
		return false
	}
	return true
}

// 修改签名成功
func ChangeSignatrueOK(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorOKCode,
		Msg:       errorcodes.ErrorOKMsg,
	}
	ctx.JSON(http.StatusOK, &resp)
}
