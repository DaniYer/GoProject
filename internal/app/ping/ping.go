package ping

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
