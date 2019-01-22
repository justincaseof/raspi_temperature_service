package consul

import (
	consulApi "github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"raspi_readtemp/logging"
	"strconv"
	"time"
)

var logger = logging.New("raspi_temperature_service_consulclient", false)
var consulAgent *consulApi.Agent

type ConsulClientConfig struct {
	ConsulServerIP	string `yaml:"consul-server-ip"`
	TTL 			string `yaml:"ttl"`
	Port 			string `yaml:"client-port"`
	Address			string `yaml:"client-address"`
	ServiceName		string `yaml:"service-name"`
}
var consulCfg *ConsulClientConfig

func Setup(cfg *ConsulClientConfig) {
	consulCfg = cfg
	// 1) init
	// yeah, i know: i could have read that via json instead of yaml
	consulClientConfig := consulApi.DefaultConfig()
	consulClientConfig.Address = consulCfg.ConsulServerIP
	consulClient, err := consulApi.NewClient(consulClientConfig)
	if err != nil {
		logger.Error("Cannot init Consul client.")
		panic(err)
	}

	// 2) TEST: Agent
	port, _ := strconv.ParseInt(consulCfg.Port, 10, 32)
	consulAgent = consulClient.Agent()
	serviceDef := &consulApi.AgentServiceRegistration{
		Name: consulCfg.ServiceName,
		Tags: []string{"raspi", "temperature"},
		Address: "dummy.hostname.domain",
		Port: int(port),
		Check: &consulApi.AgentServiceCheck{
			TTL: consulCfg.TTL,
		},
	}
	if err := consulAgent.ServiceRegister(serviceDef); err != nil {
		logger.Error("Cannot register agent with Consul server.")
		panic(err)
	}
	defer consulAgent.Leave()

	// 4) periodically notify consul about our health
	go updateConsul(func() (bool, error) { return true, nil })

}

func update(check func() (bool, error)) {
	ok, err := check()
	if !ok {
		logger.Error("service check not OK", zap.Error(err))
		if agentErr := consulAgent.FailTTL("service:" + consulCfg.ServiceName, err.Error()); agentErr != nil {
			logger.Error("Failed to notify consul", zap.Error(err))
		}
	} else {
		if agentErr := consulAgent.PassTTL("service:" + consulCfg.ServiceName, ""); agentErr != nil {
			logger.Error("Failed to notify consul", zap.Error(err))
		}
	}
}
func updateConsul(check func() (bool, error)) {
	interval, err := time.ParseDuration(consulCfg.TTL)
	if err != nil {
		panic("Invalid TTL")
	}
	ticker := time.NewTicker(interval / 2)
	for range ticker.C {
		update(check)
	}
}