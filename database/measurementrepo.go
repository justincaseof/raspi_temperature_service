package database

import (
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"raspi_temperature_service/model"
)

type IMeasurementRepository interface {
	InsertMeasurement(value float32, unit string, instanceId string) (error, *model.Measurement)
	FindMeasurementByID(measurementId string) (error, *model.Measurement)
	FindAllMeasurements() (error, []*model.Measurement)
}

type measurementRepository struct {
	xormengine *xorm.Engine
}

// Init
func NewMeasurementRepository() IMeasurementRepository {
	result := measurementRepository{
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
func (repo measurementRepository) InsertMeasurement(value float32, unit string, instanceId string) (error, *model.Measurement) {
	logger.Debug("Inserting meaurement ...",
		zap.Float32("value", value),
		zap.String("unit", unit))

	m := &model.Measurement{
		Value:      value,
		Unit:       unit,
		InstanceId: dbconfig.DeviceId,
	}
	affected, err := repo.xormengine.Insert(m)

	if err != nil {
		return err, nil
	}

	logger.Info("Measurement successfully inserted measurement.", zap.Int64("number_of_insertions", affected))
	logger.Info("  --> measurement_id", zap.Int64("measurement_id", m.Id))

	return nil, m
}

func (repo measurementRepository) FindAllMeasurements() (error, []*model.Measurement) {
	var measurements []*model.Measurement
	err := repo.xormengine.
		Where("instance_id = ?", dbconfig.DeviceId).
		//Limit(100).
		Find(&measurements)
	if err != nil {
		return nil, nil
	}
	return nil, measurements
}

func (repo measurementRepository) FindMeasurementByID(measurementId string) (error, *model.Measurement) {
	var measurements []model.Measurement
	err := repo.xormengine.
		//Where("instance_id = ? and id = ?", dbconfig.DeviceId, measurementId).
		Where("id = ?", measurementId).
		Limit(100).
		Find(&measurements)
	if err != nil {
		return err, nil
	}
	return nil, &measurements[0]
}

/**
  tableIdentifier should be the raspi's mac address
*/
func (repo measurementRepository) ensureTableExists() error {
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
