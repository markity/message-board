package service

import (
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
