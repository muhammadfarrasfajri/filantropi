package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectMySQL() {

	var HOST = os.Getenv("DB_HOST")
	var USER = os.Getenv("DB_USER")
	var NAME = os.Getenv("DB_NAME")
	var DB_USE = os.Getenv("DB_USE")
	var PORT = os.Getenv("DB_PORT")

	var err error
	dsn := fmt.Sprintf("%s:@tcp(%s:%s)/%s?parseTime=true&loc=Local", USER, HOST, PORT, NAME)
	DB, err = sql.Open(DB_USE, dsn)
	if err != nil {
		log.Fatal("DB error: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("DB unreachable: ", err)
	}

	log.Println("MySQL connected.")
}
