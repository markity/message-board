package service

import (
	"database/sql"
	"log"
	"message-board/dao"
	errorcodes "message-board/util/error_codes"
	timeconvert "message-board/util/time_convert"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type UserInfo struct {
	Username          string
	CreatedAt         time.Time
	PersonalSignature *string
}

// true, nil 代表没有这个用户
// false 代表出现错误
// true *userinfo 代表有这个用户
func TryGetUserinfo(username string) (bool, *UserInfo) {
	row := dao.DB.QueryRow("SELECT created_at, personal_signature FROM user WHERE username = ?", username)

	var createdAt string
	var personalSignature *string

	if err := row.Scan(&createdAt, &personalSignature); err != nil {
		if err == sql.ErrNoRows {
			return true, nil
		}
		log.Printf("failed to Query in TryGetUserinfo: %v\n", err)
		return false, nil
	}

	ui := UserInfo{
		Username:          username,
		CreatedAt:         timeconvert.MustStrToTime(createdAt),
		PersonalSignature: personalSignature,
	}
	return true, &ui
}

func RespNoSuchUser(ctx *gin.Context) {
	resp := errorcodes.BasicErrorResp{
		ErrorCode: errorcodes.ErrorNoSuchUserCode,
		Msg:       errorcodes.ErrorNoSuchUserMsg,
	}

	ctx.JSON(http.StatusOK, &resp)
}

type UserInfoResp struct {
	errorcodes.BasicErrorResp
	CreatedAt         time.Time `json:"created_at"`
	Username          string    `json:"username"`
	PersonalSignature *string   `json:"personal_sigature"`
}

func RespGetUserinfoOK(ctx *gin.Context, userInfo *UserInfo) {
	uip := UserInfoResp{
		CreatedAt:         userInfo.CreatedAt,
		Username:          userInfo.Username,
		PersonalSignature: userInfo.PersonalSignature,
		BasicErrorResp: errorcodes.BasicErrorResp{
			ErrorCode: errorcodes.ErrorOKCode,
			Msg:       errorcodes.ErrorOKMsg,
		},
	}

	ctx.JSON(http.StatusOK, &uip)
}
