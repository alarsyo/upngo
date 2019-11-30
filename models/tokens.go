package models

import (
	"fmt"
	"os"

	"github.com/go-chi/jwtauth"

	jwt "github.com/dgrijalva/jwt-go"
)

type Token struct {
	UserId uint
	Tkn    string
}

var TokenAuth *jwtauth.JWTAuth

func InitTokens() {
	secret := os.Getenv("JWTAuthSecret")
	if secret == "" {
		fmt.Fprintf(os.Stderr, "Please set up the JWTAuthSecret env variable\n")
		os.Exit(1)
	}
	TokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	CreateTokensTable()
}

func CreateTokensTable() {
	const query = "CREATE TABLE IF NOT EXISTS tokens ( id serial PRIMARY KEY, uid serial, token text NOT NULL UNIQUE )"
	if _, err := DB.Exec(query); err != nil {
		panic(err)
	}
}

func (t *Token) Create() error {
	_, t.Tkn, _ = TokenAuth.Encode(jwt.MapClaims{"user_id": t.UserId})
	const query = "INSERT INTO tokens (uid, token) VALUES ($1, $2)"
	if _, err := DB.Exec(query, t.UserId, t.Tkn); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create token for user %d", t.UserId)
		return err
	}
	return nil
}

func (t *Token) Delete() error {
	const query = "DELETE FROM tokens WHERE uid = $1"
	if _, err := DB.Exec(query, t.UserId); err != nil {
		fmt.Fprintf(os.Stderr, "Could not delete user %d", t.UserId)
		return err
	}
	return nil
}
