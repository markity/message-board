package dao

import "time"

type User struct {
	ID                int64
	Username          string
	PasswordCrypto    string
	CreatedAt         time.Time
	PersonalSignature *string
	Admin             bool
}
