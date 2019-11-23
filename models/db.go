package models

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDb() {
	username := os.Getenv("db_user")
	password := os.Getenv("db_password")
	name := os.Getenv("db_name")

	param := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", username, password, name)
	var err error
	DB, err = sql.Open("postgres", param)
	if err != nil {
		panic(err)
	}
	CreateUsersTable()
}
