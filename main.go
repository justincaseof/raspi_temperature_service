package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"raspi_readtemp/logging"
	"raspi_temperature_service/config"
	"raspi_temperature_service/consul"
	"raspi_temperature_service/database"
	"raspi_temperature_service/web"
	"syscall"
)

var logger = logging.New("raspi_temperature_service_main", false)

var PORT = 8083

func main() {
	// channel for receiving interruptions
	sigs := make(chan os.Signal, 1)

	logger.Info("### STARTUP")

	// INIT
	var cfg database.DBConfig
	var scfg config.ServiceConfiguration
	readDatabaseConfig(&scfg)
	readDatabaseConfig(&cfg)
	database.Open(&cfg)
	defer database.Close()

	// CONSUL registration
	consul.Setup(PORT)

	// REST stuff
	chiRouter := chi.NewMux()
	if err := web.SetupChi(chiRouter); err != nil {
		logger.Error("Error registering server", zap.Error(err))
		os.Exit(1)
	}
	portDefinition := fmt.Sprintf(":%d", PORT)
	if err := http.ListenAndServe( portDefinition, chiRouter); err != nil {
		logger.Error("Error starting http listener", zap.Error(err))

		os.Exit(1)
		//sigs <- os.Interrupt
	}

	// wait for external termination
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM) // Ctrl + c
	<-sigs
	logger.Info("### EXIT")
}
