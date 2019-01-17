package database

import (
	"github.com/go-xorm/xorm"
	"go.uber.org/zap"
	"raspi_temperature_service/model"
	"github.com/pkg/errors"
)

type MeasurementRepository struct {
	xormengine *xorm.Engine
}

// Init
func InitMeasurementRepository() MeasurementRepository {
	result := MeasurementRepository{
		// simply copy global xormengine (package 'database') reference into our local struct. dunno whether or not this is useful.
		xormengine: xormengine,
	}

	// fail hard in case of an error
	err := result.ensureTableExists()
	if err != nil {
		panic(err)
	}

	return result
}

// InsertMeasurement -- insert a measurement
func (repo MeasurementRepository) InsertMeasurement(value float32, unit string) error {
	logger.Debug("Inserting meaurement ...",
		zap.Float32("value", value),
		zap.String("unit", unit))

	m := new(model.Measurement)
	m.Value = value
	m.Unit = unit
	m.InstanceId = dbconfig.DeviceId
	affected, err := repo.xormengine.Insert(m)

	if err != nil {
		return err
	} else {
		logger.Info("Measurement successfully inserted measurement.", zap.Int64("number_of_insertions", affected))
		logger.Info("  --> measurement_id", zap.Int64("measurement_id", m.Id))
	}

	return nil
}

func (repo MeasurementRepository) FindMeasurements() {
	var measurements []model.Measurement
	err := repo.xormengine.Where("instance_id = ?", dbconfig.DeviceId).Limit(100).Find(&measurements)
	if err != nil {
		panic(err)
	}
	if measurements != nil && len(measurements) > 0 {
		logger.Info("We have been here already: Found existing measurements for our device-id.", zap.String("device-id", dbconfig.DeviceId))
	}
}

/**
  tableIdentifier should be the raspi's mac address
*/
func (repo MeasurementRepository) ensureTableExists() error {
	// simple validation
	if len(dbconfig.DeviceId) < 1 {
		return errors.New("Cannot use empty device id!")
	}

	err := repo.xormengine.Sync(new(model.Measurement))
	if err != nil {
		return err
	}
	logger.Info("Successfully synced tables")

	return nil
}

