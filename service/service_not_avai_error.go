package service

import (
	errorcodes "message-board/util/error_codes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 服务暂时不可用
func ServiceNotAvailabelError(ctx *gin.Context) {
	respStruct := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorServiceNotAvailabelCode,
		Msg:       errorcodes.ErrorServiceNotAvailabelMsg,
	}

	ctx.JSON(http.StatusOK, &respStruct)
}
