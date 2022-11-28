package service

import (
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 入参错误
func InvalidParaError(ctx *gin.Context) {
	respStruct := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorInvalidInputParametersCode,
		Msg:       errorcodes.ErrorInvalidInputParametersMsg,
	}

	ctx.JSON(http.StatusOK, &respStruct)
}
