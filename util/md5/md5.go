package md5

import "crypto/md5"

// 将字符串快速转换为md5 bytes
func ToMD5(s string) []byte {
	md5 := md5.New()
	md5.Write([]byte(s))
	return md5.Sum(nil)
}
