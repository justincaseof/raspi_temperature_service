package database

import (
	"fmt"
	"github.com/go-xorm/xorm"
	"log"
	"raspi_readtemp/logging"

	/* blank-imported Postgres driver */
	_ "github.com/lib/pq"

	"go.uber.org/zap"
)

// DBConfig -- Struct for yaml-based DB config
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     uint32 `yaml:"port"`
	DBname   string `yaml:"dbname"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	// a unique identifier for distinguishing individual devices
	DeviceId string `yaml:"device-id"`
}

var dbconfig *DBConfig
var xormengine *xorm.Engine
var logger = logging.NewDevLog("database-xorm")

// Open -- Opens a database connection according to yaml file 'dbconfig.yml'
func Open(dbconfigArg *DBConfig) {
	var err error
	dbconfig = dbconfigArg
	// fail hard in case of a stupid config
	err = connectDatabase()
	if err != nil {
		panic(err)
	}
}

func connectDatabase() error {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbconfig.Host, dbconfig.Port, dbconfig.Username, dbconfig.Password, dbconfig.DBname)
	var err error

	en, err := xorm.NewEngine("postgres", dbinfo)
	if err != nil {
		log.Println("engine creation failed", err)
	}

	err = en.Ping()
	if err != nil {
		return err
	}

	xormengine = en
	log.Println("Successfully connected")

	return nil
}

// Close -- closes the given database connection
func Close() {
	if xormengine != nil {
		err := xormengine.Close()
		if err != nil {
			logger.Info("DB connection has been shut down gracefully")
		} else {
			logger.Error("Error closing DB connection", zap.Error(err))
		}
	}
}
