package service

import (
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// jwt鉴权错误
func JWTError(ctx *gin.Context) {
	respStruct := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorInvalidUserTokenCode,
		Msg:       errorcodes.ErrorInvalidUserTokenMsg,
	}

	ctx.JSON(http.StatusOK, &respStruct)
}
