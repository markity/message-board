package service

import (
	"database/sql"
	"message-board/dao"
)

// 第二个参数用于说明新建事务是否成功
func NewTX() (*sql.Tx, bool) {
	tx, err := dao.DB.Begin()
	if err != nil {
		return nil, false
	}

	return tx, true
}
