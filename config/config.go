package config

import (
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"raspi_readtemp/logging"
	"raspi_temperature_service/consul"
	"raspi_temperature_service/database"
)

var logger = logging.New("raspi_temperature_service_config", false)

const CONFIG_FILENAME = "config.yml"

type ServiceConfiguration struct {
	DBConfig           database.DBConfig         `yaml:"dbconfig"`
	ConsulClientConfig consul.ConsulClientConfig `yaml:"consul-client"`
}

var serviceConfig ServiceConfiguration

// read config from 'config.yml'
func ReadConfig(targetConfig interface{}) {
	var err error
	var bytes []byte
	bytes, err = ioutil.ReadFile(CONFIG_FILENAME)
	if err != nil {
		logger.Error("Cannot open config file", zap.String("filename", CONFIG_FILENAME))
		panic(err)
	}
	err = yaml.Unmarshal(bytes, targetConfig)
	if err != nil {
		panic(err)
	}
	logger.Info("Config parsed.")
}
