package dao

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

var dns = "root:mark2004119@/message_board"

func init() {
	db, err := sql.Open("mysql", dns)
	if err != nil {
		log.Panicf("failed to sql.Open: %v\n", err)
	}
	DB = db
}
