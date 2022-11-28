package dao

var SentenceCreateUser = `
CREATE TABLE IF NOT EXISTS user(
	id 					INT PRIMARY KEY AUTO_INCREMENT,
	username 			VARCHAR(32) NOT NULL UNIQUE,
	password_crypto 	TINYBLOB NOT NULL COMMENT '加密后的密码, 不保存密码本身',
	created_at 			DATETIME NOT NULL,
	personal_signature 	VARCHAR(200) NULL COMMENT '如果为NULL, 则代表没有个性签名',
	admin 				TINYINT NOT NULL DEFAULT 0,
	deleted 			TINYINT NOT NULL DEFAULT 0
) COMMENT '用户表'
`

var SentenceCreateMessage = `
CREATE TABLE IF NOT EXISTS message(
	id 					INT PRIMARY KEY AUTO_INCREMENT,
	content				VARCHAR(500) NOT NULL,
	sender_user_id 		INT NOT NULL,
	parent_message_id	INT DEFAULT NULL,
	thumbs_up			INT NOT NULL DEFAULT 0,
	created_at 			DATETIME NOT NULL,
	anonymous			TINYINT NOT NULL DEFAULT 0,
	deleted				TINYINT NOT NULL DEFAULT 0
) COMMENT '消息表, 存储顶级留言以及子评论'
`

var SentenceCreateThumbMessageUser = `
CREATE TABLE IF NOT EXISTS thumb_message_user(
	id 			INT PRIMARY KEY AUTO_INCREMENT,
	user_id		INT NOT NULL,
	message_id 	INT NOT NULL
) COMMENT '建立user和message点赞的一对多关系'
`
