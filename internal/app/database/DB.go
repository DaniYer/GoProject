package database

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func PingDb(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := db.Ping(); err != nil {
			http.Error(w, "cannot connect DB: "+err.Error(), http.StatusInternalServerError)
			return
		} else {

			w.Write([]byte("Успешный запрос"))
		}
		w.WriteHeader(http.StatusOK)
	}
}

func ConnectDB(cfg *config.Config) *sql.DB {

	db, err := sql.Open("pgx", cfg.D)
	if err != nil {
		log.Panic("Cannot connect to db")
		return nil
	}

	return db
}
