package errorcodes

// 此文件定义了通用错误200xx

type BasicErrorResp struct {
	ErrorCode int    `json:"error_code"`
	Msg       string `json:"msg"`
}

// 成功
var ErrorOKCode = 20000
var ErrorOKMsg = "OK"

// 服务暂时不可用
var ErrorServiceNotAvailabelCode = 20001
var ErrorServiceNotAvailabelMsg = "service not available temporarily"

// 鉴权失败
var ErrorInvalidUserTokenCode = 20002
var ErrorInvalidUserTokenMsg = "invalid identity token"

// 不合法的入参
var ErrorInvalidInputParametersCode = 20003
var ErrorInvalidInputParametersMsg = "invalid parameters"
