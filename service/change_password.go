package service

import (
	"database/sql"
	"log"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 此函数用于尝试用userid以及oldpassword查询用户并锁住, 如果没找到, 那么说明原密码错误, 返回nil, false
// 如果error不为nil, 那么第一个返回值没有意义
// 第一个返回值用于表示用户原密码是否正确, 修改密码操作是否成功
func TryChangePassword(userid int64, oldPasswordCrypto, newPasswordCrypto string) (bool, error) {
	tx, err := dao.DB.Begin()
	if err != nil {
		log.Printf("failed to Begin in TryChangePassword: %v\n", err)
		return false, err
	}

	row := tx.QueryRow("SELECT id FROM user WHERE id = ? AND password_crypto = ? FOR UPDATE", userid, oldPasswordCrypto)
	var discard int64
	err = row.Scan(&discard)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Printf("failed to QueryRow in TryEditMessage: %v", err)
		return false, err
	}

	// 已经锁住, 现在可UPDATE
	_, err = tx.Exec("UPDATE user SET password_crypto = ? WHERE id = ?", newPasswordCrypto, userid)
	if err != nil {
		tx.Rollback()
		log.Printf("failed to Exec in TryChangePassword: %v\n", err)
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("failed to Commit in TryChangePassword: %v\n", err)
		return false, err
	}

	return true, nil
}

// 修改密码成功
func RespChangePasswordOK(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorOKCode,
		Msg:       errorcodes.ErrorOKMsg,
	}
	ctx.JSON(http.StatusOK, &resp)
}

// 原密码错误
func RespOldPasswordWrong(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorWrongFormerPasswordCode,
		Msg:       errorcodes.ErrorWrongFormerPasswordMsg,
	}
	ctx.JSON(http.StatusOK, &resp)
}
