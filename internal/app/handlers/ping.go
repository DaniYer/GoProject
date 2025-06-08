package handlers

import (
	"database/sql"
	"net/http"
)

func PingDB(db *sql.DB, w http.ResponseWriter) {
	if db == nil {
		http.Error(w, "База данных не подключена", http.StatusInternalServerError)
		return
	}

	if err := db.Ping(); err != nil {
		http.Error(w, "Ошибка соединения с БД", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Связь налажена"))
}

func PingDBInit(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		PingDB(db, w)
	}
}
