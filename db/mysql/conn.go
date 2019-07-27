package mysql

import(
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", "root:123@tcp(127.0.0.1:3306)/fileStoreServer?charset=UTF8")
	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil {
		log.Fatal("Failed to connect mysql, " + err.Error())
	}
}

func DBConn() *sql.DB {
	return db
}