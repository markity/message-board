package service

import (
	"database/sql"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 返回值第二个参数是是否正确, 第三个参数是userid, 第四个参数为是否为管理员
// 新增的两个参数用于设计JWT
func TryCheckLoginInfo(username string, passwordCrypto []byte) (bool, int64, bool, error) {
	row := dao.DB.QueryRow("SELECT id,admin FROM user WHERE username = ? AND password_crypto = ?",
		username, passwordCrypto)
	if err := row.Err(); err != nil {
		if err == sql.ErrNoRows {
			// 不存在该条目, 可能是用户未创建, 也可能是密码错误
			return false, 0, false, nil
		} else {
			// 其它错误
			return false, 0, false, err
		}
	}

	var id int64
	var admin_ int
	var admin bool
	// 前面QueryRow没错的话这里可以忽略错误
	row.Scan(&id, &admin_)
	if admin_ == 0 {
		admin = false
	} else {
		admin = true
	}

	return true, id, admin, nil
}

// 用户登录失败, 可能是账户或密码错误
func UserLoginInfoWrong(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorUserInfoWrongCode,
		Msg:       errorcodes.ErrorrUserInfoWrongMsg,
	}
	ctx.JSON(http.StatusOK, &resp)
}

// 登录成功
func UserLoginSuccess(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorOKCode,
		Msg:       errorcodes.ErrorOKMsg,
	}
	ctx.JSON(http.StatusOK, &resp)
}
