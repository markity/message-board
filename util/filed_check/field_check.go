package filedcheck

import (
	"unicode"
	"unicode/utf8"
)

// 提供字段检查

// 用户名, 要求长度[4, 20], 只允许大小写字符, 数字以及下划线
func CheckUsernameValid(username string) bool {
	for _, v := range username {
		if !unicode.IsLetter(v) && !unicode.IsDigit(v) && v != '_' {
			return false
		}
	}

	l := len(username)
	return l >= 4 && l <= 20
}

// 密码, 要求长度[6, 25], 只允许大小写字符, 数字以及下划线
func CheckPasswordValid(password string) bool {
	for _, v := range password {
		if !unicode.IsLetter(v) && !unicode.IsDigit(v) && v != '_' {
			return false
		}
	}

	l := len(password)
	return l >= 6 && l <= 25
}

// 消息, 要求长度在[10, 300]之间
func CheckMessageValid(content string) bool {
	nCount := utf8.RuneCountInString(content)
	return nCount >= 10 && nCount <= 300
}

// 个性签名, 要求长度在[10, 150]之间
func CheckPersonalSignatureValid(sign string) bool {
	nCount := utf8.RuneCountInString(sign)
	return nCount >= 10 && nCount <= 150
}
