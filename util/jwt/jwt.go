package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// JWT签名器, 提供签名, 验证解包操作
type JWTSignaturer interface {
	Signature(int64, bool, time.Duration) string
	CheckAndUnpackPayload(string) (bool, *UserJWTPayloadForJson)
}

func NewUserJWTSignaturer(cryptor Cryptor) JWTSignaturer {
	return &userJWTSignaturer{
		jti:     0,
		cryptor: cryptor,
	}
}

type userJWTSignaturer struct {
	jti       int64
	jtiLocker sync.Mutex
	cryptor   Cryptor
}

// userJWTHeaderForJson 用于生成和解析json
type UserJWTHeaderForJson struct {
	Algo string `json:"alg"`
	Type string `json:"typ"`
}

// // userJWTPayloadForJson 用于生成和解析json
type UserJWTPayloadForJson struct {
	UserID int64 `json:"userid"`
	Admin  bool  `json:"admin"`
	Expire int64 `json:"exp"`
	JTI    int64 `json:"jit"`
}

func (a *userJWTSignaturer) Signature(userid int64, admin bool, duatrion time.Duration) string {
	header_, _ := json.Marshal(UserJWTHeaderForJson{
		Algo: string(a.cryptor.GetJWTType()),
		Type: "JWT",
	})
	header := string(header_)

	a.jtiLocker.Lock()
	payload_, _ := json.Marshal(UserJWTPayloadForJson{
		UserID: userid,
		Admin:  admin,
		Expire: time.Now().Add(duatrion).Unix(),
		JTI:    a.jti,
	})
	payload := string(payload_)
	a.jti++
	a.jtiLocker.Unlock()

	headerBase64 := base64.StdEncoding.EncodeToString([]byte(header))
	payloadBase64 := base64.StdEncoding.EncodeToString([]byte(payload))

	signature := string(a.cryptor.Encrypt([]byte(headerBase64 + "." + payloadBase64)))

	return headerBase64 + "." + payloadBase64 + "." + base64.StdEncoding.EncodeToString([]byte(signature))
}

func (a *userJWTSignaturer) CheckAndUnpackPayload(signature string) (bool, *UserJWTPayloadForJson) {
	if strings.Count(signature, ".") != 2 || len(strings.Split(signature, ".")) != 3 {
		return false, nil
	}

	elements := strings.Split(signature, ".")

	payload_, err := base64.StdEncoding.DecodeString(elements[1])
	if err != nil {
		return false, nil
	}

	// 检查时间戳
	var payload UserJWTPayloadForJson
	if err := json.Unmarshal(payload_, &payload); err != nil {
		return false, nil
	}
	if payload.Expire < time.Now().Unix() {
		return false, nil
	}

	b, err := base64.StdEncoding.DecodeString(elements[2])
	if err != nil {
		return false, nil
	}
	de, ok := a.cryptor.Decrypt(b)
	if !ok {
		return false, nil
	}

	if string(de) != elements[0]+"."+elements[1] {
		return false, nil
	}

	return true, &payload
}
