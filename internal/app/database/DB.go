package database

import (
	"database/sql"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.D == "" {
			http.Error(w, "DATABASE_DSN not provided", http.StatusInternalServerError)
			return
		}

		db, err := sql.Open("pgx", cfg.D)
		if err != nil {
			http.Error(w, "cannot open DB: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			http.Error(w, "cannot connect DB: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Успешный запрос"))
		w.WriteHeader(http.StatusOK)
	}
}
