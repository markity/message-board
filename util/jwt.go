package util

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

func NewUserJWTSignaturer(cryptor Cryptor) *userJWTSignaturer {
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
type userJWTHeaderForJson struct {
	Algo string `json:"alg"`
	Type string `json:"typ"`
}

// // userJWTPayloadForJson 用于生成和解析json
type userJWTPayloadForJson struct {
	UserID int64 `json:"userid"`
	Admin  bool  `json:"admin"`
	Expire int64 `json:"exp"`
	JTI    int64 `json:"jit"`
}

func (a *userJWTSignaturer) Signature(userid int64, admin bool, duatrion time.Duration) string {
	header_, _ := json.Marshal(userJWTHeaderForJson{
		Algo: string(a.cryptor.GetJWTType()),
		Type: "JWT",
	})
	header := string(header_)

	a.jtiLocker.Lock()
	payload_, _ := json.Marshal(userJWTPayloadForJson{
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

func (a *userJWTSignaturer) Check(signature string) bool {
	if strings.Count(signature, ".") != 2 || len(strings.Split(signature, ".")) != 3 {
		return false
	}

	elements := strings.Split(signature, ".")

	payload_, err := base64.StdEncoding.DecodeString(elements[1])
	if err != nil {
		return false
	}

	// 检查时间戳
	var payload userJWTPayloadForJson
	if err := json.Unmarshal(payload_, &payload); err != nil {
		return false
	}
	if payload.Expire < time.Now().Unix() {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(elements[2])
	if err != nil {
		return false
	}
	de, ok := a.cryptor.Decrypt(b)
	if !ok {
		return false
	}

	return string(de) == elements[0]+"."+elements[1]
}
