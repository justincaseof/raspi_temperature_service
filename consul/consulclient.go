package consul

import (
	consulApi "github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"raspi_readtemp/logging"
	"time"
)

var logger = logging.New("raspi_temperature_service_consulclient", false)
const SERVICE_NAME = "raspi-temperature-service-1"
const consulTTL string = "10s"
var consulAgent *consulApi.Agent

func Setup() {
	// 1) init
	consulClientConfig := consulApi.DefaultConfig()
	consulClientConfig.Address = "192.168.171.34:8500"
	consulClient, err := consulApi.NewClient(consulClientConfig)
	if err != nil {
		logger.Error("Cannot init Consul client.")
		panic(err)
	}

	// 2) TEST
	consulAgent = consulClient.Agent()
	serviceDef := &consulApi.AgentServiceRegistration{
		Name: SERVICE_NAME,
		Check: &consulApi.AgentServiceCheck{
			TTL: consulTTL,
		},
	}
	if err := consulAgent.ServiceRegister(serviceDef); err != nil {
		panic(err)
	}

	//// 3) TEST
	//svc, err := connect.NewService("raspi-temperature-service-2", consulClient)
	//if err != nil {
	//	logger.Error("Cannot register client with Consul server.")
	//	panic(err)
	//}
	//defer svc.Close()

	// 4) periodically notify consul about our health
	go updateConsul(func() (bool, error) { return true, nil })
}

func update(check func() (bool, error)) {
	ok, err := check()
	if !ok {
		logger.Error("service check not OK", zap.Error(err))
		if agentErr := consulAgent.FailTTL("service:"+SERVICE_NAME, err.Error()); agentErr != nil {
			logger.Error("Failed to notify consul", zap.Error(err))
		}
	} else {
		if agentErr := consulAgent.PassTTL("service:"+SERVICE_NAME, ""); agentErr != nil {
			logger.Error("Failed to notify consul", zap.Error(err))
		}
	}
}
func updateConsul(check func() (bool, error)) {
	interval, err := time.ParseDuration(consulTTL)
	if err != nil {
		panic("Invalid TTL")
	}
	ticker := time.NewTicker(interval / 2)
	for range ticker.C {
		update(check)
	}
}