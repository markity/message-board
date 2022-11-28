package service

import (
	"database/sql"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 第一个bool返回值指示调用方是否出现可重试的错误, 比如网络错误, 调用方首先判断第一个参数, 来断定
// 		后面的参数是否有效
// 第二个参数指示是否查到这个条目, 如果查到, 就说明密码正确
// 第三个参数返回的是此用户的userid
// 第四个参数指示他是不是管理员
func TryCheckLoginInfo(username string, passwordCrypto []byte) (bool, bool, int64, bool) {
	row := dao.DB.QueryRow("SELECT id,admin FROM user WHERE username = ? AND password_crypto = ?",
		username, passwordCrypto)

	var id int64
	var admin_ int
	var admin bool

	// 前面QueryRow没错的话这里可以忽略错误
	err := row.Scan(&id, &admin_)
	if err != nil {
		if err == sql.ErrNoRows {
			// 不存在该条目, 可能是用户未创建, 也可能是密码错误
			// 此处最后两个参数其实没用
			return true, false, 0, false
		} else {
			// 其它错误
			// 此处最后三个参数其实没用
			return false, false, 0, false
		}
	}

	if admin_ == 0 {
		admin = false
	} else {
		admin = true
	}

	return true, true, id, admin
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
