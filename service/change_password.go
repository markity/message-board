package service

import (
	"database/sql"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 此函数用于尝试用userid以及oldpassword查询用户并锁住, 如果没找到, 那么说明密码错误, 返回nil, false
// 如果出现其它错误, 返回error, false
// 成功修改密码, 返回nil, true
func TryChangePassword(tx *sql.Tx, userid int64, oldPasswordCrypto, newPasswordCrypto string) (error, bool) {
	row := tx.QueryRow("SELECT id FROM user WHERE id = ? AND password_crypto = ? FOR UPDATE", userid, oldPasswordCrypto)
	if row.Err() == sql.ErrNoRows {
		return nil, false
	}

	// 已经锁住, 现在可UPDATE
	_, err := tx.Exec("UPDATE user SET password_crypto = ? WHERE id = ?", newPasswordCrypto, userid)
	if err != nil {
		return err, false
	}

	return nil, true
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
