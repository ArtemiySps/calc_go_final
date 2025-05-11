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
	api, err := service.NewOrkestrator(cfg, logger) // создаем экземпляр оркестратора
	if err != nil {
		log.Fatal(err.Error())
	}

	api.ConnectToServer() // коннектимся к gRPC серверу (агенту)

	logger = models.MakeLogger()
	//serverErr := make(chan error, 1)
	transport, err := h.NewTransportHttp(api, cfg.OrkestratorPort, logger)
	if err != nil {
		log.Fatal(err.Error())
	}
	transport.RunServer()

	/*go func() {
		if err := transport.RunServer(); err != nil && err != http.ErrServerClosed { // запускаем http-сервер оркестратора
			serverErr <- err
		}
	}()*/

	/*sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v\n", sig)
	case err := <-serverErr:
		log.Printf("Server error: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	} else {
		log.Println("Server gracefully stopped")
	}*/
}
