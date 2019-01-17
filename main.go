package main

import (
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"raspi_readtemp/logging"
	"raspi_temperature_service/database"
	"raspi_temperature_service/web"
	"syscall"
)

var logger = logging.New("raspi_temperature_service_main", false)

const DB_CONFIG_FILENAME = "dbconfig.yml"

func main() {
	logger.Info("### STARTUP")

	// INIT
	var cfg database.DBConfig
	readDatabaseConfig(&cfg)
	database.Open(&cfg)
	defer database.Close()

	// TEST...
	measurementRepo := database.InitMeasurementRepository()
	measurementRepo.InsertMeasurement(11, "Celsius")
	measurementRepo.FindMeasurements()

	// REST stuff
	chiRouter := chi.NewMux()
	if err := web.SetupChi(chiRouter); err != nil {
		logger.Error("Error registering server", zap.Error(err))
		os.Exit(1)
	}
	if err := http.ListenAndServe(":8083", chiRouter); err != nil {
		logger.Error("Error starting http listener", zap.Error(err))
		os.Exit(1)
	}

	// wait indefinitely until external abortion
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM) // Ctrl + c
	<-sigs
	logger.Info("### EXIT")
}

// ==== I/O and properties ====

// read config from 'dbconfig.yml'
func readDatabaseConfig(dbconfig *database.DBConfig) {
	var err error
	var bytes []byte
	bytes, err = ioutil.ReadFile(DB_CONFIG_FILENAME)
	if err != nil {
		logger.Error("Cannot open config file", zap.String("filename", DB_CONFIG_FILENAME))
		panic(err)
	}
	err = yaml.Unmarshal(bytes, dbconfig)
	if err != nil {
		panic(err)
	}
	logger.Info("DBConfig parsed.")
}
