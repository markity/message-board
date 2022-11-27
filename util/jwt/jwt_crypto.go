package jwt

// 这个文件用于声明为jwt实现的加密方法, 目前只实现了RSA-SHA256

type CryptoType string

var CryptoTypeRsaSHA256 CryptoType = "RS256"

// 通用加密器接口, 用于支持jwt加密
type Cryptor interface {
	Encrypt([]byte) []byte
	Decrypt([]byte) ([]byte, bool)
	GetJWTType() CryptoType
}
