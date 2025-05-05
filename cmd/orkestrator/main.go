package main

import (
	"log"

	"github.com/ArtemiySps/calc_go_final/internal/orkestrator/config"
	h "github.com/ArtemiySps/calc_go_final/internal/orkestrator/http"
	"github.com/ArtemiySps/calc_go_final/internal/orkestrator/service"
	"github.com/ArtemiySps/calc_go_final/pkg/models"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	logger := models.MakeLogger()
	api := service.NewOrkestrator(cfg, logger) // создаем экземпляр оркестратора

	api.ConnectToServer() // коннектимся к gRPC серверу (агенту)

	logger = models.MakeLogger()
	transport := h.NewTransportHttp(api, cfg.OrkestratorPort, logger)
	transport.RunServer() // запускаем http-сервер оркестратора
}
