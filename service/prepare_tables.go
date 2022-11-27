package service

import (
	"log"
	"message-board/dao"
)

// 用于创建所有表需要的表, 如果出错, 直接Panic
func MustPrepareTables() {
	_, err := dao.DB.Exec(dao.SentenceCreateUser)
	if err != nil {
		log.Panicf("failed to create table user: %v\n", err)
	}
	println("created table user")

	_, err = dao.DB.Exec(dao.SentenceCreateMessage)
	if err != nil {
		log.Panicf("failed to create table message: %v\n", err)
	}
	println("created table message")

	_, err = dao.DB.Exec(dao.SentenceCreateThumbMessageUser)
	if err != nil {
		log.Panicf("failed to create table thumb_message_user: %v\n", err)
	}
	println("created table thumb_message_user")
}
