package database

import (
	"fmt"
	"github.com/go-xorm/xorm"
	"log"
	"raspi_readtemp/logging"
	"time"

	/* blank-imported Postgres driver */
	_ "github.com/lib/pq"

	"github.com/pkg/errors"
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

type Measurement struct {
	Id         int64 `xorm:"pk not null autoincr"`
	Value      float32
	Unit       string
	InstanceId string    `xorm:"varchar(200)"`
	Created    time.Time `xorm:"created"`
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
	// fail hard in case of a stupid config
	err = ensureTableExists()
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

// InsertMeasurement -- insert a measurement
func InsertMeasurement(measurement Measurement) error {
	logger.Debug("Inserting meaurement ...",
		zap.Float32("value", measurement.Value),
		zap.String("unit", measurement.Unit))

	measurement.InstanceId = dbconfig.DeviceId
	affected, err := xormengine.Insert(measurement)

	if err != nil {
		return err
	} else {
		logger.Info("Measurement successfully inserted", zap.Int64("affected", affected))
		logger.Info("  --> measurement_id", zap.Int64("measurement_id", measurement.Id))
	}

	return nil
}

/**
tableIdentifier should be the raspi's mac address
*/
func ensureTableExists() error {
	// simple validation
	if len(dbconfig.DeviceId) < 1 {
		return errors.New("Cannot use empty device id!")
	}

	err := xormengine.Sync(new(Measurement))
	if err != nil {
		return err
	}
	logger.Info("Successfully synced tables")

	return nil
}
