package service

import (
	"database/sql"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 第二个参数用于说明新建事务是否成功
func NewTX() (*sql.Tx, bool) {
	tx, err := dao.DB.Begin()
	if err != nil {
		return nil, false
	}

	return tx, true
}

// jwt鉴权错误
func RespJWTError(ctx *gin.Context) {
	respStruct := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorInvalidUserTokenCode,
		Msg:       errorcodes.ErrorInvalidUserTokenMsg,
	}

	ctx.JSON(http.StatusOK, &respStruct)
}

// 服务暂时不可用
func RespServiceNotAvailabelError(ctx *gin.Context) {
	respStruct := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorServiceNotAvailabelCode,
		Msg:       errorcodes.ErrorServiceNotAvailabelMsg,
	}

	ctx.JSON(http.StatusOK, &respStruct)
}

// 入参错误
func RespInvalidParaError(ctx *gin.Context) {
	respStruct := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorInvalidInputParametersCode,
		Msg:       errorcodes.ErrorInvalidInputParametersMsg,
	}

	ctx.JSON(http.StatusOK, &respStruct)
}
