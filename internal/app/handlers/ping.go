package handlers

import (
	"database/sql"
	"net/http"
)

func PingDBInit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "db unavailable", http.StatusInternalServerError)
			return
		}

		if err := db.Ping(); err != nil {
			http.Error(w, "db ping error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
