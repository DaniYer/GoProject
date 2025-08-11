package handlers

import (
	"database/sql"
	"net/http"
)

// PingDBInit godoc
// @Summary      Проверка подключения к базе данных
// @Description  Отправляет ping к базе данных. Если соединение установлено, возвращает 200 OK.
// @Tags         health
// @Produce      plain
// @Success      200 {string} string "OK"
// @Failure      500 {string} string "db unavailable" или "db ping error"
// @Router       /ping [get]
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
