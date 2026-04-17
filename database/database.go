package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectMySQL() {

	var err error
	dsn := fmt.Sprintf("%s:@tcp(127.0.0.1:3306)/%s?parseTime=true&loc=Local", "root", "filantropi")
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("DB error: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("DB unreachable: ", err)
	}

	log.Println("MySQL connected.")
}
