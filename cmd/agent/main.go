package main

import (
	"log"

	"github.com/ArtemiySps/calc_go_final/internal/agent/config"
	a "github.com/ArtemiySps/calc_go_final/internal/agent/service"
	"github.com/ArtemiySps/calc_go_final/pkg/models"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	logger := models.MakeLogger()

	err = a.RunServer(cfg, logger) // запускаем gRPC сервер (агент)
	if err != nil {
		log.Fatal(err.Error())
	}
}
