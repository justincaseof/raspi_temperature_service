package main

import (
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/signal"
	"raspi_readtemp/logging"
	"raspi_temperature_service/database"
	"syscall"
)

var logger = logging.New("raspi_temperature_service", false)

const DB_CONFIG_FILENAME = "dbconfig.yml"

func main() {
	logger.Info("### STARTUP")

	// INIT
	var cfg database.DBConfig
	readDatabaseConfig(&cfg)
	database.Open(&cfg)
	defer database.Close()

	// TEST...
	database.InsertMeasurement(11, "Celsius")
	database.FindMeasurements()

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
