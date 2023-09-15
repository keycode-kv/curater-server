package app

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

const (
	dbHost     = "localhost"
	dbUser     = "postgres"
	dbPassword = "postgres"
	dbPort     = 5432
	dbName     = "curater"
)

var db *sqlx.DB

func Init() (err error) {
	err = initDB()
	return
}

func initDB() (err error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err = sqlx.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}
	//query()
	return
}

func GetDB() *sqlx.DB {
	return db
}

//type account struct {
//	ID string `db:"id"`
//}
//
//func query() {
//	var acc account
//	err := db.GetContext(context.Background(), &acc, "select id from account")
//	fmt.Println(acc, "error", err)
//}
