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

func main() {
	// channel for receiving interruptions
	sigs := make(chan os.Signal, 1)

	logger.Info("### STARTUP")

	// INIT
	var scfg config.ServiceConfiguration
	config.ReadConfig(&scfg)

	// DB
	database.Open(&scfg.DBConfig)
	defer database.Close()

	// CONSUL registration
	consul.Setup(&scfg.ConsulClientConfig)

	// REST stuff
	chiRouter := chi.NewMux()
	if err := web.SetupChi(chiRouter); err != nil {
		logger.Error("Error registering server", zap.Error(err))
		os.Exit(1)
	}
	listen := fmt.Sprintf(":%s", scfg.ConsulClientConfig.Port)
	if err := http.ListenAndServe( listen, chiRouter); err != nil {
		logger.Error("Error starting http listener", zap.Error(err))

		os.Exit(1)
		//sigs <- os.Interrupt
	}

	// wait for external termination
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM) // Ctrl + c
	<-sigs
	logger.Info("### EXIT")
}
