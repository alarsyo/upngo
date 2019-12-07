package routes

import (
	"encoding/json"
	"net/http"

	models "github.com/alarsyo/upngo/models"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Create(w http.ResponseWriter, r *http.Request) {
	creds := Credentials{}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user := models.User{Id: 1, Email: creds.Email, Password: creds.Password}
	err = user.Create()
	if err != nil {
		http.Error(w, "Email is already taken", http.StatusUnauthorized)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, tkn, err := models.Login(creds.Email, creds.Password)
	if err != nil {
		http.Error(w, "Bad credentials", http.StatusUnauthorized)
		return
	}

	res, err := json.Marshal(tkn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}
