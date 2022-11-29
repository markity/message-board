package service

import (
	"log"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

// 传入此函数的参数已经过检查, 此时直接操作数据库
// 如果用户已存在, 返回nil, false
// 如果发生其它错误, 返回err, false
// 如果插入成功, 返回nil, true
func TryCreateUser(username string, passwordCrypto []byte, personalSignature *string,
	createdAt time.Time, admin bool) (error, bool) {
	_, err := dao.DB.Exec("INSERT INTO user(username, password_crypto, created_at, personal_signature, admin) VALUES(?,?,?,?,?)",
		username, passwordCrypto, createdAt, personalSignature, admin)
	if err != nil {
		log.Printf("insert error in TryCreateUser: %v\n", err)
		mError := err.(*mysql.MySQLError)
		// 1062 duplicate entry, 代表该用户已存在
		if mError.Number == 1062 {
			return nil, false
		} else {
			return err, false
		}
	}

	// 成功插入
	return nil, true
}

// 告知用户该用户名已被占用
func RespUsernameOccupied(ctx *gin.Context) {
	respStruct := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorUsernameOccupiedCode,
		Msg:       errorcodes.ErrorUsernameOccupiedMsg,
	}

	ctx.JSON(http.StatusOK, &respStruct)
}

// 告知用户创建用户成功
func RespCreateUserOK(ctx *gin.Context) {
	respStruct := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorOKCode,
		Msg:       errorcodes.ErrorOKMsg,
	}

	ctx.JSON(http.StatusOK, &respStruct)
}
