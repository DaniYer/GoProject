package main

import (
	"fmt"
	"os"

	"github.com/DaniYer/GoProject.git/internal/app/initapp"
)

func main() {
	if err := initapp.InitializeApp(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка старта приложения: %v\n", err)
		os.Exit(1)
	}

}
