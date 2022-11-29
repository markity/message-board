package api

import (
	"message-board/service"
	"message-board/util/jwt"

	"github.com/gin-gonic/gin"
)

// 用于JWT签名以及鉴权
var JwtSignaturer jwt.JWTSignaturer

// 用于鉴权后传递给下一级中间件的用户信息
type UserAuthInfo struct {
	UserID int64
	Admin  bool
}

// 加载该包的时候生成一个jwt签名器
func init() {
	JwtSignaturer = jwt.NewUserJWTSignaturer(jwt.NewRsaSHA256Cryptor())
}

func MiddleWareJWTVerify(ctx *gin.Context) {
	jwtStr, err := ctx.Cookie("authtoken")
	if err != nil {
		// 未鉴权的错误
		service.RespJWTError(ctx)
		ctx.Abort()
		return
	}

	valid, payload := JwtSignaturer.CheckAndUnpackPayload(jwtStr)
	if !valid {
		// 未鉴权的错误
		service.RespJWTError(ctx)
		ctx.Abort()
		return
	}

	uaf := &UserAuthInfo{
		UserID: payload.UserID,
		Admin:  payload.Admin,
	}

	// 鉴权成功, set用户数据, 并移交给下一个中间件
	ctx.Set("user", uaf)
	ctx.Next()
}
