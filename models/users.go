package models

import (
	"database/sql"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       uint
	Email    string
	Password string
}

func CreateUsersTable() {
	const query = "CREATE TABLE IF NOT EXISTS users ( id serial PRIMARY KEY, email text NOT NULL UNIQUE, password text)"
	if _, err := DB.Exec(query); err != nil {
		panic(err)
	}
}

func (u *User) Create() {
	const query = "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id"
	var id uint
	if err := DB.QueryRow(query, u.Email, u.Password).Scan(&id); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create user %s", u.Email)
	} else {
		u.Id = id
	}
}

func (u *User) Delete() {
	const query = "DELETE FROM users WHERE email = $1"
	if _, err := DB.Exec(query, u.Email); err != nil {
		panic(err)
	}
}

func Login(email string, password string) (User, error) {
	var db_id uint
	var db_email string
	var db_password string
	const query = "SELECT id, email, password FROM users WHERE email=$1"
	row := DB.QueryRow(query, email)
	err := row.Scan(&db_id, &db_email, &db_password)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Fprintf(os.Stderr, "Could not create user %s", email)
		} else {
			panic(err)
		}
	}
	err = bcrypt.CompareHashAndPassword([]byte(db_password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return User{}, err
	} else {
		t := Token{UserId: db_id, Tkn: ""}
		if err = t.Create(); err != nil {
			return User{}, err
		}
		return User{Id: db_id, Email: db_email, Password: db_password}, nil
	}
}
