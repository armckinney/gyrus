package handlers

import (
	"encoding/json"
	"net/http"

	"example.com/template-go/internal/database"
)

func Health(db database.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonResp, _ := json.Marshal(db.Health())
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResp)
	}
}
