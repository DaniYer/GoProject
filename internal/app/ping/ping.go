package ping

import (
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/storage/database"
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

func PingDB(w http.ResponseWriter, r *http.Request, dataSN string) {
	db, err := database.InitDB("postgres", dataSN)
	if err != nil {
		sugar.Errorf("Ошибка чтения базы данных: %v", err)
		http.Error(w, "Ошибка соединения с БД", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Если соединение установлено, возвращаем сообщение (по умолчанию статус 200 OK)
	w.Write([]byte("Связь налажена"))
}
