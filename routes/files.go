package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	models "github.com/alarsyo/upngo/models"
	"github.com/go-chi/jwtauth"
)

func GetFilesInfo(w http.ResponseWriter, r *http.Request) {
	_, claims, _ := jwtauth.FromContext(r.Context())
	userId := fmt.Sprintf("%v", claims["user_id"])
	user, ok := strconv.Atoi(userId)
	if ok != nil || user < 0 {
		http.Error(w, "Wrong user id", http.StatusUnauthorized)
		return
	}
	files, err := models.GetFiles(uint(user))
	if err != nil {
		http.Error(w, "No files found", http.StatusNotFound)
		return
	}

	res, err := json.Marshal(files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}
