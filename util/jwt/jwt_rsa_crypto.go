package jwt

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

type rsaCtyptor struct {
	jwtTyp     CryptoType
	privateKey *rsa.PrivateKey
}

// Encrypt 加密bytes
func (c *rsaCtyptor) Encrypt(m []byte) []byte {
	encryptedBytes, _ := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		&c.privateKey.PublicKey,
		m,
		nil,
	)
	return encryptedBytes
}

// Decrypt 解密bytes, 如果成功第二个参数为true
func (c *rsaCtyptor) Decrypt(encryptedBytes []byte) ([]byte, bool) {
	decryptedBytes, err := c.privateKey.Decrypt(nil, encryptedBytes, &rsa.OAEPOptions{Hash: crypto.SHA256})
	if err != nil {
		return nil, false
	}
	return decryptedBytes, true
}

func (c *rsaCtyptor) GetJWTType() CryptoType {
	return c.jwtTyp
}

// NewRsaCryptor 新建一个加密器
func NewRsaSHA256Cryptor() Cryptor {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	return &rsaCtyptor{
		jwtTyp:     CryptoTypeRsaSHA256,
		privateKey: privateKey,
	}
}
