package service

import (
	"log"
	"message-board/dao"
)

// 用于创建所有表需要的表, 如果出错, 直接Panic
func MustResetTables() {
	_, err := dao.DB.Exec(dao.SentenceDropUser)
	if err != nil {
		log.Panicf("failed to drop table user: %v\n", err)
	}

	_, err = dao.DB.Exec(dao.SentenceDropMessage)
	if err != nil {
		log.Panicf("failed to drop table message: %v\n", err)
	}

	_, err = dao.DB.Exec(dao.SentenceDropThumbMessageUser)
	if err != nil {
		log.Panicf("failed to drop table thumb_message_user: %v\n", err)
	}

	_, err = dao.DB.Exec(dao.SentenceDropDistributedLock)
	if err != nil {
		log.Panicf("failed to drop table distributed_lock: %v\n", err)
	}

	// --------------建表--------------
	_, err = dao.DB.Exec(dao.SentenceCreateUser)
	if err != nil {
		log.Panicf("failed to create table user: %v\n", err)
	}

	_, err = dao.DB.Exec(dao.SentenceCreateMessage)
	if err != nil {
		log.Panicf("failed to create table message: %v\n", err)
	}

	_, err = dao.DB.Exec(dao.SentenceCreateThumbMessageUser)
	if err != nil {
		log.Panicf("failed to create table thumb_message_user: %v\n", err)
	}

	_, err = dao.DB.Exec(dao.SentenceCreateDistributedLock)
	if err != nil {
		log.Panicf("failed to create table distributed_lock: %v\n", err)
	}

	// -------------准备数据------------------

	// 准备分布式表锁, 主要是给thumb_message_user提供一致性读写, 所以提供一个事务级表锁
	_, err = dao.DB.Exec(dao.SentenceInsertTableLock)
	if err != nil {
		log.Panicf("failed to insert to distributed_lock table: %v\n", err)
	}

	log.Println("reset database ok")
}
