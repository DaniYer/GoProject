package ping

import (
	"database/sql"
	"net/http"

	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

func PingDB(db *sql.DB, w http.ResponseWriter) {
	if err := db.Ping(); err != nil {
		sugar.Errorf("Ошибка соединения с БД: %v", err)
		http.Error(w, "Ошибка соединения с БД", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Связь налажена"))
}
