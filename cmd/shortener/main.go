package main

// Package main Shortener API.
//
// Сервис для сокращения ссылок с поддержкой batch, авторизации и удаления.
// Использует in-memory, file или PostgreSQL в качестве хранилища.
//
//     Schemes: http
//     BasePath: /
//     Version: 1.0
//     Host: localhost:8080
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
import (
	"fmt"
	"os"

	"github.com/DaniYer/GoProject.git/internal/app/initapp"
)

// @title           Shortener API
// @version         1.0
// @description     Сервис сокращения ссылок.
// @host            localhost:8080
// @BasePath        /
func main() {
	if err := initapp.InitializeApp(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка старта приложения: %v\n", err)
		os.Exit(1)
	}

}
