package discovery

import (
	"log"
	"notification/internal/consul" // Your local consul package

	"github.com/hashicorp/consul/api"
)

func Register(addr string, serviceName string, port int) {
	agent := consul.NewAgent(&api.Config{Address: addr})

	serviceCfg := consul.Config{
		ServiceID:   serviceName + "-1",
		ServiceName: serviceName,
		Address:     serviceName, // Or use hostname/IP
		Tags:        []string{"grpc"},
		Port:        port,
		Check: &api.AgentServiceCheck{
			HTTP:     "http://" + serviceName + ":8080/health",
			Interval: "10s",
			Timeout:  "2s",
		},
	}

	if err := agent.RegisterService(serviceCfg); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}
}
